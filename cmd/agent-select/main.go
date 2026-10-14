// Command agent-select ranks candidate models from a JSON request on stdin
// and writes the ordering to stdout. It is a thin wrapper over package
// route, for shell scripts, hosts in other languages and for trying a policy
// without writing Go.
//
//	agent-select rank [--history FILE] [--min-runs N] < request.json
//	agent-select record --history FILE --task T --id ID --verdict pass|fail
//	agent-select classify --taxonomy TAXONOMY < prompt
//	agent-select judge --history FILE --task T --id ID < run.json
//
// FILE is an append-only outcome log. rank scores from it first and falls
// back to the request's scores table. classify and judge call Jev at
// TYPESAFE_URL, a gateway such as Aperture's /v1/systemone, or at TypeSafe
// with TYPESAFE_API_KEY: classify prints
// the task fields of a rank request, judge records a verdict Jev is sure of.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/guygrigsby/agent-select/jev"
	"github.com/guygrigsby/agent-select/route"
)

type request struct {
	Task       task                          `json:"task"`
	Candidates []candidate                   `json:"candidates"`
	Scores     map[string]map[string]float64 `json:"scores"`
}

type task struct {
	Name          string   `json:"name"`
	Mode          string   `json:"mode,omitempty"`
	Need          []string `json:"need"`
	InputTokens   int64    `json:"input_tokens"`
	OutputTokens  int64    `json:"output_tokens"`
	MaxCostUSD    *float64 `json:"max_cost_usd,omitempty"`
	MaxLatencyMS  int64    `json:"max_latency_ms,omitempty"`
	MinScore      *float64 `json:"min_score,omitempty"`
	AllowUnscored bool     `json:"allow_unscored,omitempty"`
}

type candidate struct {
	ID               string   `json:"id"`
	Caps             []string `json:"caps"`
	ContextTokens    int64    `json:"context_tokens"`
	MaxOutputTokens  int64    `json:"max_output_tokens"`
	InputUSDPerMTok  *float64 `json:"input_usd_per_mtok"`
	OutputUSDPerMTok *float64 `json:"output_usd_per_mtok"`
	LatencyMS        int64    `json:"latency_ms"`
}

type ranked struct {
	ID       string       `json:"id"`
	Score    *route.Score `json:"score,omitempty"`
	CostUSD  *float64     `json:"cost_usd,omitempty"`
	Warnings []string     `json:"warnings,omitempty"`
}

type response struct {
	Ranked   []ranked          `json:"ranked"`
	Rejected []route.Rejection `json:"rejected"`
}

const usage = `usage:
  agent-select rank [--history FILE] [--min-runs N] < request.json
  agent-select record --history FILE --task T --id ID --verdict pass|fail
  agent-select classify --taxonomy TAXONOMY < prompt
  agent-select judge --history FILE --task T --id ID < run.json`

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

// run exits 0 with a selection, classification or recorded outcome, 3 with
// no selection or an undecided judge, 2 on bad input and 1 when Jev or the
// outcome log failed.
func run(args []string, in io.Reader, out, errw io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errw, usage)
		return 2
	}
	switch args[0] {
	case "rank":
		return rank(args[1:], in, out, errw)
	case "record":
		return record(args[1:], errw)
	case "classify":
		return classify(args[1:], in, out, errw)
	case "judge":
		return judge(args[1:], in, out, errw)
	}
	fmt.Fprintln(errw, usage)
	return 2
}

func record(args []string, errw io.Writer) int {
	fl := flag.NewFlagSet("record", flag.ContinueOnError)
	fl.SetOutput(errw)
	path := fl.String("history", "", "outcome log to append to")
	var o route.Outcome
	fl.StringVar(&o.Task, "task", "", "task name, as given to rank")
	fl.StringVar(&o.ID, "id", "", "candidate id that ran")
	verdict := fl.String("verdict", "", "pass or fail, from a mechanical check")
	if fl.Parse(args) != nil || *path == "" || o.Task == "" || o.ID == "" || fl.NArg() > 0 || (*verdict != "pass" && *verdict != "fail") {
		fmt.Fprintln(errw, usage)
		return 2
	}
	o.Pass = *verdict == "pass"
	return appendOutcome(*path, o, errw)
}

func appendOutcome(path string, o route.Outcome, errw io.Writer) int {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 1
	}
	err = route.AppendOutcome(f, o)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 1
	}
	return 0
}

// jevTimeout bounds one classify or judge, retries included.
const jevTimeout = 2 * time.Minute

func jevClient(errw io.Writer) *jev.Client {
	c := &jev.Client{Key: os.Getenv("TYPESAFE_API_KEY"), URL: os.Getenv("TYPESAFE_URL")}
	if c.Key == "" && c.URL == "" {
		fmt.Fprintln(errw, "agent-select: set TYPESAFE_URL to a gateway, or TYPESAFE_API_KEY")
		return nil
	}
	return c
}

func writeJSON(out, errw io.Writer, v any) int {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 1
	}
	return 0
}

func classify(args []string, in io.Reader, out, errw io.Writer) int {
	fl := flag.NewFlagSet("classify", flag.ContinueOnError)
	fl.SetOutput(errw)
	path := fl.String("taxonomy", "", "JSON file of task names and capability tags with their meanings")
	if fl.Parse(args) != nil || *path == "" || fl.NArg() > 0 {
		fmt.Fprintln(errw, usage)
		return 2
	}
	var tx jev.Taxonomy
	b, err := os.ReadFile(*path)
	if err == nil {
		err = json.Unmarshal(b, &tx)
	}
	if err != nil {
		fmt.Fprintf(errw, "agent-select: bad taxonomy: %v\n", err)
		return 2
	}
	prompt, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 2
	}
	c := jevClient(errw)
	if c == nil {
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), jevTimeout)
	defer cancel()
	res, err := c.Classify(ctx, string(prompt), tx)
	if err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 1
	}
	return writeJSON(out, errw, struct {
		Task           task    `json:"task"`
		TaskConfidence float64 `json:"task_confidence"`
		Model          string  `json:"model"`
	}{task{Name: res.Task.Name, Need: res.Task.Need, InputTokens: res.Task.InputTokens, OutputTokens: res.Task.OutputTokens},
		res.TaskConfidence, res.Model})
}

func judge(args []string, in io.Reader, out, errw io.Writer) int {
	fl := flag.NewFlagSet("judge", flag.ContinueOnError)
	fl.SetOutput(errw)
	path := fl.String("history", "", "outcome log to append a decided verdict to")
	var o route.Outcome
	fl.StringVar(&o.Task, "task", "", "task name, as given to rank")
	fl.StringVar(&o.ID, "id", "", "candidate id that ran")
	if fl.Parse(args) != nil || *path == "" || o.Task == "" || o.ID == "" || fl.NArg() > 0 {
		fmt.Fprintln(errw, usage)
		return 2
	}
	var r jev.Run
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		fmt.Fprintf(errw, "agent-select: bad run: %v\n", err)
		return 2
	}
	c := jevClient(errw)
	if c == nil {
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), jevTimeout)
	defer cancel()
	v, err := c.Judge(ctx, r)
	if err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 1
	}
	if code := writeJSON(out, errw, struct {
		Pass    bool    `json:"pass"`
		Decided bool    `json:"decided"`
		P       float64 `json:"p"`
		Model   string  `json:"model"`
	}{v.Pass, v.Decided, v.P, v.Model}); code != 0 {
		return code
	}
	if !v.Decided {
		return 3
	}
	o.Pass, o.Judge = v.Pass, v.Model
	return appendOutcome(*path, o, errw)
}

// history loads the outcome log. A missing log is no history yet; an
// unreadable one or bad lines warn and rank on what remains, since a ranking
// without history beats no ranking.
func history(path string, minRuns int, errw io.Writer) *route.History {
	h := route.NewHistory(minRuns)
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return h
	}
	if err != nil {
		fmt.Fprintf(errw, "agent-select: history ignored: %v\n", err)
		return h
	}
	defer f.Close()
	skipped, err := h.ReadLog(f)
	if skipped > 0 {
		fmt.Fprintf(errw, "agent-select: %d malformed history lines skipped\n", skipped)
	}
	if err != nil {
		fmt.Fprintf(errw, "agent-select: history read stopped early: %v\n", err)
	}
	return h
}

func rank(args []string, in io.Reader, out, errw io.Writer) int {
	fl := flag.NewFlagSet("rank", flag.ContinueOnError)
	fl.SetOutput(errw)
	path := fl.String("history", "", "outcome log scored ahead of the request's scores")
	minRuns := fl.Int("min-runs", 5, "outcomes needed before history scores a candidate")
	if fl.Parse(args) != nil || fl.NArg() > 0 {
		fmt.Fprintln(errw, usage)
		return 2
	}
	var req request
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fmt.Fprintf(errw, "agent-select: bad request: %v\n", err)
		return 2
	}
	var scorer route.Scorer = route.Table(req.Scores)
	if *path != "" {
		scorer = route.Chain{history(*path, *minRuns, errw), scorer}
	}
	cands := make([]route.Candidate, len(req.Candidates))
	for i, c := range req.Candidates {
		cands[i] = route.Candidate{ID: c.ID, Caps: c.Caps, ContextTokens: c.ContextTokens, MaxOutputTokens: c.MaxOutputTokens,
			InputUSDPerMTok: c.InputUSDPerMTok, OutputUSDPerMTok: c.OutputUSDPerMTok, Latency: time.Duration(c.LatencyMS) * time.Millisecond}
	}
	t := req.Task
	res := route.Rank(cands, route.Task{Name: t.Name, Mode: route.Mode(t.Mode), Need: t.Need, InputTokens: t.InputTokens,
		OutputTokens: t.OutputTokens, MaxCostUSD: t.MaxCostUSD, MaxLatency: time.Duration(t.MaxLatencyMS) * time.Millisecond,
		MinScore: t.MinScore, AllowUnscored: t.AllowUnscored}, scorer)
	resp := response{Ranked: []ranked{}, Rejected: res.Rejected}
	if resp.Rejected == nil {
		resp.Rejected = []route.Rejection{}
	}
	for _, r := range res.Ranked {
		resp.Ranked = append(resp.Ranked, ranked{ID: r.ID, Score: r.Score, CostUSD: r.CostUSD, Warnings: r.Warnings})
	}
	if writeJSON(out, errw, resp) != 0 {
		return 2
	}
	if len(resp.Ranked) == 0 {
		return 3
	}
	return 0
}
