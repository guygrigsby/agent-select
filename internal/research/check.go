package research

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CheckResult is the structured outcome for --json.
type CheckResult struct {
	Result       string
	OutputHashes map[string]string
	Diagnostics  []string
}

// parseHashManifest parses source-hashes.sha256 into an ordered path->hash
// list, validating the sha256sum format.
func parseHashManifest(raw []byte) ([][2]string, error) {
	var out [][2]string
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	for i, ln := range lines {
		if ln == "" {
			continue
		}
		if len(ln) < 68 || ln[64] != ' ' || ln[65] != ' ' {
			return nil, inconsf("source-hashes.sha256 line %d: not sha256sum format", i+1)
		}
		h, p := ln[:64], ln[66:]
		if !validSHA(h) {
			return nil, inconsf("source-hashes.sha256 line %d: bad hex digest", i+1)
		}
		out = append(out, [2]string{p, h})
	}
	paths := make([]string, len(out))
	for i, e := range out {
		paths[i] = e[0]
	}
	if !sort.StringsAreSorted(paths) {
		return nil, inconsf("source-hashes.sha256: lines must be sorted by path")
	}
	return out, nil
}

// Check is the research check command (§1): validates durable inputs and the
// published artifact set read-only under a shared lock.
func Check(root string) (*CheckResult, error) {
	lk, err := acquireLock(root, false, LockTimeout)
	if err != nil {
		return nil, err
	}
	defer lk.Release()

	// interrupted publication: report read-only, exit 3
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(JournalPath))); err == nil {
		return nil, inconsf("publication journal %s exists: an interrupted publication must be recovered by research rebuild", JournalPath)
	}

	// durable inputs (full load incl. assemble-level validation)
	_, inv, _, _, _, err := loadInputs(root)
	if err != nil {
		return nil, err
	}

	// published artifact set exists
	for _, name := range OutputNames {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			return nil, inconsf("published artifact %s is absent; run research rebuild", name)
		}
	}
	outHashes := map[string]string{}
	for _, name := range OutputNames {
		h, err := HashFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, iof("hash %s: %v", name, err)
		}
		outHashes[name] = h
	}

	// gates 1-3 on the published DB
	db, err := sql.Open("sqlite", filepath.Join(root, filepath.FromSlash(ArtifactDB)))
	if err != nil {
		return nil, iof("open published db: %v", err)
	}
	defer db.Close()
	if err := RunGatesDB(root, db); err != nil {
		return nil, err
	}
	// metadata schema_version must match the DDL the tool carries
	var sv string
	if err := db.QueryRow(`SELECT value FROM metadata WHERE key='schema_version'`).Scan(&sv); err != nil {
		return nil, inconsf("published db: metadata schema_version unreadable: %v", err)
	}
	if sv != fmt.Sprintf("%d", SchemaVersion) {
		return nil, inconsf("published db schema_version %s but tool carries %d", sv, SchemaVersion)
	}
	if err := db.Close(); err != nil {
		return nil, iof("close published db: %v", err)
	}

	// manifest bytes must equal a fresh regeneration from the inventory
	tsvWant := BuildManifestTSV(inv)
	tsvGot, err := readBounded(filepath.Join(root, filepath.FromSlash(ArtifactManifest)))
	if err != nil {
		return nil, iof("read %s: %v", ArtifactManifest, err)
	}
	if string(tsvGot) != string(tsvWant) {
		return nil, inconsf("%s does not match the capture inventory; run research rebuild", ArtifactManifest)
	}

	// hash manifest: recompute every covered file
	hraw, err := readBounded(filepath.Join(root, filepath.FromSlash(ArtifactHashes)))
	if err != nil {
		return nil, iof("read %s: %v", ArtifactHashes, err)
	}
	entries, err := parseHashManifest(hraw)
	if err != nil {
		return nil, err
	}
	want, _, err := BuildHashManifest(root, inv, filepath.Join(root, filepath.FromSlash(ArtifactDB)), filepath.Join(root, filepath.FromSlash(ArtifactManifest)))
	if err != nil {
		return nil, err
	}
	if string(hraw) != string(want) {
		// name the first divergence
		wantLines := strings.Split(strings.TrimSuffix(string(want), "\n"), "\n")
		gotLines := strings.Split(strings.TrimSuffix(string(hraw), "\n"), "\n")
		for i := 0; i < len(wantLines) || i < len(gotLines); i++ {
			var w, g string
			if i < len(wantLines) {
				w = wantLines[i]
			}
			if i < len(gotLines) {
				g = gotLines[i]
			}
			if w != g {
				path := w
				if len(g) > 66 {
					path = g[66:]
				} else if len(w) > 66 {
					path = w[66:]
				}
				return nil, inconsf("%s diverges from current inputs at %s; run research rebuild", ArtifactHashes, path)
			}
		}
		return nil, inconsf("%s diverges from current inputs; run research rebuild", ArtifactHashes)
	}
	_ = entries

	return &CheckResult{Result: "ok", OutputHashes: outHashes}, nil
}
