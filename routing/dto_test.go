package routing

import (
	"strings"
	"testing"
	"time"
)

// mutating helpers over the fixture

func snapshotWith(t *testing.T, mutate func(*SnapshotData)) SnapshotData {
	t.Helper()
	d := fixtureSnapshotData()
	mutate(&d)
	return d
}

func TestNewSnapshotAcceptsFixture(t *testing.T) {
	s, err := NewSnapshot(fixtureSnapshotData())
	if err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
	if s.Hash() == "" {
		t.Fatal("hash empty")
	}
}

func TestNewSnapshotRejectsInvalidData(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*SnapshotData)
		want   string
	}{
		{"schema", func(d *SnapshotData) { d.Schema = 2 }, "schema"},
		{"policy", func(d *SnapshotData) { d.Policy = "other" }, "policy"},
		{"distribution", func(d *SnapshotData) { d.Distribution = "public" }, "distribution"},
		{"zero created", func(d *SnapshotData) { d.CreatedAt = time.Time{} }, "created_at"},
		{"non-utc created", func(d *SnapshotData) {
			d.CreatedAt = fixtureTime.In(time.FixedZone("X", 3600))
		}, "UTC"},
		{"unaligned created", func(d *SnapshotData) {
			d.CreatedAt = fixtureTime.Add(1500 * time.Microsecond)
		}, "millisecond-aligned"},
		{"zero cutoff", func(d *SnapshotData) { d.KnowledgeCutoff = time.Time{} }, "knowledge_cutoff"},
		{"cutoff after created", func(d *SnapshotData) {
			d.KnowledgeCutoff = fixtureTime
			d.CreatedAt = fixtureTime.Add(-time.Second)
		}, "knowledge_cutoff cannot exceed"},
		{"too many offers", func(d *SnapshotData) {
			d.Offers = make([]OfferData, MaxOffers+1)
			for i := range d.Offers {
				d.Offers[i] = OfferData{ID: "p/m", Provider: "p", UpstreamID: "m"}
			}
		}, "offers"},
		{"input empty observation", func(d *SnapshotData) { d.Inputs[0].ObservationID = "" }, "empty"},
		{"input empty source", func(d *SnapshotData) { d.Inputs[0].SourceID = "" }, "empty"},
		{"input empty policy", func(d *SnapshotData) { d.Inputs[0].PolicyVersion = "" }, "empty"},
		{"duplicate input observation", func(d *SnapshotData) {
			d.Inputs[1].ObservationID = d.Inputs[0].ObservationID
		}, "duplicate input"},
		{"unsorted inputs", func(d *SnapshotData) {
			d.Inputs[0], d.Inputs[1] = d.Inputs[1], d.Inputs[0]
		}, "sorted"},
		{"offer id mismatch", func(d *SnapshotData) { d.Offers[0].ID = "wrong" }, "id must equal"},
		{"offer empty provider", func(d *SnapshotData) {
			d.Offers[0].Provider = ""
			d.Offers[0].ID = d.Offers[0].Provider + "/" + d.Offers[0].UpstreamID
		}, "provider is empty"},
		{"offer id whitespace", func(d *SnapshotData) {
			d.Offers[0].Provider = " prov"
			d.Offers[0].ID = d.Offers[0].Provider + "/" + d.Offers[0].UpstreamID
		}, "whitespace"},
		{"offer id control char", func(d *SnapshotData) {
			d.Offers[0].Provider = "pro\x02v"
			d.Offers[0].ID = d.Offers[0].Provider + "/" + d.Offers[0].UpstreamID
		}, "control"},
		{"offer id over 512 bytes", func(d *SnapshotData) {
			d.Offers[0].Provider = strings.Repeat("a", 513)
			d.Offers[0].ID = d.Offers[0].Provider + "/" + d.Offers[0].UpstreamID
		}, "512"},
		{"unknown capability", func(d *SnapshotData) {
			d.Offers[0].Capabilities = append(d.Offers[0].Capabilities, "telepathy")
		}, "unknown capability"},
		{"duplicate capability", func(d *SnapshotData) {
			d.Offers[0].Capabilities = append(d.Offers[0].Capabilities, "tool_use")
		}, "duplicate capability"},
		{"unknown limit resource", func(d *SnapshotData) {
			d.Offers[0].Limits[0].Resource = "gpus"
		}, "unknown limit"},
		{"duplicate limit", func(d *SnapshotData) {
			d.Offers[0].Limits = append(d.Offers[0].Limits, LimitData{Resource: "input_tokens", Maximum: 5})
		}, "duplicate limit"},
		{"zero limit", func(d *SnapshotData) { d.Offers[0].Limits[0].Maximum = 0 }, "positive"},
		{"negative limit", func(d *SnapshotData) { d.Offers[0].Limits[0].Maximum = -3 }, "positive"},
		{"limit over 53-bit", func(d *SnapshotData) {
			d.Offers[0].Limits[0].Maximum = canonicalIntMax + 1
		}, "canonical integer range"},
		{"unknown price component", func(d *SnapshotData) {
			d.Offers[0].Prices[0].Component = "gpu_seconds"
		}, "unknown price"},
		{"duplicate price", func(d *SnapshotData) {
			d.Offers[0].Prices = append(d.Offers[0].Prices, PriceData{Component: "input_tokens", USDPerMillion: "1"})
		}, "duplicate price"},
		{"negative price", func(d *SnapshotData) { d.Offers[0].Prices[0].USDPerMillion = "-1" }, "decimal syntax"},
		{"price bad syntax", func(d *SnapshotData) { d.Offers[0].Prices[0].USDPerMillion = "1." }, "decimal syntax"},
		{"price over 64 chars", func(d *SnapshotData) {
			d.Offers[0].Prices[0].USDPerMillion = "1." + strings.Repeat("0", 63)
		}, "64"},
		{"price exponent over 18", func(d *SnapshotData) {
			d.Offers[0].Prices[0].USDPerMillion = "1e19"
		}, "±18"},
		{"price 19 fractional digits", func(d *SnapshotData) {
			d.Offers[0].Prices[0].USDPerMillion = "0.1234567890123456789"
		}, "fractional"},
		{"price float expansion 19 digits", func(d *SnapshotData) {
			d.Offers[0].Prices[0].USDPerMillion = "1.23456789012345678e-2"
		}, "fractional"},
		{"signal id empty", func(d *SnapshotData) { d.Offers[0].Signals[0].ID = "" }, "signal id"},
		{"signal unknown task", func(d *SnapshotData) { d.Offers[0].Signals[0].Task = "vibes" }, "unknown signal task"},
		{"signal value high", func(d *SnapshotData) { d.Offers[0].Signals[0].Value = 10001 }, "outside 0..10000"},
		{"signal value negative", func(d *SnapshotData) { d.Offers[0].Signals[0].Value = -1 }, "outside 0..10000"},
		{"signal weight zero", func(d *SnapshotData) { d.Offers[0].Signals[0].Weight = 0 }, "outside 1..10000"},
		{"signal weight high", func(d *SnapshotData) { d.Offers[0].Signals[0].Weight = 10001 }, "outside 1..10000"},
		{"signal unknown basis", func(d *SnapshotData) { d.Offers[0].Signals[0].Basis = "rumor" }, "unknown signal basis"},
		{"signal empty evidence", func(d *SnapshotData) { d.Offers[0].Signals[0].EvidenceID = "" }, "evidence id"},
		{"signal revision conflict", func(d *SnapshotData) {
			d.Offers[0].Signals[0].Revision = "different"
		}, "conflicts with offer revision"},
		{"signal empty revision allowed", func(d *SnapshotData) {
			d.Offers[0].Signals[0].Revision = ""
		}, ""}, // must NOT error; handled by no-error path below
		{"warning bad token", func(d *SnapshotData) {
			d.Offers[0].Signals[1].Warnings = []string{"Bad Token"}
		}, "bounded token"},
		{"warning raw text", func(d *SnapshotData) {
			d.Offers[0].Signals[1].Warnings = []string{"the source said \"trust me\""}
		}, "bounded token"},
		{"too many warnings", func(d *SnapshotData) {
			d.Offers[0].Signals[1].Warnings = make([]string, 65)
			for i := range d.Offers[0].Signals[1].Warnings {
				d.Offers[0].Signals[1].Warnings[i] = "w_1"
			}
			// duplicate tokens are fine for the count check first? no:
			// duplicates are also invalid, so use distinct valid tokens.
			for i := range d.Offers[0].Signals[1].Warnings {
				d.Offers[0].Signals[1].Warnings[i] = strings.Repeat("w", 1) + "_"
				// still duplicates; craft distinct instead below
			}
		}, ""},
		{"over signal cap", func(d *SnapshotData) {
			d.Offers[0].Signals = make([]SignalData, MaxSignalsPerOffer+1)
			for i := range d.Offers[0].Signals {
				d.Offers[0].Signals[i] = SignalData{
					ID: strings.Repeat("a", 1) + "_", Task: "code", Value: 1, Weight: 1,
					Basis: "published", EvidenceID: "e",
				}
			}
		}, "per-offer bound"},
		{"negative latency", func(d *SnapshotData) { d.Offers[0].LatencyMillis = -1 }, "negative"},
		{"confirmed after cutoff", func(d *SnapshotData) {
			d.Offers[0].ConfirmedAt = fixtureTime.Add(time.Second)
		}, "confirmed_at cannot exceed"},
		{"duplicate offer id", func(d *SnapshotData) {
			d.Offers[1].ID = d.Offers[0].ID
			d.Offers[1].Provider = d.Offers[0].Provider
			d.Offers[1].UpstreamID = d.Offers[0].UpstreamID
		}, "duplicate offer"},
	}
	for _, tc := range cases {
		d := snapshotWith(t, tc.mutate)
		_, err := NewSnapshot(d)
		if tc.name == "signal empty revision allowed" {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tc.name, err)
			}
			continue
		}
		if tc.name == "too many warnings" {
			continue // covered separately with distinct tokens below
		}
		if err == nil {
			t.Errorf("%s: accepted", tc.name)
			continue
		}
		if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.name, err.Error(), tc.want)
		}
	}
}

func TestWarningBounds(t *testing.T) {
	d := fixtureSnapshotData()
	ws := make([]string, maxWarnings+1)
	for i := range ws {
		// distinct valid tokens: w0, w1, ... all match [a-z][a-z0-9_]{0,63}
		ws[i] = "w" + string(rune('0'+i%10)) + string(rune('a'+i/10))
	}
	d.Offers[0].Signals[1].Warnings = ws
	if _, err := NewSnapshot(d); err == nil {
		t.Fatal("65 warnings accepted")
	}
	ws = ws[:maxWarnings]
	d.Offers[0].Signals[1].Warnings = ws
	if _, err := NewSnapshot(d); err != nil {
		t.Fatalf("64 distinct warnings rejected: %v", err)
	}
}

func TestPriceBoundsAccept(t *testing.T) {
	accepted := []string{"0", "0.0", "1e-18", "1e18", "0.123456789012345678", "1.234567890123456789e1", "999999", "1.5E+2"}
	for _, p := range accepted {
		d := fixtureSnapshotData()
		d.Offers[0].Prices[0].USDPerMillion = p
		if _, err := NewSnapshot(d); err != nil {
			t.Errorf("price %q rejected: %v", p, err)
		}
	}
}

func TestPriceBoundsReject(t *testing.T) {
	rejected := []string{"1e-19", "1e19", "0.1234567890123456789", "1.23456789012345678901e-1", " 1", "1 ", "+1", "1e", "0x10", "1.2.3"}
	for _, p := range rejected {
		d := fixtureSnapshotData()
		d.Offers[0].Prices[0].USDPerMillion = p
		if _, err := NewSnapshot(d); err == nil {
			t.Errorf("price %q accepted", p)
		}
	}
}
