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
	"math"
	"runtime"
	"sort"
	"sync"
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
var perfEngineSink *Engine

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

// routeAllocBytes measures process-wide allocated bytes per Route call.
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

// ---------------------------------------------------------------------
// Latency percentiles and cold construction (verification-gap slice).
//
// The benchmarks above report means (ns/op) only; a tail latency class
// was missing. Everything below measures the actual Engine.Route and the
// cold constructors, reports observed p50/p95/p99 plus allocation
// counts, and asserts no performance thresholds: numbers below are
// measurements taken on this machine at run time, not promises. Where a
// distribution is summarized, the nearest-rank method is used: rank
// ceil(p/100*N) over an ascending-sorted sample slice, 1-based, clamped
// to the sample count.
// ---------------------------------------------------------------------

// perfPercentile returns the nearest-rank percentile of an ascending-
// sorted duration slice.
func perfPercentile(sorted []time.Duration, pct float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(pct / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

// collectRouteSamples warms the pinned route 20 times, then measures calls
// including result checks, clock overhead and any GC pauses. Allocation
// totals exclude timing storage but can include background runtime work.
func collectRouteSamples(tb testing.TB, eng *Engine, req Request, samples int) (sorted []time.Duration, bytes, allocs uint64) {
	tb.Helper()
	for range 20 {
		if _, err := eng.Route(req); err != nil {
			tb.Fatalf("Route warmup: %v", err)
		}
	}
	return collectPerfSamples(tb, samples, func(tb testing.TB) {
		d, err := eng.Route(req)
		if err != nil {
			tb.Fatalf("Route: %v", err)
		}
		if d.Selected == nil {
			tb.Fatal("fixture selected no offer")
		}
		perfSink = d
	})
}

// BenchmarkRoutePercentiles measures the per-call latency distribution of
// the real Engine.Route path at 1000 and 10000 snapshot offers crossed
// with 1000 and 10000 runtime offers (the request runtime registry is the
// per-call workload; the snapshot is pinned and immutable). Run through
// `make perf`, which passes -benchtime=1x so each sub-benchmark executes
// exactly one fixed 200-call sample batch. With a default benchtime the
// framework repeats batches; metrics aggregate every measured operation.
// Framework N counts batches, while ns/op, B/op and allocs/op describe one
// Route call. Percentiles use nearest rank. No thresholds are asserted.
func BenchmarkRoutePercentiles(b *testing.B) {
	const samples = 200
	for _, tc := range []struct {
		snapOffers    int
		runtimeOffers int
	}{
		{1000, 1000},
		{1000, 10000},
		{10000, 1000},
		{10000, 10000},
	} {
		b.Run(fmt.Sprintf("Snap%d_Runtime%d", tc.snapOffers, tc.runtimeOffers), func(b *testing.B) {
			eng, req := perfRoutePinned(b, tc.snapOffers, tc.runtimeOffers)
			benchmarkPerfSamples(b, "route", func() ([]time.Duration, uint64, uint64) {
				return collectRouteSamples(b, eng, req, samples)
			})
		})
	}
}

// collectPerfSamples measures an operation, including its result checks.
// Allocate timing storage before MemStats so it is not charged to the operation.
// Process-wide allocation deltas can include background runtime activity.
func collectPerfSamples(tb testing.TB, samples int, op func(tb testing.TB)) (sorted []time.Duration, bytes, allocs uint64) {
	tb.Helper()
	ds := make([]time.Duration, 0, samples)
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range samples {
		t0 := time.Now()
		op(tb)
		ds = append(ds, time.Since(t0))
	}
	runtime.ReadMemStats(&after)
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	return ds, after.TotalAlloc - before.TotalAlloc, after.Mallocs - before.Mallocs
}

// benchmarkPerfSamples reports per-operation metrics, not sample-batch costs.
func benchmarkPerfSamples(b *testing.B, unit string, sample func() ([]time.Duration, uint64, uint64)) {
	b.Helper()
	var all []time.Duration
	var bytes, allocs uint64
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ds, bt, at := sample()
		all = append(all, ds...)
		bytes += bt
		allocs += at
	}
	b.StopTimer()
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	var total time.Duration
	for _, d := range all {
		total += d
	}
	n := float64(len(all))
	b.ReportMetric(float64(total)/n, "ns/op")
	b.ReportMetric(float64(perfPercentile(all, 50)), "p50-ns/"+unit)
	b.ReportMetric(float64(perfPercentile(all, 95)), "p95-ns/"+unit)
	b.ReportMetric(float64(perfPercentile(all, 99)), "p99-ns/"+unit)
	b.ReportMetric(float64(bytes)/n, "B/op")
	b.ReportMetric(float64(allocs)/n, "allocs/op")
	b.ReportMetric(n, "samples")
}

// BenchmarkColdConstructionPercentiles measures one-time acceptance and
// adoption costs separately from the warm Route path: NewSnapshot
// (validation, deep copy, canonical encode, hash), NewRequest
// (validation plus deep copy), DecodeSnapshot (the Engine.Adopt decode
// path) and NewEngine. Run through `make perf` (-benchtime=1x) so every
// operation executes exactly its fixed sample count. NewEngine is a
// single atomic pointer store, so its measured wall clock is dominated by
// the time.Now clock-read overhead rather than the operation itself; it is
// reported anyway, with that caveat, because "cold construction" without
// the engine step would be incomplete.
func BenchmarkColdConstructionPercentiles(b *testing.B) {
	perfSnap1000 := perfSnapshotData(1000)
	perfSnap10000 := perfSnapshotData(10000)
	perfReq1000 := perfRequestData(perfRuntimeOffers(1000))
	perfReq10000 := perfRequestData(perfRuntimeOffers(10000))
	snap1000, err := NewSnapshot(perfSnap1000)
	if err != nil {
		b.Fatalf("NewSnapshot: %v", err)
	}
	snap10000, err := NewSnapshot(perfSnap10000)
	if err != nil {
		b.Fatalf("NewSnapshot: %v", err)
	}
	perfDec1000, err := EncodeSnapshot(snap1000)
	if err != nil {
		b.Fatalf("EncodeSnapshot: %v", err)
	}
	perfDec10000, err := EncodeSnapshot(snap10000)
	if err != nil {
		b.Fatalf("EncodeSnapshot: %v", err)
	}

	for _, tc := range []struct {
		name    string
		samples int
		op      func(tb testing.TB)
	}{
		{"NewSnapshot1000", 150, func(tb testing.TB) {
			if _, err := NewSnapshot(perfSnap1000); err != nil {
				tb.Fatalf("NewSnapshot: %v", err)
			}
		}},
		{"NewSnapshot10000", 30, func(tb testing.TB) {
			if _, err := NewSnapshot(perfSnap10000); err != nil {
				tb.Fatalf("NewSnapshot: %v", err)
			}
		}},
		{"NewRequest_RT1000", 300, func(tb testing.TB) {
			if _, err := NewRequest(perfReq1000); err != nil {
				tb.Fatalf("NewRequest: %v", err)
			}
		}},
		{"NewRequest_RT10000", 100, func(tb testing.TB) {
			if _, err := NewRequest(perfReq10000); err != nil {
				tb.Fatalf("NewRequest: %v", err)
			}
		}},
		{"DecodeSnapshot1000", 40, func(tb testing.TB) {
			if _, err := DecodeSnapshot(perfDec1000); err != nil {
				tb.Fatalf("DecodeSnapshot: %v", err)
			}
		}},
		{"DecodeSnapshot10000", 10, func(tb testing.TB) {
			if _, err := DecodeSnapshot(perfDec10000); err != nil {
				tb.Fatalf("DecodeSnapshot: %v", err)
			}
		}},
		{"NewEngine", 2000, func(tb testing.TB) {
			perfEngineSink = NewEngine(snap1000)
		}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			benchmarkPerfSamples(b, "op", func() ([]time.Duration, uint64, uint64) {
				return collectPerfSamples(b, tc.samples, tc.op)
			})
		})
	}
}

// TestRouteUnderConcurrentAdoption exercises the existing public API
// only (NewEngine, EncodeSnapshot, Engine.Adopt, Engine.Route) with Route
// readers running concurrently against an Adopt writer, which is the
// documented usage shape: adopters swap generations while callers pin
// one generation per call. Invariants under churn: every Route call
// succeeds, selects an offer, and reports a SnapshotHash belonging to one
// of the two known valid generations (each call must observe one atomic
// generation, never a torn mix); every Adopt of valid bytes succeeds. Run
// with -race (the Makefile race target does) to catch data races; the
// timing distribution it logs is contention-distorted by design and is an
// observation, not a comparable latency figure and not a threshold.
func TestRouteUnderConcurrentAdoption(t *testing.T) {
	if testing.Short() {
		t.Skip("concurrent adoption test skipped in short mode")
	}
	// Two valid generations of different sizes so each Adopt pays a real
	// DecodeSnapshot cost; both contain every runtime offer ID below.
	s300, err := NewSnapshot(perfSnapshotData(300))
	if err != nil {
		t.Fatalf("NewSnapshot: %v", err)
	}
	s600, err := NewSnapshot(perfSnapshotData(600))
	if err != nil {
		t.Fatalf("NewSnapshot: %v", err)
	}
	enc300, err := EncodeSnapshot(s300)
	if err != nil {
		t.Fatalf("EncodeSnapshot: %v", err)
	}
	enc600, err := EncodeSnapshot(s600)
	if err != nil {
		t.Fatalf("EncodeSnapshot: %v", err)
	}
	req, err := NewRequest(perfRequestData(perfRuntimeOffers(50)))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	valid := map[string]bool{s300.Hash(): true, s600.Hash(): true}
	eng := NewEngine(s300)

	const (
		workers        = 4
		itersPerWorker = 100
		adoptIters     = 12
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range adoptIters {
			enc := enc300
			if i%2 == 1 {
				enc = enc600
			}
			if err := eng.Adopt(enc); err != nil {
				t.Errorf("adopt of valid generation failed: %v", err)
				return
			}
		}
	}()

	perWorker := make([][]time.Duration, workers)
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			ds := make([]time.Duration, 0, itersPerWorker)
			for range itersPerWorker {
				t0 := time.Now()
				d, err := eng.Route(req)
				el := time.Since(t0)
				if err != nil {
					t.Errorf("Route under adoption failed: %v", err)
					return
				}
				if d.Selected == nil {
					t.Errorf("Route under adoption selected no offer")
					return
				}
				if !valid[d.SnapshotHash] {
					t.Errorf("Route observed unknown generation hash %.12s...", d.SnapshotHash)
					return
				}
				ds = append(ds, el)
			}
			perWorker[w] = ds
		}(w)
	}
	wg.Wait()
	if t.Failed() {
		return
	}

	var all []time.Duration
	for _, ds := range perWorker {
		all = append(all, ds...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	t.Logf("Route under concurrent adoption: n=%d p50=%v p95=%v p99=%v min=%v max=%v (observed under adopter contention and possibly -race; not comparable to idle percentiles, no threshold implied)",
		len(all),
		perfPercentile(all, 50), perfPercentile(all, 95), perfPercentile(all, 99),
		all[0], all[len(all)-1])
}
