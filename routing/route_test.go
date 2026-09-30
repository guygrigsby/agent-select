package routing

import (
	"reflect"
	"testing"
	"time"
)

// Route engine tests driven by synthetic owned fixtures.

func routeFixture(t *testing.T, sd SnapshotData, rd RequestData) (Decision, error) {
	t.Helper()
	snap, err := NewSnapshot(sd)
	if err != nil {
		t.Fatalf("snapshot fixture invalid: %v", err)
	}
	req, err := NewRequest(rd)
	if err != nil {
		t.Fatalf("request fixture invalid: %v", err)
	}
	return NewEngine(snap).Route(req)
}

func TestRouteUnavailableEngine(t *testing.T) {
	req, err := NewRequest(fixtureRequestData())
	if err != nil {
		t.Fatal(err)
	}
	d, err := NewEngine(nil).Route(req)
	if d.RequestID != "" || err == nil {
		t.Fatalf("want zero decision and error, got %+v %v", d, err)
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeSnapshotUnavailable {
		t.Fatalf("want snapshot_unavailable, got %v", err)
	}
}

func TestRouteZeroValueRequest(t *testing.T) {
	snap, err := NewSnapshot(fixtureSnapshotData())
	if err != nil {
		t.Fatal(err)
	}
	d, err := NewEngine(snap).Route(Request{})
	if d.RequestID != "" || err == nil {
		t.Fatalf("want zero decision and error, got %+v %v", d, err)
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeInvalidRequest {
		t.Fatalf("want invalid_request, got %v", err)
	}
}

func TestRouteAdmitsAndRanks(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	if dec.Selected == nil {
		t.Fatalf("no selection: %+v", dec.Rejections)
	}
	if dec.Selected.OfferID != "prov/alpha" {
		t.Fatalf("want prov/alpha, got %s", dec.Selected.OfferID)
	}
	if dec.SnapshotHash == "" || dec.Policy != "educated-v1" || dec.RuntimeGeneration != "rg-1" {
		t.Fatalf("decision metadata wrong: %+v", dec)
	}
	// Estimate: alpha has code (6200*40) and general (5500*20):
	// floor((248000+110000)/60) = floor(5966.67) = 5966.
	if !dec.Selected.Estimate.Known || dec.Selected.Estimate.Value != 5966 {
		t.Fatalf("estimate wrong: %+v", dec.Selected.Estimate)
	}
	if dec.Selected.Estimate.Basis != "heuristic" {
		t.Fatalf("basis wrong: %s", dec.Selected.Estimate.Basis)
	}
	// Cost: 1000*2.50/1e6 USD + 500*10/1e6 USD = 7500 micros + 5000 micros.
	if !dec.Selected.Cost.Known || dec.Selected.Cost.UpperMicros != 7500 {
		t.Fatalf("cost wrong: %+v", dec.Selected.Cost)
	}
	// The general signal carries a mixed_protocol warning.
	if !containsStr(dec.Selected.Estimate.Warnings, "mixed_protocol") {
		t.Fatalf("estimate warnings missing mixed_protocol: %v", dec.Selected.Estimate.Warnings)
	}
}

// containsStr reports membership without importing slices helpers.
func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// betaRuntime returns a runtime entry for the snapshot-only beta offer.
func betaRuntime() RuntimeOfferData {
	return RuntimeOfferData{
		ID: "prov/beta", Provider: "prov", UpstreamID: "beta", Eligible: true,
		Capabilities: []string{"text_input", "text_output", "image_input"},
		Region:       "us-east", Privacy: "standard",
	}
}

func TestRouteRejectionReasons(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = append(rd.Runtime, betaRuntime())
	rd.InputTokens = 300000 // exceeds alpha's input limit
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("beta should absorb: %v", err)
	}
	if dec.Selected == nil || dec.Selected.OfferID != "prov/beta" {
		t.Fatalf("want beta admitted via context bound: %+v", dec)
	}
	// Alpha rejected for its own bound; beta cost is unknown (only the
	// input quote exists) but admitted non-strict with a warning.
	if !containsStr(dec.Selected.Warnings, "cost_unknown") {
		t.Fatalf("beta should disclose cost_unknown: %v", dec.Selected.Warnings)
	}

	rd2 := fixtureRequestData()
	rd2.Runtime = append(rd2.Runtime, betaRuntime())
	rd2.AllowedProviders = []string{"other"}
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil || dec2.Selected != nil {
		t.Fatalf("want no-admissible, got %+v %v", dec2, err2)
	}
	if e, ok := err2.(*Error); !ok || e.Code != CodeNoAdmissibleOffer {
		t.Fatalf("want no_admissible_offer, got %v", err2)
	}
	if len(dec2.Rejections) != 2 || dec2.Rejections[0].OfferID != "prov/alpha" ||
		!reflect.DeepEqual(dec2.Rejections[0].Reasons, []string{"provider_disallowed"}) ||
		!reflect.DeepEqual(dec2.Rejections[1].Reasons, []string{"provider_disallowed"}) {
		t.Fatalf("rejections wrong: %+v", dec2.Rejections)
	}
}

func TestRouteRegionPrivacyFailClosed(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Region = "us-east"
	dec, err := routeFixture(t, sd, rd)
	if err != nil || dec.Selected == nil {
		t.Fatalf("runtime region should satisfy: %v", err)
	}
	// Runtime privacy emptied: snapshot privacy must NOT be consulted.
	rd2 := fixtureRequestData()
	rd2.Runtime[0].Privacy = ""
	rd2.Privacy = "standard"
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("empty runtime privacy must fail closed")
	}
	found := false
	for _, r := range dec2.Rejections {
		if r.OfferID == "prov/alpha" && containsStr(r.Reasons, "privacy_unknown") {
			found = true
		}
	}
	if !found {
		t.Fatalf("want privacy_unknown on alpha: %+v", dec2.Rejections)
	}
}

func TestRouteRevisionConflictDiscardsProfile(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime[0].Revision = "r2" // differs from snapshot r1
	rd.Runtime[0].Limits = []LimitData{
		{Resource: "input_tokens", Maximum: 200000},
		{Resource: "output_tokens", Maximum: 64000},
	}
	rd.Runtime[0].Prices = []PriceData{
		{Component: "input_tokens", USDPerMillion: "3"},
		{Component: "output_tokens", USDPerMillion: "12"},
	}
	rd.AllowUnscored = true
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("runtime facts still admit: %v", err)
	}
	if dec.Selected == nil {
		t.Fatalf("expected unscored admission: %+v", dec.Rejections)
	}
	// Estimate must be unscored (profile discarded), disclosed by warning.
	if dec.Selected.Estimate.Known {
		t.Fatalf("estimate must be unknown after profile discard: %+v", dec.Selected.Estimate)
	}
	if !containsStr(dec.Selected.Warnings, "revision_conflict") {
		t.Fatalf("selection must disclose revision_conflict: %v", dec.Selected.Warnings)
	}
	// Snapshot prices are discarded too: cost unknown with warning.
	// Runtime prices supply the cost (9000); the discarded snapshot quote
	// (7500) must not bleed through.
	if !dec.Selected.Cost.Known || dec.Selected.Cost.UpperMicros != 9000 {
		t.Fatalf("runtime cost must be 9000 after profile discard: %+v", dec.Selected.Cost)
	}
}

func TestRouteUnmappedOfferUnscored(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = append(rd.Runtime, RuntimeOfferData{
		ID: "prov/gamma", Provider: "prov", UpstreamID: "gamma", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1"},
			{Component: "output_tokens", USDPerMillion: "2"},
		},
		Region: "us-east", Privacy: "standard",
	})
	rd.AllowUnscored = true
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	// Cost mode: both costs known; gamma (2000 micros) undercuts alpha.
	if dec.Selected.OfferID != "prov/gamma" {
		t.Fatalf("cheapest known cost must win: %+v", dec.Selected)
	}
	if len(dec.Alternatives) != 1 || dec.Alternatives[0].OfferID != "prov/alpha" {
		t.Fatalf("alternatives wrong: %+v", dec.Alternatives)
	}
	if dec.Selected.Estimate.Known || dec.Selected.Estimate.Basis != "unscored" {
		t.Fatalf("unmapped offer must be unscored: %+v", dec.Selected.Estimate)
	}
	// Unscored without AllowUnscored is rejected.
	rd2 := fixtureRequestData()
	rd2.Runtime = rd.Runtime[:len(rd.Runtime)-1]
	rd2.Runtime = []RuntimeOfferData{rd.Runtime[len(rd.Runtime)-1]}
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("unscored without allow_unscored must reject")
	}
	if len(dec2.Rejections) != 1 || !containsStr(dec2.Rejections[0].Reasons, "estimate_unknown") {
		t.Fatalf("want estimate_unknown: %+v", dec2.Rejections)
	}
}

func TestRouteMinimumEstimate(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.MinimumEstimate = 6000 // alpha scores 5966
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("alpha below minimum must reject")
	}
	want := []string{"estimate_below_minimum"}
	if !reflect.DeepEqual(dec.Rejections[0].Reasons, want) {
		t.Fatalf("want %v, got %v", want, dec.Rejections[0].Reasons)
	}
	// Unknown cannot pass a positive minimum even with AllowUnscored.
	rd2 := fixtureRequestData()
	rd2.MinimumEstimate = 1
	rd2.AllowUnscored = true
	rd2.Runtime = []RuntimeOfferData{{
		ID: "prov/gamma", Provider: "prov", UpstreamID: "gamma", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1"},
			{Component: "output_tokens", USDPerMillion: "1"},
		},
		Region: "us-east", Privacy: "standard",
	}}
	_, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("unknown estimate cannot pass positive minimum")
	}
}

func TestRouteSignalDedupAndRevisionExclusion(t *testing.T) {
	sd := fixtureSnapshotData()
	// Duplicate evidence/task: higher-value twin must not double-count.
	sd.Offers[0].Signals = append(sd.Offers[0].Signals,
		SignalData{ID: "sig-code-2", Task: "code", Value: 10000, Weight: 100, Basis: "published", EvidenceID: "ev-1", Revision: "r1"},
		// Same evidence, different task: allowed, general is a separate task.
		SignalData{ID: "sig-gen-2", Task: "general", Value: 100, Weight: 100, Basis: "published", EvidenceID: "ev-2", Revision: "r1"})
	rd := fixtureRequestData()
	rd.Task = "code"
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	// code: sig-code counts; sig-code-2 (same evidence/task, larger ID)
	// drops. general: "sig-gen-2" < "sig-general" bytewise, so sig-gen-2
	// (100*100) wins and sig-general drops.
	// floor((248000 + 10000)/(40+100)) = floor(1842.86) = 1842.
	if dec.Selected.Estimate.Value != 1842 {
		t.Fatalf("dedup estimate wrong: %+v", dec.Selected.Estimate)
	}
	if !reflect.DeepEqual(dec.Selected.Estimate.EvidenceIDs, []string{"ev-1", "ev-2"}) {
		t.Fatalf("evidence ids wrong: %v", dec.Selected.Estimate.EvidenceIDs)
	}
}

func TestRouteSignalRevisionExclusionDisclosed(t *testing.T) {
	sd := fixtureSnapshotData()
	// Offer revisions are validated at construction, so a stale signal
	// revision can only exist on an unresolved offer (beta). The host
	// asserts r1; the signal pinned to rX is excluded with a warning.
	sd.Offers[1].Signals[0].Revision = "rX"
	rd := fixtureRequestData()
	rd.AllowUnscored = true
	beta := betaRuntime()
	beta.Revision = "r1"
	rd.Runtime = append(rd.Runtime, beta)
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	var bsel *Selection
	if dec.Selected.OfferID == "prov/beta" {
		bsel = dec.Selected
	} else {
		for i := range dec.Alternatives {
			if dec.Alternatives[i].OfferID == "prov/beta" {
				bsel = &dec.Alternatives[i]
			}
		}
	}
	if bsel == nil {
		t.Fatalf("beta must be admitted unscored: %+v", dec.Rejections)
	}
	if bsel.Estimate.Known {
		t.Fatalf("beta signal must be excluded: %+v", bsel.Estimate)
	}
	if !containsStr(bsel.Estimate.Warnings, "revision_conflict") {
		t.Fatalf("exclusion must be disclosed: %v", bsel.Estimate.Warnings)
	}
}

func TestRouteProvisionalAssociation(t *testing.T) {
	// beta is unresolved (no offer revision); its signal association is
	// provisional even with an assertive host.
	sd := fixtureSnapshotData()
	rd2 := fixtureRequestData()
	rd2.Runtime = append(rd2.Runtime, betaRuntime())
	rd2.InputTokens = 300000 // alpha rejected; beta admitted
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 != nil {
		t.Fatal(err2)
	}
	if dec2.Selected.OfferID != "prov/beta" {
		t.Fatalf("want beta: %+v", dec2.Selected)
	}
	// beta signal has empty revision on an unresolved offer: provisional.
	if !containsStr(dec2.Selected.Estimate.Warnings, "provisional_association") {
		t.Fatalf("beta estimate must be provisional: %+v", dec2.Selected.Estimate)
	}
}

func TestRouteBudgetAndStrictCost(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.BudgetMicros = 7000 // alpha cost 7500
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("known budget is a hard cap even non-strict")
	}
	if !containsStr(dec.Rejections[0].Reasons, "budget_exceeded") {
		t.Fatalf("want budget_exceeded: %v", dec.Rejections[0].Reasons)
	}
	// Strict cost rejects unknown (beta) outright.
	rd2 := fixtureRequestData()
	rd2.Runtime = append(rd2.Runtime, betaRuntime())
	rd2.InputTokens = 300000 // alpha limit-rejected
	rd2.StrictCost = true
	_, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("strict cost must reject unknown cost")
	}
}

func TestRouteZeroBudget(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.BudgetKnown = true
	rd.BudgetMicros = 0 // true zero budget: any positive known cost exceeds
	_, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("zero budget must reject positive costs")
	}
}

func TestRouteFreshness(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.MaxAgeSeconds = 100000 // ~27.8h; alpha confirmed 2026-09-29 08:30
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.OfferID != "prov/alpha" {
		t.Fatalf("fresh alpha expected: %+v", dec.Selected)
	}
	// Effective 2026-10-30: age 31 days exceeds.
	rd2 := fixtureRequestData()
	rd2.MaxAgeSeconds = 86400
	rd2.EffectiveAt = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	_, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("stale confirmation must reject")
	}
	// Future confirmation.
	rd3 := fixtureRequestData()
	rd3.MaxAgeSeconds = 86400
	rd3.EffectiveAt = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	_, err3 := routeFixture(t, sd, rd3)
	if err3 == nil {
		t.Fatal("future confirmation must reject")
	}
}

func TestRouteLatencyStrict(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.DeadlineMillis = 800
	rd.StrictLatency = true // alpha latency 900 exceeds; beta unknown
	_, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("strict latency must reject")
	}
	// Non-strict excess is disclosed, not rejected.
	rd2 := fixtureRequestData()
	rd2.DeadlineMillis = 800
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 != nil {
		t.Fatalf("non-strict excess must admit: %v", err2)
	}
	if dec2.Selected.LatencyMillis != 900 {
		t.Fatalf("latency must be disclosed: %d", dec2.Selected.LatencyMillis)
	}
}

func TestRouteImageOperationAdmission(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.BudgetMicros = 200000000
	rd.ImageOperation = "generate"
	rd.ImageCount, rd.ImageWidth, rd.ImageHeight, rd.ImageBytes = 2, 1024, 768, 4000000
	rd.Runtime = []RuntimeOfferData{{
		ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha", Eligible: true,
		Capabilities: []string{"text_input", "text_output", "image_output"},
		Region:       "us-east", Privacy: "standard",
	}}
	rd.RequiredCapabilities = []string{"text_input", "text_output"}
	// No image limits or prices: rejected for unknown bounds and missing
	// image quotes.
	_, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("generate without image limits/prices must reject")
	}
	// Supply host limits and prices: admitted.
	rd.Runtime[0].Limits = []LimitData{
		{Resource: "input_tokens", Maximum: 200000},
		{Resource: "output_tokens", Maximum: 64000},
		{Resource: "image_width", Maximum: 2048},
		{Resource: "image_height", Maximum: 2048},
		{Resource: "image_bytes", Maximum: 8000000},
		{Resource: "output_images", Maximum: 4},
	}
	rd.Runtime[0].Prices = []PriceData{
		{Component: "input_tokens", USDPerMillion: "2.50"},
		{Component: "output_tokens", USDPerMillion: "10"},
		{Component: "output_images", USDPerMillion: "8000000"},
	}
	dec, err2 := routeFixture(t, sd, rd)
	if err2 != nil {
		t.Fatalf("complete host facts must admit: %v %v", err2, dec.Rejections)
	}
	// Cost: 1000*2.5 + 500*10 + 2*8000000 micros = 16007500.
	if dec.Selected.Cost.UpperMicros != 16007500 {
		t.Fatalf("image cost wrong: %+v", dec.Selected.Cost)
	}
}

func TestRouteCostRoundingUp(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	// 1 input token at 0.000001 USD/M = 0.000001 micro; plus output 0.
	rd.InputTokens, rd.OutputTokens = 1, 0
	rd.Runtime = []RuntimeOfferData{{
		ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "0.000001"},
			{Component: "output_tokens", USDPerMillion: "0"},
		},
		Region: "us-east", Privacy: "standard",
	}}
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if !dec.Selected.Cost.Known || dec.Selected.Cost.UpperMicros != 1 {
		t.Fatalf("fractional micros must round up to 1: %+v", dec.Selected.Cost)
	}
}

func TestRoutePerformanceOrdering(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Mode = "performance"
	rd.AllowUnscored = true
	rd.Runtime = []RuntimeOfferData{
		{ // unmapped, cheap, complete quotes, unscored
			ID: "prov/aaa", Provider: "prov", UpstreamID: "aaa", Eligible: true,
			Capabilities: []string{"text_input", "text_output"},
			Limits: []LimitData{
				{Resource: "input_tokens", Maximum: 200000},
				{Resource: "output_tokens", Maximum: 64000},
			},
			Prices: []PriceData{
				{Component: "input_tokens", USDPerMillion: "0.001"},
				{Component: "output_tokens", USDPerMillion: "0.001"},
			},
			Region: "us-east", Privacy: "standard",
		},
	}
	rd.Runtime = append(rd.Runtime, fixtureRequestData().Runtime...)
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	// Performance mode: known estimate (alpha) precedes unscored cheap aaa.
	if dec.Selected.OfferID != "prov/alpha" {
		t.Fatalf("performance mode must prefer known estimate: %+v", dec.Selected)
	}
	if len(dec.Alternatives) != 1 || dec.Alternatives[0].OfferID != "prov/aaa" {
		t.Fatalf("alternatives wrong: %+v", dec.Alternatives)
	}
	// Cost mode: aaa's known cost (ceil(1.5)=2 micros) undercuts alpha.
	rd.Mode = "cost"
	dec2, err2 := routeFixture(t, sd, rd)
	if err2 != nil {
		t.Fatal(err2)
	}
	if dec2.Selected.OfferID != "prov/aaa" {
		t.Fatalf("cost mode must prefer cheapest known: %+v", dec2.Selected)
	}
}

func TestRouteSnapshotOnlyOfferNeverCandidate(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = []RuntimeOfferData{} // empty registry: nothing admissible
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("empty runtime must yield no candidates")
	}
	if len(dec.Rejections) != 0 || dec.Selected != nil {
		t.Fatalf("snapshot-only offers must not appear: %+v", dec)
	}
}

func TestRouteCapabilityUnionAndNegatives(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	// Host drops tool_use but keeps text; snapshot image_input on beta is
	// irrelevant. Required text satisfied by union.
	rd.Runtime[0].Capabilities = []string{"text_input"} // snapshot adds text_output, tool_use
	rd.Runtime[0].UnsupportedCapabilities = []string{"tool_use"}
	rd.RequiredCapabilities = []string{"text_input", "text_output"}
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("union should satisfy: %v", err)
	}
	_ = dec
	// Host negative overrides snapshot positive.
	rd2 := fixtureRequestData()
	rd2.Runtime[0].Capabilities = []string{}
	rd2.Runtime[0].UnsupportedCapabilities = []string{"text_output"}
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("host negative must reject required capability")
	}
	if !containsStr(rejectionFor(dec2, "prov/alpha"), "capability_missing:text_output") {
		t.Fatalf("expected capability_missing:text_output: %+v", dec2.Rejections)
	}
}

func rejectionFor(dec Decision, id string) []string {
	for _, r := range dec.Rejections {
		if r.OfferID == id {
			return r.Reasons
		}
	}
	return nil
}

func TestRouteRuntimeOverridesSnapshot(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	// Runtime halves the input price: 1000*1.25 + 500*10 = 6250.
	rd.Runtime[0].Prices = []PriceData{{Component: "input_tokens", USDPerMillion: "1.25"}}
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.Cost.UpperMicros != 6250 {
		t.Fatalf("runtime price must override: %+v", dec.Selected.Cost)
	}
	// Runtime limit overrides snapshot (tighter input bound rejects).
	rd2 := fixtureRequestData()
	rd2.Runtime[0].Limits = []LimitData{{Resource: "input_tokens", Maximum: 999}}
	_, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("runtime limit must override and reject")
	}
}

func TestRouteWarningsDedupSorted(t *testing.T) {
	sd := fixtureSnapshotData()
	// Two signals with the same warning token.
	sd.Offers[0].Signals = append(sd.Offers[0].Signals,
		SignalData{ID: "sig-w", Task: "code", Value: 1000, Weight: 10, Basis: "published", EvidenceID: "ev-w", Revision: "r1", Warnings: []string{"mixed_protocol"}})
	rd := fixtureRequestData()
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	w := dec.Selected.Estimate.Warnings
	if !reflect.DeepEqual(w, []string{"mixed_protocol", "provisional_association"}) {
		t.Fatalf("warnings must dedup sorted: %v", w)
	}
}

func TestRouteContextBoundSatisfiesTokens(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = []RuntimeOfferData{betaRuntime()} // context-only bounds
	rd.InputTokens = 400000                        // no own bound; context covers
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("context bound should satisfy: %v %+v", err, dec.Rejections)
	}
	if dec.Selected.OfferID != "prov/beta" {
		t.Fatalf("want beta: %+v", dec.Selected)
	}
	// Context bound also rejects when the sum exceeds.
	rd2 := fixtureRequestData()
	rd2.Runtime = []RuntimeOfferData{betaRuntime()}
	rd2.InputTokens = 900000
	rd2.OutputTokens = 200000
	_, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("context exceedance must reject")
	}
}
