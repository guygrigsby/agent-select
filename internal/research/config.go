package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config is build.json per §2.2.
type Config struct {
	ArtifactVersion string            `json:"artifact_version"`
	SchemaVersion   int               `json:"schema_version"`
	Phase           string            `json:"phase"`
	PublicationPerm string            `json:"publication_permission"`
	BaselineURL     string            `json:"baseline_url"`
	PricingCrossref string            `json:"pricing_crossref_url"`
	CollectionDate  string            `json:"collection_date_utc"`
	Driver          DriverPin         `json:"driver"`
	Scope           ScopeRule         `json:"scope"`
	Expected        Expected          `json:"expected"`
	PermVocabulary  []string          `json:"publication_permission_vocabulary"`
	BatchFormats    map[string]string `json:"batch_formats"`

	// Raw keeps the file bytes for fingerprinting.
	Raw []byte `json:"-"`
}

// DriverPin pins the SQLite driver module and Go toolchain minor version.
type DriverPin struct {
	SQLiteModule  string `json:"sqlite_driver_module"`
	SQLiteVersion string `json:"sqlite_driver_version"`
	GoToolchain   string `json:"go_toolchain_minor"`
}

// IsPlaceholder reports whether a driver pin has not been filled in yet
// (phase 4 replaces the seeded placeholders). A placeholder is never a
// mismatch because there is nothing to compare against. The phase-3 seed
// literal is recognized alongside "placeholder" so pre-fill configs keep
// loading after toolchain setup.
func (d DriverPin) IsPlaceholder() bool {
	const seed = "pinned at phase 4 toolchain setup"
	return d.SQLiteVersion == "" || stringsContains(d.SQLiteVersion, "placeholder") ||
		stringsContains(d.SQLiteVersion, seed) ||
		stringsContains(d.GoToolchain, "placeholder") ||
		stringsContains(d.GoToolchain, seed) || d.GoToolchain == ""
}

// ScopeRule records which catalog kinds are in research scope.
type ScopeRule struct {
	Rule     string `json:"rule"`
	Excludes string `json:"excludes"`
}

// Expected holds fixture expectations compared against derived counts.
type Expected struct {
	CatalogModels        int `json:"catalog_models"`
	ExactCatalogJoins    int `json:"exact_catalog_joins"`
	ModelResearch        int `json:"model_research"`
	PriceFacts           int `json:"price_facts"`
	PricingDisagreements int `json:"pricing_disagreements"`
}

// Adapters enumerated in §2.4.
const (
	AdapterModelBatch       = "model_batch"
	AdapterUnresolvedOnly   = "unresolved_only"
	AdapterCoverageFollowup = "coverage_followup"
	AdapterBenchmarkSources = "benchmark_source_inventory"
	AdapterCatalogSources   = "catalog_sources"
)

var validAdapters = map[string]bool{
	AdapterModelBatch: true, AdapterUnresolvedOnly: true,
	AdapterCoverageFollowup: true, AdapterBenchmarkSources: true,
	AdapterCatalogSources: true,
}

// LoadConfig reads and validates build.json from root/data/research-inputs/.
func LoadConfig(root string) (*Config, error) {
	p := filepath.Join(root, "data", "research-inputs", "build.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, iof("read %s: %v", p, err)
	}
	if err := sizeBound(len(raw), p); err != nil {
		return nil, err
	}
	var c Config
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&c); err != nil {
		return nil, invalidf("%s: malformed JSON: %v", rel(root, p), err)
	}
	if c.ArtifactVersion == "" || c.SchemaVersion == 0 {
		return nil, invalidf("%s: artifact_version and schema_version are required", rel(root, p))
	}
	if c.SchemaVersion != SchemaVersion {
		return nil, inconsf("%s: schema_version %d does not match the tool DDL schema %d",
			rel(root, p), c.SchemaVersion, SchemaVersion)
	}
	if len(c.PermVocabulary) == 0 {
		return nil, invalidf("%s: publication_permission_vocabulary is required", rel(root, p))
	}
	for _, v := range c.PermVocabulary {
		if v != "unknown" && v != "allowed" && v != "denied" {
			return nil, invalidf("%s: publication_permission_vocabulary value %q outside unknown|allowed|denied", rel(root, p), v)
		}
	}
	if !permInVocab(c.PublicationPerm, c.PermVocabulary) {
		return nil, invalidf("%s: publication_permission %q not in vocabulary", rel(root, p), c.PublicationPerm)
	}
	if len(c.BatchFormats) == 0 {
		return nil, invalidf("%s: batch_formats mapping is required", rel(root, p))
	}
	for dir, adapter := range c.BatchFormats {
		if !validAdapters[adapter] {
			return nil, invalidf("%s: batch_formats[%s]: unknown adapter %q", rel(root, p), dir, adapter)
		}
	}
	c.Raw = raw
	return &c, nil
}

func permInVocab(v string, vocab []string) bool {
	for _, x := range vocab {
		if x == v {
			return true
		}
	}
	return false
}

// MetadataPairs returns the exact §7.2 metadata rows from the config.
func (c *Config) MetadataPairs() [][2]string {
	return [][2]string{
		{"artifact_version", c.ArtifactVersion},
		{"schema_version", fmt.Sprintf("%d", c.SchemaVersion)},
		{"phase", c.Phase},
		{"baseline_url", c.BaselineURL},
		{"pricing_crossref_url", c.PricingCrossref},
		{"collection_date_utc", c.CollectionDate},
		{"publication_permission", c.PublicationPerm},
		{"expected_catalog_models", fmt.Sprintf("%d", c.Expected.CatalogModels)},
		{"expected_exact_joins", fmt.Sprintf("%d", c.Expected.ExactCatalogJoins)},
		{"expected_text_and_image_research_models", fmt.Sprintf("%d", c.Expected.ModelResearch)},
	}
}

func sizeBound(n int, p string) error {
	const perFile = 64 << 20
	if n > perFile {
		return invalidf("%s: file exceeds 64 MiB bound (%d bytes)", p, n)
	}
	return nil
}

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

func stringsContains(s, sub string) bool { return strings.Contains(s, sub) }

// AdapterOf returns the batch -> adapter mapping from batch_formats.
func (c *Config) AdapterOf() map[string]string {
	out := make(map[string]string, len(c.BatchFormats))
	for b, a := range c.BatchFormats {
		out[b] = a
	}
	return out
}
