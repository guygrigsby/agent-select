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
	if !strings.Contains(summary, "29 tasks") {
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
		"950 unresolved measurements",
		"no image-specific limit facts",
		"host-registry-discovery",
	} {
		if !strings.Contains(string(report), want) {
			t.Fatalf("report missing %q", want)
		}
	}
	// Task status counts come from the durable task file, never a pinned
	// total here: the report must render exactly what the tasks carry
	// (§10.2 explicit status contract; the drift gate catches divergence).
	_, inv, _, _, tables, err := loadInputs(root)
	if err != nil {
		t.Fatalf("loadInputs: %v", err)
	}
	tasks, err := LoadTasks(root, inv, researchBatchNames(tables))
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	sc := map[string]int{}
	for _, tk := range tasks {
		sc[tk.Status]++
	}
	wantTasks := fmt.Sprintf("%d tasks: %s %d, %s %d, %s %d, %s %d",
		len(tasks), TaskStatusOpen, sc[TaskStatusOpen], TaskStatusInProgress, sc[TaskStatusInProgress],
		TaskStatusBlocked, sc[TaskStatusBlocked], TaskStatusDone, sc[TaskStatusDone])
	if !strings.Contains(string(report), wantTasks) {
		t.Fatalf("report missing derived task counts %q", wantTasks)
	}
	// Status and basis appear per task; LiveCodeBench must render blocked.
	for _, ln := range strings.Split(string(report), "\n") {
		if strings.HasPrefix(ln, "| livecodebench |") {
			if !strings.Contains(ln, "| blocked |") {
				t.Fatalf("livecodebench row must render status blocked: %s", ln)
			}
			if !strings.Contains(ln, "policies/livecodebench.json") {
				t.Fatalf("livecodebench status_basis must cite the policy evidence: %s", ln)
			}
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

// TestBacklogTaskStatus drives the §10.2 status contract: every task must
// carry an explicit status from the closed enum with a trimmed nonempty
// status_basis citing repo evidence; absent, null and unknown statuses and
// empty or whitespace-only bases are all rejected ExitInvalid.
func TestBacklogTaskStatus(t *testing.T) {
	root := copyCorpus(t)
	_, inv, _, _, tables, err := loadInputs(root)
	if err != nil {
		t.Fatalf("loadInputs: %v", err)
	}
	batches := researchBatchNames(tables)
	p := filepath.Join(root, "data", "research-inputs", "backlog-tasks.json")
	orig, _ := os.ReadFile(p)
	defer os.WriteFile(p, orig, 0o644)
	restore := func(t *testing.T) {
		t.Helper()
		if err := os.WriteFile(p, orig, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("explicit statuses preserved", func(t *testing.T) {
		tasks, err := LoadTasks(root, inv, batches)
		if err != nil {
			t.Fatalf("LoadTasks: %v", err)
		}
		byID := map[string]BacklogTask{}
		for _, tk := range tasks {
			byID[tk.ID] = tk
		}
		if tk := byID["geneval"]; tk.Status != TaskStatusDone || tk.StatusBasis == "" {
			t.Fatalf("geneval status = %q basis = %q, want done with a basis", tk.Status, tk.StatusBasis)
		}
		if tk := byID["livecodebench"]; tk.Status != TaskStatusBlocked || tk.StatusBasis == "" {
			t.Fatalf("livecodebench status = %q basis = %q, want blocked with a basis", tk.Status, tk.StatusBasis)
		}
		if tk := byID["arena"]; tk.Status != TaskStatusBlocked || !strings.Contains(tk.StatusBasis, "policies/arena.json") {
			t.Fatalf("arena status = %q basis = %q, want blocked citing the policy", tk.Status, tk.StatusBasis)
		}
	})
	t.Run("missing status rejected", func(t *testing.T) {
		restore(t)
		editTasks(t, p, "geneval", func(tk map[string]any) { delete(tk, "status") })
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid {
			t.Fatalf("missing status must fail ExitInvalid under research-artifact.md §10.2, got %v", err)
		}
	})
	t.Run("null status rejected", func(t *testing.T) {
		restore(t)
		editTasks(t, p, "geneval", func(tk map[string]any) { tk["status"] = nil })
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid {
			t.Fatalf("null status must fail ExitInvalid, got %v", err)
		}
	})
	t.Run("invalid status rejected", func(t *testing.T) {
		restore(t)
		editTasks(t, p, "geneval", func(tk map[string]any) { tk["status"] = "closed" })
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "open|in_progress|blocked|done") {
			t.Fatalf("error = %v, want ExitInvalid naming the status enum", err)
		}
	})
	t.Run("empty status_basis rejected", func(t *testing.T) {
		restore(t)
		editTasks(t, p, "geneval", func(tk map[string]any) {
			tk["status"] = "done"
			tk["status_basis"] = ""
		})
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "status_basis") {
			t.Fatalf("error = %v, want ExitInvalid naming status_basis", err)
		}
	})
	t.Run("whitespace-only status_basis rejected", func(t *testing.T) {
		restore(t)
		editTasks(t, p, "geneval", func(tk map[string]any) {
			tk["status"] = "done"
			tk["status_basis"] = " \t\n "
		})
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "status_basis") {
			t.Fatalf("error = %v, want ExitInvalid naming status_basis", err)
		}
	})
}

func TestBacklogMissingStatusRejected(t *testing.T) {
	root := copyCorpus(t)
	_, inv, _, _, tables, err := loadInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	editTasks(t, filepath.Join(root, "data", "research-inputs", "backlog-tasks.json"), "geneval", func(task map[string]any) {
		delete(task, "status")
	})
	_, err = LoadTasks(root, inv, researchBatchNames(tables))
	if e, ok := err.(*Error); !ok || e.Code != ExitInvalid {
		t.Fatalf("missing status must fail ExitInvalid under research-artifact.md §10.2, got %v", err)
	}
}

// editTasks rewrites one task object in backlog-tasks.json (test copies
// only) so status-contract mutations do not depend on key order.
func editTasks(t *testing.T, p, id string, fn func(tk map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, tk := range doc.Tasks {
		if tk["id"] == id {
			fn(tk)
		}
	}
	out, err := json.Marshal(&doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBacklogStatusDrift pins the drift-guard half of the status contract:
// the generated report must expose status and basis so a status change in
// the durable input makes the committed report stale under --check.
func TestBacklogStatusDrift(t *testing.T) {
	root := copyCorpus(t)
	_, report, _, err := Backlog(root)
	if err != nil {
		t.Fatalf("Backlog: %v", err)
	}
	p := filepath.Join(root, "docs", "research", "backlog.md")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, report, 0o644)
	if err := BacklogCheck(root, "docs/research/backlog.md"); err != nil {
		t.Fatalf("fresh report must pass --check: %v", err)
	}
	editTasks(t, filepath.Join(root, "data", "research-inputs", "backlog-tasks.json"), "aider", func(tk map[string]any) {
		tk["status"] = "in_progress"
		tk["status_basis"] = "probe basis"
	})
	err = BacklogCheck(root, "docs/research/backlog.md")
	if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("status change must make the committed report stale, got %v", err)
	}
}

// TestBacklogTaskValidation drives LoadTasks rejections: uninventoried
// evidence file, and a research batch left without a claiming task.
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

	t.Run("uninventoried evidence", func(t *testing.T) {
		os.WriteFile(p, []byte(strings.Replace(string(orig), `"data/research/anthropic/findings.json"`, `"data/research/nope/findings.json"`, 1)), 0o644)
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid {
			t.Fatalf("error = %v, want ExitInvalid", err)
		}
	})
	t.Run("unclaimed research batch", func(t *testing.T) {
		// Point the mixedbread task at anthropic so mixedbread (a batch with
		// model_research rows) has no claiming task; kind rules still hold.
		os.WriteFile(p, []byte(strings.Replace(string(orig), `"batches": ["mixedbread"]`, `"batches": ["anthropic"]`, 1)), 0o644)
		_, err := LoadTasks(root, inv, batches)
		if ce, ok := err.(*Error); err == nil || !ok || ce.Code != ExitInvalid || !strings.Contains(err.Error(), "mixedbread") {
			t.Fatalf("error = %v, want ExitInvalid naming mixedbread", err)
		}
	})
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
