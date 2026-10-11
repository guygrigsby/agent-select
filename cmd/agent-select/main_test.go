package main

import (
	"bytes"
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
