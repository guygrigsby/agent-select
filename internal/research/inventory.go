package research

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// InvRow is one capture-inventory row (§2.1).
type InvRow struct {
	Path                  string
	Origin                string // captured_remote | local_derived | header_snapshot
	URL                   string
	SHA256                string
	Bytes                 int64
	ObservedAt            string
	TimestampBasis        string
	License               string
	PublicationPermission string
	RecoveryBasis         string
}

// Inventory is the parsed capture-inventory.tsv.
type Inventory struct {
	Rows   []*InvRow
	byPath map[string]*InvRow
}

// Valid origin values.
const (
	OriginCaptured = "captured_remote"
	OriginLocal    = "local_derived"
	OriginHeader   = "header_snapshot"
)

func validOrigin(o string) bool {
	return o == OriginCaptured || o == OriginLocal || o == OriginHeader
}

func validSHA(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// LoadInventory reads and validates capture-inventory.tsv (§2.1 shape rules;
// disk-vs-inventory hash verification is VerifyDisk per §5 precedence).
func LoadInventory(root string) (*Inventory, error) {
	p := filepath.Join(root, "data", "research-inputs", "capture-inventory.tsv")
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, iof("read %s: %v", p, err)
	}
	if err := sizeBound(len(raw), p); err != nil {
		return nil, err
	}
	if bytes.ContainsRune(raw, '\r') {
		return nil, invalidf("%s: CR found; LF line endings required", rel(root, p))
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) < 2 {
		return nil, invalidf("%s: header row plus at least one data row required", rel(root, p))
	}
	const header = "path\torigin\turl\tsha256\tbytes\tobserved_at\ttimestamp_basis\tlicense\tpublication_permission\trecovery_basis"
	if lines[0] != header {
		return nil, invalidf("%s: header mismatch: %q", rel(root, p), lines[0])
	}
	inv := &Inventory{byPath: map[string]*InvRow{}}
	for i, ln := range lines[1:] {
		rowNum := i + 2 // 1-based, header is row 1
		c := strings.Split(ln, "\t")
		if len(c) != 10 {
			return nil, invalidf("%s row %d: expected 10 columns, got %d", rel(root, p), rowNum, len(c))
		}
		r := &InvRow{
			Path: c[0], Origin: c[1], URL: c[2], SHA256: c[3],
			Bytes: 0, ObservedAt: c[5], TimestampBasis: c[6],
			License: c[7], PublicationPermission: c[8], RecoveryBasis: c[9],
		}
		var n int64
		if _, err := fmt.Sscanf(c[4], "%d", &n); err != nil || fmt.Sprintf("%d", n) != c[4] {
			return nil, invalidf("%s row %d: bytes %q is not a non-negative integer", rel(root, p), rowNum, c[4])
		}
		r.Bytes = n
		if !ValidEvidencePath(r.Path) {
			return nil, invalidf("%s row %d: path %q violates path rules", rel(root, p), rowNum, r.Path)
		}
		if strings.HasPrefix(r.Path, "data/research-inputs/") || isControlPath(r.Path) {
			return nil, invalidf("%s row %d: %q names a control file; control files are exempt from inventory membership", rel(root, p), rowNum, r.Path)
		}
		if underStaging(r.Path) {
			return nil, invalidf("%s row %d: %q is inside the staging lane", rel(root, p), rowNum, r.Path)
		}
		if !validOrigin(r.Origin) {
			return nil, invalidf("%s row %d: origin %q outside captured_remote|local_derived|header_snapshot", rel(root, p), rowNum, r.Origin)
		}
		if !validSHA(r.SHA256) {
			return nil, invalidf("%s row %d: sha256 %q is not 64 hex characters", rel(root, p), rowNum, r.SHA256)
		}
		switch r.Origin {
		case OriginLocal:
			if r.URL != "" {
				return nil, invalidf("%s row %d: local_derived row %q must have empty url", rel(root, p), rowNum, r.Path)
			}
		case OriginCaptured, OriginHeader:
			if r.URL == "" && r.RecoveryBasis == "" {
				return nil, invalidf("%s row %d: %s row %q has neither url nor recovery_basis (gate 5)", rel(root, p), rowNum, r.Origin, r.Path)
			}
		}
		if r.License == "" || r.PublicationPermission == "" {
			return nil, invalidf("%s row %d: license and publication_permission are required", rel(root, p), rowNum)
		}
		if r.PublicationPermission != "unknown" && r.PublicationPermission != "allowed" && r.PublicationPermission != "denied" {
			return nil, invalidf("%s row %d: publication_permission %q outside unknown|allowed|denied", rel(root, p), rowNum, r.PublicationPermission)
		}
		if r.TimestampBasis == "" || r.ObservedAt == "" {
			return nil, invalidf("%s row %d: observed_at and timestamp_basis are required", rel(root, p), rowNum)
		}
		if prev, dup := inv.byPath[r.Path]; dup {
			return nil, invalidf("%s row %d: duplicate path %q (first at bytes %d)", rel(root, p), rowNum, r.Path, prev.Bytes)
		}
		inv.byPath[r.Path] = r
		inv.Rows = append(inv.Rows, r)
	}
	if !sort.StringsAreSorted(invPaths(inv)) {
		return nil, invalidf("%s: rows must be sorted by path", rel(root, p))
	}
	return inv, nil
}

func invPaths(inv *Inventory) []string {
	out := make([]string, len(inv.Rows))
	for i, r := range inv.Rows {
		out[i] = r.Path
	}
	return out
}

// Get returns the row for path or nil.
func (inv *Inventory) Get(path string) *InvRow { return inv.byPath[path] }

// Hash returns the inventory sha for path, or "".
func (inv *Inventory) Hash(path string) string {
	if r := inv.byPath[path]; r != nil {
		return r.SHA256
	}
	return ""
}

// VerifyDisk compares every inventory row's sha256/bytes against the file on
// disk (§5 load-time disk-vs-inventory, exit 1; runs before pin gates).
// totalInputBound is the §6 bound on summed evidence input bytes.
const totalInputBound = 2 << 30 // 2 GiB

func (inv *Inventory) VerifyDisk(root string) error {
	var total int64
	for _, r := range inv.Rows {
		p := filepath.Join(root, filepath.FromSlash(r.Path))
		fi, err := os.Lstat(p)
		if err != nil {
			return invalidf("inventory path %s: %v", r.Path, err)
		}
		if !fi.Mode().IsRegular() {
			return invalidf("inventory path %s: not a regular file", r.Path)
		}
		if fi.Size() != r.Bytes {
			return invalidf("inventory path %s: size %d, inventory says %d", r.Path, fi.Size(), r.Bytes)
		}
		h := sha256.New()
		f, err := os.Open(p)
		if err != nil {
			return iof("open %s: %v", r.Path, err)
		}
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return iof("hash %s: %v", r.Path, err)
		}
		_ = f.Close()
		if got := hex.EncodeToString(h.Sum(nil)); got != r.SHA256 {
			return invalidf("inventory path %s: disk sha256 %s, inventory pins %s", r.Path, got, r.SHA256)
		}
		// Symlink escape check (§6): the path resolved to a regular file via
		// Lstat (no symlink following), and the path itself passed the
		// character-set rules; a symlink at an inventory path is rejected.
		if fi.Mode()&os.ModeSymlink != 0 {
			return invalidf("inventory path %s: symlinks are rejected", r.Path)
		}
		total += fi.Size()
	}
	if total > totalInputBound {
		return invalidf("total evidence input size %d bytes exceeds the 2 GiB bound", total)
	}
	return nil
}

// CheckReadPath validates that an evidence path read by the exporter is
// inventoried and well-formed (gate 5, exit 1).
func (inv *Inventory) CheckReadPath(path string) error {
	if strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return invalidf("read path %s: absolute paths and .. are rejected", path)
	}
	if underStaging(path) {
		return invalidf("read path %s: staging lane is not read by the exporter", path)
	}
	if isControlPath(path) {
		return invalidf("read path %s: control file", path)
	}
	if inv.Get(path) == nil {
		return invalidf("read path %s: not in capture inventory", path)
	}
	return nil
}
