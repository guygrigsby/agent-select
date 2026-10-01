package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRepo creates a minimal repository skeleton under t.TempDir() and
// returns its root path.
func writeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for p, content := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestPathRules(t *testing.T) {
	valid := []string{
		"data/raw/models-dev.json",
		"data/research/anthropic/findings.json",
		"data/reviews/source-evidence-2026-09-30.md",
		"data/research/images/sources/hf_gemma.md",
	}
	for _, p := range valid {
		if !ValidEvidencePath(p) {
			t.Errorf("expected valid: %s", p)
		}
	}
	invalid := []string{
		"/etc/passwd",
		"data/../secret.txt",
		"data/research-staging/x/findings.json",
		"data/research/a b/findings.json",
		"data//double",
		"other/dir/file.txt",
		"",
	}
	for _, p := range invalid {
		if ValidEvidencePath(p) {
			t.Errorf("expected invalid: %s", p)
		}
	}
	if !isControlPath("data/research-inputs/build.json") ||
		!isControlPath("data/research-inputs/capture-inventory.tsv") ||
		!isControlPath("data/research-inputs/fetch-logs/anthropic.log") ||
		!isControlPath("data/reviews/decisions/google.v1.json") {
		t.Error("control path predicate failed on known control paths")
	}
	for _, p := range []string{"data/research-inputs/other.json", "data/research-inputs/fetch-logs/../x.log", "data/reviews/decisions/x.md"} {
		if isControlPath(p) {
			t.Errorf("expected non-control: %s", p)
		}
	}
}

var goodInventory = "path\torigin\turl\tsha256\tbytes\tobserved_at\ttimestamp_basis\tlicense\tpublication_permission\trecovery_basis\n" +
	"data/raw/a.json\tcaptured_remote\thttps://x.test/a\t" + "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a" + "\t2\t2026-09-30T00:00:00Z\tlocal_file_mtime_not_http_retrieval\tunknown\tunknown\t\n"

func TestInventoryValid(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"data/research-inputs/capture-inventory.tsv": goodInventory,
		"data/raw/a.json": "{}",
	})
	inv, err := LoadInventory(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(inv.Rows) != 1 || inv.Rows[0].Path != "data/raw/a.json" {
		t.Fatalf("rows: %+v", inv.Rows)
	}
	if err := inv.VerifyDisk(root); err != nil {
		t.Fatalf("verify disk: %v", err)
	}
}

func TestInventoryRejectsTamperedBytes(t *testing.T) {
	root := writeRepo(t, map[string]string{
		"data/research-inputs/capture-inventory.tsv": goodInventory,
		"data/raw/a.json": "XX", // bytes differ from pinned sha
	})
	inv, err := LoadInventory(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	err = inv.VerifyDisk(root)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 on disk-vs-inventory mismatch, got %v", err)
	}
}

func TestInventoryGate5(t *testing.T) {
	// captured_remote with neither url nor recovery_basis -> exit 1
	bad := strings.Replace(goodInventory, "https://x.test/a\t"+"44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", 1)
	root := writeRepo(t, map[string]string{
		"data/research-inputs/capture-inventory.tsv": bad,
	})
	_, err := LoadInventory(root)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 for captured row without url or basis, got %v", err)
	}
	// local_derived with a url -> exit 1
	bad2 := strings.Replace(goodInventory, "captured_remote", "local_derived", 1)
	root = writeRepo(t, map[string]string{
		"data/research-inputs/capture-inventory.tsv": bad2,
	})
	_, err = LoadInventory(root)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 for local_derived row with url, got %v", err)
	}
}

func TestInventoryRejectsControlRow(t *testing.T) {
	row := "data/research-inputs/build.json\tlocal_derived\t\t" + "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a" + "\t1\t2026-09-30T00:00:00Z\tb\tunknown\tunknown\t\n"
	root := writeRepo(t, map[string]string{
		"data/research-inputs/capture-inventory.tsv": goodInventory + row,
	})
	_, err := LoadInventory(root)
	if err == nil || !strings.Contains(err.Error(), "control file") {
		t.Fatalf("expected control-file rejection, got %v", err)
	}
}

var goodDecision = `{
  "batch": "google",
  "scope": "published_measurements",
  "state": "source_validated",
  "reason": "review",
  "review_path": "data/reviews/r.md",
  "pins": [{"path": "data/research/google/findings.json", "sha256": "` + strings.Repeat("b", 64) + `"}],
  "decision_version": 1,
  "decided_at": "2026-09-30T03:34:48Z"
}`

func decisionRepo(t *testing.T, decision string) string {
	return writeRepo(t, map[string]string{
		"data/reviews/decisions/google.v1.json": decision,
	})
}

func TestDecisionValid(t *testing.T) {
	l, err := LoadDecisions(decisionRepo(t, goodDecision))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if l.Current("google", ScopeMeasurements) == nil {
		t.Fatal("expected current decision")
	}
}

func TestDecisionSupersessionOrder(t *testing.T) {
	v2 := strings.Replace(strings.Replace(goodDecision, `"decision_version": 1`, `"decision_version": 2`, 1), "2026-09-30T03:34:48Z", "2026-09-30T04:30:00Z", 1)
	root := writeRepo(t, map[string]string{
		"data/reviews/decisions/google.v1.json": goodDecision,
		"data/reviews/decisions/google.v2.json": v2,
	})
	l, err := LoadDecisions(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := l.Current("google", ScopeMeasurements); got == nil || got.DecisionVersion != 2 {
		t.Fatalf("expected v2 current, got %+v", got)
	}
}

func TestDecisionUnknownScope(t *testing.T) {
	bad := strings.Replace(goodDecision, `"published_measurements"`, `"benchmark_acceptance"`, 1)
	_, err := LoadDecisions(decisionRepo(t, bad))
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 for unknown scope, got %v", err)
	}
}

func TestPinGate(t *testing.T) {
	inv := &Inventory{byPath: map[string]*InvRow{
		"data/research/google/findings.json": {Path: "data/research/google/findings.json", SHA256: strings.Repeat("c", 64)},
		"data/reviews/r.md":                  {Path: "data/reviews/r.md", SHA256: strings.Repeat("d", 64)},
	}}
	l, err := LoadDecisions(decisionRepo(t, goodDecision))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	bf := map[string]string{"google": AdapterModelBatch}
	bm := map[string]bool{"google": true}

	// pin disagrees with inventory -> exit 2 (stale)
	err = l.CheckPinGate(inv, bf, bm)
	if err == nil || err.(*Error).Code != ExitStale {
		t.Fatalf("expected exit 2, got %v", err)
	}

	// pin path not inventoried -> exit 1
	fixed := strings.Replace(goodDecision, strings.Repeat("b", 64), strings.Repeat("c", 64), 1)
	l2, _ := LoadDecisions(decisionRepo(t, fixed))
	err = l2.CheckPinGate(&Inventory{byPath: map[string]*InvRow{}}, bf, bm)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 uninventoried pin, got %v", err)
	}

	// matching pin -> ok
	invOK := &Inventory{byPath: map[string]*InvRow{
		"data/research/google/findings.json": {SHA256: strings.Repeat("c", 64)},
		"data/reviews/r.md":                  {SHA256: strings.Repeat("d", 64)},
	}}
	if err := l2.CheckPinGate(invOK, bf, bm); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}

	// scope/adapter binding: catalog scope on a model_batch batch -> exit 1
	cat := strings.Replace(strings.Replace(goodDecision, `"published_measurements"`, `"catalog_sources"`, 1), `"source_validated"`, `"accepted"`, 1)
	l3, _ := LoadDecisions(decisionRepo(t, cat))
	err = l3.CheckPinGate(invOK, bf, bm)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 scope/adapter mismatch, got %v", err)
	}

	// published_measurements scope on a batch without measurements -> exit 1
	bmNo := map[string]bool{}
	err = l2.CheckPinGate(invOK, bf, bmNo)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 batch without measurements, got %v", err)
	}
}

func TestConfigLoad(t *testing.T) {
	cfg := `{
  "artifact_version": "0.2.0", "schema_version": 5,
  "phase": "research_only_not_routing_authority", "publication_permission": "unknown",
  "baseline_url": "u", "pricing_crossref_url": "p", "collection_date_utc": "2026-09-30",
  "publication_permission_vocabulary": ["unknown","allowed","denied"],
  "batch_formats": {"google": "model_batch"},
  "expected": {"catalog_models":395,"exact_catalog_joins":394,"model_research":302,"price_facts":31659,"pricing_disagreements":12}
}`
	root := writeRepo(t, map[string]string{"data/research-inputs/build.json": cfg})
	c, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(c.MetadataPairs()) != 10 {
		t.Fatalf("metadata pairs: %d", len(c.MetadataPairs()))
	}
	if c.Expected.PriceFacts != 31659 {
		t.Fatal("expected counts not parsed")
	}

	// bad vocabulary value -> exit 1
	bad := strings.Replace(cfg, `"allowed"`, `"maybe"`, 1)
	root = writeRepo(t, map[string]string{"data/research-inputs/build.json": bad})
	_, err = LoadConfig(root)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 vocabulary, got %v", err)
	}

	// schema mismatch -> exit 3
	bad2 := strings.Replace(cfg, `"schema_version": 5`, `"schema_version": 4`, 1)
	root = writeRepo(t, map[string]string{"data/research-inputs/build.json": bad2})
	_, err = LoadConfig(root)
	if err == nil || err.(*Error).Code != ExitInconsist {
		t.Fatalf("expected exit 3 schema mismatch, got %v", err)
	}

	// unknown adapter -> exit 1
	bad3 := strings.Replace(cfg, `"model_batch"`, `"csv_reader"`, 1)
	root = writeRepo(t, map[string]string{"data/research-inputs/build.json": bad3})
	_, err = LoadConfig(root)
	if err == nil || err.(*Error).Code != ExitInvalid {
		t.Fatalf("expected exit 1 adapter, got %v", err)
	}
}
