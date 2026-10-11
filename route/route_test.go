package route

import (
	"slices"
	"testing"
	"time"
)

func f(v float64) *float64 { return &v }

var (
	free   = Candidate{ID: "local", Caps: []string{"tools"}, ContextTokens: 131072, InputUSDPerMTok: f(0), OutputUSDPerMTok: f(0)}
	cheap  = Candidate{ID: "cheap", Caps: []string{"tools"}, ContextTokens: 1_000_000, InputUSDPerMTok: f(0.15), OutputUSDPerMTok: f(0.5)}
	strong = Candidate{ID: "strong", Caps: []string{"tools"}, ContextTokens: 1_000_000, InputUSDPerMTok: f(2), OutputUSDPerMTok: f(10)}
	noTool = Candidate{ID: "notool", ContextTokens: 1_000_000, InputUSDPerMTok: f(0), OutputUSDPerMTok: f(0)}
)

func ids(r Result) []string {
	var out []string
	for _, c := range r.Ranked {
		out = append(out, c.ID)
	}
	return out
}

func TestCostModeOrdersByCostThenScore(t *testing.T) {
	r := Rank([]Candidate{strong, cheap, free}, Task{Name: "triage", Mode: Cost, InputTokens: 1000, OutputTokens: 100, AllowUnscored: true}, nil)
	if got := ids(r); !slices.Equal(got, []string{"local", "cheap", "strong"}) {
		t.Errorf("order = %v", got)
	}
}

func TestQualityModeOrdersByScoreFirst(t *testing.T) {
	s := Table{"strong": {"code": 0.9}, "cheap": {"code": 0.6}, "local": {"code": 0.4}}
	r := Rank([]Candidate{free, cheap, strong}, Task{Name: "code", Mode: Quality}, s)
	if got := ids(r); !slices.Equal(got, []string{"strong", "cheap", "local"}) {
		t.Errorf("order = %v", got)
	}
	if r.Ranked[0].Score == nil || r.Ranked[0].Score.Value != 0.9 {
		t.Errorf("score = %+v", r.Ranked[0].Score)
	}
}

func TestUnscoredRejectedUnlessAllowed(t *testing.T) {
	r := Rank([]Candidate{cheap}, Task{Name: "code", Mode: Quality}, Table{})
	if len(r.Ranked) != 0 || r.Rejected[0].Reasons[0] != "unscored" {
		t.Errorf("result = %+v", r)
	}
	r = Rank([]Candidate{cheap}, Task{Name: "code", Mode: Quality, AllowUnscored: true}, Table{})
	if len(r.Ranked) != 1 || !slices.Contains(r.Ranked[0].Warnings, "unscored") {
		t.Errorf("result = %+v", r)
	}
}

func TestScoredPrecedeUnscored(t *testing.T) {
	r := Rank([]Candidate{free, strong}, Task{Name: "code", Mode: Quality, AllowUnscored: true}, Table{"strong": {"code": 0.1}})
	if got := ids(r); !slices.Equal(got, []string{"strong", "local"}) {
		t.Errorf("order = %v", got)
	}
}

func TestMinScoreAppliesToScoredOnly(t *testing.T) {
	r := Rank([]Candidate{cheap, strong}, Task{Name: "code", Mode: Quality, MinScore: f(0.7)}, Table{"cheap": {"code": 0.6}, "strong": {"code": 0.9}})
	if got := ids(r); !slices.Equal(got, []string{"strong"}) || r.Rejected[0].Reasons[0] != "score_below_min" {
		t.Errorf("ranked %v rejected %+v", got, r.Rejected)
	}
}

func TestHardConstraints(t *testing.T) {
	small := Candidate{ID: "small", Caps: []string{"tools"}, ContextTokens: 1000}
	unknown := Candidate{ID: "unknown", Caps: []string{"tools"}}
	task := Task{Name: "x", Mode: Cost, Need: []string{"tools"}, InputTokens: 5000, OutputTokens: 500, AllowUnscored: true}
	r := Rank([]Candidate{noTool, small, unknown, cheap}, task, nil)
	if got := ids(r); !slices.Equal(got, []string{"cheap"}) {
		t.Fatalf("ranked %v", got)
	}
	want := map[string]string{"notool": "missing:tools", "small": "context_exceeded", "unknown": "context_unknown"}
	for _, rj := range r.Rejected {
		if rj.Reasons[0] != want[rj.ID] {
			t.Errorf("%s rejected for %v, want %s", rj.ID, rj.Reasons, want[rj.ID])
		}
	}
}

func TestOutputCapAndBudget(t *testing.T) {
	capped := Candidate{ID: "capped", ContextTokens: 1_000_000, MaxOutputTokens: 100, InputUSDPerMTok: f(1), OutputUSDPerMTok: f(1)}
	r := Rank([]Candidate{capped, strong}, Task{Name: "x", Mode: Cost, InputTokens: 1000, OutputTokens: 1000, MaxCostUSD: f(0.001), AllowUnscored: true}, nil)
	if len(r.Ranked) != 0 {
		t.Fatalf("ranked %v", ids(r))
	}
	got := map[string]string{}
	for _, rj := range r.Rejected {
		got[rj.ID] = rj.Reasons[0]
	}
	if got["capped"] != "output_exceeded" || got["strong"] != "over_budget" {
		t.Errorf("reasons = %v", got)
	}
}

func TestUnknownCostRanksAfterKnownAndWarns(t *testing.T) {
	unpriced := Candidate{ID: "unpriced", ContextTokens: 1_000_000}
	r := Rank([]Candidate{unpriced, strong}, Task{Name: "x", Mode: Cost, InputTokens: 10, OutputTokens: 10, AllowUnscored: true}, nil)
	if got := ids(r); !slices.Equal(got, []string{"strong", "unpriced"}) || !slices.Contains(r.Ranked[1].Warnings, "cost_unknown") {
		t.Errorf("ranked %v %+v", got, r.Ranked)
	}
	r = Rank([]Candidate{unpriced}, Task{Name: "x", Mode: Cost, InputTokens: 10, MaxCostUSD: f(1), AllowUnscored: true}, nil)
	if len(r.Ranked) != 0 || r.Rejected[0].Reasons[0] != "cost_unknown" {
		t.Errorf("a budget must reject unknown cost: %+v", r)
	}
}

func TestLatencyBound(t *testing.T) {
	slow := Candidate{ID: "slow", Latency: 5 * time.Second}
	fast := Candidate{ID: "fast", Latency: time.Second}
	r := Rank([]Candidate{slow, fast}, Task{Name: "x", Mode: Cost, MaxLatency: 2 * time.Second, AllowUnscored: true}, nil)
	if got := ids(r); !slices.Equal(got, []string{"fast"}) {
		t.Errorf("ranked %v", got)
	}
}

func TestDeterministicTieBreakByID(t *testing.T) {
	a, b := Candidate{ID: "a"}, Candidate{ID: "b"}
	r1 := Rank([]Candidate{b, a}, Task{Name: "x", Mode: Cost, AllowUnscored: true}, nil)
	r2 := Rank([]Candidate{a, b}, Task{Name: "x", Mode: Cost, AllowUnscored: true}, nil)
	if !slices.Equal(ids(r1), ids(r2)) || ids(r1)[0] != "a" {
		t.Errorf("%v vs %v", ids(r1), ids(r2))
	}
}

func TestInputNotMutated(t *testing.T) {
	in := []Candidate{strong, free}
	Rank(in, Task{Name: "x", Mode: Cost, AllowUnscored: true}, nil)
	if in[0].ID != "strong" {
		t.Error("Rank reordered the caller's slice")
	}
}
