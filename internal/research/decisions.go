package research

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Decision is one append-only review decision record (§2.3).
type Decision struct {
	Batch           string `json:"batch"`
	Scope           string `json:"scope"`
	State           string `json:"state"`
	Reason          string `json:"reason"`
	ReviewPath      string `json:"review_path"`
	Pins            []Pin  `json:"pins"`
	DecisionVersion int    `json:"decision_version"`
	DecidedAt       string `json:"decided_at"`
	Supersedes      string `json:"supersedes,omitempty"`

	File string `json:"-"` // repository-relative decision file path
}

// Pin names bytes a decision covers.
type Pin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// Scopes and states (§2.3 table).
const (
	ScopeMeasurements = "published_measurements"
	ScopeCatalog      = "catalog_sources"

	StateSourceValidated = "source_validated"
	StateAccepted        = "accepted"
	StateRevoked         = "revoked"
)

// DefaultQuarantinedReason is the frozen default for batches with no current
// decision (byte-exact frozen literal, §2.3).
const DefaultQuarantinedReason = "Provisional research evidence. Composite review failed or batch not separately approved; no routing admission."

// Ledger is the loaded decision set with supersession resolved.
type Ledger struct {
	records []*Decision          // all, sorted by supersession key
	current map[string]*Decision // (batch, scope) -> current
	Files   []string
}

// LoadDecisions reads every data/reviews/decisions/*.json record.
func LoadDecisions(root string) (*Ledger, error) {
	dir := filepath.Join(root, "data", "reviews", "decisions")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, iof("read %s: %v", dir, err)
	}
	l := &Ledger{current: map[string]*Decision{}}
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		rel := "data/reviews/decisions/" + e.Name()
		if !isControlPath(rel) {
			return nil, invalidf("%s: decision file path violates control path rules", rel)
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, iof("read %s: %v", rel, err)
		}
		var d Decision
		dec := json.NewDecoder(bytes.NewReader(raw))
		if err := dec.Decode(&d); err != nil {
			return nil, invalidf("%s: malformed JSON: %v", rel, err)
		}
		d.File = rel
		if err := l.validate(&d); err != nil {
			return nil, err
		}
		l.records = append(l.records, &d)
		l.Files = append(l.Files, rel)
	}
	l.resolveSupersession()
	return l, nil
}

func (l *Ledger) validate(d *Decision) error {
	if d.Batch == "" || d.Scope == "" || d.State == "" || d.Reason == "" {
		return invalidf("%s: batch, scope, state and reason are required", d.File)
	}
	if !ValidEvidencePath("data/research/"+d.Batch+"/findings.json") || strings.Contains(d.Batch, "/") {
		return invalidf("%s: batch %q is not a batch directory name", d.File, d.Batch)
	}
	if underStaging(d.Batch) || strings.HasPrefix(d.Batch, "research-staging") {
		return invalidf("%s: batch %q is in the staging lane", d.File, d.Batch)
	}
	switch d.Scope {
	case ScopeMeasurements:
		if d.State != StateSourceValidated && d.State != StateRevoked {
			return invalidf("%s: scope published_measurements allows state source_validated|revoked, got %q", d.File, d.State)
		}
	case ScopeCatalog:
		if d.State != StateAccepted && d.State != StateRevoked {
			return invalidf("%s: scope catalog_sources allows state accepted|revoked, got %q", d.File, d.State)
		}
	default:
		return invalidf("%s: unknown scope %q", d.File, d.Scope)
	}
	if !ValidEvidencePath(d.ReviewPath) {
		return invalidf("%s: review_path %q violates path rules", d.File, d.ReviewPath)
	}
	if underStaging(d.ReviewPath) {
		return invalidf("%s: review_path %q is in the staging lane", d.File, d.ReviewPath)
	}
	if d.DecisionVersion < 1 {
		return invalidf("%s: decision_version must be a positive integer", d.File)
	}
	if d.DecidedAt == "" || !strings.HasSuffix(d.DecidedAt, "Z") {
		return invalidf("%s: decided_at %q must be UTC with Z suffix", d.File, d.DecidedAt)
	}
	for _, p := range d.Pins {
		if p.Path == "" || !validSHA(p.SHA256) {
			return invalidf("%s: pin {%s %s}: path required and sha256 must be 64 hex", d.File, p.Path, p.SHA256)
		}
		if strings.HasPrefix(p.Path, "data/research-staging/") || strings.Contains(p.Path, "research-staging") {
			return invalidf("%s: pin %s: staging-lane paths are rejected", d.File, p.Path)
		}
		if isControlPath(p.Path) {
			return invalidf("%s: pin %s: control files cannot be pinned", d.File, p.Path)
		}
		if strings.HasPrefix(p.Path, "/") || strings.Contains(p.Path, "..") {
			return invalidf("%s: pin %s: absolute and .. paths are rejected", d.File, p.Path)
		}
	}
	return nil
}

// resolveSupersession orders records by (decided_at, decision_version, file
// path) and keeps the maximum per (batch, scope) current (§2.3).
func (l *Ledger) resolveSupersession() {
	sort.Slice(l.records, func(i, j int) bool {
		a, b := l.records[i], l.records[j]
		if a.DecidedAt != b.DecidedAt {
			return a.DecidedAt < b.DecidedAt
		}
		if a.DecisionVersion != b.DecisionVersion {
			return a.DecisionVersion < b.DecisionVersion
		}
		return a.File < b.File
	})
	for _, r := range l.records {
		l.current[r.Batch+"\x00"+r.Scope] = r
	}
}

// Current returns the current decision for (batch, scope) or nil.
func (l *Ledger) Current(batch, scope string) *Decision {
	return l.current[batch+"\x00"+scope]
}

// CurrentMeasurements returns the current published_measurements decision
// for batch, or nil (revoked or absent).
func (l *Ledger) CurrentMeasurements(batch string) *Decision {
	d := l.Current(batch, ScopeMeasurements)
	if d == nil || d.State == StateRevoked {
		return nil
	}
	return d
}

// CatalogAccepted returns the catalog_sources acceptance for batch or nil.
func (l *Ledger) CatalogAccepted(batch string) *Decision {
	d := l.Current(batch, ScopeCatalog)
	if d == nil || d.State == StateRevoked {
		return nil
	}
	return d
}

// CheckPinGate runs gate 4: every current decision's pins match the capture
// inventory hashes (exit 2 on mismatch, exit 1 on uninventoried pin path).
// Callers must run inventory.VerifyDisk first (§5 precedence).
func (l *Ledger) CheckPinGate(inv *Inventory, batchFormats map[string]string, batchesWithMeasurements map[string]bool) error {
	for _, r := range l.records {
		if l.current[r.Batch+"\x00"+r.Scope] != r {
			continue // superseded records are history, not gated
		}
		adapter, ok := batchFormats[r.Batch]
		if !ok {
			return invalidf("%s: batch %q has no batch_formats mapping", r.File, r.Batch)
		}
		// Scope binds to adapter at load (§2.3).
		switch r.Scope {
		case ScopeMeasurements:
			if !batchesWithMeasurements[r.Batch] {
				return invalidf("%s: scope published_measurements but batch %q carries no measurement rows", r.File, r.Batch)
			}
		case ScopeCatalog:
			if adapter != AdapterCatalogSources {
				return invalidf("%s: scope catalog_sources binds to the catalog_sources adapter, batch %q uses %q", r.File, r.Batch, adapter)
			}
		}
		if r.State == StateRevoked {
			continue // revoked current decisions pin nothing actionable
		}
		for _, p := range r.Pins {
			row := inv.Get(p.Path)
			if row == nil {
				return invalidf("%s: pin path %s is not in the capture inventory", r.File, p.Path)
			}
			if row.SHA256 != p.SHA256 {
				return stalef("%s: pin %s sha256 %s but inventory pins %s; a newer decision must cover the changed bytes",
					r.File, p.Path, p.SHA256[:8], row.SHA256[:8])
			}
		}
	}
	return nil
}

// CheckReviewPathsInInventory ensures every current decision's review report
// is an inventoried evidence input.
func (l *Ledger) CheckReviewPathsInInventory(inv *Inventory) error {
	for key, r := range l.current {
		if l.current[key] != r || r.State == StateRevoked {
			continue
		}
		if inv.Get(r.ReviewPath) == nil {
			return invalidf("%s: review_path %s is not in the capture inventory", r.File, r.ReviewPath)
		}
	}
	return nil
}

// ReviewBatches returns the sorted batches carrying a current (non-revoked)
// published_measurements decision — the evidence_reviews row set.
func (l *Ledger) ReviewBatches() []string {
	var out []string
	for _, d := range l.current {
		if d.Scope == ScopeMeasurements && d.State != StateRevoked {
			out = append(out, d.Batch)
		}
	}
	sort.Strings(out)
	return out
}

// CheckPinSet enforces the §2.3 pin-set equality: the current
// published_measurements decision's pins are exactly the computed set
// (review report + findings + measurement source paths).
func (l *Ledger) CheckPinSet(batch string, computed map[string]bool) error {
	d := l.CurrentMeasurements(batch)
	if d == nil {
		return invalidf("decision: batch %q has no current published_measurements decision", batch)
	}
	pinned := map[string]bool{}
	for _, p := range d.Pins {
		pinned[p.Path] = true
	}
	var missing, extra []string
	for p := range computed {
		if !pinned[p] {
			missing = append(missing, p)
		}
	}
	for p := range pinned {
		if !computed[p] {
			extra = append(extra, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		return invalidf("decision for batch %q: pin set mismatch: missing %v; extra %v", batch, missing, extra)
	}
	return nil
}
