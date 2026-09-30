package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBacklogCorpus runs the full backlog path on a corpus copy: rows for
// every catalog model, gates, TSV shape, report determinism and the drift
// check.
func TestBacklogCorpus(t *testing.T) {
	root := copyCorpus(t)
	tsv, report, summary, err := Backlog(root)
	if err != nil {
		t.Fatalf("Backlog: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(tsv), "\n"), "\n")
	if len(lines) != 396 { // header + 395 models
		t.Fatalf("backlog TSV rows = %d, want 396 (header + 395)", len(lines))
	}
	if lines[0] != BacklogHeader {
		t.Fatalf("header mismatch:\n%s\nwant\n%s", lines[0], BacklogHeader)
	}
	inScope, imageOut := 0, 0
	for _, l := range lines[1:] {
		c := strings.Split(l, "\t")
		if len(c) != 19 {
			t.Fatalf("row %q has %d columns, want 19", c[0], len(c))
		}
		if c[2] == "in_scope" {
			inScope++
		}
		if c[9] == "t" {
			imageOut++
		}
	}
	if inScope != 302 {
		t.Fatalf("in-scope models = %d, want 302", inScope)
	}
	if imageOut != 41 {
		t.Fatalf("image-output models = %d, want 41", imageOut)
	}
	if !strings.Contains(summary, "28 tasks") {
		t.Fatalf("summary %q missing task count", summary)
	}

	// Deterministic report.
	_, report2, _, err := Backlog(root)
	if err != nil {
		t.Fatalf("second Backlog: %v", err)
	}
	if string(report) != string(report2) {
		t.Fatal("backlog report is not deterministic")
	}
	for _, want := range []string{
		"catalog models: 395 (in scope 302, out of scope 93)",
		"image-output models: 41",
		"unlocated claims: 56 total",
		"252 unresolved measurements",
		"no image-specific limit facts",
		"host-registry-discovery",
	} {
		if !strings.Contains(string(report), want) {
			t.Fatalf("report missing %q", want)
		}
	}
}

// TestBacklogDriftCheck verifies the drift gate: a matching file passes, an
// edited one fails with exit 1.
func TestBacklogDriftCheck(t *testing.T) {
	root := copyCorpus(t)
	_, report, _, err := Backlog(root)
	if err != nil {
		t.Fatalf("Backlog: %v", err)
	}
	p := filepath.Join(root, "docs", "research", "backlog.md")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, report, 0o644)
	if err := BacklogCheck(root, "docs/research/backlog.md"); err != nil {
		t.Fatalf("drift check on fresh report: %v", err)
	}
	os.WriteFile(p, append([]byte("# edited\n"), report...), 0o644)
	err = BacklogCheck(root, "docs/research/backlog.md")
	if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid {
		t.Fatalf("stale report error = %v, want ExitInvalid", err)
	}
}

// TestBacklogTaskValidation drives LoadTasks rejections: uninventoried
// evidence file, unclaimed research batch, empty closure.
func TestBacklogTaskValidation(t *testing.T) {
	root := copyCorpus(t)
	_, inv, _, _, tables, err := loadInputs(root)
	if err != nil {
		t.Fatalf("loadInputs: %v", err)
	}
	batches := researchBatchNames(tables)

	p := filepath.Join(root, "data", "research-inputs", "backlog-tasks.json")
	orig, _ := os.ReadFile(p)
	defer os.WriteFile(p, orig, 0o644)

	for _, tc := range []struct {
		name, repl string
	}{
		{"uninventoried evidence", `"data/research/anthropic/findings.json"`},
		{"unclaimed batch", `"batches": ["anthropic"]`},
	} {
		switch tc.name {
		case "uninventoried evidence":
			os.WriteFile(p, []byte(strings.Replace(string(orig), `"data/research/anthropic/findings.json"`, `"data/research/nope/findings.json"`, 1)), 0o644)
		case "unclaimed batch":
			os.WriteFile(p, []byte(strings.Replace(string(orig), `"batches": ["anthropic"]`, `"batches": []`, 1)), 0o644)
		}
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid {
			t.Fatalf("%s: error = %v, want ExitInvalid", tc.name, err)
		}
	}
}

// TestBacklogImageGate adds a language model to the images batch in the
// findings file (with the inventory re-pinned, no other gate fires) so the
// images-batch assignment disagrees with modality facts; the gate must fire.
func TestBacklogImageGate(t *testing.T) {
	root := copyCorpus(t)
	fp := "data/research/images/findings.json"
	target := filepath.Join(root, filepath.FromSlash(fp))
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read findings: %v", err)
	}
	var d struct {
		Models []map[string]any `json:"models"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatalf("parse findings: %v", err)
	}
	// A language model that is in scope and not image-output.
	d.Models = append(d.Models, map[string]any{
		"id": "anthropic/claude-opus-4.8", "benchmark_status": "unresearched", "notes": "gate probe",
	})
	out, err := json.Marshal(&d)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, out, 0o644); err != nil {
		t.Fatal(err)
	}
	repinInventory(t, root, fp, out)

	_, _, _, err = Backlog(root)
	if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "image-output") {
		t.Fatalf("image gate error = %v, want ExitInvalid mentioning image-output", err)
	}
}

// repinInventory updates the inventory row (sha256, bytes) for a mutated
// durable input so hash verification passes and later gates are exercised.
func repinInventory(t *testing.T, root, p string, content []byte) {
	t.Helper()
	newSHA, err := HashFile(filepath.Join(root, filepath.FromSlash(p)))
	if err != nil {
		t.Fatal(err)
	}
	invPath := filepath.Join(root, "data", "research-inputs", "capture-inventory.tsv")
	inv, _ := os.ReadFile(invPath)
	lines := strings.Split(string(inv), "\n")
	for i, ln := range lines {
		c := strings.Split(ln, "\t")
		if len(c) == 10 && c[0] == p {
			c[3] = newSHA
			c[4] = fmt.Sprintf("%d", len(content))
			lines[i] = strings.Join(c, "\t")
		}
	}
	os.WriteFile(invPath, []byte(strings.Join(lines, "\n")), 0o644)
}
