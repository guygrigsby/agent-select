package routing

import (
	"fmt"
	"reflect"
	"testing"
)

// Disclosure tests pinning the accepted educated-routing amendments:
// output warning lists are the complete deduplicated union (no 64-entry
// truncation), the unknown-estimate gate reports estimate_unknown, and
// operation-derived capabilities include the text roles. All tests drive
// the public NewSnapshot -> NewRequest -> Engine.Route path.

// emptySnapshotData is a synthetic owned snapshot with no inputs and no
// offers (educated-routing.md permits fixtures with no inputs).
func emptySnapshotData() SnapshotData {
	return SnapshotData{
		Schema:          1,
		CreatedAt:       fixtureTime,
		KnowledgeCutoff: fixtureTime,
		Policy:          "educated-v1",
		Distribution:    "local",
	}
}

// TestRouteAggregateWarningsNotTruncated pins amendment 1: 70 retained
// signals each carrying one distinct contract-valid warning token produce
// 70 aggregate warnings; the output estimate warning list must keep all of
// them (no truncation to the 64-entry input bound) and every evidence ID
// must be preserved.
func TestRouteAggregateWarningsNotTruncated(t *testing.T) {
	sd := fixtureSnapshotData()
	signals := make([]SignalData, 70)
	for i := range signals {
		signals[i] = SignalData{
			ID:         fmt.Sprintf("sig-%03d", i),
			Task:       "code",
			Value:      1000,
			Weight:     1,
			Basis:      "published",
			EvidenceID: fmt.Sprintf("ev-%03d", i),
			Revision:   "r1",
			Warnings:   []string{fmt.Sprintf("w%02d", i)},
		}
	}
	sd.Offers = []OfferData{{
		ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha",
		ModelID: "alpha-2026-09", Revision: "r1",
		Capabilities: []string{"text_input", "text_output"},
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1"},
			{Component: "output_tokens", USDPerMillion: "1"},
		},
		Signals: signals,
	}}
	rd := fixtureRequestData()
	rd.Runtime = []RuntimeOfferData{{
		ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha",
		Revision: "r1", Eligible: true,
		Capabilities: []string{"text_input", "text_output"},
		Region:       "us-east", Privacy: "standard",
	}}
	dec, err := routeFixture(t, sd, rd)
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	if dec.Selected == nil {
		t.Fatalf("no selection: %+v", dec.Rejections)
	}
	w := dec.Selected.Estimate.Warnings
	if len(w) != 70 {
		t.Fatalf("aggregate warnings must not be truncated to 64: got %d: %v", len(w), w)
	}
	for i := 0; i < 70; i++ {
		if !containsStr(w, fmt.Sprintf("w%02d", i)) {
			t.Fatalf("warning w%02d dropped: %v", i, w)
		}
	}
	if !reflect.DeepEqual(w, func() []string {
		out := make([]string, 70)
		for i := range out {
			out[i] = fmt.Sprintf("w%02d", i)
		}
		return out
	}()) {
		t.Fatalf("warnings must be the deduplicated sorted union: %v", w)
	}
	if len(dec.Selected.Estimate.EvidenceIDs) != 70 {
		t.Fatalf("all 70 evidence IDs must be preserved: got %d", len(dec.Selected.Estimate.EvidenceIDs))
	}
	for i := 0; i < 70; i++ {
		if !containsStr(dec.Selected.Estimate.EvidenceIDs, fmt.Sprintf("ev-%03d", i)) {
			t.Fatalf("evidence ID ev-%03d dropped", i)
		}
	}
	if !reflect.DeepEqual(dec.Warnings, []string{}) {
		t.Fatalf("decision warnings must be an empty list: %v", dec.Warnings)
	}
}

// TestRouteTextTaskRejectsImageOnlyModel pins amendment 3: with
// ImageOperation none and no RequiredCapabilities, text_input and
// text_output are operation-derived hard requirements, so an image-only
// runtime offer must be rejected with capability_missing codes for both
// text roles instead of being silently selected.
func TestRouteTextTaskRejectsImageOnlyModel(t *testing.T) {
	sd := emptySnapshotData()
	rd := RequestData{
		ID: "req-f3", Task: "code", Mode: "cost",
		EffectiveAt: fixtureTime, RuntimeGeneration: "rg-f3",
		Runtime: []RuntimeOfferData{{
			ID: "prov/img", Provider: "prov", UpstreamID: "img", Eligible: true,
			Capabilities: []string{"image_input", "image_output"},
			Limits: []LimitData{
				{Resource: "input_tokens", Maximum: 200000},
				{Resource: "output_tokens", Maximum: 64000},
			},
			Prices: []PriceData{
				{Component: "input_tokens", USDPerMillion: "1"},
				{Component: "output_tokens", USDPerMillion: "1"},
			},
			Region: "us-east", Privacy: "standard",
		}},
		InputTokens:    1000,
		OutputTokens:   500,
		ImageOperation: "none",
		BudgetKnown:    true,
		BudgetMicros:   5000000,
		AllowUnscored:  true,
	}
	dec, err := routeFixture(t, sd, rd)
	if err == nil || dec.Selected != nil {
		t.Fatalf("image-only model must not serve a text task: %+v %v", dec.Selected, err)
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeNoAdmissibleOffer {
		t.Fatalf("want no_admissible_offer, got %v", err)
	}
	reasons := rejectionFor(dec, "prov/img")
	if !containsStr(reasons, "capability_missing:text_input") {
		t.Fatalf("want capability_missing:text_input: %v", reasons)
	}
	if !containsStr(reasons, "capability_missing:text_output") {
		t.Fatalf("want capability_missing:text_output: %v", reasons)
	}
	if containsStr(reasons, "estimate_below_minimum") {
		t.Fatalf("unexpected estimate code: %v", reasons)
	}
}

// TestRouteUnknownEstimatePositiveMinimumCode pins amendment 2: an unknown
// estimate with a positive MinimumEstimate is rejected as estimate_unknown,
// never estimate_below_minimum (the zero sentinel is not a low score).
func TestRouteUnknownEstimatePositiveMinimumCode(t *testing.T) {
	sd := emptySnapshotData()
	rd := RequestData{
		ID: "req-f2", Task: "code", Mode: "cost",
		EffectiveAt: fixtureTime, RuntimeGeneration: "rg-f2",
		Runtime: []RuntimeOfferData{{
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
		}},
		InputTokens:     1000,
		OutputTokens:    500,
		ImageOperation:  "none",
		BudgetKnown:     true,
		BudgetMicros:    5000000,
		AllowUnscored:   true,
		MinimumEstimate: 1,
	}
	dec, err := routeFixture(t, sd, rd)
	if err == nil || dec.Selected != nil {
		t.Fatalf("unknown estimate must not pass a positive minimum: %+v %v", dec.Selected, err)
	}
	reasons := rejectionFor(dec, "prov/gamma")
	if !containsStr(reasons, "estimate_unknown") {
		t.Fatalf("want estimate_unknown: %v", reasons)
	}
	if containsStr(reasons, "estimate_below_minimum") {
		t.Fatalf("unknown sentinel must not be labeled estimate_below_minimum: %v", reasons)
	}
}

// imageRuntime returns one runtime image offer with complete limits and
// prices so that admission turns only on capabilities and text roles.
func imageRuntime(id string, caps ...string) RuntimeOfferData {
	return RuntimeOfferData{
		ID: "prov/" + id, Provider: "prov", UpstreamID: id, Eligible: true,
		Capabilities: caps,
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
			{Resource: "image_width", Maximum: 2048},
			{Resource: "image_height", Maximum: 2048},
			{Resource: "image_bytes", Maximum: 8000000},
			{Resource: "input_images", Maximum: 4},
			{Resource: "output_images", Maximum: 4},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "1"},
			{Component: "output_tokens", USDPerMillion: "1"},
			{Component: "input_images", USDPerMillion: "1000000"},
			{Component: "output_images", USDPerMillion: "1000000"},
		},
		Region: "us-east", Privacy: "standard",
	}
}

// imageRequest builds one image-operation request against a runtime offer.
func imageRequest(id, op, task string, inTok, outTok int64, caps ...string) RequestData {
	return RequestData{
		ID: "req-img-" + id, Task: task, Mode: "cost",
		EffectiveAt: fixtureTime, RuntimeGeneration: "rg-img",
		Runtime:        []RuntimeOfferData{imageRuntime(id, caps...)},
		InputTokens:    inTok,
		OutputTokens:   outTok,
		ImageOperation: op,
		ImageCount:     2,
		ImageWidth:     1024,
		ImageHeight:    768,
		ImageBytes:     4000000,
		BudgetKnown:    true,
		BudgetMicros:   10000000000000,
		AllowUnscored:  true,
	}
}

// TestRouteOperationDerivedTextRoles pins amendment 3 for every operation:
// none and vision require text_input and text_output (vision additionally
// image_input); generate requires text_input and image_output; edit
// requires image_input, image_output and image_edit, plus text_input when
// InputTokens is positive; positive OutputTokens on generate/edit also
// requires text_output. Zero-token image edits stay admitted on image caps
// alone (bounds are not silently relaxed, and not over-constrained).
func TestRouteOperationDerivedTextRoles(t *testing.T) {
	sd := emptySnapshotData()
	cases := []struct {
		name   string
		op     string
		task   string
		inTok  int64
		outTok int64
		caps   []string
		want   string // "" means admitted
	}{
		{"generate image-only lacks text_input", "generate", "image_generation", 0, 0,
			[]string{"image_output"}, "capability_missing:text_input"},
		{"generate zero tokens no text_output needed", "generate", "image_generation", 0, 0,
			[]string{"text_input", "image_output"}, ""},
		{"generate positive output tokens needs text_output", "generate", "image_generation", 0, 500,
			[]string{"text_input", "image_output"}, "capability_missing:text_output"},
		{"generate full text roles admitted", "generate", "image_generation", 0, 500,
			[]string{"text_input", "text_output", "image_output"}, ""},
		{"edit image-only zero tokens admitted", "edit", "image_editing", 0, 0,
			[]string{"image_input", "image_output", "image_edit"}, ""},
		{"edit positive input tokens needs text_input", "edit", "image_editing", 1000, 0,
			[]string{"image_input", "image_output", "image_edit"}, "capability_missing:text_input"},
		{"edit positive output tokens needs text_output", "edit", "image_editing", 0, 500,
			[]string{"image_input", "image_output", "image_edit"}, "capability_missing:text_output"},
		{"edit full text roles admitted", "edit", "image_editing", 1000, 0,
			[]string{"text_input", "image_input", "image_output", "image_edit"}, ""},
		{"vision image-only lacks text roles", "vision", "vision", 0, 0,
			[]string{"image_input"}, "capability_missing:text_input"},
		{"vision with text roles admitted", "vision", "vision", 0, 0,
			[]string{"text_input", "text_output", "image_input"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rd := imageRequest(tc.op, tc.op, tc.task, tc.inTok, tc.outTok, tc.caps...)
			dec, err := routeFixture(t, sd, rd)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("want admitted, got %v (%+v)", err, dec.Rejections)
				}
				return
			}
			if err == nil || dec.Selected != nil {
				t.Fatalf("want rejected, got selected %+v %v", dec.Selected, err)
			}
			reasons := rejectionFor(dec, "prov/"+tc.op)
			if !containsStr(reasons, tc.want) {
				t.Fatalf("want %s in %v", tc.want, reasons)
			}
		})
	}
}

// TestDecodeAdoptRouteDeterministic pins the public decode/adopt path:
// routing through a decoded artifact and through Adopt must produce the
// same decision as routing through the directly constructed snapshot.
func TestDecodeAdoptRouteDeterministic(t *testing.T) {
	sd := fixtureSnapshotData()
	snap, err := NewSnapshot(sd)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := EncodeSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}
	rd := fixtureRequestData()
	req, err := NewRequest(rd)
	if err != nil {
		t.Fatal(err)
	}
	direct, err := NewEngine(snap).Route(req)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSnapshot(enc)
	if err != nil {
		t.Fatal(err)
	}
	viaDecode, err := NewEngine(decoded).Route(req)
	if err != nil {
		t.Fatal(err)
	}
	viaAdopt := NewEngine(nil)
	if err := viaAdopt.Adopt(enc); err != nil {
		t.Fatal(err)
	}
	if viaAdopt.ActiveHash() != snap.Hash() {
		t.Fatalf("adopted hash mismatch: %s vs %s", viaAdopt.ActiveHash(), snap.Hash())
	}
	adopted, err := viaAdopt.Route(req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(direct, viaDecode) {
		t.Fatalf("decoded snapshot decision diverges: %+v vs %+v", direct, viaDecode)
	}
	if !reflect.DeepEqual(direct, adopted) {
		t.Fatalf("adopted snapshot decision diverges: %+v vs %+v", direct, adopted)
	}
}

// TestRouteDecisionCopiedSurvivesAdopt pins step 6: a returned Decision is
// a detached copy that stays stable across a later generation swap.
func TestRouteDecisionCopiedSurvivesAdopt(t *testing.T) {
	sd := fixtureSnapshotData()
	snap, err := NewSnapshot(sd)
	if err != nil {
		t.Fatal(err)
	}
	// Generation B: same offer key, different signal values and model.
	sd2 := fixtureSnapshotData()
	sd2.Offers[0].ModelID = "alpha-changed"
	sd2.Offers[0].Signals[0].Value = 1000
	snap2, err := NewSnapshot(sd2)
	if err != nil {
		t.Fatal(err)
	}
	encB, err := EncodeSnapshot(snap2)
	if err != nil {
		t.Fatal(err)
	}
	req, err := NewRequest(fixtureRequestData())
	if err != nil {
		t.Fatal(err)
	}
	e := NewEngine(snap)
	before, err := e.Route(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Adopt(encB); err != nil {
		t.Fatal(err)
	}
	after, err := e.Route(req)
	if err != nil {
		t.Fatal(err)
	}
	// The old decision must be untouched by the swap.
	if before.SnapshotHash != snap.Hash() || before.SnapshotHash == after.SnapshotHash {
		t.Fatalf("snapshot hash pinned wrong: %s %s", before.SnapshotHash, after.SnapshotHash)
	}
	if before.Selected == nil || before.Selected.OfferID != "prov/alpha" ||
		before.Selected.Estimate.Value != 5966 || !before.Selected.Estimate.Known {
		t.Fatalf("pinned decision drifted: %+v", before.Selected)
	}
	if !reflect.DeepEqual(before.Selected.Estimate.Warnings, []string{"mixed_protocol", "provisional_association"}) {
		t.Fatalf("pinned estimate warnings drifted: %v", before.Selected.Estimate.Warnings)
	}
	// The new generation routes with the new facts.
	if after.Selected == nil || after.Selected.Estimate.Value != 2500 {
		t.Fatalf("adopted generation must re-estimate: %+v", after.Selected)
	}
}
