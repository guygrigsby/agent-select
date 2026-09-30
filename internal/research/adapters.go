package research

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// decimalRe matches exact decimal lexemes (no NaN/Inf, no hex).
var decimalRe = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$`)

// ValidDecimal reports whether s parses as an exact finite decimal lexeme.
func ValidDecimal(s string) bool { return decimalRe.MatchString(s) }

// ModelRow is one models[] item (model_batch).
type ModelRow struct {
	ID              string
	BenchmarkStatus string
	Notes           string
	Strengths       []ClaimItem
	Weaknesses      []ClaimItem
}

// ClaimItem is one strengths[]/weaknesses[] entry.
type ClaimItem struct {
	Text       string `json:"text"`
	Basis      string `json:"basis"`
	SourceURL  string `json:"source_url"`
	SourceFile string `json:"source_file"`
	Locator    string `json:"locator"`
}

// BenchRow is one benchmarks[] item (model_batch).
type BenchRow struct {
	ModelID          string
	Benchmark        string
	BenchmarkVersion string
	Metric           string
	Value            string // exact lexeme
	Unit             string
	Configuration    string
	SourceURL        string
	SourceFile       string
	Locator          string
	EvidenceClass    string
	IdentityBasis    string
}

// UnresolvedRow is one measurements[] item (unresolved_only, coverage_followup).
type UnresolvedRow struct {
	ReportedModel    string
	Benchmark        string
	BenchmarkVersion string
	Metric           string
	Value            string
	Unit             string
	Configuration    string
	SourceURL        string
	SourceFile       string
	Locator          string
	EvidenceClass    string
	IdentityStatus   string
	ProposedOfferIDs []string
}

// SourceRow is one sources[] item (projected to source_access).
type SourceRow struct {
	URL     string
	File    string
	Status  string
	Purpose string
	License string
	// coverage_followup extras (accepted, not projected)
	SHA256 string
	Commit string
}

// Batch is the adapter-normalized content of one findings file.
type Batch struct {
	Name         string
	Adapter      string
	Models       []ModelRow
	Benchmarks   []BenchRow
	Measurements []UnresolvedRow
	Sources      []SourceRow
	Gaps         []string
}

// HasMeasurements reports whether the batch carries benchmark or
// unresolved measurement rows (the §2.3.1 decision gate).
func (b *Batch) HasMeasurements() bool { return len(b.Benchmarks) > 0 || len(b.Measurements) > 0 }

// Adapter top-level key sets (§2.4). Descriptive keys are data, not closed
// over, for catalog-sources; benchmark-sources keys are enumerated.
var adapterTopKeys = map[string]map[string]bool{
	AdapterModelBatch:       {"models": true, "benchmarks": true, "sources": true, "gaps": true},
	AdapterUnresolvedOnly:   {"measurements": true, "gaps": true},
	AdapterCoverageFollowup: {"measurements": true, "sources": true, "gaps": true},
	AdapterBenchmarkSources: {"assigned_models": true, "baseline": true, "benchmark_rows": true, "note": true, "stage": true},
	AdapterCatalogSources: {
		"research_title": true, "collected": true, "evidence": true,
		"headers": true, "counts": true, "units": true, "pricing_comparison": true,
		"vercel_priced_but_models_dev_cost_missing": true, "image_pricing": true,
		"vercel_pricing_keys": true, "licensing": true, "notes": true,
	},
}

// itemFieldSets enumerates the closed per-item fields (§2.4 table).
var itemFieldSets = map[string][]string{
	"model_batch:models":        {"id", "strengths", "weaknesses", "benchmark_status", "notes"},
	"model_batch:benchmarks":    {"model_id", "benchmark", "benchmark_version", "metric", "value", "unit", "configuration", "source_url", "source_file", "locator", "evidence_class", "identity_basis"},
	"model_batch:sources":       {"url", "file", "status", "purpose", "license"},
	"coverage_followup:sources": {"url", "file", "status", "purpose", "license", "sha256", "commit"},
	"measurements":              {"reported_model", "benchmark", "benchmark_version", "metric", "value", "unit", "configuration", "source_url", "source_file", "locator", "evidence_class", "identity_status", "proposed_offer_ids"},
	"claims":                    {"text", "basis", "source_url", "source_file", "locator"},
}

// decodeFindings parses a findings file with exact number lexemes preserved
// and depth bounded (§6).
func decodeFindings(root, path string) (map[string]json.RawMessage, error) {
	full := filepath.Join(root, filepath.FromSlash(path))
	raw, err := os.ReadFile(full)
	if err != nil {
		return nil, iof("read %s: %v", path, err)
	}
	if err := sizeBound(len(raw), path); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var top map[string]json.RawMessage
	if err := dec.Decode(&top); err != nil {
		return nil, invalidf("%s: malformed JSON: %v", path, err)
	}
	if depthCheck(raw) > 64 {
		return nil, invalidf("%s: JSON nesting depth exceeds 64", path)
	}
	return top, nil
}

// depthCheck returns maximum brace/bracket nesting depth of the raw bytes.
func depthCheck(raw []byte) int {
	depth, max := 0, 0
	inStr, esc := false, false
	for _, b := range raw {
		switch {
		case esc:
			esc = false
		case b == '\\':
			esc = true
		case b == '"':
			inStr = !inStr
		case inStr:
		case b == '{' || b == '[':
			depth++
			if depth > max {
				max = depth
			}
		case b == '}' || b == ']':
			depth--
		}
	}
	return max
}

// checkItemFields enforces the closed item-field vocabulary.
func checkItemFields(path string, list string, raw []json.RawMessage, allowed []string) error {
	set := map[string]bool{}
	for _, k := range allowed {
		set[k] = true
	}
	for i, item := range raw {
		var m map[string]json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(item))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return invalidf("%s %s[%d]: malformed object: %v", path, list, i, err)
		}
		for k := range m {
			if !set[k] {
				return invalidf("%s %s[%d]: unknown field %q (adapter closure)", path, list, i, k)
			}
		}
	}
	return nil
}

// LoadBatch reads one batch findings file through its adapter.
func LoadBatch(root, batch, adapter string) (*Batch, error) {
	path := "data/research/" + batch + "/findings.json"
	top, err := decodeFindings(root, path)
	if err != nil {
		return nil, err
	}
	allowed, ok := adapterTopKeys[adapter]
	if !ok {
		return nil, invalidf("build.json: batch %q: unknown adapter %q", batch, adapter)
	}
	for k := range top {
		if !allowed[k] {
			return nil, invalidf("%s: unknown top-level key %q for adapter %s", path, k, adapter)
		}
	}
	b := &Batch{Name: batch, Adapter: adapter}

	if adapter == AdapterModelBatch || adapter == AdapterUnresolvedOnly || adapter == AdapterCoverageFollowup {
		if err := b.loadGaps(path, top); err != nil {
			return nil, err
		}
	}
	switch adapter {
	case AdapterModelBatch:
		if err := b.loadModels(path, top); err != nil {
			return nil, err
		}
		if err := b.loadBenchmarks(path, top); err != nil {
			return nil, err
		}
		if err := b.loadSources(path, top, itemFieldSets["model_batch:sources"]); err != nil {
			return nil, err
		}
	case AdapterUnresolvedOnly:
		if err := b.loadMeasurements(path, top); err != nil {
			return nil, err
		}
	case AdapterCoverageFollowup:
		if err := b.loadMeasurements(path, top); err != nil {
			return nil, err
		}
		if err := b.loadSources(path, top, itemFieldSets["coverage_followup:sources"]); err != nil {
			return nil, err
		}
	case AdapterBenchmarkSources, AdapterCatalogSources:
		// Descriptive rows only; no measurement/claim projection. Shapes are
		// pinned by the findings decision pin and the frozen round-trip oracle.
	}
	return b, nil
}

func rawList(top map[string]json.RawMessage, key string) ([]json.RawMessage, bool, error) {
	raw, ok := top[key]
	if !ok || string(raw) == "null" {
		return nil, false, nil
	}
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, true, invalidf("key %s: expected a list: %v", key, err)
	}
	return list, true, nil
}

func (b *Batch) loadGaps(path string, top map[string]json.RawMessage) error {
	list, _, err := rawList(top, "gaps")
	if err != nil {
		return invalidf("%s: %v", path, err)
	}
	for i, r := range list {
		var s string
		if err := json.Unmarshal(r, &s); err != nil {
			return invalidf("%s gaps[%d]: must be a plain string, not an object", path, i)
		}
		b.Gaps = append(b.Gaps, s)
	}
	return nil
}

func str(m map[string]json.RawMessage, key, dflt string) string {
	if raw, ok := m[key]; ok && string(raw) != "null" {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
	}
	return dflt
}

func (b *Batch) loadModels(path string, top map[string]json.RawMessage) error {
	list, _, err := rawList(top, "models")
	if err != nil {
		return invalidf("%s: %v", path, err)
	}
	if err := checkItemFields(path, "models", list, itemFieldSets["model_batch:models"]); err != nil {
		return err
	}
	seen := map[string]bool{}
	for i, r := range list {
		var m map[string]json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(r))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return invalidf("%s models[%d]: %v", path, i, err)
		}
		id := str(m, "id", "")
		if id == "" {
			return invalidf("%s models[%d]: id is required", path, i)
		}
		if seen[id] {
			return invalidf("%s models[%d]: duplicate model id %q", path, i, id)
		}
		seen[id] = true
		status := str(m, "benchmark_status", "")
		if status != "found" && status != "not_found_in_checked_sources" && status != "unresearched" {
			return invalidf("%s models[%d] %s: benchmark_status %q outside vocabulary", path, i, id, status)
		}
		row := ModelRow{ID: id, BenchmarkStatus: status, Notes: str(m, "notes", "")}
		for pol, dst := range map[string]*[]ClaimItem{"strengths": &row.Strengths, "weaknesses": &row.Weaknesses} {
			items, ok, err := rawList(map[string]json.RawMessage(m), pol)
			if err != nil {
				return invalidf("%s models[%d].%s: %v", path, i, pol, err)
			}
			if !ok {
				continue
			}
			if err := checkItemFields(path, "models[]."+pol, items, itemFieldSets["claims"]); err != nil {
				return err
			}
			for j, it := range items {
				var cm map[string]json.RawMessage
				dec := json.NewDecoder(bytes.NewReader(it))
				if err := dec.Decode(&cm); err != nil {
					return invalidf("%s models[%d].%s[%d]: %v", path, i, pol, j, err)
				}
				*dst = append(*dst, ClaimItem{
					Text:       str(cm, "text", ""),
					Basis:      str(cm, "basis", ""),
					SourceURL:  str(cm, "source_url", ""),
					SourceFile: str(cm, "source_file", ""),
					Locator:    str(cm, "locator", ""),
				})
			}
		}
		b.Models = append(b.Models, row)
	}
	return nil
}

func (b *Batch) loadBenchmarks(path string, top map[string]json.RawMessage) error {
	list, _, err := rawList(top, "benchmarks")
	if err != nil {
		return invalidf("%s: %v", path, err)
	}
	if err := checkItemFields(path, "benchmarks", list, itemFieldSets["model_batch:benchmarks"]); err != nil {
		return err
	}
	for i, r := range list {
		var m map[string]json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(r))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return invalidf("%s benchmarks[%d]: %v", path, i, err)
		}
		val := str(m, "value", "")
		if !ValidDecimal(val) {
			return invalidf("%s benchmarks[%d]: value %q is not an exact finite decimal", path, i, val)
		}
		row := BenchRow{
			ModelID:          str(m, "model_id", ""),
			Benchmark:        str(m, "benchmark", ""),
			BenchmarkVersion: str(m, "benchmark_version", ""),
			Metric:           str(m, "metric", ""),
			Value:            val,
			Unit:             str(m, "unit", ""),
			Configuration:    str(m, "configuration", ""),
			SourceURL:        str(m, "source_url", ""),
			SourceFile:       str(m, "source_file", ""),
			Locator:          str(m, "locator", ""),
			EvidenceClass:    str(m, "evidence_class", ""),
			IdentityBasis:    str(m, "identity_basis", ""),
		}
		if row.ModelID == "" || row.Benchmark == "" || row.SourceFile == "" || row.SourceURL == "" {
			return invalidf("%s benchmarks[%d]: model_id, benchmark, source_file and source_url are required", path, i)
		}
		b.Benchmarks = append(b.Benchmarks, row)
	}
	return nil
}

func (b *Batch) loadMeasurements(path string, top map[string]json.RawMessage) error {
	list, _, err := rawList(top, "measurements")
	if err != nil {
		return invalidf("%s: %v", path, err)
	}
	if err := checkItemFields(path, "measurements", list, itemFieldSets["measurements"]); err != nil {
		return err
	}
	for i, r := range list {
		var m map[string]json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(r))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return invalidf("%s measurements[%d]: %v", path, i, err)
		}
		val := str(m, "value", "")
		if !ValidDecimal(val) {
			return invalidf("%s measurements[%d]: value %q is not an exact finite decimal", path, i, val)
		}
		status := str(m, "identity_status", "")
		if status == "" {
			status = "unresolved"
		}
		if status != "unresolved" {
			return invalidf("%s measurements[%d]: identity_status must be %q", path, i, "unresolved")
		}
		var proposals []string
		if raw, ok := m["proposed_offer_ids"]; ok && string(raw) != "null" {
			if err := json.Unmarshal(raw, &proposals); err != nil {
				return invalidf("%s measurements[%d].proposed_offer_ids: %v", path, i, err)
			}
		}
		row := UnresolvedRow{
			ReportedModel:    str(m, "reported_model", ""),
			Benchmark:        str(m, "benchmark", ""),
			BenchmarkVersion: str(m, "benchmark_version", ""),
			Metric:           str(m, "metric", ""),
			Value:            val,
			Unit:             str(m, "unit", ""),
			Configuration:    str(m, "configuration", ""),
			SourceURL:        str(m, "source_url", ""),
			SourceFile:       str(m, "source_file", ""),
			Locator:          str(m, "locator", ""),
			EvidenceClass:    str(m, "evidence_class", ""),
			IdentityStatus:   status,
			ProposedOfferIDs: proposals,
		}
		if row.ReportedModel == "" || row.Benchmark == "" || row.SourceFile == "" || row.SourceURL == "" {
			return invalidf("%s measurements[%d]: reported_model, benchmark, source_file and source_url are required (schema requires both)", path, i)
		}
		b.Measurements = append(b.Measurements, row)
	}
	return nil
}

func (b *Batch) loadSources(path string, top map[string]json.RawMessage, allowed []string) error {
	list, _, err := rawList(top, "sources")
	if err != nil {
		return invalidf("%s: %v", path, err)
	}
	if err := checkItemFields(path, "sources", list, allowed); err != nil {
		return err
	}
	for i, r := range list {
		var m map[string]json.RawMessage
		dec := json.NewDecoder(bytes.NewReader(r))
		if err := dec.Decode(&m); err != nil {
			return invalidf("%s sources[%d]: %v", path, i, err)
		}
		b.Sources = append(b.Sources, SourceRow{
			URL:     str(m, "url", ""),
			File:    str(m, "file", ""),
			Status:  str(m, "status", ""),
			Purpose: str(m, "purpose", ""),
			License: str(m, "license", "unknown"),
			SHA256:  str(m, "sha256", ""),
			Commit:  str(m, "commit", ""),
		})
	}
	return nil
}

// ResolveSourceFile implements the §2.4 normalization precedence: an exact
// repository-relative hit wins, then the batch-relative path, then the
// reference is unresolvable. Resolved paths must be inventoried.
func (b *Batch) ResolveSourceFile(root, ref string, inv *Inventory) (string, bool, error) {
	if ref == "" {
		return "", false, nil
	}
	if strings.HasPrefix(ref, "/") || strings.Contains(ref, "..") {
		return "", false, invalidf("batch %s: source_file %q: absolute and .. paths are rejected", b.Name, ref)
	}
	cand1 := filepath.ToSlash(filepath.Clean(ref))
	cand2 := "data/research/" + b.Name + "/" + cand1
	for _, cand := range []string{cand1, cand2} {
		if !ValidEvidencePath(cand) {
			continue
		}
		if fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(cand))); err == nil && fi.Mode().IsRegular() {
			if inv.Get(cand) == nil {
				return "", false, invalidf("batch %s: resolved source path %s is not in the capture inventory (gate 5)", b.Name, cand)
			}
			return cand, true, nil
		}
	}
	return "", false, nil
}

// SortedBatchNames returns the deterministic batch processing order.
func SortedBatchNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
