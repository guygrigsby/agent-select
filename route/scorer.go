package route

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Score is a scorer's opinion of one candidate for one task.
type Score struct {
	// Value is 0..1, higher is better. Scorers need not agree on a scale
	// across sources, so a Chain mixes them only by first-hit.
	Value float64
	// Basis says where it came from: "table", "history n=12", "benchmark X".
	Basis string
}

// Scorer is where quality opinions come from. ok is false when the scorer
// has none, which is "unscored", not "bad".
type Scorer interface {
	Score(task string, c Candidate) (s Score, ok bool)
}

// Chain asks each scorer in turn and returns the first opinion.
type Chain []Scorer

func (ch Chain) Score(task string, c Candidate) (Score, bool) {
	for _, s := range ch {
		if sc, ok := s.Score(task, c); ok {
			return sc, true
		}
	}
	return Score{}, false
}

// Table is an operator-declared score: candidate ID, then task name, to a
// value. "*" matches any task under an ID and any ID for a task; the exact
// ID and task win.
type Table map[string]map[string]float64

func (t Table) Score(task string, c Candidate) (Score, bool) {
	for _, id := range []string{c.ID, "*"} {
		for _, name := range []string{task, "*"} {
			if v, ok := t[id][name]; ok {
				return Score{Value: v, Basis: "table"}, true
			}
		}
	}
	return Score{}, false
}

// History scores by recorded outcomes: a Laplace-smoothed pass rate,
// (passes+1)/(runs+2), once a candidate has at least MinRuns for the task.
// It is safe for concurrent use and persists as JSON.
type History struct {
	MinRuns int
	mu      sync.Mutex
	counts  map[string]map[string]*tally // task -> id -> tally
}

type tally struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
}

func NewHistory(minRuns int) *History {
	return &History{MinRuns: minRuns, counts: map[string]map[string]*tally{}}
}

// Record adds one outcome. Callers record only outcomes a mechanical check
// decided (tests, CI), never the model's own claim.
func (h *History) Record(task, id string, pass bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.counts == nil {
		h.counts = map[string]map[string]*tally{}
	}
	if h.counts[task] == nil {
		h.counts[task] = map[string]*tally{}
	}
	t := h.counts[task][id]
	if t == nil {
		t = &tally{}
		h.counts[task][id] = t
	}
	if pass {
		t.Pass++
	} else {
		t.Fail++
	}
}

func (h *History) Score(task string, c Candidate) (Score, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	t := h.counts[task][c.ID]
	if t == nil || t.Pass+t.Fail < max(h.MinRuns, 1) {
		return Score{}, false
	}
	n := t.Pass + t.Fail
	return Score{Value: float64(t.Pass+1) / float64(n+2), Basis: fmt.Sprintf("history n=%d", n)}, true
}

func (h *History) MarshalJSON() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return json.Marshal(h.counts)
}

func (h *History) UnmarshalJSON(b []byte) error {
	var c map[string]map[string]*tally
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.counts = c
	return nil
}

// Outcome is one recorded verdict, a line of an outcome log.
type Outcome struct {
	Task string `json:"task"`
	ID   string `json:"id"`
	Pass bool   `json:"pass"`
}

// AppendOutcome writes o as one JSON line in a single Write, so concurrent
// appenders to a file opened with O_APPEND do not interleave.
func AppendOutcome(w io.Writer, o Outcome) error {
	if o.Task == "" || o.ID == "" {
		return errors.New("outcome needs a task and an id")
	}
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

// ReadLog records every outcome in an outcome log. Blank lines are ignored;
// malformed ones (a torn final write, a hand edit) are skipped and counted,
// so one bad line does not discard the rest of the evidence.
func (h *History) ReadLog(r io.Reader) (skipped int, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var o Outcome
		if json.Unmarshal(sc.Bytes(), &o) != nil || o.Task == "" || o.ID == "" {
			skipped++
			continue
		}
		h.Record(o.Task, o.ID, o.Pass)
	}
	return skipped, sc.Err()
}
