package jev

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fake serves canned answers and keeps the last request body.
func fake(t *testing.T, answers string, last *map[string]any) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("auth header %q", r.Header.Get("Authorization"))
		}
		if last != nil {
			json.NewDecoder(r.Body).Decode(last)
		}
		w.Write([]byte(`{"model":"jev-1.13.0","answers":` + answers + `}`))
	}))
	t.Cleanup(srv.Close)
	return &Client{Key: "k", URL: srv.URL}
}

var tx = Taxonomy{
	Tasks: map[string]string{"code": "write or change code", "triage": "sort an issue"},
	Caps:  map[string]string{"tools": "must call tools", "vision": "must read images"},
}

func TestClassify(t *testing.T) {
	var req map[string]any
	c := fake(t, `{
		"task":{"type":"choice","choice":"code","probabilities":{"code":0.9,"triage":0.1,"none":0},"confidence":0.8},
		"need:tools":{"type":"noul","noul":0.92},
		"need:vision":{"type":"noul","noul":0.1},
		"output":{"type":"score","score":1.2,"probabilities":{"0":0.05,"1":0.7,"2":0.2,"3":0.05},"confidence":0.6}
	}`, &req)
	got, err := c.Classify(context.Background(), strings.Repeat("x", 4000), tx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Task.Name != "code" || !slices.Equal(got.Task.Need, []string{"tools"}) {
		t.Errorf("task %+v", got.Task)
	}
	if got.Task.InputTokens != 1000 {
		t.Errorf("input tokens %d, want 1000", got.Task.InputTokens)
	}
	// Cumulative 0.05, 0.75, 0.95: the 90th percentile lands on level 2.
	if got.Task.OutputTokens != outputLevels[2].tokens {
		t.Errorf("output tokens %d, want %d", got.Task.OutputTokens, outputLevels[2].tokens)
	}
	if got.Model != "jev-1.13.0" {
		t.Errorf("model %q", got.Model)
	}
	if req["model"] != "jev-latest" {
		t.Errorf("request model %v", req["model"])
	}
}

func TestClassifyUnsureLeavesNameUnknown(t *testing.T) {
	c := fake(t, `{
		"task":{"type":"choice","choice":"code","probabilities":{"code":0.45,"triage":0.4,"none":0.15},"confidence":0.2},
		"need:tools":{"type":"noul","noul":0.5},
		"need:vision":{"type":"noul","noul":0.4},
		"output":{"type":"score","score":0,"probabilities":{"0":1},"confidence":1}
	}`, nil)
	got, err := c.Classify(context.Background(), "hi", tx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Task.Name != "" || got.TaskConfidence != 0.2 {
		t.Errorf("unsure choice must leave the name unknown: %+v", got)
	}
	if !slices.Equal(got.Task.Need, []string{"tools"}) {
		t.Errorf("a coin-flip need is required, since missing it picks a model that cannot do the work: %v", got.Task.Need)
	}
}

func TestClassifyNoneIsUnknown(t *testing.T) {
	c := fake(t, `{
		"task":{"type":"choice","choice":"none","probabilities":{"none":1},"confidence":1},
		"output":{"type":"score","score":0,"probabilities":{"0":1},"confidence":1}
	}`, nil)
	got, err := c.Classify(context.Background(), "hi", Taxonomy{Tasks: tx.Tasks})
	if err != nil {
		t.Fatal(err)
	}
	if got.Task.Name != "" {
		t.Errorf("none must not become a task name: %q", got.Task.Name)
	}
}

func TestClassifyRejectsEmptyTaxonomy(t *testing.T) {
	if _, err := (&Client{Key: "k"}).Classify(context.Background(), "hi", Taxonomy{}); err == nil {
		t.Error("a taxonomy without tasks must be an error")
	}
}

func TestJudge(t *testing.T) {
	cases := []struct {
		p           float64
		decided, ok bool
	}{{0.95, true, true}, {0.05, true, false}, {0.5, false, false}}
	for _, cs := range cases {
		c := fake(t, `{"done":{"type":"noul","noul":`+jsonNum(cs.p)+`}}`, nil)
		v, err := c.Judge(context.Background(), Run{Request: "add a test", Result: "added TestX"})
		if err != nil {
			t.Fatal(err)
		}
		if v.Decided != cs.decided || v.Pass != cs.ok || v.P != cs.p || v.Model != "jev-1.13.0" {
			t.Errorf("p=%v: %+v", cs.p, v)
		}
	}
}

func TestRetriesThrottle(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(529)
			return
		}
		w.Write([]byte(`{"model":"m","answers":{"done":{"type":"noul","noul":1}}}`))
	}))
	defer srv.Close()
	c := &Client{Key: "k", URL: srv.URL, Backoff: time.Millisecond}
	if _, err := c.Judge(context.Background(), Run{Request: "a", Result: "b"}); err != nil {
		t.Fatal(err)
	}
	if n.Load() != 3 {
		t.Errorf("%d attempts, want 3", n.Load())
	}
}

func TestNoRetryOnClientError(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		http.Error(w, `{"detail":"bad"}`, 422)
	}))
	defer srv.Close()
	c := &Client{Key: "k", URL: srv.URL, Backoff: time.Millisecond}
	if _, err := c.Judge(context.Background(), Run{Request: "a", Result: "b"}); err == nil || n.Load() != 1 {
		t.Errorf("err %v after %d attempts", err, n.Load())
	}
}

func TestRetryAfter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := map[string]time.Duration{
		"":                              0,
		"7":                             7 * time.Second,
		"Thu, 01 Jan 2026 00:00:03 GMT": 3 * time.Second,
		"Wed, 31 Dec 2025 23:00:00 GMT": 0,
		"soon":                          0,
	}
	for h, want := range cases {
		if got := retryAfter(h, now); got != want {
			t.Errorf("%q: %v, want %v", h, got, want)
		}
	}
}

func jsonNum(f float64) string { b, _ := json.Marshal(f); return string(b) }

func TestBadAnswersAreErrors(t *testing.T) {
	cases := map[string]string{
		"missing noul":   `{"model":"m","answers":{"done":{"type":"noul"}}}`,
		"null noul":      `{"model":"m","answers":{"done":{"type":"noul","noul":null}}}`,
		"out of range":   `{"model":"m","answers":{"done":{"type":"noul","noul":7}}}`,
		"wrong type":     `{"model":"m","answers":{"done":{"type":"error"}}}`,
		"no model":       `{"answers":{"done":{"type":"noul","noul":1}}}`,
		"missing answer": `{"model":"m","answers":{}}`,
	}
	for name, body := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		if v, err := (&Client{Key: "k", URL: srv.URL}).Judge(context.Background(), Run{Request: "a", Result: "b"}); err == nil {
			t.Errorf("%s: want an error, got %+v", name, v)
		}
		srv.Close()
	}
}

func TestClassifyMissingNeedIsError(t *testing.T) {
	c := fake(t, `{
		"task":{"type":"choice","choice":"code","probabilities":{"code":1},"confidence":1},
		"need:tools":{"type":"noul","noul":0.9},
		"output":{"type":"score","score":0,"probabilities":{"0":1},"confidence":1}
	}`, nil)
	if _, err := c.Classify(context.Background(), "hi", tx); err == nil {
		t.Error("a missing need answer must not mean the capability is not needed")
	}
}

func TestKeyNeverSentInClear(t *testing.T) {
	var hits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(`{"model":"m","answers":{"done":{"type":"noul","noul":1}}}`))
	}))
	defer target.Close()
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redir.Close()
	if _, err := (&Client{Key: "k", URL: redir.URL}).Judge(context.Background(), Run{Request: "a", Result: "b"}); err == nil || hits.Load() != 0 {
		t.Errorf("redirect must not be followed: err %v, %d hits", err, hits.Load())
	}
	if _, err := (&Client{Key: "k", URL: "http://example.com/v1"}).Judge(context.Background(), Run{Request: "a", Result: "b"}); err == nil {
		t.Error("plain http to a remote host must be refused")
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGatewayWithoutKey(t *testing.T) {
	hc := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if h := r.Header.Get("Authorization"); h != "" {
			t.Errorf("no key must send no Authorization, got %q", h)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{},
			Body: io.NopCloser(strings.NewReader(`{"model":"m","answers":{"done":{"type":"noul","noul":1}}}`))}, nil
	})}
	if _, err := (&Client{URL: "http://ai/v1/systemone", HTTP: hc}).Judge(context.Background(), Run{Request: "a", Result: "b"}); err != nil {
		t.Errorf("a keyless gateway on plain http carries no secret: %v", err)
	}
}
