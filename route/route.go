// Package route ranks candidate models for a task. The host supplies what
// it can run right now (Candidate), what the task needs (Task) and,
// optionally, where quality scores come from (Scorer). Rank is pure: no
// network, clock or state, so a result replays from its inputs.
//
// Cost is an estimate for ordering and budget gates, computed in float64 USD
// from per-million-token rates. It is not a billing figure.
package route

import (
	"cmp"
	"slices"
	"time"
)

// Candidate is one thing the host can run right now. Zero values and nil
// rates mean unknown, never free or unlimited: a free local model says so
// with explicit zero rates.
type Candidate struct {
	// ID is the host's key and what a Scorer is asked about.
	ID string
	// Caps are free-form capability tags the host asserts ("tools",
	// "vision"). Task.Need is matched against them exactly.
	Caps            []string
	ContextTokens   int64 // combined input and output window; 0 unknown
	MaxOutputTokens int64 // 0 unknown
	// Rates are USD per million tokens; nil unknown.
	InputUSDPerMTok  *float64
	OutputUSDPerMTok *float64
	Latency          time.Duration // typical; 0 unknown
}

// Mode picks the primary ordering.
type Mode string

const (
	// Cost orders by estimated cost first, then score.
	Cost Mode = "cost"
	// Quality orders by score first, then cost.
	Quality Mode = "quality"
)

// Task is what the caller needs done.
type Task struct {
	// Name is handed to the Scorer ("code", "triage"); the router reads
	// nothing into it.
	Name string
	Mode Mode
	// Need lists capability tags a candidate must carry.
	Need                      []string
	InputTokens, OutputTokens int64
	// MaxCostUSD and MaxLatency are hard bounds. A candidate whose cost or
	// latency is unknown fails a set bound.
	MaxCostUSD *float64
	MaxLatency time.Duration
	// MinScore rejects scored candidates below it.
	MinScore *float64
	// AllowUnscored admits candidates the Scorer has no opinion on; they
	// rank after every scored one and carry an "unscored" warning.
	AllowUnscored bool
}

// Ranked is an admitted candidate with the numbers it was ordered on.
type Ranked struct {
	Candidate
	Score    *Score   // nil when unscored
	CostUSD  *float64 // nil when unknown
	Warnings []string
}

// Rejection says why a candidate was excluded. Reasons are short codes.
type Rejection struct {
	ID      string
	Reasons []string
}

// Result is the full ordering and everything excluded.
type Result struct {
	Ranked   []Ranked
	Rejected []Rejection
}

// Best returns the first ranked candidate.
func (r Result) Best() (Ranked, bool) {
	if len(r.Ranked) == 0 {
		return Ranked{}, false
	}
	return r.Ranked[0], true
}

// Rank filters candidates by the task's hard constraints and orders the
// rest. s may be nil (everything is unscored). It does not modify cands.
func Rank(cands []Candidate, t Task, s Scorer) Result {
	var res Result
	for _, c := range cands {
		r := Ranked{Candidate: c}
		if s != nil {
			if sc, ok := s.Score(t.Name, c); ok {
				r.Score = &sc
			}
		}
		r.CostUSD = estimate(c, t)
		if reasons := reject(c, t, r); len(reasons) > 0 {
			res.Rejected = append(res.Rejected, Rejection{ID: c.ID, Reasons: reasons})
			continue
		}
		if r.Score == nil {
			r.Warnings = append(r.Warnings, "unscored")
		}
		if r.CostUSD == nil {
			r.Warnings = append(r.Warnings, "cost_unknown")
		}
		res.Ranked = append(res.Ranked, r)
	}
	slices.SortFunc(res.Ranked, func(a, b Ranked) int { return compare(t.Mode, a, b) })
	slices.SortFunc(res.Rejected, func(a, b Rejection) int { return cmp.Compare(a.ID, b.ID) })
	return res
}

func estimate(c Candidate, t Task) *float64 {
	if c.InputUSDPerMTok == nil || c.OutputUSDPerMTok == nil {
		return nil
	}
	x := (float64(t.InputTokens)**c.InputUSDPerMTok + float64(t.OutputTokens)**c.OutputUSDPerMTok) / 1e6
	return &x
}

func reject(c Candidate, t Task, r Ranked) []string {
	var out []string
	for _, need := range t.Need {
		if !slices.Contains(c.Caps, need) {
			out = append(out, "missing:"+need)
		}
	}
	if total := t.InputTokens + t.OutputTokens; total > 0 {
		switch {
		case c.ContextTokens == 0:
			out = append(out, "context_unknown")
		case total > c.ContextTokens:
			out = append(out, "context_exceeded")
		}
	}
	if t.OutputTokens > 0 && c.MaxOutputTokens > 0 && t.OutputTokens > c.MaxOutputTokens {
		out = append(out, "output_exceeded")
	}
	if t.MaxCostUSD != nil {
		switch {
		case r.CostUSD == nil:
			out = append(out, "cost_unknown")
		case *r.CostUSD > *t.MaxCostUSD:
			out = append(out, "over_budget")
		}
	}
	if t.MaxLatency > 0 {
		switch {
		case c.Latency == 0:
			out = append(out, "latency_unknown")
		case c.Latency > t.MaxLatency:
			out = append(out, "too_slow")
		}
	}
	switch {
	case r.Score == nil && !t.AllowUnscored:
		out = append(out, "unscored")
	case r.Score != nil && t.MinScore != nil && r.Score.Value < *t.MinScore:
		out = append(out, "score_below_min")
	}
	return out
}

func compare(m Mode, a, b Ranked) int {
	byScore := func() int { return cmpScore(a.Score, b.Score) }
	byCost := func() int { return cmpPtr(a.CostUSD, b.CostUSD, false) }
	first, second := byScore, byCost
	if m == Cost {
		first, second = byCost, byScore
	}
	if c := first(); c != 0 {
		return c
	}
	if c := second(); c != 0 {
		return c
	}
	if c := cmpDur(a.Latency, b.Latency); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}

// cmpScore orders scored before unscored, higher first.
func cmpScore(a, b *Score) int {
	var av, bv *float64
	if a != nil {
		av = &a.Value
	}
	if b != nil {
		bv = &b.Value
	}
	return cmpPtr(av, bv, true)
}

// cmpPtr orders known before unknown, then ascending (descending when desc).
func cmpPtr(a, b *float64, desc bool) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	}
	c := cmp.Compare(*a, *b)
	if desc {
		c = -c
	}
	return c
}

// cmpDur orders known latency before unknown, then ascending.
func cmpDur(a, b time.Duration) int {
	switch {
	case a == 0 && b == 0:
		return 0
	case a == 0:
		return 1
	case b == 0:
		return -1
	}
	return cmp.Compare(a, b)
}
