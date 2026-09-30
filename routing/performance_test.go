package routing

// Performance benchmarks for the accepted pure offline library actual path:
// construct NewSnapshot/NewRequest once outside the timer, then exercise
// Engine.Route repeatedly. No data is mutated inside benchmark loops and no
// construction or codec cost is folded into warm-route metrics.
//
// All fixture data is synthetic and owned by this repository: unique
// synthetic IDs, explicit UTC fixture time, closed-vocabulary capabilities,
// limits, prices, tasks and bases. No real model identifiers, no
// credentials, no inference.

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// perfTime is the explicit UTC fixture time for this file.
var perfTime = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

// perfOffer builds one synthetic snapshot offer. IDs are unique synthetic
// constructions, never real models.
func perfOffer(i int) OfferData {
	id := fmt.Sprintf("perf/m%05d", i)
	confirmed := time.Date(2026, 9, 29, 8, 30, 0, 0, time.UTC)
	return OfferData{
		ID:           id,
		Provider:     "perf",
		UpstreamID:   id[len("perf/"):],
		ModelID:      id + "-2026-09",
		Revision:     "r1",
		Capabilities: []string{"text_input", "text_output", "tool_use"},
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "2.50"},
			{Component: "output_tokens", USDPerMillion: "10"},
		},
		Signals: []SignalData{
			{ID: "sig-code", Task: "code", Value: 6200, Weight: 40, Basis: "published", EvidenceID: "ev-perf-code", Revision: "r1"},
			{ID: "sig-general", Task: "general", Value: 5500, Weight: 20, Basis: "published", EvidenceID: "ev-perf-general", Revision: "r1"},
		},
		LatencyMillis: 900,
		Region:        "us-east",
		Privacy:       "standard",
		ConfirmedAt:   confirmed,
	}
}

// perfSnapshotData builds a valid snapshot with n synthetic offers.
func perfSnapshotData(n int) SnapshotData {
	offers := make([]OfferData, n)
	for i := range offers {
		offers[i] = perfOffer(i)
	}
	return SnapshotData{
		Schema:          SnapshotSchema,
		CreatedAt:       perfTime,
		KnowledgeCutoff: perfTime,
		Policy:          SnapshotPolicy,
		Distribution:    "local",
		Inputs: []InputData{
			{ObservationID: "obs-perf-1", SourceID: "src-perf-1", PolicyVersion: "local-research-v1"},
		},
		Offers: offers,
	}
}

// perfRuntimeOffer builds one live host runtime offer matching snapshot
// offer i (host assertions are trusted local facts).
func perfRuntimeOffer(i int) RuntimeOfferData {
	id := fmt.Sprintf("perf/m%05d", i)
	return RuntimeOfferData{
		ID:           id,
		Provider:     "perf",
		UpstreamID:   id[len("perf/"):],
		Revision:     "r1",
		Eligible:     true,
		Capabilities: []string{"text_input", "text_output"},
		Limits: []LimitData{
			{Resource: "input_tokens", Maximum: 200000},
			{Resource: "output_tokens", Maximum: 64000},
		},
		Prices: []PriceData{
			{Component: "input_tokens", USDPerMillion: "2.50"},
			{Component: "output_tokens", USDPerMillion: "10"},
		},
	}
}

// perfRuntimeOffers builds n distinct runtime offers.
func perfRuntimeOffers(n int) []RuntimeOfferData {
	rts := make([]RuntimeOfferData, n)
	for i := range rts {
		rts[i] = perfRuntimeOffer(i)
	}
	return rts
}

// perfRequestData builds a valid routing request over the given runtime.
func perfRequestData(rts []RuntimeOfferData) RequestData {
	return RequestData{
		ID:                   "req-perf-1",
		Task:                 "code",
		Mode:                 "cost",
		EffectiveAt:          perfTime,
		RuntimeGeneration:    "gen-perf-1",
		Runtime:              rts,
		RequiredCapabilities: []string{"text_input", "text_output"},
		InputTokens:          1000,
		OutputTokens:         500,
		ImageOperation:       "none",
	}
}

// perfSink prevents dead-code elimination of benchmark results.
var perfSink Decision

// perfRoutePinned constructs the full warm-route fixture once: snapshot and
// request are accepted via the library constructors outside any timer.
func perfRoutePinned(b testing.TB, snapOffers, runtimeOffers int) (*Engine, Request) {
	b.Helper()
	snap, err := NewSnapshot(perfSnapshotData(snapOffers))
	if err != nil {
		b.Fatalf("NewSnapshot: %v", err)
	}
	req, err := NewRequest(perfRequestData(perfRuntimeOffers(runtimeOffers)))
	if err != nil {
		b.Fatalf("NewRequest: %v", err)
	}
	return NewEngine(snap), req
}

// BenchmarkRoute is the warm actual path: one pinned generation, repeated
// Route calls. Snapshot construction (including canonicalization and
// hashing) happens once outside the timer.
func BenchmarkRoute(b *testing.B) {
	for _, tc := range []struct {
		snapOffers    int
		runtimeOffers int
	}{
		{10000, 1},
		{10000, 100},
		{10000, 1000},
	} {
		b.Run(fmt.Sprintf("Snap%d_Runtime%d", tc.snapOffers, tc.runtimeOffers), func(b *testing.B) {
			eng, req := perfRoutePinned(b, tc.snapOffers, tc.runtimeOffers)
			// Sanity: the fixture must route successfully once.
			d, err := eng.Route(req)
			if err != nil {
				b.Fatalf("Route: %v", err)
			}
			if d.Selected == nil {
				b.Fatal("fixture selected no offer")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				perfSink, _ = eng.Route(req)
			}
		})
	}
}

// BenchmarkRouteScale isolates how Route scales with pinned snapshot size
// for a fixed single runtime offer. The expected shape is flat: a pinned
// immutable snapshot should not force per-call work proportional to its
// offer count.
func BenchmarkRouteScale(b *testing.B) {
	for _, snapOffers := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("Snap%d_Runtime1", snapOffers), func(b *testing.B) {
			eng, req := perfRoutePinned(b, snapOffers, 1)
			if _, err := eng.Route(req); err != nil {
				b.Fatalf("Route: %v", err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				perfSink, _ = eng.Route(req)
			}
		})
	}
}

// BenchmarkNewSnapshot measures the one-time acceptance cost of the
// realistic 10000-offer fixture (validation, deep copy, canonical encode,
// hash). This cost sits outside the warm-route timers above.
func BenchmarkNewSnapshot(b *testing.B) {
	b.Run("Snap10000", func(b *testing.B) {
		data := perfSnapshotData(10000)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := NewSnapshot(data); err != nil {
				b.Fatalf("NewSnapshot: %v", err)
			}
		}
	})
}

// BenchmarkDecodeSnapshot measures DecodeSnapshot on the canonical bytes of
// the same realistic 10000-offer fixture.
func BenchmarkDecodeSnapshot(b *testing.B) {
	b.Run("Snap10000", func(b *testing.B) {
		snap, err := NewSnapshot(perfSnapshotData(10000))
		if err != nil {
			b.Fatalf("NewSnapshot: %v", err)
		}
		data, err := EncodeSnapshot(snap)
		if err != nil {
			b.Fatalf("EncodeSnapshot: %v", err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := DecodeSnapshot(data); err != nil {
				b.Fatalf("DecodeSnapshot: %v", err)
			}
		}
	})
}

// BenchmarkNewRequest measures acceptance of the realistic request DTO with
// a 100-offer runtime registry (validation plus deep copy).
func BenchmarkNewRequest(b *testing.B) {
	b.Run("Runtime100", func(b *testing.B) {
		data := perfRequestData(perfRuntimeOffers(100))
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := NewRequest(data); err != nil {
				b.Fatalf("NewRequest: %v", err)
			}
		}
	})
}

// routeAllocBytes returns the cumulative bytes allocated per Route call for
// the given pinned fixture, measured with runtime.MemStats.TotalAlloc
// (monotonic, GC-independent).
func routeAllocBytes(b testing.TB, eng *Engine, req Request, iters int) uint64 {
	b.Helper()
	// Warm the path once so lazily built state cannot skew the delta.
	if _, err := eng.Route(req); err != nil {
		b.Fatalf("Route warmup: %v", err)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range iters {
		perfSink, _ = eng.Route(req)
	}
	runtime.ReadMemStats(&after)
	return (after.TotalAlloc - before.TotalAlloc) / uint64(iters)
}

// TestRouteAllocsPerCallIsSnapshotIndependent is the allocation regression
// test. Deterministic logical bound: a Route call's allocations are a
// function of the runtime registry size, not of the pinned immutable
// snapshot size. One runtime offer must not allocate a full 10000-offer
// lookup structure per call. The bound is 16 KiB of allocation per call,
// measured on a snapshot of 10 offers and asserted against a snapshot of
// 10000 offers; the small-snapshot path sits far below it.
// The lookup index belongs to the pinned snapshot, never an individual call.
func TestRouteAllocsPerCallIsSnapshotIndependent(t *testing.T) {
	if testing.Short() {
		t.Skip("allocation regression test skipped in short mode")
	}
	const iters = 500
	engSmall, reqSmall := perfRoutePinned(t, 10, 1)
	small := routeAllocBytes(t, engSmall, reqSmall, iters)

	engBig, reqBig := perfRoutePinned(t, 10000, 1)
	big := routeAllocBytes(t, engBig, reqBig, iters)

	t.Logf("route allocs per call: snapshot=10 offers -> %d B/op, snapshot=10000 offers -> %d B/op", small, big)

	// Bound: 16 KiB per call for one runtime offer, independent of the
	// pinned snapshot size.
	const maxBytesPerCall = uint64(16 << 10)
	if big > maxBytesPerCall {
		t.Errorf("Route with one runtime offer allocated %d B/call against a 10000-offer snapshot (bound %d B/call); per-call work must not scale with pinned snapshot size", big, maxBytesPerCall)
	}
	if big > small+maxBytesPerCall {
		t.Errorf("Route allocations scale with snapshot size: %d B/call at 10000 offers vs %d B/call at 10 offers", big, small)
	}
}
