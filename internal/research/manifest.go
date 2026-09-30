package research

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Artifact file names and the staging area (§3).
const (
	ArtifactDB       = "data/research.sqlite"
	ArtifactManifest = "data/source-manifest.tsv"
	ArtifactHashes   = "data/source-hashes.sha256"

	BuildStagingDir = "data/.research-staging"
	BackupDir       = "data/.research-backup"
	BackupManifest  = "data/.research-backup/manifest.json"
	JournalPath     = "data/.research-rebuild.journal"
	LockPath        = "data/.research-rebuild.lock"
)

// OutputNames are the three artifact paths in rename order (manifest, TSV,
// DB: cheapest-to-regenerate first).
var OutputNames = []string{ArtifactManifest, ArtifactHashes, ArtifactDB}

// BuildManifestTSV renders source-manifest.tsv bytes from the inventory
// (§3: 8 columns, path-sorted — the inventory is path-sorted at load).
func BuildManifestTSV(inv *Inventory) []byte {
	var b strings.Builder
	b.WriteString("path\turl\tsha256\tbytes\tobserved_at\ttimestamp_basis\tlicense\tpublication_permission\n")
	for _, r := range inv.Rows {
		b.WriteString(r.Path)
		b.WriteByte('\t')
		b.WriteString(r.URL)
		b.WriteByte('\t')
		b.WriteString(r.SHA256)
		b.WriteByte('\t')
		fmt.Fprintf(&b, "%d", r.Bytes)
		b.WriteByte('\t')
		b.WriteString(r.ObservedAt)
		b.WriteByte('\t')
		b.WriteString(r.TimestampBasis)
		b.WriteByte('\t')
		b.WriteString(r.License)
		b.WriteByte('\t')
		b.WriteString(r.PublicationPermission)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// ControlFiles enumerates the control files under data/ that the hash
// manifest covers (§6 control-path regex): capture-inventory.tsv, build.json,
// every fetch log, every decision record.
func ControlFiles(root string) ([]string, error) {
	var out []string
	fixed := []string{
		"data/research-inputs/capture-inventory.tsv",
		"data/research-inputs/build.json",
	}
	for _, p := range fixed {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err != nil {
			return nil, iof("stat %s: %v", p, err)
		}
		out = append(out, p)
	}
	logDir := filepath.Join(root, "data", "research-inputs", "fetch-logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, iof("read %s: %v", logDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return nil, invalidf("%s: subdirectories are not valid control files", "data/research-inputs/fetch-logs/"+e.Name())
		}
		p := "data/research-inputs/fetch-logs/" + e.Name()
		if !ValidControlPath(p) {
			return nil, invalidf("%s: fetch log name violates control path rules", p)
		}
		out = append(out, p)
	}
	decDir := filepath.Join(root, "data", "reviews", "decisions")
	entries, err = os.ReadDir(decDir)
	if err != nil {
		return nil, iof("read %s: %v", decDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return nil, invalidf("%s: subdirectories are not valid control files", "data/reviews/decisions/"+e.Name())
		}
		p := "data/reviews/decisions/" + e.Name()
		if !ValidControlPath(p) {
			return nil, invalidf("%s: decision file name violates control path rules", p)
		}
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// HashFile returns the hex sha256 of the file at p.
func HashFile(p string) (string, error) {
	raw, err := readBounded(p)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// BuildHashManifest renders source-hashes.sha256 covering every inventoried
// evidence path, every control file, the completed DB and the generated TSV
// (§3). Paths are the repository-relative slash form.
func BuildHashManifest(root string, inv *Inventory, dbPath, tsvPath string) ([]byte, []string, error) {
	paths := make([]string, 0, len(inv.Rows)+40)
	for _, r := range inv.Rows {
		paths = append(paths, r.Path)
	}
	ctl, err := ControlFiles(root)
	if err != nil {
		return nil, nil, err
	}
	paths = append(paths, ctl...)
	// outputs, keyed by artifact name
	abs := map[string]string{
		ArtifactDB:       dbPath,
		ArtifactManifest: tsvPath,
	}
	paths = append(paths, ArtifactDB, ArtifactManifest)
	sort.Strings(paths)

	type hp struct{ path, sum string }
	hashed := make([]hp, 0, len(paths))
	seen := map[string]bool{}
	for _, p := range paths {
		if seen[p] {
			return nil, nil, invalidf("hash manifest: duplicate path %s", p)
		}
		seen[p] = true
		var fp string
		if a, ok := abs[p]; ok {
			fp = a
		} else {
			fp = filepath.Join(root, filepath.FromSlash(p))
		}
		sum, err := HashFile(fp)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil, invalidf("hash manifest: inventoried path %s missing on disk", p)
			}
			return nil, nil, iof("hash %s: %v", p, err)
		}
		hashed = append(hashed, hp{p, sum})
	}
	var b strings.Builder
	for _, h := range hashed {
		b.WriteString(h.sum)
		b.WriteString("  ")
		b.WriteString(h.path)
		b.WriteByte('\n')
	}
	order := make([]string, len(hashed))
	for i, h := range hashed {
		order[i] = h.path
	}
	return []byte(b.String()), order, nil
}
