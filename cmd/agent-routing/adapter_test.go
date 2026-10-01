// CLI adapter verification cases (cli-adapter.md "Verification cases"):
// golden determinism, the exit-code table per error class, no-admissible
// stdout with the full rejection set, one-line stderr forms, adoption
// rollback and wire strictness. The adapter performs no network access
// anywhere in its imports, so every test here is a network-disabled run.
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/guygrigsby/agent-routing/routing"
)

var adapterFixtureTime = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

// adapterSnapshot builds a two-offer snapshot: one scored candidate and one
// candidate lacking tool_use so a tool-requiring request rejects it.
func adapterSnapshot(t *testing.T) routing.SnapshotData {
	t.Helper()
	confirmed := time.Date(2026, 9, 29, 8, 30, 0, 0, time.UTC)
	return routing.SnapshotData{
		Schema:          1,
		CreatedAt:       adapterFixtureTime,
		KnowledgeCutoff: adapterFixtureTime,
		Policy:          "educated-v1",
		Distribution:    "local",
		Inputs: []routing.InputData{
			{ObservationID: "obs-1", SourceID: "src-1", PolicyVersion: "local-research-v1"},
		},
		Offers: []routing.OfferData{
			{
				ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha",
				ModelID: "alpha-2026-09", Revision: "r1",
				Capabilities: []string{"text_input", "text_output", "tool_use"},
				Limits: []routing.LimitData{
					{Resource: "input_tokens", Maximum: 200000},
					{Resource: "output_tokens", Maximum: 64000},
				},
				Prices: []routing.PriceData{
					{Component: "input_tokens", USDPerMillion: "2.50"},
					{Component: "output_tokens", USDPerMillion: "10"},
				},
				Signals: []routing.SignalData{
					{ID: "sig-code", Task: "code", Value: 6200, Weight: 40, Basis: "published", EvidenceID: "ev-1", Revision: "r1"},
				},
				LatencyMillis: 900,
				Region:        "us-east",
				Privacy:       "standard",
				ConfirmedAt:   confirmed,
			},
			{
				ID: "prov/beta", Provider: "prov", UpstreamID: "beta",
				Capabilities: []string{"text_input", "text_output"},
				Limits:       []routing.LimitData{{Resource: "context_tokens", Maximum: 1000000}},
				Prices:       []routing.PriceData{{Component: "input_tokens", USDPerMillion: "0"}},
				Signals: []routing.SignalData{
					{ID: "sig-code", Task: "code", Value: 4100, Weight: 30, Basis: "publisher_claim", EvidenceID: "ev-3"},
				},
			},
		},
	}
}

func adapterArtifact(t *testing.T) []byte {
	t.Helper()
	snap, err := routing.NewSnapshot(adapterSnapshot(t))
	if err != nil {
		t.Fatalf("fixture snapshot: %v", err)
	}
	b, err := routing.EncodeSnapshot(snap)
	if err != nil {
		t.Fatalf("fixture artifact: %v", err)
	}
	return b
}

// adapterRequest builds the request DTO whose wire form the tests write.
func adapterRequest(task string, required []string) routing.RequestData {
	return routing.RequestData{
		ID:                   "req-1",
		Task:                 task,
		Mode:                 "performance",
		EffectiveAt:          adapterFixtureTime.Add(time.Hour),
		RuntimeGeneration:    "gen-1",
		Runtime:              adapterRuntime(),
		ImageOperation:       "none",
		RequiredCapabilities: required,
		InputTokens:          1000,
		OutputTokens:         2000,
	}
}

func adapterRuntime() []routing.RuntimeOfferData {
	return []routing.RuntimeOfferData{
		{
			ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha", Revision: "r1",
			Eligible:     true,
			Capabilities: []string{"text_input", "text_output", "tool_use"},
			Limits:       []routing.LimitData{{Resource: "input_tokens", Maximum: 200000}},
			Prices:       []routing.PriceData{{Component: "input_tokens", USDPerMillion: "2.50"}},
			Region:       "us-east", Privacy: "standard",
		},
		{
			ID: "prov/beta", Provider: "prov", UpstreamID: "beta",
			Eligible:     true,
			Capabilities: []string{"text_input", "text_output"},
			Prices:       []routing.PriceData{{Component: "input_tokens", USDPerMillion: "0"}},
		},
	}
}

func requestWire(t *testing.T, rd routing.RequestData) []byte {
	t.Helper()
	// The wire form carries every member; nil slices would encode as null,
	// which canonical-v1 forbids.
	if rd.RequiredCapabilities == nil {
		rd.RequiredCapabilities = []string{}
	}
	if rd.AllowedProviders == nil {
		rd.AllowedProviders = []string{}
	}
	for i := range rd.Runtime {
		o := &rd.Runtime[i]
		if o.Capabilities == nil {
			o.Capabilities = []string{}
		}
		if o.UnsupportedCapabilities == nil {
			o.UnsupportedCapabilities = []string{}
		}
		if o.Limits == nil {
			o.Limits = []routing.LimitData{}
		}
		if o.Prices == nil {
			o.Prices = []routing.PriceData{}
		}
	}
	b, err := routing.CanonicalEncode(rd)
	if err != nil {
		t.Fatalf("fixture request encode: %v", err)
	}
	return append(b, '\n')
}

// captureRun runs run(args) with stdout and stderr swapped for pipes.
func captureRun(t *testing.T, args []string) (stdout, stderr string, exit int) {
	t.Helper()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	exit = run(args)
	outW.Close()
	errW.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	var ob, eb bytes.Buffer
	io.Copy(&ob, outR)
	io.Copy(&eb, errR)
	return ob.String(), eb.String(), exit
}

func writeFileT(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRouteGoldenDeterministic(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	req := filepath.Join(dir, "request.json")
	writeFileT(t, req, requestWire(t, adapterRequest("code", []string{"text_input", "text_output"})))
	args := []string{"route", "--snapshot", art, "--request", req}

	stdout1, stderr1, exit1 := captureRun(t, args)
	if exit1 != adapterExitOK {
		t.Fatalf("exit = %d, stderr = %q", exit1, stderr1)
	}
	if stderr1 != "" {
		t.Fatalf("exit 0 wrote stderr: %q", stderr1)
	}
	stdout2, _, exit2 := captureRun(t, args)
	if exit2 != adapterExitOK || stdout1 != stdout2 {
		t.Fatalf("nondeterministic: exit %d vs %d, stdout identical = %v", exit1, exit2, stdout1 == stdout2)
	}
	if !strings.HasSuffix(stdout1, "\n") || strings.HasSuffix(stdout1, "\n\n") {
		t.Fatalf("stdout must carry exactly one trailing newline: %q", stdout1)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSuffix(stdout1, "\n")), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v", err)
	}
	if _, ok := doc["selected"]; !ok {
		t.Fatalf("exit-0 decision must carry selected: %s", stdout1)
	}
	if _, ok := doc["rejections"]; !ok {
		t.Fatalf("decision must always carry rejections: %s", stdout1)
	}
}

func TestRouteNoAdmissibleOfferFullRejections(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	req := filepath.Join(dir, "request.json")
	writeFileT(t, req, requestWire(t, adapterRequest("code", []string{"structured_output"})))
	stdout, stderr, exit := captureRun(t, []string{"route", "--snapshot", art, "--request", req})
	if exit != adapterExitNoOffer {
		t.Fatalf("exit = %d, want 3; stderr = %q", exit, stderr)
	}
	if stderr != "" {
		t.Fatalf("exit 3 must not write stderr: %q", stderr)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSuffix(stdout, "\n")), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v", err)
	}
	if _, ok := doc["selected"]; ok {
		t.Fatalf("exit-3 decision must omit selected: %s", stdout)
	}
	rej, ok := doc["rejections"].([]any)
	if !ok || len(rej) == 0 {
		t.Fatalf("exit-3 decision must list the full rejection set: %s", stdout)
	}
	alts, ok := doc["alternatives"].([]any)
	if !ok {
		t.Fatalf("alternatives must encode as [], never null: %s", stdout)
	}
	if len(alts) != 0 {
		t.Fatalf("exit-3 alternatives must be empty: %s", stdout)
	}
	// No null anywhere in the document.
	var walk func(v any) bool
	walk = func(v any) bool {
		switch x := v.(type) {
		case nil:
			return true
		case map[string]any:
			for _, vv := range x {
				if walk(vv) {
					return true
				}
			}
		case []any:
			for _, vv := range x {
				if walk(vv) {
					return true
				}
			}
		}
		return false
	}
	if walk(doc) {
		t.Fatalf("exit-3 decision contains a null value: %s", stdout)
	}
}

func TestRouteExitTable(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	goodReq := filepath.Join(dir, "request.json")
	writeFileT(t, goodReq, requestWire(t, adapterRequest("code", nil)))

	cases := []struct {
		name       string
		args       []string
		wantExit   int
		wantStderr bool
		wantCode   string
	}{
		{"unknown flag", []string{"route", "--snapshot", art, "--request", goodReq, "--bogus"}, adapterExitUsage, true, ""},
		{"missing flag", []string{"route", "--request", goodReq}, adapterExitUsage, true, ""},
		{"unknown command", []string{"frobnicate"}, adapterExitUsage, true, ""},
		{"extra positional", []string{"route", "--snapshot", art, "--request", goodReq, "junk"}, adapterExitUsage, true, ""},
		{"unknown JSON member", []string{"route", "--snapshot", art, "--request", writeReq(t, dir, "member", `{"id":"req-1","task":"code","mode":"performance","effective_at":"2026-09-30T13:00:00.000Z","runtime_generation":"gen-1","runtime":[],"mystery":1}`+"\n")}, adapterExitUsage, true, "invalid_request_wire"},
		{"duplicate JSON member", []string{"route", "--snapshot", art, "--request", writeReq(t, dir, "dup", `{"id":"req-1","id":"req-2","task":"code","mode":"performance","effective_at":"2026-09-30T13:00:00.000Z","runtime_generation":"gen-1","runtime":[]}`+"\n")}, adapterExitUsage, true, "invalid_request_wire"},
		{"no trailing newline", []string{"route", "--snapshot", art, "--request", writeReq(t, dir, "nonl", `{"id":"req-1","task":"code","mode":"performance","effective_at":"2026-09-30T13:00:00.000Z","runtime_generation":"gen-1","runtime":[]}`)}, adapterExitUsage, true, "invalid_request_wire"},
		{"trailing garbage", []string{"route", "--snapshot", art, "--request", writeReq(t, dir, "trail", `{"id":"req-1","task":"code","mode":"performance","effective_at":"2026-09-30T13:00:00.000Z","runtime_generation":"gen-1","runtime":[]} {}`+"\n")}, adapterExitUsage, true, "invalid_request_wire"},
		{"domain invalid request", []string{"route", "--snapshot", art, "--request", writeReq(t, dir, "badtask", `{"id":"req-1","task":"crypto","mode":"performance","effective_at":"2026-09-30T13:00:00.000Z","runtime_generation":"gen-1","runtime":[]}`+"\n")}, adapterExitUsage, true, "invalid_request"},
		{"unreadable request", []string{"route", "--snapshot", art, "--request", filepath.Join(dir, "absent.json")}, adapterExitUsage, true, "invalid_request_wire"},
		{"truncated artifact", []string{"route", "--snapshot", truncateArtifact(t, dir), "--request", goodReq}, adapterExitSnapshot, true, "invalid_snapshot"},
		{"missing artifact", []string{"route", "--snapshot", filepath.Join(dir, "absent.bin"), "--request", goodReq}, adapterExitSnapshot, true, "snapshot_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exit := captureRun(t, tc.args)
			if exit != tc.wantExit {
				t.Fatalf("exit = %d, want %d; stderr = %q", exit, tc.wantExit, stderr)
			}
			if stdout != "" {
				t.Fatalf("nonzero exit wrote stdout: %q", stdout)
			}
			if !tc.wantStderr || stderr == "" {
				t.Fatalf("expected stderr, got %q", stderr)
			}
			lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
			if len(lines) != 1 {
				t.Fatalf("stderr must be exactly one line: %q", stderr)
			}
			if tc.wantCode != "" {
				if !strings.HasPrefix(lines[0], "agent-routing: "+tc.wantCode+":") {
					t.Fatalf("stderr code: want %q, got %q", tc.wantCode, lines[0])
				}
			} else if !strings.HasPrefix(lines[0], "usage: ") {
				t.Fatalf("usage error must print the usage line: %q", lines[0])
			}
		})
	}
}

func writeReq(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, "req-"+name+".json")
	writeFileT(t, p, []byte(content))
	return p
}

func truncateArtifact(t *testing.T, dir string) string {
	t.Helper()
	full := adapterArtifact(t)
	cut := full[:len(full)/2]
	p := filepath.Join(dir, "truncated.bin")
	writeFileT(t, p, cut)
	return p
}

func TestSnapshotCheck(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	snap, err := routing.NewSnapshot(adapterSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, exit := captureRun(t, []string{"snapshot", "check", "--artifact", art})
	if exit != 0 || stderr != "" {
		t.Fatalf("exit = %d stderr = %q", exit, stderr)
	}
	if got := strings.TrimSpace(stdout); got != snap.Hash() {
		t.Fatalf("check reported %q, want artifact hash %q", got, snap.Hash())
	}
	_, stderr2, exit2 := captureRun(t, []string{"snapshot", "check", "--artifact", filepath.Join(dir, "absent.bin")})
	if exit2 != adapterExitSnapshot || !strings.HasPrefix(stderr2, "agent-routing: snapshot_unavailable:") {
		t.Fatalf("absent artifact: exit %d stderr %q", exit2, stderr2)
	}
}

func TestSnapshotAdoptRetainsPriorGeneration(t *testing.T) {
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	art1 := filepath.Join(dir, "gen1.bin")
	writeFileT(t, art1, adapterArtifact(t))
	snap1, err := routing.NewSnapshot(adapterSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exit := captureRun(t, []string{"snapshot", "adopt", "--artifact", art1, "--store", store})
	if exit != 0 || stderr != "" {
		t.Fatalf("first adopt: exit %d stderr %q", exit, stderr)
	}
	if got := strings.TrimSpace(stdout); got != snap1.Hash() {
		t.Fatalf("adopt reported %q, want %q", got, snap1.Hash())
	}

	// Corrupted artifact must fail without touching CURRENT.
	corrupt := filepath.Join(dir, "corrupt.bin")
	writeFileT(t, corrupt, bytes.Repeat([]byte{0x7f}, 512))
	_, stderr2, exit2 := captureRun(t, []string{"snapshot", "adopt", "--artifact", corrupt, "--store", store})
	if exit2 != adapterExitSnapshot || !strings.HasPrefix(stderr2, "agent-routing: invalid_snapshot:") {
		t.Fatalf("corrupt adopt: exit %d stderr %q", exit2, stderr2)
	}
	current, err := os.ReadFile(filepath.Join(store, "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(current)); got != snap1.Hash() {
		t.Fatalf("failed adoption changed CURRENT to %q, want retained %q", got, snap1.Hash())
	}

	// A valid second generation adopts cleanly and the prior bytes survive
	// retention.
	art2 := filepath.Join(dir, "gen2.bin")
	data2 := adapterSnapshot(t)
	data2.Offers[0].LatencyMillis = 800
	snap2, err := routing.NewSnapshot(data2)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := routing.EncodeSnapshot(snap2)
	if err != nil {
		t.Fatal(err)
	}
	writeFileT(t, art2, b2)
	_, stderr3, exit3 := captureRun(t, []string{"snapshot", "adopt", "--artifact", art2, "--store", store})
	if exit3 != 0 || stderr3 != "" {
		t.Fatalf("second adopt: exit %d stderr %q", exit3, stderr3)
	}
	if _, err := os.Stat(filepath.Join(store, "snapshots", snap1.Hash()+".bin")); err != nil {
		t.Fatalf("prior generation file must survive retention: %v", err)
	}
	current2, err := os.ReadFile(filepath.Join(store, "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(current2)); got != snap2.Hash() {
		t.Fatalf("CURRENT = %q, want %q", got, snap2.Hash())
	}
}

func TestRouteReadsStdinRequest(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	wire := requestWire(t, adapterRequest("code", nil))

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	go func() { w.Write(wire); w.Close() }()
	defer func() { os.Stdin = oldStdin }()

	stdout, stderr, exit := captureRun(t, []string{"route", "--snapshot", art, "--request", "-"})
	if exit != adapterExitOK || stderr != "" {
		t.Fatalf("exit = %d stderr = %q", exit, stderr)
	}
	if !strings.Contains(stdout, `"selected"`) {
		t.Fatalf("stdin request did not produce a selection: %s", stdout)
	}
}
