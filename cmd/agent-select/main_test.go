package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRankExitCodes(t *testing.T) {
	cases := []struct {
		name, in string
		want     int
		contains string
	}{
		{"selection", `{"task":{"name":"code","mode":"quality"},"candidates":[{"id":"a"},{"id":"b"}],"scores":{"b":{"code":0.9},"a":{"code":0.2}}}`, 0, `"id": "b"`},
		{"none admissible", `{"task":{"name":"code","mode":"quality"},"candidates":[{"id":"a"}]}`, 3, `"unscored"`},
		{"unknown field", `{"nope":1}`, 2, ""},
	}
	for _, c := range cases {
		var out, errw bytes.Buffer
		if got := run([]string{"rank"}, strings.NewReader(c.in), &out, &errw); got != c.want {
			t.Errorf("%s: exit %d, want %d (%s)", c.name, got, c.want, errw.String())
		}
		if !strings.Contains(out.String(), c.contains) {
			t.Errorf("%s: stdout %q lacks %q", c.name, out.String(), c.contains)
		}
	}
	if run(nil, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}) != 2 {
		t.Error("no subcommand must be a usage error")
	}
}

func TestRecordThenRankWithHistory(t *testing.T) {
	log := filepath.Join(t.TempDir(), "outcomes.jsonl")
	rec := func(id, verdict string) int {
		return run([]string{"record", "--history", log, "--task", "code", "--id", id, "--verdict", verdict}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	}
	for range 5 {
		if got := rec("a", "pass"); got != 0 {
			t.Fatalf("record exit %d", got)
		}
		rec("b", "fail")
	}
	if rec("a", "maybe") != 2 {
		t.Error("an unknown verdict must be a usage error")
	}
	// The table prefers b; five recorded runs each override it.
	req := `{"task":{"name":"code","mode":"quality"},"candidates":[{"id":"a"},{"id":"b"}],"scores":{"a":{"code":0.1},"b":{"code":0.9}}}`
	var out, errw bytes.Buffer
	if got := run([]string{"rank", "--history", log}, strings.NewReader(req), &out, &errw); got != 0 {
		t.Fatalf("rank exit %d: %s", got, errw.String())
	}
	if first := strings.Index(out.String(), `"id": "a"`); first < 0 || first > strings.Index(out.String(), `"id": "b"`) {
		t.Errorf("history should rank a first:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "history n=5") {
		t.Errorf("basis should name the history:\n%s", out.String())
	}
	out.Reset()
	if got := run([]string{"rank", "--history", filepath.Join(t.TempDir(), "absent.jsonl")}, strings.NewReader(req), &out, &errw); got != 0 || !strings.Contains(out.String(), `"Basis": "table"`) {
		t.Errorf("a missing log is no history, not an error: exit %d\n%s", got, out.String())
	}
}

// jevServer answers every request with answers and points the CLI at it.
func jevServer(t *testing.T, answers string) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"model":"jev-1.13.0","answers":` + answers + `}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TYPESAFE_URL", srv.URL)
	t.Setenv("TYPESAFE_API_KEY", "k")
}

func TestClassify(t *testing.T) {
	jevServer(t, `{"task":{"type":"choice","choice":"code","confidence":0.9},"need:tools":{"type":"noul","noul":0.9},
		"output":{"type":"score","probabilities":{"0":1},"confidence":1}}`)
	tx := filepath.Join(t.TempDir(), "tx.json")
	os.WriteFile(tx, []byte(`{"tasks":{"code":"write code"},"caps":{"tools":"call tools"}}`), 0o644)
	var out, errw bytes.Buffer
	if got := run([]string{"classify", "--taxonomy", tx}, strings.NewReader("fix the test"), &out, &errw); got != 0 {
		t.Fatalf("exit %d: %s", got, errw.String())
	}
	for _, want := range []string{`"name": "code"`, `"need": [`, `"tools"`, `"output_tokens": 500`, `"input_tokens": 3`, `"model": "jev-1.13.0"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("stdout lacks %s:\n%s", want, out.String())
		}
	}
	t.Setenv("TYPESAFE_API_KEY", "")
	if got := run([]string{"classify", "--taxonomy", tx}, strings.NewReader("x"), &out, &errw); got != 2 {
		t.Errorf("no key: exit %d, want 2", got)
	}
}

func TestJudgeRecords(t *testing.T) {
	log := filepath.Join(t.TempDir(), "outcomes.jsonl")
	judge := func(p string) int {
		jevServer(t, `{"done":{"type":"noul","noul":`+p+`}}`)
		var out, errw bytes.Buffer
		return run([]string{"judge", "--history", log, "--task", "code", "--id", "a"},
			strings.NewReader(`{"request":"add a test","result":"added TestX"}`), &out, &errw)
	}
	if got := judge("0.95"); got != 0 {
		t.Fatalf("decided exit %d", got)
	}
	if got := judge("0.5"); got != 3 {
		t.Fatalf("undecided exit %d, want 3", got)
	}
	b, _ := os.ReadFile(log)
	if want := `{"task":"code","id":"a","pass":true,"judge":"jev-1.13.0"}` + "\n"; string(b) != want {
		t.Errorf("log %q, want only the decided verdict %q", b, want)
	}
}
