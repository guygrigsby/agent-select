package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Source policy records (§2.5). One current decision per captured source
// URL, mirroring the production source-access contract.
type PolicyEvidence struct {
	Path        string `json:"path"`
	RetrievedAt string `json:"retrieved_at"`
	Outcome     string `json:"outcome"`
}

type PolicyRecord struct {
	URL            string         `json:"url"`
	SourceKind     string         `json:"source_kind"`
	TermsURL       string         `json:"terms_url"`
	Research       string         `json:"research"`
	Extraction     string         `json:"extraction"`
	Redistribution string         `json:"redistribution"`
	LicenseName    string         `json:"license_name"`
	LicenseText    string         `json:"license_text"`
	Evidence       PolicyEvidence `json:"evidence"`
	Notes          string         `json:"notes"`
}

var (
	policyKindRE = regexp.MustCompile(`^(provider_docs|huggingface_model|github_repo|arxiv_paper|benchmark_site|other)$`)
	policyPermRE = regexp.MustCompile(`^(allowed|denied|unknown)$`)
	policyOutRE  = regexp.MustCompile(`^(success|http_error|source_blocked)$`)
	policyTSRE   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)
)

// Policies is the parsed content of all policies files, keyed by file name.
type Policies struct {
	ByBatch map[string][]PolicyRecord // batch -> sorted records
	order   []string                  // file order for deterministic loads
}

// Batches returns the policy file batches in file order.
func (p *Policies) Batches() []string { return p.order }

// CheckPolicyBatches enforces the §2.5 load-time unknown-batch rule: every
// policies file must name a batch that has findings. The assembled-table
// restatement (policies batch absent from source_access) still runs at
// gate 6.
func CheckPolicyBatches(p *Policies, knownBatches []string) error {
	known := make(map[string]bool, len(knownBatches))
	for _, b := range knownBatches {
		known[b] = true
	}
	for _, batch := range p.Batches() {
		if !known[batch] {
			return invalidf("policies/%s.json: names a batch absent from data/research/ (unknown batch name)", batch)
		}
	}
	return nil
}

func strictDecodePolicy(raw []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if dec.More() {
		return fmt.Errorf("trailing content after JSON value")
	}
	return nil
}

// LoadPolicies reads and validates every policies file (§2.5 load-time
// shape rules; violations are exit 1 naming file and record).
func LoadPolicies(root string) (*Policies, error) {
	dir := filepath.Join(root, "data", "research-inputs", "policies")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, iof("read %s: %v", rel(root, dir), err)
	}
	// the dir listing is name-sorted, so the first sighting of a URL is
	// deterministic and the duplicate diagnostic names a stable pair.
	seenURL := map[string]string{}
	p := &Policies{ByBatch: map[string][]PolicyRecord{}}
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			return nil, invalidf("policies: unexpected entry %q (only <batch>.json files)", name)
		}
		batch := strings.TrimSuffix(name, ".json")
		if !controlPathRe.MatchString("data/research-inputs/policies/" + name) {
			return nil, invalidf("policies: file name %q violates path rules", name)
		}
		fp := filepath.Join(dir, name)
		raw, err := os.ReadFile(fp)
		if err != nil {
			return nil, iof("read %s: %v", rel(root, fp), err)
		}
		if err := sizeBound(len(raw), rel(root, fp)); err != nil {
			return nil, err
		}
		if bytes.ContainsRune(raw, '\r') {
			return nil, invalidf("policies/%s: CR found; LF line endings required", name)
		}
		var recs []PolicyRecord
		if err := strictDecodePolicy(raw, &recs); err != nil {
			return nil, invalidf("policies/%s: %v", name, err)
		}
		if len(recs) == 0 {
			return nil, invalidf("policies/%s: empty array", name)
		}
		sorted := append([]string(nil), "")
		sorted = sorted[:0]
		for i, r := range recs {
			where := fmt.Sprintf("policies/%s record %d", name, i+1)
			if !policyKindRE.MatchString(r.SourceKind) {
				return nil, invalidf("%s: source_kind %q outside enum", where, r.SourceKind)
			}
			if !policyPermRE.MatchString(r.Research) || !policyPermRE.MatchString(r.Extraction) || !policyPermRE.MatchString(r.Redistribution) {
				return nil, invalidf("%s: permission outside allowed|denied|unknown", where)
			}
			if !policyOutRE.MatchString(r.Evidence.Outcome) {
				return nil, invalidf("%s: evidence.outcome %q outside enum", where, r.Evidence.Outcome)
			}
			if !strings.HasPrefix(r.URL, "https://") || !strings.HasPrefix(r.TermsURL, "https://") {
				return nil, invalidf("%s: url and terms_url must be https URLs", where)
			}
			if !policyTSRE.MatchString(r.Evidence.RetrievedAt) {
				return nil, invalidf("%s: evidence.retrieved_at %q must match the ms-aligned Z format", where, r.Evidence.RetrievedAt)
			}
			if len(r.LicenseText) > 65536 {
				return nil, invalidf("%s: license_text exceeds 65536 UTF-8 bytes", where)
			}
			sorted = append(sorted, r.URL)
			if prev, dup := seenURL[r.URL]; dup {
				return nil, invalidf("policies/%s record %d: duplicate url %q already recorded in policies/%s", name, i+1, r.URL, prev)
			}
			seenURL[r.URL] = name
		}
		if !sort.StringsAreSorted(sorted) {
			return nil, invalidf("policies/%s: records not sorted by url", name)
		}
		p.ByBatch[batch] = recs
		p.order = append(p.order, batch)
	}
	return p, nil
}
