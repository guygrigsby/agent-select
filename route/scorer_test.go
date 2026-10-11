package route

import (
	"strings"
	"testing"
)

func TestTableWildcards(t *testing.T) {
	tb := Table{"m": {"code": 0.8, "*": 0.5}, "*": {"code": 0.3}}
	for _, c := range []struct {
		id, task string
		want     float64
		ok       bool
	}{
		{"m", "code", 0.8, true},
		{"m", "triage", 0.5, true},
		{"other", "code", 0.3, true},
		{"other", "triage", 0, false},
	} {
		s, ok := tb.Score(c.task, Candidate{ID: c.id})
		if ok != c.ok || (ok && s.Value != c.want) {
			t.Errorf("%s/%s = %+v %v", c.id, c.task, s, ok)
		}
	}
}

func TestChainFirstHitWins(t *testing.T) {
	h := NewHistory(2)
	tb := Table{"m": {"code": 0.5}}
	chain := Chain{h, tb}
	if s, _ := chain.Score("code", Candidate{ID: "m"}); s.Basis != "table" {
		t.Errorf("below the sample floor the table answers: %+v", s)
	}
	h.Record("code", "m", true)
	h.Record("code", "m", true)
	if s, _ := chain.Score("code", Candidate{ID: "m"}); !strings.HasPrefix(s.Basis, "history") {
		t.Errorf("history with enough samples answers first: %+v", s)
	}
}

func TestHistorySmoothedRateAndPersistence(t *testing.T) {
	h := NewHistory(3)
	for range 8 {
		h.Record("code", "m", true)
	}
	h.Record("code", "m", false)
	h.Record("code", "m", false)
	s, ok := h.Score("code", Candidate{ID: "m"})
	if !ok || s.Value != 9.0/12.0 { // (8+1)/(10+2)
		t.Fatalf("score = %+v %v", s, ok)
	}
	b, err := h.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	h2 := NewHistory(3)
	if err := h2.UnmarshalJSON(b); err != nil {
		t.Fatal(err)
	}
	if s2, _ := h2.Score("code", Candidate{ID: "m"}); s2.Value != s.Value {
		t.Errorf("round trip = %+v", s2)
	}
}
