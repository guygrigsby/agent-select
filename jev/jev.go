// Package jev asks TypeSafe's Jev model the two questions a host cannot
// answer in code: what task a prompt is (Classify) and whether a run did
// what was asked (Judge). It calls the network, so it lives beside route,
// never inside Rank (ADR 0003).
package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/guygrigsby/agent-select/route"
)

const (
	defaultURL     = "https://api.typesafe.ai/v1/systemone"
	defaultModel   = "jev-latest"
	defaultBackoff = 500 * time.Millisecond
	attempts       = 5
	maxWait        = time.Minute
	userAgent      = "agent-select (+https://github.com/guygrigsby/agent-select)"
)

// Client calls a System One evaluation endpoint: TypeSafe's with a Key, or
// a gateway that authenticates the caller itself with none.
type Client struct {
	Key   string       // sent only over https or to loopback
	URL   string       // default TypeSafe's public endpoint
	Model string       // default jev-latest
	HTTP  *http.Client // default one with a 30s timeout
	// Backoff is the first retry delay on 429 or 529 without Retry-After;
	// it doubles per attempt.
	Backoff time.Duration
}

type question struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions"`
	Criteria     any    `json:"criteria,omitempty"`
}

type answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Noul          *float64           `json:"noul"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

// valid reports whether a is a usable answer to a question of type typ. A
// missing or out-of-range value is an error, never read as zero.
func (a answer) valid(typ string) bool {
	in := func(f float64) bool { return f >= 0 && f <= 1 }
	if a.Type != typ {
		return false
	}
	switch typ {
	case "noul":
		return a.Noul != nil && in(*a.Noul)
	case "choice":
		if a.Choice == "" {
			return false
		}
	}
	for _, p := range a.Probabilities {
		if !in(p) {
			return false
		}
	}
	return in(a.Confidence)
}

// ask sends one request and returns the answers and the model that gave
// them. It retries throttling and overload, honoring Retry-After; any other
// failure returns at once, since resending the same body cannot fix it.
func (c *Client) ask(ctx context.Context, state any, qs map[string]question) (map[string]answer, string, error) {
	body, err := json.Marshal(map[string]any{"state": state, "model": or(c.Model, defaultModel), "questions": qs})
	if err != nil {
		return nil, "", err
	}
	endpoint := or(c.URL, defaultURL)
	if u, err := url.Parse(endpoint); err != nil || (c.Key != "" && u.Scheme != "https" && !(u.Scheme == "http" && isLoopback(u.Hostname()))) {
		return nil, "", fmt.Errorf("jev: a key goes only to an https endpoint: %q", endpoint)
	}
	hc := http.Client{Timeout: 30 * time.Second}
	if c.HTTP != nil {
		hc = *c.HTTP
	}
	// A redirect would carry the key to wherever it points.
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	wait := c.Backoff
	if wait == 0 {
		wait = defaultBackoff
	}
	for attempt := range attempts {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, "", err
		}
		if c.Key != "" {
			req.Header.Set("Authorization", "Bearer "+c.Key)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", userAgent)
		resp, err := hc.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("jev: %w", err)
		}
		b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil {
			return nil, "", fmt.Errorf("jev: reading response: %w", err)
		}
		switch {
		case resp.StatusCode == http.StatusOK:
			var out struct {
				Model   string            `json:"model"`
				Answers map[string]answer `json:"answers"`
			}
			if err := json.Unmarshal(b, &out); err != nil {
				return nil, "", fmt.Errorf("jev: bad response: %w", err)
			}
			if out.Model == "" {
				return nil, "", errors.New("jev: response names no model")
			}
			for id, q := range qs {
				if a, ok := out.Answers[id]; !ok || !a.valid(q.Type) {
					return nil, "", fmt.Errorf("jev: no valid answer for %q", id)
				}
			}
			return out.Answers, out.Model, nil
		case (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 529) && attempt < attempts-1:
			d := max(wait<<attempt, retryAfter(resp.Header.Get("Retry-After"), time.Now()))
			if d > maxWait {
				return nil, "", fmt.Errorf("jev: %s, retry after %v", resp.Status, d)
			}
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(d):
			}
		default:
			// The body can quote the request back; keep enough to diagnose.
			return nil, "", fmt.Errorf("jev: %s: %s", resp.Status, bytes.TrimSpace(b[:min(len(b), 256)]))
		}
	}
	return nil, "", errors.New("jev: still throttled after retries")
}

// retryAfter reads either form of the header, seconds or an HTTP date.
// Anything unreadable or past is zero.
func retryAfter(h string, now time.Time) time.Duration {
	if s, err := strconv.Atoi(h); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// Taxonomy is the host's vocabulary: the task names its Scorer knows and
// the capability tags its candidates carry, each with what it means.
type Taxonomy struct {
	Tasks map[string]string `json:"tasks"`
	Caps  map[string]string `json:"caps"`
	// MinConfidence is the task choice confidence below which the name is
	// left unknown. Zero means 0.5.
	MinConfidence float64 `json:"min_confidence,omitempty"`
}

// none is the escape option: without it Jev must force a fit.
const none = "none"

// outputLevels are the answer sizes Jev picks between, with the tokens a
// candidate must be able to emit for each.
var outputLevels = []struct {
	desc   string
	tokens int64
}{
	{"A short answer: a sentence, a yes or no, a one-line fix", 500},
	{"A paragraph or a small code change in one function", 2_000},
	{"A page of text or a change across several functions or one file", 8_000},
	{"A long document or a change across many files", 32_000},
}

// Classification is a partly filled route.Task: Name, Need, InputTokens
// and OutputTokens. Mode and bounds stay the host's policy.
type Classification struct {
	Task           route.Task
	TaskConfidence float64
	Model          string
}

// Classify asks, in one request, which task the prompt is, which
// capabilities it needs and how long its answer will be. A task Jev is
// unsure of, or that fits none of the host's, comes back with no name:
// unknown, never a guess. A capability is required at even odds, because
// missing one routes to a model that cannot do the work. Output tokens are
// the size Jev is 90% sure the answer fits in; input tokens are estimated in
// code at four bytes a token.
func (c *Client) Classify(ctx context.Context, prompt string, tx Taxonomy) (Classification, error) {
	if len(tx.Tasks) == 0 {
		return Classification{}, errors.New("jev: taxonomy has no tasks")
	}
	if prompt == "" {
		return Classification{}, errors.New("jev: empty prompt")
	}
	opts := map[string]string{none: "None of the other options describes this work"}
	maps.Copy(opts, tx.Tasks)
	levels := make([]string, len(outputLevels))
	for i, l := range outputLevels {
		levels[i] = l.desc
	}
	qs := map[string]question{
		"task":   {Type: "choice", Instructions: "Which kind of work does `prompt` ask for?", Criteria: opts},
		"output": {Type: "score", Instructions: "How long will a complete response to `prompt` be?", Criteria: levels},
	}
	for tag, desc := range tx.Caps {
		qs["need:"+tag] = question{Type: "noul",
			Instructions: map[string]string{"capability": desc, "question": "Does doing the work `prompt` asks for require `capability`?"},
			Criteria:     map[string]string{"true": "The work cannot be done without it", "false": "The work can be done fully without it"}}
	}
	ans, model, err := c.ask(ctx, map[string]string{"prompt": prompt}, qs)
	if err != nil {
		return Classification{}, err
	}
	min := tx.MinConfidence
	if min == 0 {
		min = 0.5
	}
	out := Classification{Model: model, TaskConfidence: ans["task"].Confidence}
	if t := ans["task"]; t.Choice != none && t.Confidence >= min {
		if _, known := tx.Tasks[t.Choice]; known {
			out.Task.Name = t.Choice
		}
	}
	for tag := range tx.Caps {
		if *ans["need:"+tag].Noul >= 0.5 {
			out.Task.Need = append(out.Task.Need, tag)
		}
	}
	slices.Sort(out.Task.Need)
	out.Task.InputTokens = int64(len(prompt)+3) / 4
	out.Task.OutputTokens = percentile(ans["output"].Probabilities, 0.9)
	return out, nil
}

// percentile returns the tokens of the smallest level whose cumulative
// probability reaches q, or of the largest level if it never does.
func percentile(p map[string]float64, q float64) int64 {
	var sum float64
	for i, l := range outputLevels {
		sum += p[strconv.Itoa(i)]
		if sum >= q {
			return l.tokens
		}
	}
	return outputLevels[len(outputLevels)-1].tokens
}

// Run is a finished piece of work to judge.
type Run struct {
	Request string `json:"request"`
	Result  string `json:"result"`
}

// Verdict is Jev's opinion of a run. Decided is false between the
// thresholds: an unsure judge records nothing rather than a guess.
type Verdict struct {
	Pass, Decided bool
	P             float64 // probability the run did what was asked
	Model         string
}

// Pass and fail thresholds. Calibrate against mechanical verdicts before
// moving them.
const passAt, failAt = 0.8, 0.2

// Judge asks whether the run's result does what its request asked.
func (c *Client) Judge(ctx context.Context, r Run) (Verdict, error) {
	if r.Request == "" || r.Result == "" {
		return Verdict{}, errors.New("jev: a run needs a request and a result")
	}
	ans, model, err := c.ask(ctx, r, map[string]question{"done": {Type: "noul",
		Instructions: "Does `result` fully accomplish what `request` asked for?",
		Criteria: map[string]string{
			"true":  "Every part of the request is done correctly, with nothing missing or broken",
			"false": "Something asked for is missing, wrong or unfinished, or the result claims success without showing it",
		}}})
	if err != nil {
		return Verdict{}, err
	}
	p := *ans["done"].Noul
	return Verdict{Pass: p >= passAt, Decided: p >= passAt || p <= failAt, P: p, Model: model}, nil
}
