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
