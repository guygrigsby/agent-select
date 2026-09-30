package research

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repoRoot locates the real repository root for corpus tests.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data", "research-inputs", "build.json")); err != nil {
		t.Skipf("real corpus not available: %v", err)
	}
	return dir
}

// copyTree copies src into dst (deep).
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, info.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

// copyCorpus copies the real corpus inputs (everything except runtime state
// and published outputs) into a temp root: a first-generation scenario.
func copyCorpus(t *testing.T) string {
	t.Helper()
	src := repoRoot(t)
	dst := t.TempDir()
	copyTree(t, filepath.Join(src, "data"), filepath.Join(dst, "data"))
	for _, p := range []string{
		filepath.Join(dst, "data", "research.sqlite"),
		filepath.Join(dst, "data", "source-manifest.tsv"),
		filepath.Join(dst, "data", "source-hashes.sha256"),
	} {
		os.Remove(p)
	}
	os.RemoveAll(filepath.Join(dst, "data", ".research-backup"))
	os.Remove(filepath.Join(dst, "data", ".research-rebuild.journal"))
	os.Remove(filepath.Join(dst, "data", ".research-rebuild.lock"))
	// Test binaries don't expose linked dependency versions via build info,
	// so sandboxed runs use placeholder driver pins (placeholders never
	// mismatch; pins are exercised by the real CLI in make check).
	bp := filepath.Join(dst, "data", "research-inputs", "build.json")
	raw, err := os.ReadFile(bp)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if d, ok := cfg["driver"].(map[string]any); ok {
		d["sqlite_driver_module"] = "modernc.org/sqlite"
		d["sqlite_driver_version"] = "placeholder"
		d["go_toolchain_minor"] = "placeholder"
	}
	out, _ := json.Marshal(cfg)
	if err := os.WriteFile(bp, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// TestDeterminismTwoBuilds proves §4: two builds from the same inputs are
// byte-identical, and — by rebuilding after touching every input file's
// mtime — that output bytes never depend on mtimes (the recovered
// exporter's defect).
func TestDeterminismTwoBuilds(t *testing.T) {
	root := copyCorpus(t)
	p1 := filepath.Join(t.TempDir(), "a.sqlite")
	p2 := filepath.Join(t.TempDir(), "b.sqlite")
	_, _, _, _, tables1, err := loadInputs(root)
	if err != nil {
		t.Fatalf("load 1: %v", err)
	}
	if err := writeDB(p1, tables1); err != nil {
		t.Fatalf("build 1: %v", err)
	}
	var touched int
	filepath.Walk(filepath.Join(root, "data"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.Contains(p, ".research") {
			future := info.ModTime().Add(37 * time.Minute)
			os.Chtimes(p, future, future)
			touched++
		}
		return nil
	})
	if touched == 0 {
		t.Fatal("no inputs touched")
	}
	_, _, _, _, tables2, err := loadInputs(root)
	if err != nil {
		t.Fatalf("load 2: %v", err)
	}
	if err := writeDB(p2, tables2); err != nil {
		t.Fatalf("build 2: %v", err)
	}
	if h1, h2 := mustHash(p1), mustHash(p2); h1 != h2 {
		t.Fatalf("determinism: DB bytes differ after mtime touch: %s vs %s", h1, h2)
	}
}

// TestRebuildFirstGenerationNoopAndCheck runs the full protocol against a
// sandboxed corpus copy: first-generation publish, byte-identical no-op
// rebuild, then check.
func TestRebuildFirstGenerationNoopAndCheck(t *testing.T) {
	root := copyCorpus(t)
	r1, err := Rebuild(root)
	if err != nil {
		t.Fatalf("first rebuild: %v", err)
	}
	if r1.Noop {
		t.Fatal("first rebuild on empty outputs reported no-op")
	}
	for name, h := range r1.OutputHashes {
		got, err := HashFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("hash %s: %v", name, err)
		}
		if got != h {
			t.Fatalf("%s hash mismatch after publish", name)
		}
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(JournalPath))); !os.IsNotExist(err) {
		t.Fatal("journal not cleared after successful publication")
	}
	r2, err := Rebuild(root)
	if err != nil {
		t.Fatalf("second rebuild: %v", err)
	}
	if !r2.Noop {
		t.Fatal("identical rebuild was not detected as no-op")
	}
	for name, h := range r1.OutputHashes {
		if r2.OutputHashes[name] != h {
			t.Fatalf("%s hash changed on no-op rebuild", name)
		}
	}
	if _, err := Check(root); err != nil {
		t.Fatalf("check after rebuild: %v", err)
	}

	// coverage parity with the SQL coverage view
	_, _, _, res, tables, err := loadInputs(root)
	if err != nil {
		t.Fatalf("loadInputs: %v", err)
	}
	rows := ComputeCoverage(res, tables)
	db, err := sql.Open("sqlite", filepath.Join(root, filepath.FromSlash(ArtifactDB)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	qrows, err := db.Query(`SELECT id, kind, CASE WHEN EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=m.id) THEN 'scores_found' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id AND r.benchmark_status='not_found_in_checked_sources') THEN 'no_scores_in_checked_sources' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id) THEN 'unresearched' ELSE 'out_of_research_scope' END FROM catalog_models m ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer qrows.Close()
	i := 0
	for qrows.Next() {
		var id, kind, cov string
		if err := qrows.Scan(&id, &kind, &cov); err != nil {
			t.Fatal(err)
		}
		if i >= len(rows) {
			t.Fatal("coverage TSV has fewer rows than the view")
		}
		r := rows[i]
		if r.ModelID != id || r.Kind != kind || r.BenchmarkCoverage != cov {
			t.Fatalf("coverage divergence at %s: TSV(%s,%s,%s) vs view(%s,%s,%s)", id, r.ModelID, r.Kind, r.BenchmarkCoverage, id, kind, cov)
		}
		i++
	}
	if i != len(rows) {
		t.Fatalf("coverage row count: TSV %d vs view %d", len(rows), i)
	}
}

// decisionForMutation finds a current superseding decision file.
func decisionForMutation(t *testing.T, root string) string {
	t.Helper()
	decDir := filepath.Join(root, "data", "reviews", "decisions")
	entries, err := os.ReadDir(decDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".v3.json") || strings.HasSuffix(e.Name(), ".v2.json") {
			return filepath.Join(decDir, e.Name())
		}
	}
	t.Fatal("no superseding decision found")
	return ""
}

func firstPinPath(t *testing.T, decisionRaw string) string {
	t.Helper()
	i := strings.Index(decisionRaw, `"path": "`)
	if i < 0 {
		t.Fatal("no pin path found in decision")
	}
	rest := decisionRaw[i+len(`"path": "`):]
	return rest[:strings.Index(rest, `"`)]
}

func reviewPathOf(t *testing.T, decisionRaw string) string {
	t.Helper()
	i := strings.Index(decisionRaw, `"review_path": "`)
	if i < 0 {
		t.Fatal("no review_path in decision")
	}
	rest := decisionRaw[i+len(`"review_path": "`):]
	return rest[:strings.Index(rest, `"`)]
}

// TestMutationMatrix tampers each durable input class on a fresh corpus copy
// and asserts the exit code per §1 with no output replacement.
func TestMutationMatrix(t *testing.T) {
	cases := []struct {
		name string
		code int
		mut  func(t *testing.T, root string)
	}{
		{"tamper reviewed evidence bytes", ExitInvalid, func(t *testing.T, root string) {
			dec := decisionForMutation(t, root)
			raw, _ := os.ReadFile(dec)
			p := firstPinPath(t, string(raw))
			target := filepath.Join(root, filepath.FromSlash(p))
			cur, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			os.WriteFile(target, append(cur, []byte("\n// tampered")...), 0o644)
		}},
		{"stale decision pin after evidence change", ExitStale, func(t *testing.T, root string) {
			// the real stale flow: evidence bytes change, the inventory is
			// updated to match, but no superseding decision exists yet
			dec := decisionForMutation(t, root)
			raw, _ := os.ReadFile(dec)
			p := firstPinPath(t, string(raw))
			target := filepath.Join(root, filepath.FromSlash(p))
			cur, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, append(cur, []byte("\n// superseded evidence revision")...), 0o644); err != nil {
				t.Fatal(err)
			}
			newSHA, err := HashFile(target)
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
					c[4] = fmt.Sprintf("%d", len(cur)+len("\n// superseded evidence revision"))
					lines[i] = strings.Join(c, "\t")
				}
			}
			os.WriteFile(invPath, []byte(strings.Join(lines, "\n")), 0o644)
		}},
		{"drop review_path from inventory", ExitInvalid, func(t *testing.T, root string) {
			dec := decisionForMutation(t, root)
			raw, _ := os.ReadFile(dec)
			rp := reviewPathOf(t, string(raw))
			invPath := filepath.Join(root, "data", "research-inputs", "capture-inventory.tsv")
			inv, _ := os.ReadFile(invPath)
			var out []string
			for _, ln := range strings.Split(string(inv), "\n") {
				if strings.HasPrefix(ln, rp+"\t") {
					continue
				}
				out = append(out, ln)
			}
			os.WriteFile(invPath, []byte(strings.Join(out, "\n")), 0o644)
		}},
		{"wrong schema_version in build.json", ExitInconsist, func(t *testing.T, root string) {
			bp := filepath.Join(root, "data", "research-inputs", "build.json")
			raw, _ := os.ReadFile(bp)
			var cfg map[string]any
			if err := json.Unmarshal(raw, &cfg); err != nil {
				t.Fatal(err)
			}
			cfg["schema_version"] = 99
			out, _ := json.Marshal(cfg)
			os.WriteFile(bp, out, 0o644)
		}},
		{"unmapped batch directory", ExitInvalid, func(t *testing.T, root string) {
			dir := filepath.Join(root, "data", "research", "rogue-batch")
			os.MkdirAll(dir, 0o755)
			os.WriteFile(filepath.Join(dir, "findings.json"), []byte(`{}`), 0o644)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := copyCorpus(t)
			r, err := Rebuild(root)
			if err != nil {
				t.Fatalf("baseline rebuild: %v", err)
			}
			baseline := map[string]string{}
			for name := range r.OutputHashes {
				h, _ := HashFile(filepath.Join(root, filepath.FromSlash(name)))
				baseline[name] = h
			}
			tc.mut(t, root)
			_, err = Rebuild(root)
			ce, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *Error, got %v", err)
			}
			if ce.Code != tc.code {
				t.Fatalf("exit code %d, want %d (%v)", ce.Code, tc.code, err)
			}
			for name, h := range baseline {
				got, _ := HashFile(filepath.Join(root, filepath.FromSlash(name)))
				if got != h {
					t.Fatalf("%s replaced despite failure", name)
				}
			}
		})
	}
}

// TestFindingsMustBeInventoried is the regression for the review's gate-5
// fail-open: a batch findings.json missing from the capture inventory must
// exit 1, not silently drop the file from the hash manifest.
func TestFindingsMustBeInventoried(t *testing.T) {
	root := copyCorpus(t)
	if _, err := Rebuild(root); err != nil {
		t.Fatalf("baseline rebuild: %v", err)
	}
	invPath := filepath.Join(root, "data", "research-inputs", "capture-inventory.tsv")
	inv, _ := os.ReadFile(invPath)
	var out []string
	for _, ln := range strings.Split(string(inv), "\n") {
		if strings.HasPrefix(ln, "data/research/images/findings.json\t") {
			continue
		}
		out = append(out, ln)
	}
	os.WriteFile(invPath, []byte(strings.Join(out, "\n")), 0o644)
	_, err := Rebuild(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInvalid {
		t.Fatalf("uninventoried findings read must exit 1, got %v", err)
	}
	if !strings.Contains(ce.Msg, "gate 5") {
		t.Fatalf("diagnostic should name gate 5: %v", err)
	}
	// check must fail the same way, read-only
	_, err = Check(root)
	if ce, ok := err.(*Error); !ok || ce.Code != ExitInvalid {
		t.Fatalf("check must also exit 1, got %v", err)
	}
}

// TestMeasurementWithoutCurrentDecision exercises the §2.3.1 hard gate:
// removing the measurement-scope decision for a measurement-bearing batch
// must exit 3 with the batch named.
func TestMeasurementWithoutCurrentDecision(t *testing.T) {
	root := copyCorpus(t)
	if _, err := Rebuild(root); err != nil {
		t.Fatalf("baseline rebuild: %v", err)
	}
	// google carries benchmark measurements and a published_measurements
	// decision chain v1..v3; delete every google record
	decDir := filepath.Join(root, "data", "reviews", "decisions")
	entries, _ := os.ReadDir(decDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "google.") {
			os.Remove(filepath.Join(decDir, e.Name()))
		}
	}
	_, err := Rebuild(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("measurements without a current decision must exit 3, got %v", err)
	}
	if !strings.Contains(ce.Msg, "google") {
		t.Fatalf("diagnostic should name the batch: %v", err)
	}
}

// TestControlFilesRejectsUnexpectedNames pins the symmetric control-dir
// rule: a fetch log or decision with a non-conforming name is exit 1, and
// subdirectories are rejected in both control areas.
func TestControlFilesRejectsUnexpectedNames(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "data", "research-inputs", "fetch-logs"), 0o755)
	os.MkdirAll(filepath.Join(root, "data", "reviews", "decisions"), 0o755)
	os.WriteFile(filepath.Join(root, "data", "research-inputs", "capture-inventory.tsv"), []byte("path\torigin\n"), 0o644)
	os.WriteFile(filepath.Join(root, "data", "research-inputs", "build.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(root, "data", "research-inputs", "fetch-logs", "bad name.log"), []byte("x"), 0o644)
	if _, err := ControlFiles(root); err == nil {
		t.Fatal("fetch log with a space in the name must be rejected")
	}
	os.Remove(filepath.Join(root, "data", "research-inputs", "fetch-logs", "bad name.log"))
	os.WriteFile(filepath.Join(root, "data", "research-inputs", "fetch-logs", "ok.log"), []byte("x"), 0o644)
	if _, err := ControlFiles(root); err != nil {
		t.Fatalf("conforming fetch log must pass: %v", err)
	}
	os.MkdirAll(filepath.Join(root, "data", "reviews", "decisions", "sub"), 0o755)
	if _, err := ControlFiles(root); err == nil {
		t.Fatal("subdirectory under decisions must be rejected")
	}
}
