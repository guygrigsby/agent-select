// Command agent-select ranks candidate models from a JSON request on stdin
// and writes the ordering to stdout. It is a thin wrapper over package
// route, for shell scripts and for trying a policy without writing Go.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/guygrigsby/agent-select/route"
)

type request struct {
	Task       task                          `json:"task"`
	Candidates []candidate                   `json:"candidates"`
	Scores     map[string]map[string]float64 `json:"scores"`
}

type task struct {
	Name          string   `json:"name"`
	Mode          string   `json:"mode"`
	Need          []string `json:"need"`
	InputTokens   int64    `json:"input_tokens"`
	OutputTokens  int64    `json:"output_tokens"`
	MaxCostUSD    *float64 `json:"max_cost_usd"`
	MaxLatencyMS  int64    `json:"max_latency_ms"`
	MinScore      *float64 `json:"min_score"`
	AllowUnscored bool     `json:"allow_unscored"`
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

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

// run exits 0 with a selection, 3 with none, 2 on bad input.
func run(args []string, in io.Reader, out, errw io.Writer) int {
	if len(args) != 1 || args[0] != "rank" {
		fmt.Fprintln(errw, "usage: agent-select rank < request.json")
		return 2
	}
	var req request
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fmt.Fprintf(errw, "agent-select: bad request: %v\n", err)
		return 2
	}
	cands := make([]route.Candidate, len(req.Candidates))
	for i, c := range req.Candidates {
		cands[i] = route.Candidate{ID: c.ID, Caps: c.Caps, ContextTokens: c.ContextTokens, MaxOutputTokens: c.MaxOutputTokens,
			InputUSDPerMTok: c.InputUSDPerMTok, OutputUSDPerMTok: c.OutputUSDPerMTok, Latency: time.Duration(c.LatencyMS) * time.Millisecond}
	}
	t := req.Task
	res := route.Rank(cands, route.Task{Name: t.Name, Mode: route.Mode(t.Mode), Need: t.Need, InputTokens: t.InputTokens,
		OutputTokens: t.OutputTokens, MaxCostUSD: t.MaxCostUSD, MaxLatency: time.Duration(t.MaxLatencyMS) * time.Millisecond,
		MinScore: t.MinScore, AllowUnscored: t.AllowUnscored}, route.Table(req.Scores))
	resp := response{Ranked: []ranked{}, Rejected: res.Rejected}
	if resp.Rejected == nil {
		resp.Rejected = []route.Rejection{}
	}
	for _, r := range res.Ranked {
		resp.Ranked = append(resp.Ranked, ranked{ID: r.ID, Score: r.Score, CostUSD: r.CostUSD, Warnings: r.Warnings})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		fmt.Fprintf(errw, "agent-select: %v\n", err)
		return 2
	}
	if len(resp.Ranked) == 0 {
		return 3
	}
	return 0
}
