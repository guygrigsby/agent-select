package routing

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func fixtureRequestData() RequestData {
	return RequestData{
		ID: "req-1", Task: "code", Mode: "cost",
		EffectiveAt: fixtureTime, RuntimeGeneration: "rg-1",
		Runtime: []RuntimeOfferData{
			{
				ID: "prov/alpha", Provider: "prov", UpstreamID: "alpha", Eligible: true,
				Capabilities: []string{"text_input", "text_output", "tool_use"},
				Region:       "us-east", Privacy: "standard",
			},
		},
		RequiredCapabilities: []string{"text_input", "text_output"},
		InputTokens:          1000, OutputTokens: 500,
		ImageOperation: "none",
		BudgetKnown:    true, BudgetMicros: 5000000,
	}
}

func TestNewRequestAcceptsFixture(t *testing.T) {
	if _, err := NewRequest(fixtureRequestData()); err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
}

func requestWith(mutate func(*RequestData)) RequestData {
	d := fixtureRequestData()
	mutate(&d)
	return d
}

func TestNewRequestRejectsInvalid(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RequestData)
		want   string
	}{
		{"empty id", func(d *RequestData) { d.ID = "" }, "empty"},
		{"unknown task", func(d *RequestData) { d.Task = "daydream" }, "unknown task"},
		{"unknown mode", func(d *RequestData) { d.Mode = "balanced" }, "unknown mode"},
		{"zero time", func(d *RequestData) { d.EffectiveAt = time.Time{} }, "effective_at"},
		{"non-utc time", func(d *RequestData) {
			d.EffectiveAt = fixtureTime.In(time.FixedZone("X", 3600))
		}, "not UTC"},
		{"empty runtime generation", func(d *RequestData) { d.RuntimeGeneration = "" }, "empty"},
		{"unknown required capability", func(d *RequestData) {
			d.RequiredCapabilities = append(d.RequiredCapabilities, "flight")
		}, "unknown"},
		{"duplicate required capability", func(d *RequestData) {
			d.RequiredCapabilities = append(d.RequiredCapabilities, "text_input")
		}, "duplicate"},
		{"negative input tokens", func(d *RequestData) { d.InputTokens = -1 }, "negative"},
		{"negative output tokens", func(d *RequestData) { d.OutputTokens = -1 }, "negative"},
		{"negative image bytes", func(d *RequestData) { d.ImageBytes = -1 }, "negative"},
		{"negative references", func(d *RequestData) { d.ReferenceCount = -1 }, "negative"},
		{"negative searches", func(d *RequestData) { d.Searches = -1 }, "negative"},
		{"negative budget", func(d *RequestData) { d.BudgetKnown = true; d.BudgetMicros = -1 }, "negative"},
		{"negative deadline", func(d *RequestData) { d.DeadlineMillis = -1 }, "negative"},
		{"negative max age", func(d *RequestData) { d.MaxAgeSeconds = -1 }, "negative"},
		{"budget value without known", func(d *RequestData) {
			d.BudgetKnown = false
		}, "requires budget_known"},
		{"zero budget allowed", func(d *RequestData) {
			d.BudgetMicros = 0
		}, ""},
		{"strict cost without budget", func(d *RequestData) {
			d.BudgetKnown = false
			d.BudgetMicros = 0
			d.StrictCost = true
		}, "requires a known budget"},
		{"strict latency without deadline", func(d *RequestData) {
			d.DeadlineMillis = 0
			d.StrictLatency = true
		}, "strict_latency false"},
		{"deadline without strict latency ok", func(d *RequestData) { d.DeadlineMillis = 5000 }, ""},
		{"minimum estimate high", func(d *RequestData) { d.MinimumEstimate = 10001 }, "0..10000"},
		{"minimum estimate negative", func(d *RequestData) { d.MinimumEstimate = -1 }, "0..10000"},
		{"unknown image operation", func(d *RequestData) { d.ImageOperation = "rotate" }, "unknown image operation"},
		{"none with image count", func(d *RequestData) { d.ImageCount = 1 }, "quantities zero"},
		{"none with mask", func(d *RequestData) { d.Mask = true }, "mask flags false"},
		{"vision without bytes", func(d *RequestData) {
			d.ImageOperation = "vision"
			d.ImageCount, d.ImageWidth, d.ImageHeight = 1, 64, 64
		}, "positive"},
		{"generate zero width", func(d *RequestData) {
			d.ImageOperation = "generate"
			d.ImageCount, d.ImageHeight, d.ImageBytes = 1, 64, 1000
		}, "positive"},
		{"edit ok", func(d *RequestData) {
			d.ImageOperation = "edit"
			d.ImageCount, d.ImageWidth, d.ImageHeight, d.ImageBytes = 1, 64, 64, 1000
		}, ""},
		{"mask without edit", func(d *RequestData) {
			d.ImageOperation = "generate"
			d.ImageCount, d.ImageWidth, d.ImageHeight, d.ImageBytes = 1, 64, 64, 1000
			d.Mask = true
		}, "edit operation"},
		{"alpha without mask", func(d *RequestData) {
			d.ImageOperation = "edit"
			d.ImageCount, d.ImageWidth, d.ImageHeight, d.ImageBytes = 1, 64, 64, 1000
			d.MaskAlpha = true
		}, "requires mask"},
		{"mask with edit ok", func(d *RequestData) {
			d.ImageOperation = "edit"
			d.ImageCount, d.ImageWidth, d.ImageHeight, d.ImageBytes = 1, 64, 64, 1000
			d.Mask, d.MaskAlpha = true, true
		}, ""},
		{"runtime id mismatch", func(d *RequestData) { d.Runtime[0].ID = "wrong" }, "id must equal"},
		{"runtime duplicate", func(d *RequestData) {
			d.Runtime = append(d.Runtime, d.Runtime[0])
		}, "duplicate runtime"},
		{"runtime unknown capability", func(d *RequestData) {
			d.Runtime[0].Capabilities = append(d.Runtime[0].Capabilities, "magic")
		}, "unknown"},
		{"runtime lists overlap", func(d *RequestData) {
			d.Runtime[0].UnsupportedCapabilities = []string{"tool_use"}
		}, "both"},
		{"runtime unsupported unknown token", func(d *RequestData) {
			d.Runtime[0].UnsupportedCapabilities = []string{"magic"}
		}, "unknown"},
		{"runtime bad limit", func(d *RequestData) {
			d.Runtime[0].Limits = []LimitData{{Resource: "input_tokens", Maximum: 0}}
		}, "positive"},
		{"runtime bad price", func(d *RequestData) {
			d.Runtime[0].Prices = []PriceData{{Component: "input_tokens", USDPerMillion: "-1"}}
		}, "decimal syntax"},
		{"allowed provider empty", func(d *RequestData) { d.AllowedProviders = []string{""} }, "empty"},
		{"allowed provider duplicate", func(d *RequestData) {
			d.AllowedProviders = []string{"prov", "prov"}
		}, "duplicate"},
	}
	for _, tc := range cases {
		_, err := NewRequest(requestWith(tc.mutate))
		if tc.want == "" {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tc.name, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%s: accepted", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.name, err.Error(), tc.want)
		}
	}
}

func TestRequestZeroValueRejectedByShape(t *testing.T) {
	// A zero-value Request is not validated; NewRequest on zero data must
	// fail, which is what Route will check in phase 11.
	if _, err := NewRequest(RequestData{}); err == nil {
		t.Fatal("zero RequestData accepted")
	}
}

func TestRequestDataImmutable(t *testing.T) {
	d := fixtureRequestData()
	r, err := NewRequest(d)
	if err != nil {
		t.Fatal(err)
	}
	out := r.Data()
	out.Runtime[0].Capabilities[0] = "image_input"
	out.RequiredCapabilities[0] = "image_input"
	if r.Data().Runtime[0].Capabilities[0] != "text_input" {
		t.Fatal("runtime capabilities not copied")
	}
	if r.Data().RequiredCapabilities[0] != "text_input" {
		t.Fatal("required capabilities not copied")
	}
	// Mutating the caller's original must not affect the request either.
	d.Runtime[0].Capabilities[0] = "image_input"
	if r.Data().Runtime[0].Capabilities[0] != "text_input" {
		t.Fatal("request aliases caller slices")
	}
}

func TestSnapshotDataImmutable(t *testing.T) {
	orig := fixtureSnapshotData()
	s := mustSnapshot(t)
	out := s.Data()
	out.Offers[0].Capabilities[0] = "image_output"
	out.Offers[0].Signals[0].ID = "tampered"
	if s.Data().Offers[0].Capabilities[0] != "text_input" {
		t.Fatal("capabilities not copied")
	}
	if s.Data().Offers[0].Signals[0].ID != "sig-code" {
		t.Fatal("signals not copied")
	}
	if !reflect.DeepEqual(orig.Inputs, fixtureSnapshotData().Inputs) {
		t.Fatal("fixture mutated")
	}
	// Mutating the caller's input after construction must not leak in.
	orig.Offers[0].Capabilities[0] = "image_output"
	if s.Data().Offers[0].Capabilities[0] != "text_input" {
		t.Fatal("snapshot aliases caller slices")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	s := mustSnapshot(t)
	enc, err := EncodeSnapshot(s)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := DecodeSnapshot(enc)
	if err != nil {
		t.Fatalf("decode rejected own encoding: %v", err)
	}
	if dec.Hash() != s.Hash() {
		t.Fatalf("hash mismatch %s != %s", dec.Hash(), s.Hash())
	}
	want := fixtureSnapshotData()
	// Constructor normalization turns nil warning slices into [], so compare
	// against the normalized form.
	for i := range want.Offers {
		for j := range want.Offers[i].Signals {
			want.Offers[i].Signals[j].Warnings = copyStrings(want.Offers[i].Signals[j].Warnings)
		}
	}
	got := dec.Data()
	if !reflect.DeepEqual(want.Inputs, got.Inputs) {
		t.Fatal("inputs differ")
	}
	if !reflect.DeepEqual(want.Offers, got.Offers) {
		t.Fatalf("offers differ:\n got %+v\nwant %+v", got.Offers, want.Offers)
	}
	if !got.Offers[0].ConfirmedAt.Equal(fixtureSnapshotData().Offers[0].ConfirmedAt) {
		t.Fatal("confirmed_at lost")
	}
}

func TestEncodeNilSnapshot(t *testing.T) {
	if _, err := EncodeSnapshot(nil); err == nil {
		t.Fatal("nil snapshot encoded")
	}
	var zero *Snapshot
	if zero.Hash() != "" {
		t.Fatal("nil hash not empty")
	}
	if zero.Data().Schema != 0 {
		t.Fatal("nil data not zero")
	}
}

func TestEmptyCollectionsEncodeAsArrays(t *testing.T) {
	d := fixtureSnapshotData()
	d.Inputs = nil
	d.Offers[1].Capabilities = nil
	s, err := NewSnapshot(d)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := EncodeSnapshot(s)
	if err != nil {
		t.Fatal(err)
	}
	str := string(enc)
	if strings.Contains(str, "null") {
		t.Fatalf("null present in encoding: %s", str)
	}
	if !strings.Contains(str, `"inputs":[]`) {
		t.Fatalf("nil inputs did not encode as []: %s", str)
	}
}
