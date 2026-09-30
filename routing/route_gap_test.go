package routing

import (
	"testing"
)

// Review round 1 gap coverage: cost arithmetic, ordering tiebreaks and
// rejection-code placements the main suite did not pin.

func TestGapCostMissingComponents(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = []RuntimeOfferData{betaRuntime()} // beta prices: input_tokens only
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.OfferID != "prov/beta" {
		t.Fatalf("want beta: %+v", dec.Selected)
	}
	if dec.Selected.Cost.Known || !equalStr(dec.Selected.Cost.Missing, []string{"output_tokens"}) {
		t.Fatalf("Missing must list output_tokens: %+v", dec.Selected.Cost)
	}
}

func equalStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGapCostRoundsUpOnceAfterSumming(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Runtime = []RuntimeOfferData{{
		ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1.2"},
			{Component: "output_tokens", USDPerMillion: "1.2"},
		},
		Region: "us-east", Privacy: "standard",
	}}
	rd.InputTokens, rd.OutputTokens = 1, 1
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	// 1.2 + 1.2 = 2.4 micros; per-component rounding would give 2+2 = 4.
	if !dec.Selected.Cost.Known || dec.Selected.Cost.UpperMicros != 3 {
		t.Fatalf("sum-then-round-once must give 3: %+v", dec.Selected.Cost)
	}
}

func TestGapOrderingTiebreaks(t *testing.T) {
	sd := fixtureSnapshotData()
	// Two identical unmapped offers differing only by ID and latency.
	mk := func(up string, lat int64) OfferData {
		return OfferData{
			ID: "prov/" + up, Provider: "prov", UpstreamID: up,
			Capabilities: []string{"text_input", "text_output"},
			Limits: []LimitData{
				{Resource: "input_tokens", Maximum: 200000},
				{Resource: "output_tokens", Maximum: 64000},
			},
			Prices: []PriceData{
				{Component: "input_tokens", USDPerMillion: "1"},
				{Component: "output_tokens", USDPerMillion: "1"},
			},
			LatencyMillis: lat,
			ConfirmedAt:   fixtureTime,
		}
	}
	sd.Offers = []OfferData{mk("zzz", 0), mk("aaa", 500)}
	rd := fixtureRequestData()
	rd.AllowUnscored = true
	rd.Runtime = []RuntimeOfferData{
		{ID: "prov/zzz", Provider: "prov", UpstreamID: "zzz", Eligible: true,
			Capabilities: []string{"text_input", "text_output"}, Region: "us-east", Privacy: "standard"},
		{ID: "prov/aaa", Provider: "prov", UpstreamID: "aaa", Eligible: true,
			Capabilities: []string{"text_input", "text_output"}, Region: "us-east", Privacy: "standard"},
	}
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.OfferID != "prov/aaa" {
		t.Fatalf("known latency must precede unknown: %+v", dec.Selected)
	}
	// Strip both latencies: bytewise ID tiebreak.
	sd.Offers[1].LatencyMillis = 0
	dec, err = routeFixture(t, sd, rd)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.OfferID != "prov/aaa" || dec.Alternatives[0].OfferID != "prov/zzz" {
		t.Fatalf("ID tiebreak wrong: %s then %v", dec.Selected.OfferID, dec.Alternatives)
	}
}

func TestGapFreshnessAndLatencyUnknownCodes(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.MaxAgeSeconds = 100000
	// Unmapped runtime offer: no confirmation, no latency.
	rd.Runtime = []RuntimeOfferData{{
		ID: "prov/gamma", Provider: "prov", UpstreamID: "gamma", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1"},
			{Component: "output_tokens", USDPerMillion: "1"},
		},
		Region: "us-east", Privacy: "standard",
	}}
	rd.AllowUnscored = true
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("unmapped offer with freshness bound must reject")
	}
	if !containsStr(dec.Rejections[0].Reasons, "freshness_unknown") {
		t.Fatalf("want freshness_unknown: %v", dec.Rejections[0].Reasons)
	}
	// Strict latency rejects unknown latency.
	rd2 := rd
	rd2.MaxAgeSeconds = 0
	rd2.DeadlineMillis = 1000
	rd2.StrictLatency = true
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("strict latency must reject unknown latency")
	}
	if !containsStr(dec2.Rejections[0].Reasons, "latency_unknown") {
		t.Fatalf("want latency_unknown: %v", dec2.Rejections[0].Reasons)
	}
}

func TestGapRegionCodes(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.Region = "eu-west"
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("region mismatch must reject")
	}
	if !containsStr(dec.Rejections[0].Reasons, "region_mismatch") {
		t.Fatalf("want region_mismatch: %v", dec.Rejections[0].Reasons)
	}
	rd2 := fixtureRequestData()
	rd2.Runtime[0].Region = ""
	rd2.Region = "us-east"
	dec2, err2 := routeFixture(t, sd, rd2)
	if err2 == nil {
		t.Fatal("unknown region must fail closed")
	}
	if !containsStr(dec2.Rejections[0].Reasons, "region_unknown") {
		t.Fatalf("want region_unknown: %v", dec2.Rejections[0].Reasons)
	}
}

func TestGapStrictCostOverBudgetCodeOrder(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.StrictCost = true
	rd.BudgetMicros = 1000 // known cost 7500 exceeds: budget_exceeded, not cost_unknown
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("over budget must reject")
	}
	if !containsStr(dec.Rejections[0].Reasons, "budget_exceeded") {
		t.Fatalf("strict over-budget must be budget_exceeded: %v", dec.Rejections[0].Reasons)
	}
	if containsStr(dec.Rejections[0].Reasons, "cost_unknown") {
		t.Fatalf("known cost must not also report cost_unknown: %v", dec.Rejections[0].Reasons)
	}
}

func TestGapMinimumEstimateInclusiveBoundary(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	rd.MinimumEstimate = 5966 // exactly the alpha estimate: admitted
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("minimum is inclusive: %v", err)
	}
	if dec.Selected.OfferID != "prov/alpha" {
		t.Fatalf("want alpha: %+v", dec.Selected)
	}
}

func TestGapContextSumAtMaxQuantities(t *testing.T) {
	sd := fixtureSnapshotData()
	rd := fixtureRequestData()
	// 2^53-1 each: the sum is 2^54-2, which must not wrap. Beta's
	// context_tokens bound (1e6) must be exceeded by the true sum.
	rd.InputTokens = 9007199254740991
	rd.OutputTokens = 9007199254740991
	rd.Runtime = []RuntimeOfferData{betaRuntime()}
	dec, err := routeFixture(t, sd, rd)
	if err == nil {
		t.Fatal("max quantities must exceed beta context bound")
	}
	if !containsStr(dec.Rejections[0].Reasons, "limit_exceeded:context_tokens") {
		t.Fatalf("want limit_exceeded:context_tokens (no wraparound): %v", dec.Rejections[0].Reasons)
	}
}
