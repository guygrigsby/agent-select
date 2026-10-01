package collector

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ArtifactStore is the local filesystem projection of snapshot_publications
// (delivery-storage V8): one immutable file per artifact generation under
// <root>/snapshots, plus an atomically swapped CURRENT pointer naming the
// current generation. The database is the durable intent; everything here is
// derived and rebuildable from snapshots.artifact_bytes.
type ArtifactStore struct {
	root string
	keep int // total generations retained, floor 2 (current + keep-1 predecessors)
}

// NewArtifactStore validates the root and retention floor.
func NewArtifactStore(root string, keepGenerations int) (*ArtifactStore, error) {
	if root == "" {
		return nil, errf(CodeInvalidRequest, "store root is required")
	}
	if keepGenerations < 2 {
		return nil, errf(CodeInvalidRequest, "keep_generations floor is 2, got %d", keepGenerations)
	}
	if err := os.MkdirAll(filepath.Join(root, "snapshots"), 0o755); err != nil {
		return nil, storageErr(err)
	}
	return &ArtifactStore{root: root, keep: keepGenerations}, nil
}

func (s *ArtifactStore) artifactPath(hashHex string) string {
	return filepath.Join(s.root, "snapshots", hashHex+".bin")
}

func (s *ArtifactStore) currentPath() string {
	return filepath.Join(s.root, "CURRENT")
}

// Stage writes the exact artifact bytes to their final path via a temporary
// file, fsync and rename: the file exists whole at its final name or not at
// all. Staging the same hash with identical bytes is a no-op; a hash/bytes
// disagreement is invalid_request (the hash pins the bytes).
func (s *ArtifactStore) Stage(hashHex string, artifact []byte) error {
	if !isHexHash(hashHex) {
		return errf(CodeInvalidRequest, "artifact hash must be lowercase hex sha-256")
	}
	if len(artifact) == 0 {
		return errf(CodeInvalidRequest, "artifact bytes are required")
	}
	sum := sha256.Sum256(artifact)
	if hexDigest(sum[:]) != hashHex {
		return errf(CodeInvalidRequest, "artifact bytes do not hash to %s", hashHex)
	}
	target := s.artifactPath(hashHex)
	if existing, err := os.ReadFile(target); err == nil {
		if bytes.Equal(existing, artifact) {
			return nil
		}
		// The hash pins the bytes, so a differing existing file is a torn
		// or corrupted write, not a competing generation: restage repairs
		// it. A valid artifact with different bytes cannot exist.
	} else if !errors.Is(err, os.ErrNotExist) {
		return storageErr(err)
	}
	if err := writeFileSync(target, artifact); err != nil {
		return storageErr(err)
	}
	return nil
}

// writeFileSync stages bytes through a temp file in the same directory, an
// fsync of the file, a rename and an fsync of the directory, so a crash
// leaves either nothing or the complete final file.
func writeFileSync(target string, data []byte) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".stage-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		tmp.Close()
		os.Remove(tmpName)
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return err
	}
	if d, err := os.Open(dir); err == nil {
		defer d.Close()
		d.Sync()
	}
	return nil
}

// Current reads the CURRENT pointer. An absent pointer is not an error: it
// reports the empty hash, meaning no generation has been adopted yet.
func (s *ArtifactStore) Current() (string, error) {
	raw, err := os.ReadFile(s.currentPath())
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", storageErr(err)
	}
	hex := strings.TrimSpace(string(raw))
	if hex == "" {
		return "", nil
	}
	if !isHexHash(hex) {
		return "", storageErr(fmt.Errorf("store: CURRENT pointer holds malformed hash %q", hex))
	}
	return hex, nil
}

// Adopt verifies the artifact is staged and byte-pinned to hashHex, then
// atomically swaps CURRENT to name it.
func (s *ArtifactStore) Adopt(hashHex string) error {
	if !isHexHash(hashHex) {
		return errf(CodeInvalidRequest, "artifact hash must be lowercase hex sha-256")
	}
	artifact, err := os.ReadFile(s.artifactPath(hashHex))
	if errors.Is(err, os.ErrNotExist) {
		return errf(CodeNotFound, "artifact %s is not staged", hashHex)
	}
	if err != nil {
		return storageErr(err)
	}
	sum := sha256.Sum256(artifact)
	if hexDigest(sum[:]) != hashHex {
		return errf(CodeConflict, "staged artifact %s fails its hash pin", hashHex)
	}
	if err := writeFileSync(s.currentPath(), []byte(hashHex+"\n")); err != nil {
		return storageErr(err)
	}
	return nil
}

// Prune deletes artifact files beyond the retention window: the current
// generation plus keep-1 newest predecessors survive, ordered by file
// modification time (staging time) with a name tiebreak. It never touches
// files named by CURRENT. Call only after a successful adoption.
func (s *ArtifactStore) Prune() error {
	current, err := s.Current()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(filepath.Join(s.root, "snapshots"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return storageErr(err)
	}
	type gen struct {
		name string
		mod  int64
	}
	var gens []gen
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".bin") {
			continue
		}
		hashHex := strings.TrimSuffix(e.Name(), ".bin")
		if hashHex == current {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return storageErr(err)
		}
		gens = append(gens, gen{name: hashHex, mod: info.ModTime().UnixNano()})
	}
	// newest first; mtime is staging time and staging happens once per hash
	sort.Slice(gens, func(i, j int) bool {
		if gens[i].mod != gens[j].mod {
			return gens[i].mod > gens[j].mod
		}
		return gens[i].name > gens[j].name
	})
	if excess := len(gens) - (s.keep - 1); excess > 0 {
		for _, g := range gens[len(gens)-excess:] {
			if err := os.Remove(s.artifactPath(g.name)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return storageErr(err)
			}
		}
	}
	return nil
}

// Staged returns the artifact bytes for a generation, for verification.
func (s *ArtifactStore) Staged(hashHex string) ([]byte, error) {
	b, err := os.ReadFile(s.artifactPath(hashHex))
	if errors.Is(err, os.ErrNotExist) {
		return nil, errf(CodeNotFound, "artifact %s is not staged", hashHex)
	}
	if err != nil {
		return nil, storageErr(err)
	}
	return b, nil
}
