package research

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openFixtureDB creates an in-memory schema-3 database.
func openFixtureDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1) // :memory: is per-connection; keep one schema instance
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(SchemaDDL); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func mustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

// seedMinimal inserts the metadata row set gates need.
func seedMetadata(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, kv := range [][2]string{
		{"expected_catalog_models", "1"}, {"expected_exact_joins", "0"},
		{"expected_text_and_image_research_models", "1"},
	} {
		mustExec(t, db, "INSERT INTO metadata VALUES(?,?)", kv[0], kv[1])
	}
}

// TestInverseStatusGateRed is the red test for the four MiniMax/Kimi defects:
// measurements present but benchmark_status='unresearched' must be rejected
// (§5.2). The frozen corpus carried exactly this contradiction undetected.
func TestInverseStatusGateRed(t *testing.T) {
	db := openFixtureDB(t)
	seedMetadata(t, db)
	mustExec(t, db, `INSERT INTO source_files VALUES('data/research/x/findings.json','https://example.invalid/f','`+sha0+`',10,'2026-09-30T00:00:00+00:00','basis','unknown','unknown')`)
	mustExec(t, db, `INSERT INTO catalog_models VALUES('m','m','o','language','d','data/research/x/findings.json')`)
	mustExec(t, db, `INSERT INTO model_research VALUES('m','x','unresearched','n')`)
	mustExec(t, db, `INSERT INTO benchmark_measurements(model_id,batch,benchmark,benchmark_version,metric,raw_value,unit,configuration,source_url,source_path,locator,evidence_class,identity_basis) VALUES('m','x','b','v','acc','1.0','pct','','https://example.invalid/s','data/research/x/findings.json','loc','provider_reported','exact_id')`)
	if err := RunGate2(db); err == nil {
		t.Fatal("inverse gate accepted a model with measurements marked unresearched")
	}
}

// TestAccessStatusGate covers §5.3: failure status vs success-format bytes,
// the API-error-envelope exception, and the bound (non-JSON/CSV skipped).
func TestAccessStatusGate(t *testing.T) {
	root := t.TempDir()
	db := openFixtureDB(t)

	mk := func(p, content string) {
		full := filepath.Join(root, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte(content), 0o644)
	}
	// a genuine API 404 envelope under a .json name: honest failure, not flagged
	mk("data/research/x/sources/gone.json", `{"message":"Not Found","documentation_url":"https://docs.github.com/rest","status":"404"}`)
	// a valid scores object declared not_found: contradiction
	mk("data/research/x/sources/scores.json", `{"results":[{"model":"m","acc":0.5}]}`)
	// valid CSV declared blocked: contradiction
	mk("data/research/x/sources/table.csv", "model,acc\nm,0.5\n")
	// HTML capture declared blocked: outside the gate's bound
	mk("data/research/x/sources/page.html", "<html>fine</html>")
	// literal 404 text under .json: fails parse, not flagged
	mk("data/research/x/sources/literal.json", "404: Not Found")

	ins := `INSERT INTO source_access VALUES(?,?,?,?,?,?)`
	mustExec(t, db, ins, "x", "https://e.invalid/gone", "data/research/x/sources/gone.json", "not_found", "p", "unknown")
	mustExec(t, db, ins, "x", "https://e.invalid/html", "data/research/x/sources/page.html", "blocked", "p", "unknown")
	mustExec(t, db, ins, "x", "https://e.invalid/literal", "data/research/x/sources/literal.json", "not_found", "p", "unknown")

	if err := RunGate3(root, db); err != nil {
		t.Fatalf("honest failure rows must pass: %v", err)
	}

	mustExec(t, db, ins, "x", "https://e.invalid/scores", "data/research/x/sources/scores.json", "not_found", "p", "unknown")
	err := RunGate3(root, db)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("success-format JSON declared not_found must exit 3, got %v", err)
	}
	// remove the offending row, add the CSV contradiction
	mustExec(t, db, `DELETE FROM source_access WHERE path='data/research/x/sources/scores.json'`)
	mustExec(t, db, ins, "x", "https://e.invalid/csv", "data/research/x/sources/table.csv", "blocked_404", "p", "unknown")
	if err := RunGate3(root, db); err == nil {
		t.Fatal("success-format CSV declared blocked must fail")
	}
}

// TestGate1EnforcesPinsAndViews seeds a measurement without its review input
// pin and expects the unpinned-input gate to fire.
func TestGate1EnforcesPinsAndViews(t *testing.T) {
	db := openFixtureDB(t)
	seedMetadata(t, db)
	mustExec(t, db, `INSERT INTO source_files VALUES('data/research/x/findings.json','https://example.invalid/f','`+sha0+`',10,'2026-09-30T00:00:00+00:00','basis','unknown','unknown')`)
	mustExec(t, db, `INSERT INTO catalog_models VALUES('m','m','o','language','d','data/research/x/findings.json')`)
	mustExec(t, db, `INSERT INTO model_research VALUES('m','x','found','n')`)
	mustExec(t, db, `INSERT INTO review_results VALUES('x','quarantined','r')`)
	mustExec(t, db, `INSERT INTO benchmark_measurements(model_id,batch,benchmark,benchmark_version,metric,raw_value,unit,configuration,source_url,source_path,locator,evidence_class,identity_basis) VALUES('m','x','b','v','acc','1.0','pct','','https://example.invalid/s','data/research/x/findings.json','loc','provider_reported','exact_id')`)
	// measurement exists but no evidence_reviews row: the source-validated
	// equality gate must fire (stored score lacks source review)
	err := RunGate1(db)
	if err == nil {
		t.Fatal("gate 1 accepted a stored score without source review")
	}
	mustExec(t, db, `INSERT INTO evidence_reviews VALUES('x','published_measurements','source_validated','data/research/x/findings.json')`)
	err = RunGate1(db)
	if err == nil {
		t.Fatal("gate 1 accepted a review without its findings pin row")
	}
	mustExec(t, db, `INSERT INTO evidence_review_inputs VALUES('x','data/research/x/findings.json','`+sha0+`')`)
	if err := RunGate1(db); err != nil {
		t.Fatalf("fully pinned fixture must pass gate 1: %v", err)
	}
}

const sha0 = "0000000000000000000000000000000000000000000000000000000000000000"

// TestEnvelopeDetector locks the envelope exception semantics.
func TestEnvelopeDetector(t *testing.T) {
	envelopes := []string{
		`{"message":"Not Found","documentation_url":"https://x","status":"404"}`,
		`{"error":"Invalid username or password."}`,
		`{"message":"Invalid username or password."}`,
	}
	for _, e := range envelopes {
		if parsesAsSuccessFormat("a.json", []byte(e)) {
			t.Fatalf("envelope must not count as success format: %s", e)
		}
	}
	success := []string{
		`{"results":[1,2,3]}`,
		`[{"model":"m"}]`,
	}
	for _, s := range success {
		if !parsesAsSuccessFormat("a.json", []byte(s)) {
			t.Fatalf("valid JSON must parse as success format: %s", s)
		}
	}
	if parsesAsSuccessFormat("a.json", []byte("404: Not Found")) {
		t.Fatal("literal 404 text must not parse")
	}
	if parsesAsSuccessFormat("a.csv", []byte("model\nm\n")) {
		t.Fatal("single-column CSV must not count as success format")
	}
	if !parsesAsSuccessFormat("a.csv", []byte("model,acc\nm,0.5\n")) {
		t.Fatal("two-column CSV with data row must parse")
	}
}

// TestPolicyLoadDiagnostics covers the §2.5 load-time diagnostics from the
// source-policies review: duplicate URLs across files and unknown batch
// names must surface as exit 1 before any database insert runs.
func TestPolicyLoadDiagnostics(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "research-inputs", "policies")
	os.MkdirAll(dir, 0o755)
	rec := func(u string) string {
		return `[{"url":"` + u + `","source_kind":"provider_docs","terms_url":"https://e.invalid/t","research":"unknown","extraction":"unknown","redistribution":"unknown","license_name":"","license_text":"","evidence":{"path":"data/research/x/y.json","retrieved_at":"2026-09-30T00:00:00.000Z","outcome":"success"}}]`
	}
	os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(rec("https://e.invalid/a")), 0o644)
	os.WriteFile(filepath.Join(dir, "beta.json"), []byte(rec("https://e.invalid/b")), 0o644)
	os.WriteFile(filepath.Join(dir, "gamma.json"), []byte(rec("https://e.invalid/a")), 0o644)

	pol, err := LoadPolicies(root)
	if err == nil {
		t.Fatal("duplicate URL across files must fail load")
	}
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInvalid {
		t.Fatalf("duplicate URL must be exit 1, got %v", err)
	}
	if !strings.Contains(err.Error(), "alpha.json") || !strings.Contains(err.Error(), "gamma.json") {
		t.Fatalf("duplicate diagnostic must name both files, got %v", err)
	}

	// drop the duplicate; unknown batch names are load-time exit 1
	os.Remove(filepath.Join(dir, "gamma.json"))
	pol, err = LoadPolicies(root)
	if err != nil {
		t.Fatalf("clean pair must load: %v", err)
	}
	if err := CheckPolicyBatches(pol, []string{"alpha"}); err == nil {
		t.Fatal("beta.json names a batch without findings and must fail")
	} else if ce, ok := err.(*Error); !ok || ce.Code != ExitInvalid {
		t.Fatalf("unknown batch must be exit 1, got %v", err)
	}
	if err := CheckPolicyBatches(pol, []string{"alpha", "beta"}); err != nil {
		t.Fatalf("known batches must pass: %v", err)
	}
}

// seedQualificationGreen seeds a minimal schema-5 database whose
// qualification tables equal recomputation: one exact-joined model with a
// linked measurement, one unresolved measurement, one limit, one price pair.
func seedQualificationGreen(t *testing.T, db *sql.DB) {
	t.Helper()
	mustExec(t, db, `INSERT INTO source_files VALUES('cap','https://e.invalid/cap','aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',1,'t1','mtime','unknown','unknown')`)
	mustExec(t, db, `INSERT INTO catalog_models VALUES('m','n','o','language','d','cap')`)
	mustExec(t, db, `INSERT INTO exact_catalog_joins VALUES('m','vercel','m')`)
	mustExec(t, db, `INSERT INTO offer_binding_qualifications VALUES('m','exact_join','vercel','m','t1','t1','vercel catalog exact-id join to models.dev entry')`)
	mustExec(t, db, `INSERT INTO benchmark_measurements VALUES(1,'m','b','bench','v1','acc','1','pct','5-shot; no tools','https://e.invalid/b','cap','l','provider_reported','exact naming')`)
	mustExec(t, db, `INSERT INTO unresolved_measurements VALUES(1,'b','label','bench','unknown','acc','1','pct','reasoning-effort high','https://e.invalid/u','cap','l','third_party_reported','unresolved')`)
	for _, tc := range []struct {
		scope string
		id    int64
		cfg   string
		ver   string
	}{
		{"linked", 1, "5-shot; no tools", "v1"},
		{"unresolved", 1, "reasoning-effort high", "unknown"},
	} {
		settings := deriveSettings(tc.cfg, tc.ver)
		blocking := ""
		for _, s := range settings {
			mustExec(t, db, `INSERT INTO measurement_settings VALUES(?,?,?,?,?)`, tc.scope, tc.id, s.component, s.knowledge, s.valueOrReason)
			if s.knowledge == "unknown" {
				if blocking != "" {
					blocking += ","
				}
				blocking += s.component
			}
		}
		eligible := 0
		if tc.scope == "linked" && blocking == "" {
			eligible = 1
		}
		mustExec(t, db, `INSERT INTO comparison_qualifications VALUES(?,?,?,?)`, tc.scope, tc.id, eligible, blocking)
	}
	mustExec(t, db, `INSERT INTO catalog_limits VALUES('m','context_window',128000)`)
	mustExec(t, db, `INSERT INTO limit_qualifications VALUES('m','context_window',128000,'unknown','unknown','unknown','catalog limit field carries no direction/operation/accounting semantics')`)
	mustExec(t, db, `INSERT INTO price_facts VALUES('cap','vercel','m','/input','0.25','USD_per_million_tokens')`)
	class, basis := classifyPriceDimension("USD_per_million_tokens", "/input")
	mustExec(t, db, `INSERT INTO price_dimension_interpretations VALUES(?,?,?,?)`, "USD_per_million_tokens", "/input", class, basis)
}

func TestQualificationGateGreen(t *testing.T) {
	db := openFixtureDB(t)
	seedQualificationGreen(t, db)
	if err := RunGate7(db); err != nil {
		t.Fatalf("green fixture must pass gate 7: %v", err)
	}
}

func TestQualificationGateRed(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"missing settings row", `DELETE FROM measurement_settings WHERE scope='linked' AND component='split'`},
		{"tampered blocking list", `UPDATE comparison_qualifications SET blocking_components='' WHERE scope='unresolved'`},
		{"hand-entered eligibility", `UPDATE comparison_qualifications SET eligible=1 WHERE scope='unresolved'`},
		{"extra settings row", `INSERT INTO measurement_settings VALUES('linked',99,'split','unknown','x')`},
		{"unjoined with provider", `UPDATE offer_binding_qualifications SET binding='unjoined' WHERE model_id='m'`},
		{"binding without join", `INSERT INTO catalog_models VALUES('m2','n2','o','language','d','cap'); INSERT INTO offer_binding_qualifications VALUES('m2','exact_join','vercel','x','t1','t1','b')`},
		{"limit value drift", `UPDATE limit_qualifications SET value=1000`},
		{"price class drift", `UPDATE price_dimension_interpretations SET dimension_class='threshold_count'`},
		{"price pair uncovered", `INSERT INTO price_facts VALUES('cap','vercel','m','/output','1','USD_per_million_tokens')`},
		{"known value absent from configuration", `UPDATE measurement_settings SET value_or_reason='8-shot' WHERE scope='linked' AND component='tool_policy'`},
		{"component renamed off the common list", `UPDATE measurement_settings SET component='custom_thing' WHERE component='split'`},
		{"joined model flipped to unjoined", `UPDATE offer_binding_qualifications SET binding='unjoined', provider='', upstream_id='' WHERE model_id='m'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := openFixtureDB(t)
			seedQualificationGreen(t, db)
			mustExec(t, db, c.query)
			if err := RunGate7(db); err == nil {
				t.Fatalf("%s must fail gate 7", c.name)
			}
		})
	}
}

func TestDeriveSettingsRules(t *testing.T) {
	get := func(ss []settingRow, comp string) settingRow {
		for _, s := range ss {
			if s.component == comp {
				return s
			}
		}
		t.Fatalf("component %s missing", comp)
		return settingRow{}
	}
	// unknown-sentinel version stays unknown (spec §12.2)
	ss := deriveSettings("no tools", "unknown")
	if s := get(ss, "benchmark_release"); s.knowledge != "unknown" || s.valueOrReason != unknownReason {
		t.Fatalf("unknown sentinel version must stay unknown: %+v", s)
	}
	// concrete version is known verbatim
	ss = deriveSettings("with tools; no test-time compute; reasoning-effort low", "v2")
	if s := get(ss, "benchmark_release"); s.knowledge != "known" || s.valueOrReason != "v2" {
		t.Fatalf("concrete version must be known verbatim: %+v", s)
	}
	if s := get(ss, "tool_policy"); s.knowledge != "known" || s.valueOrReason != "with tools" {
		t.Fatalf("with tools must match: %+v", s)
	}
	if s := get(ss, "output_compute_budget"); s.knowledge != "known" || s.valueOrReason != "no test-time compute" {
		t.Fatalf("no test-time compute must match: %+v", s)
	}
	if s := get(ss, "reasoning_effort"); s.knowledge != "known" || s.valueOrReason != "reasoning-effort" {
		t.Fatalf("reasoning-effort must match: %+v", s)
	}
	// keyword match is case-insensitive
	if s := get(deriveSettings("With Tools", ""), "tool_policy"); s.knowledge != "known" {
		t.Fatalf("case-insensitive match failed: %+v", s)
	}
	// undocumented components stay unknown with the fixed reason
	for _, comp := range []string{"split", "protocol_harness_version", "prompt_template", "shots", "decoding_sampling_policy", "seed_distribution", "generation_count", "precision_quantization"} {
		if s := get(ss, comp); s.knowledge != "unknown" || s.valueOrReason != unknownReason {
			t.Fatalf("%s must be unknown with fixed reason: %+v", comp, s)
		}
	}
}

func TestClassifyPriceDimension(t *testing.T) {
	cases := []struct {
		unit, path, class string
	}{
		{"USD_per_million_tokens", "/input", "token_input_monetary"},
		{"USD_per_token", "/output", "token_output_monetary"},
		{"USD_per_million_tokens", "/cache_read", "token_cache_read_monetary"},
		{"USD_per_million_tokens", "/input_cache_read", "token_cache_read_monetary"},
		{"USD_per_million_tokens", "/cache_write", "token_cache_write_monetary"},
		{"USD_per_million_tokens", "/context_over_200k/input", "token_input_monetary"},
		{"source_specific_unresolved", "/tiers/0/tier/size", "threshold_count"},
		{"source_specific_unresolved", "/reasoning", "unresolved"},
		{"USD_per_million_tokens", "/input_audio", "unresolved"},
	}
	for _, c := range cases {
		class, basis := classifyPriceDimension(c.unit, c.path)
		if class != c.class {
			t.Fatalf("%s %s: class %s want %s", c.unit, c.path, class, c.class)
		}
		if c.class == "unresolved" && basis != "path/unit semantics not documented in retained catalog bytes" {
			t.Fatalf("%s %s: unresolved basis %q", c.unit, c.path, basis)
		}
		if c.class != "unresolved" && basis == "" {
			t.Fatalf("%s %s: classified rows need a basis", c.unit, c.path)
		}
	}
	// tiered path names the retained condition
	_, basis := classifyPriceDimension("USD_per_million_tokens", "/context_over_200k/input")
	if !strings.Contains(basis, "tier condition retained verbatim in field_path") {
		t.Fatalf("tiered basis must note the condition: %q", basis)
	}
}
