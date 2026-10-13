package main

import (
	"bytes"
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
