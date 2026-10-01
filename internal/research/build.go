package research

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Frozen reason literals (byte-exact). The default quarantined reason lives
// in decisions.go as DefaultQuarantinedReason.
const (
	UnlocatedNoFile = "No saved evidence file supplied; quarantined"
	UnlocatedNoURL  = "no_source_url"
)

// Staging directory (§3 step 4).
const stagingDir = "data/.research-staging"

// Staged file names inside the staging directory.
const (
	StagedDB       = "research.sqlite"
	StagedManifest = "source-manifest.tsv"
	StagedHashes   = "source-hashes.sha256"
)

// BuildResult summarizes one staged build.
type BuildResult struct {
	SourceFiles            int
	CatalogModels          int
	ProviderOffers         int
	PriceFacts             int
	ExactJoins             int
	PricingDisagreements   int
	ModelResearch          int
	Claims                 int
	UnlocatedClaims        int
	BenchmarkMeasurements  int
	UnresolvedMeasurements int
	SourceAccess           int
	ResearchGaps           int
	EvidenceReviews        int
}

// tableRows accumulates one table's rows for sorted insertion (§4).
type tableRows struct {
	name  string
	cols  int
	kinds []byte // 'S' text, 'I' integer
	genID bool   // first column is a generated integer id
	rows  [][]any
}

func (t *tableRows) add(vals ...any) {
	if len(vals) != t.cols-(func() int {
		if t.genID {
			return 1
		}
		return 0
	}()) {
		panic(fmt.Sprintf("%s: row arity %d want %d", t.name, len(vals), t.cols))
	}
	if t.genID {
		t.rows = append(t.rows, append([]any{int64(0)}, vals...))
	} else {
		t.rows = append(t.rows, vals)
	}
}

// less compares two rows by the non-generated column tuple in schema order
// (text as UTF-8 bytes, integers numerically).
func (t *tableRows) less(a, b []any) bool {
	start := 0
	if t.genID {
		start = 1
	}
	for i := start; i < t.cols; i++ {
		switch t.kinds[i] {
		case 'I':
			av, aok := a[i].(int64)
			bv, bok := b[i].(int64)
			if aok && bok {
				if av != bv {
					return av < bv
				}
				continue
			}
			fallthrough
		default:
			as, _ := a[i].(string)
			bs, _ := b[i].(string)
			if as != bs {
				return as < bs
			}
		}
	}
	return false
}

func (t *tableRows) insertInto(tx *sql.Tx) error {
	sort.SliceStable(t.rows, func(i, j int) bool { return t.less(t.rows[i], t.rows[j]) })
	for id, row := range t.rows {
		if t.genID {
			row[0] = int64(id + 1)
		}
		ph := strings.TrimSuffix(strings.Repeat("?,", t.cols), ",")
		if _, err := tx.Exec(fmt.Sprintf("INSERT INTO %s VALUES(%s)", t.name, ph), row...); err != nil {
			return iof("insert %s row %d: %v", t.name, id+1, err)
		}
	}
	return nil
}

func newTable(name, ddlName string, genID bool, kinds string) *tableRows {
	return &tableRows{name: ddlName, cols: len(kinds), kinds: []byte(kinds), genID: genID}
}

// assemble builds every table's rows from the durable inputs.
func assemble(root string, cfg *Config, inv *Inventory, led *Ledger, pol *Policies) (*BuildResult, []*tableRows, error) {
	res := &BuildResult{}

	// schema order (§4)
	tSourceFiles := newTable("source_files", "source_files", false, "SSSISSSS")
	tCatalogModels := newTable("catalog_models", "catalog_models", false, "SSSSSS")
	tCatalogModalities := newTable("catalog_modalities", "catalog_modalities", false, "SSS")
	tCatalogLimits := newTable("catalog_limits", "catalog_limits", false, "SSI")
	tCatalogParams := newTable("catalog_parameters", "catalog_parameters", false, "SS")
	tCatalogTags := newTable("catalog_tags", "catalog_tags", false, "SS")
	tProviderOffers := newTable("provider_offers", "provider_offers", false, "SSSS")
	tCanonicalLabels := newTable("canonical_labels", "canonical_labels", false, "SSS")
	tExactJoins := newTable("exact_catalog_joins", "exact_catalog_joins", false, "SSS")
	tPriceFacts := newTable("price_facts", "price_facts", false, "SSSSSS")
	tDisagreements := newTable("pricing_disagreements", "pricing_disagreements", false, "SSSS")
	tReviewResults := newTable("review_results", "review_results", false, "SSS")
	tModelResearch := newTable("model_research", "model_research", false, "SSSS")
	tClaims := newTable("claims", "claims", true, "ISSSSSSSS")
	tUnlocated := newTable("unlocated_claims", "unlocated_claims", false, "SSSSSSSS")
	tBenchmarks := newTable("benchmark_measurements", "benchmark_measurements", true, "ISSSSSSSSSSSSS")
	tUnresolved := newTable("unresolved_measurements", "unresolved_measurements", true, "ISSSSSSSSSSSSS")
	tProposals := newTable("unresolved_identity_proposals", "unresolved_identity_proposals", false, "IS")
	tSourceAccess := newTable("source_access", "source_access", false, "SSSSSS")
	tGaps := newTable("research_gaps", "research_gaps", false, "SSI")
	tEvidenceReviews := newTable("evidence_reviews", "evidence_reviews", false, "SSSS")
	tEvidenceInputs := newTable("evidence_review_inputs", "evidence_review_inputs", false, "SSS")
	tMetadata := newTable("metadata", "metadata", false, "SS")

	// 1. source_files
	for _, row := range inv.Rows {
		tSourceFiles.add(row.Path, row.URL, row.SHA256, int64(row.Bytes), row.ObservedAt, row.TimestampBasis, row.License, row.PublicationPermission)
	}
	res.SourceFiles = len(inv.Rows)

	// 2. catalog projection
	cat, err := LoadCatalog(root, inv)
	if err != nil {
		return nil, nil, err
	}
	for _, m := range cat.Models {
		tCatalogModels.add(m.ID, m.Name, m.OwnedBy, m.Type, m.Description, vercelCapture)
		modsAny, _ := m.Entry.get("modalities")
		mods, ok := modsAny.(*oobj)
		if !ok {
			return nil, nil, invalidf("%s: model %s: modalities object is required by the frozen corpus", vercelCapture, m.ID)
		}
		for _, dir := range mods.keysInOrder() {
			listAny, _ := mods.get(dir)
			list, ok := listAny.(*oarr)
			if !ok {
				return nil, nil, invalidf("%s: model %s: modality direction %q is not a list", vercelCapture, m.ID, dir)
			}
			seen := map[string]bool{}
			for _, mv := range list.items {
				s, _ := mv.(string)
				if seen[s] {
					return nil, nil, invalidf("%s: model %s: duplicate modality %q in direction %q", vercelCapture, m.ID, s, dir)
				}
				seen[s] = true
				tCatalogModalities.add(m.ID, dir, s)
			}
		}
		for _, key := range []string{"context_window", "max_tokens"} {
			if v, ok := m.Entry.get(key); ok {
				n, ok := v.(jsonNumber)
				if !ok {
					return nil, nil, invalidf("%s: model %s: %s is not a number", vercelCapture, m.ID, key)
				}
				i, err := n.Int64()
				if err != nil {
					return nil, nil, invalidf("%s: model %s: %s %q is not an integer", vercelCapture, m.ID, key, n.String())
				}
				tCatalogLimits.add(m.ID, key, i)
			}
		}
		if paramsAny, ok := m.Entry.get("supported_parameters"); ok {
			if params, ok := paramsAny.(*oarr); ok {
				seen := map[string]bool{}
				for _, pv := range params.items {
					s, _ := pv.(string)
					if seen[s] {
						continue
					}
					seen[s] = true
					tCatalogParams.add(m.ID, s)
				}
			}
		}
		if tagsAny, ok := m.Entry.get("tags"); ok {
			if tags, ok := tagsAny.(*oarr); ok {
				seen := map[string]bool{}
				for _, tv := range tags.items {
					s, _ := tv.(string)
					if seen[s] {
						continue
					}
					seen[s] = true
					tCatalogTags.add(m.ID, s)
				}
			}
		}
	}
	for _, o := range cat.Offers {
		tProviderOffers.add(o.Provider, o.UID, o.Name, modelsDevCapture)
		if o.Canonical {
			tCanonicalLabels.add(o.Provider, o.UID, o.CanonicalLabel)
		}
	}
	for _, pf := range cat.PriceFacts {
		tPriceFacts.add(pf.SourcePath, pf.Provider, pf.UpstreamID, pf.FieldPath, pf.RawDecimal, pf.Unit)
	}
	for _, j := range cat.ExactJoins {
		tExactJoins.add(j.ModelID, j.Provider, j.UpstreamID)
	}
	for _, d := range cat.Disagreements {
		tDisagreements.add(d.ModelID, d.Component, d.VercelPerMillion, d.ModelsDevPerMillion)
	}
	res.CatalogModels = len(cat.Models)
	res.ProviderOffers = len(cat.Offers)
	res.PriceFacts = len(cat.PriceFacts)
	res.ExactJoins = len(cat.ExactJoins)
	res.PricingDisagreements = len(cat.Disagreements)

	// 3. review_results: one row per batch directory (sorted)
	batchDirs, err := listBatchDirs(root)
	if err != nil {
		return nil, nil, err
	}
	adapterOf := cfg.AdapterOf()
	for _, batch := range batchDirs {
		if adapterOf[batch] == "" {
			return nil, nil, invalidf("build.json: batch %q has no batch_formats mapping", batch)
		}
		grade, reason := "quarantined", DefaultQuarantinedReason
		if d := led.CatalogAccepted(batch); d != nil {
			grade, reason = "accepted", d.Reason
		}
		tReviewResults.add(batch, grade, reason)
	}

	// 4. batch rows
	var proposalRows [][2]any // (unresolved row pointer, model_id) resolved after ids
	type unresRef struct {
		idx int // index into tUnresolved.rows
		mid string
	}
	var refs []unresRef
	for _, batch := range batchDirs {
		adapter := adapterOf[batch]
		b, err := LoadBatch(root, batch, adapter)
		if err != nil {
			return nil, nil, err
		}
		if b.HasMeasurements() {
			// §2.3.1 hard gate
			if led.CurrentMeasurements(batch) == nil {
				return nil, nil, &Error{Code: ExitInconsist, Msg: fmt.Sprintf("batch %q carries measurement rows without a current published_measurements decision; record the review decision or move the batch to the staging lane", batch)}
			}
		}
		for _, r := range b.Models {
			tModelResearch.add(r.ID, batch, r.BenchmarkStatus, r.Notes)
			for _, pol := range []struct {
				items []ClaimItem
				des   string
			}{{r.Strengths, "strength"}, {r.Weaknesses, "weakness"}} {
				for _, a := range pol.items {
					source, url, loc := a.SourceFile, a.SourceURL, a.Locator
					resolved := false
					if a.Basis == "catalog_claim" {
						source, url, loc = vercelCapture, baselineCatalogURL, "data entry id="+r.ID+"; "+a.Locator
						resolved = true
					} else {
						source, resolved, err = b.ResolveSourceFile(root, a.SourceFile, inv)
						if err != nil {
							return nil, nil, err
						}
					}
					if !resolved {
						tUnlocated.add(r.ID, batch, pol.des, a.Basis, a.Text, url, loc, UnlocatedNoFile)
						res.UnlocatedClaims++
						continue
					}
					if url == "" {
						tUnlocated.add(r.ID, batch, pol.des, a.Basis, a.Text, url, loc, UnlocatedNoURL)
						res.UnlocatedClaims++
						continue
					}
					tClaims.add(r.ID, batch, pol.des, a.Basis, a.Text, url, source, loc)
					res.Claims++
				}
			}
		}
		res.ModelResearch += len(b.Models)
		for _, bm := range b.Benchmarks {
			source, resolved, err := b.ResolveSourceFile(root, bm.SourceFile, inv)
			if err != nil {
				return nil, nil, err
			}
			if !resolved {
				return nil, nil, invalidf("data/research/%s/findings.json: benchmark for %s: source_file %q is unresolvable (schema requires it)", batch, bm.ModelID, bm.SourceFile)
			}
			if bm.SourceURL == "" {
				return nil, nil, invalidf("data/research/%s/findings.json: benchmark for %s: source_url is required (schema requires it)", batch, bm.ModelID)
			}
			tBenchmarks.add(bm.ModelID, batch, bm.Benchmark, bm.BenchmarkVersion, bm.Metric, bm.Value, bm.Unit, bm.Configuration, bm.SourceURL, source, bm.Locator, bm.EvidenceClass, bm.IdentityBasis)
			res.BenchmarkMeasurements++
		}
		for _, um := range b.Measurements {
			source, resolved, err := b.ResolveSourceFile(root, um.SourceFile, inv)
			if err != nil {
				return nil, nil, err
			}
			if !resolved {
				return nil, nil, invalidf("data/research/%s/findings.json: measurement for %s: source_file %q is unresolvable (schema requires it)", batch, um.ReportedModel, um.SourceFile)
			}
			if um.SourceURL == "" {
				return nil, nil, invalidf("data/research/%s/findings.json: measurement for %s: source_url is required (schema requires it)", batch, um.ReportedModel)
			}
			tUnresolved.add(batch, um.ReportedModel, um.Benchmark, um.BenchmarkVersion, um.Metric, um.Value, um.Unit, um.Configuration, um.SourceURL, source, um.Locator, um.EvidenceClass, um.IdentityStatus)
			idx := len(tUnresolved.rows) - 1
			for _, mid := range um.ProposedOfferIDs {
				refs = append(refs, unresRef{idx: idx, mid: mid})
			}
			res.UnresolvedMeasurements++
		}
		for _, s := range b.Sources {
			tSourceAccess.add(batch, s.URL, s.File, s.Status, s.Purpose, s.License)
			res.SourceAccess++
		}
		for i, g := range b.Gaps {
			tGaps.add(batch, int64(i), g)
		}
		res.ResearchGaps += len(b.Gaps)
	}

	// 4b. unresolved_measurements gets deterministic generated ids: sort by
	// non-generated tuple now (writeDB re-sorts stably, a no-op) and assign
	// ids so identity proposals can reference them.
	sort.SliceStable(tUnresolved.rows, func(i, j int) bool { return tUnresolved.less(tUnresolved.rows[i], tUnresolved.rows[j]) })
	for i := range tUnresolved.rows {
		tUnresolved.rows[i][0] = int64(i + 1)
	}
	for _, r := range refs {
		tProposals.add(tUnresolved.rows[r.idx][0].(int64), r.mid)
	}

	// 5. evidence reviews + inputs
	reviewBatches := led.ReviewBatches()
	for _, batch := range reviewBatches {
		d := led.CurrentMeasurements(batch)
		if d == nil {
			continue
		}
		tEvidenceReviews.add(batch, "published_measurements", "source_validated", d.ReviewPath)
		computed := map[string]bool{d.ReviewPath: true, "data/research/" + batch + "/findings.json": true}
		for _, sp := range measurementSourcePaths(tUnresolved, tBenchmarks, batch) {
			computed[sp] = true
		}
		if err := led.CheckPinSet(batch, computed); err != nil {
			return nil, nil, err
		}
		for _, p := range d.Pins {
			tEvidenceInputs.add(batch, p.Path, p.SHA256)
		}
		res.EvidenceReviews++
	}

	_ = proposalRows

	// 6. metadata
	for _, kv := range cfg.MetadataPairs() {
		tMetadata.add(kv[0], kv[1])
	}

	// 7. source policies (§2.5): current decision per captured URL; the
	// derived timestamps follow the spec recipe (checked_at = effective_at =
	// evidence.retrieved_at, expires_at = +90 days). Files load in name
	// order; rows within a file are pre-sorted by url, so insertion order is
	// deterministic.
	tSourcePolicies := newTable("source_policies", "source_policies", false, "SSSSSSSSSSSSSSSS")
	for _, batch := range pol.Batches() {
		for _, r := range pol.ByBatch[batch] {
			tt, err := time.Parse("2006-01-02T15:04:05.000Z", r.Evidence.RetrievedAt)
			if err != nil {
				return nil, nil, invalidf("policies/%s.json: evidence.retrieved_at %q unparseable", batch, r.Evidence.RetrievedAt)
			}
			expires := tt.AddDate(0, 0, 90).UTC().Format("2006-01-02T15:04:05.000Z")
			tSourcePolicies.add(r.URL, batch, r.SourceKind, r.TermsURL, r.Research, r.Extraction,
				r.Redistribution, r.LicenseName, r.LicenseText, r.Evidence.Path,
				r.Evidence.RetrievedAt, r.Evidence.Outcome, r.Evidence.RetrievedAt,
				r.Evidence.RetrievedAt, expires, r.Notes)
		}
	}

	// 24. qualification layer (§12, schema 5) — derived, no new inputs.
	obsAt := make(map[string]string, len(inv.Rows))
	for _, row := range inv.Rows {
		obsAt[row.Path] = row.ObservedAt
	}
	tOfferBindings := newTable("offer_binding_qualifications", "offer_binding_qualifications", false, "SSSSSSS")
	joinByModel := make(map[string]ExactJoin, len(cat.ExactJoins))
	for _, j := range cat.ExactJoins {
		joinByModel[j.ModelID] = j
	}
	for _, m := range cat.Models {
		if j, ok := joinByModel[m.ID]; ok {
			tOfferBindings.add(m.ID, "exact_join", j.Provider, j.UpstreamID, obsAt[vercelCapture], obsAt[modelsDevCapture], "vercel catalog exact-id join to models.dev entry")
		} else {
			tOfferBindings.add(m.ID, "unjoined", "", "", obsAt[vercelCapture], "", "no exact-id join in the models.dev capture; unmatched catalog entry")
		}
	}
	tSettings := newTable("measurement_settings", "measurement_settings", false, "SISSS")
	tComparisons := newTable("comparison_qualifications", "comparison_qualifications", false, "SIIS")
	// benchmark_measurements ids are assigned by insert order over the sorted
	// tuple; pre-sort and pre-assign so derivation references the same ids
	// (writeDB's stable re-sort is a no-op on pre-sorted rows, mirroring 4b).
	sort.SliceStable(tBenchmarks.rows, func(i, j int) bool { return tBenchmarks.less(tBenchmarks.rows[i], tBenchmarks.rows[j]) })
	for i := range tBenchmarks.rows {
		tBenchmarks.rows[i][0] = int64(i + 1)
	}
	for _, r := range tBenchmarks.rows {
		config, _ := r[8].(string)
		version, _ := r[4].(string)
		settings := deriveSettings(config, version)
		blocking := ""
		for _, s := range settings {
			tSettings.add("linked", r[0], s.component, s.knowledge, s.valueOrReason)
			if s.knowledge == "unknown" {
				if blocking != "" {
					blocking += ","
				}
				blocking += s.component
			}
		}
		eligible := 0
		if blocking == "" {
			eligible = 1
		}
		tComparisons.add("linked", r[0], eligible, blocking)
	}
	for _, r := range tUnresolved.rows {
		config, _ := r[8].(string)
		version, _ := r[4].(string)
		settings := deriveSettings(config, version)
		blocking := ""
		for _, s := range settings {
			tSettings.add("unresolved", r[0], s.component, s.knowledge, s.valueOrReason)
			if s.knowledge == "unknown" {
				if blocking != "" {
					blocking += ","
				}
				blocking += s.component
			}
		}
		tComparisons.add("unresolved", r[0], 0, blocking)
	}
	tLimitQuals := newTable("limit_qualifications", "limit_qualifications", false, "SSISSSS")
	for _, r := range tCatalogLimits.rows {
		tLimitQuals.add(r[0], r[1], r[2], "unknown", "unknown", "unknown", "catalog limit field carries no direction/operation/accounting semantics")
	}
	tPriceInterp := newTable("price_dimension_interpretations", "price_dimension_interpretations", false, "SSSS")
	seenPair := map[[2]string]bool{}
	for _, r := range tPriceFacts.rows {
		unit, _ := r[5].(string)
		path, _ := r[3].(string)
		key := [2]string{unit, path}
		if seenPair[key] {
			continue
		}
		seenPair[key] = true
		class, basis := classifyPriceDimension(unit, path)
		tPriceInterp.add(unit, path, class, basis)
	}

	all := []*tableRows{tSourceFiles, tCatalogModels, tCatalogModalities, tCatalogLimits,
		tCatalogParams, tCatalogTags, tProviderOffers, tCanonicalLabels, tExactJoins,
		tPriceFacts, tDisagreements, tReviewResults, tModelResearch, tClaims, tUnlocated,
		tBenchmarks, tUnresolved, tProposals, tSourceAccess, tGaps, tEvidenceReviews,
		tEvidenceInputs, tMetadata, tSourcePolicies, tOfferBindings, tSettings,
		tComparisons, tLimitQuals, tPriceInterp}
	return res, all, nil
}

// commonComponents is the fixed §12.2 component order.
var commonComponents = []string{
	"benchmark_release", "split", "protocol_harness_version", "prompt_template",
	"shots", "decoding_sampling_policy", "seed_distribution", "generation_count",
	"output_compute_budget", "reasoning_effort", "tool_policy", "precision_quantization",
}

const unknownReason = "not documented in retained source bytes"

type settingRow struct {
	component     string
	knowledge     string
	valueOrReason string
}

// deriveSettings applies the §12.2 keyword rules to one measurement's
// configuration sentence and benchmark_version.
func deriveSettings(config, version string) []settingRow {
	lower := strings.ToLower(config)
	keyword := func(needle string) (bool, string) {
		if strings.Contains(lower, needle) {
			return true, needle
		}
		return false, ""
	}
	out := make([]settingRow, 0, len(commonComponents))
	for _, comp := range commonComponents {
		switch comp {
		case "benchmark_release":
			if version == "" || version == "unknown" {
				out = append(out, settingRow{comp, "unknown", unknownReason})
			} else {
				out = append(out, settingRow{comp, "known", version})
			}
		case "tool_policy":
			if ok, v := keyword("no tools"); ok {
				out = append(out, settingRow{comp, "known", v})
			} else if ok, v := keyword("with tools"); ok {
				out = append(out, settingRow{comp, "known", v})
			} else {
				out = append(out, settingRow{comp, "unknown", unknownReason})
			}
		case "output_compute_budget":
			if ok, v := keyword("no test-time compute"); ok {
				out = append(out, settingRow{comp, "known", v})
			} else {
				out = append(out, settingRow{comp, "unknown", unknownReason})
			}
		case "reasoning_effort":
			if ok, v := keyword("reasoning-effort"); ok {
				out = append(out, settingRow{comp, "known", v})
			} else {
				out = append(out, settingRow{comp, "unknown", unknownReason})
			}
		default:
			out = append(out, settingRow{comp, "unknown", unknownReason})
		}
	}
	return out
}

// classifyPriceDimension applies the §12.5 terminal-segment rules.
func classifyPriceDimension(unit, fieldPath string) (class, basis string) {
	seg := fieldPath
	if i := strings.LastIndex(fieldPath, "/"); i >= 0 {
		seg = fieldPath[i+1:]
	}
	if seg == "size" {
		return "threshold_count", "tier size is a count threshold, not a monetary value"
	}
	if unit != "USD_per_million_tokens" && unit != "USD_per_token" {
		return "unresolved", "path/unit semantics not documented in retained catalog bytes"
	}
	scale := "per-million"
	if unit == "USD_per_token" {
		scale = "per-token"
	}
	tier := ""
	if strings.Contains(fieldPath, "tiers/") || strings.Contains(fieldPath, "context_over_200k") {
		tier = "; tier condition retained verbatim in field_path"
	}
	switch seg {
	case "input":
		return "token_input_monetary", "monetary " + scale + " token input" + tier
	case "output":
		return "token_output_monetary", "monetary " + scale + " token output" + tier
	case "cache_read", "input_cache_read":
		return "token_cache_read_monetary", "monetary " + scale + " token cache read" + tier
	case "cache_write":
		return "token_cache_write_monetary", "monetary " + scale + " token cache write" + tier
	default:
		return "unresolved", "path/unit semantics not documented in retained catalog bytes"
	}
}

// measurementSourcePaths returns distinct source paths for a batch from the
// assembled rows (the pin-set basis).
func measurementSourcePaths(tUnresolved, tBenchmarks *tableRows, batch string) []string {
	seen := map[string]bool{}
	for _, r := range tBenchmarks.rows {
		if r[2] == batch { // model_id,batch,...
			if s, ok := r[10].(string); ok {
				seen[s] = true
			}
		}
	}
	for _, r := range tUnresolved.rows {
		if r[1] == batch { // batch is col 1 after generated id
			if s, ok := r[10].(string); ok {
				seen[s] = true
			}
		}
	}
	var out []string
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// writeDB creates a fresh SQLite database at path with pinned pragmas and
// all tables inserted per §4 ordering.
func writeDB(path string, tables []*tableRows) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return iof("mkdir staging: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return iof("open staged db: %v", err)
	}
	defer db.Close()
	for _, pragma := range []string{
		`PRAGMA page_size=4096`,
		`PRAGMA journal_mode=OFF`,
		`PRAGMA encoding='UTF-8'`,
		`PRAGMA foreign_keys=ON`,
	} {
		if _, err := db.Exec(pragma); err != nil {
			return iof("pragma %s: %v", pragma, err)
		}
	}
	if _, err := db.Exec(SchemaDDL); err != nil {
		return iof("staged schema: %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		return iof("begin: %v", err)
	}
	defer tx.Rollback()
	for _, t := range tables {
		if err := t.insertInto(tx); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return iof("commit: %v", err)
	}
	if err := db.Close(); err != nil {
		return iof("close staged db: %v", err)
	}
	return nil
}

func listBatchDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "data", "research"))
	if err != nil {
		return nil, iof("read data/research: %v", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "terms" { // terms/ is the §2.5 evidence area, not a batch
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func firstWords(q string) string {
	f := strings.Fields(q)
	if len(f) > 3 {
		f = f[:3]
	}
	return strings.Join(f, " ")
}
