package routing

import (
	"math/big"
	"sort"
)

// Route evaluates one request against the pinned current snapshot using the
// educated-routing estimate and selection algorithm. It is pure: no events,
// no writes, no wall clock. A nil selection is returned together with a
// no_admissible_offer error when every candidate is rejected.
func (e *Engine) Route(r Request) (Decision, error) {
	if !r.validated || r.data.ID == "" {
		return Decision{}, invalidRequest("request is not a validated NewRequest value")
	}
	pinned := e.pinned()
	if pinned == nil || pinned.hash == "" {
		return Decision{}, newError(CodeSnapshotUnavailable, "engine has no adopted snapshot")
	}
	d := r.data

	// Stable ID order for the whole iteration.
	rts := make([]RuntimeOfferData, len(d.Runtime))
	copy(rts, d.Runtime)
	sort.Slice(rts, func(i, j int) bool { return rts[i].ID < rts[j].ID })

	// Immutable per-generation offer index, built once by the snapshot
	// constructor and never mutated afterwards.
	snapByID := pinned.byID

	allowed := make(map[string]bool, len(d.AllowedProviders))
	for _, p := range d.AllowedProviders {
		allowed[p] = true
	}

	cands := make([]*candidate, 0, len(rts))
	rejections := make([]Rejection, 0, len(rts))
	for i := range rts {
		c := evaluateCandidate(&rts[i], snapByID, allowed, &d)
		if len(c.reasons) > 0 {
			rejections = append(rejections, Rejection{OfferID: c.rt.ID, Reasons: sortedCopy(c.reasons)})
			continue
		}
		cands = append(cands, c)
	}

	sortCands(cands, d.Mode)

	decision := Decision{
		RequestID:         d.ID,
		Rejections:        rejections,
		SnapshotHash:      pinned.hash,
		Policy:            pinned.data.Policy,
		RuntimeGeneration: d.RuntimeGeneration,
		Warnings:          []string{},
	}
	if len(cands) == 0 {
		decision.Selected = nil
		decision.Alternatives = []Selection{}
		return decision, newError(CodeNoAdmissibleOffer, "no admissible candidate")
	}
	sel := cands[0].selection()
	decision.Selected = &sel
	alts := make([]Selection, 0, len(cands)-1)
	for _, c := range cands[1:] {
		alts = append(alts, c.selection())
	}
	decision.Alternatives = alts
	return decision, nil
}

// candidate is one runtime offer under evaluation.
type candidate struct {
	rt              RuntimeOfferData
	snap            *OfferData // nil when unmapped or profile discarded
	profileConflict bool
	est             Estimate
	cost            CostEstimate
	latency         int64 // 0 means unknown
	warnings        []string
	reasons         []string
}

func (c *candidate) selection() Selection {
	warnings := c.warnings
	if c.profileConflict {
		// The snapshot profile was discarded for a revision conflict.
		warnings = append(append([]string(nil), warnings...), "revision_conflict")
	}
	return Selection{
		OfferID:       c.rt.ID,
		Provider:      c.rt.Provider,
		UpstreamID:    c.rt.UpstreamID,
		Estimate:      c.est,
		Cost:          c.cost,
		LatencyMillis: c.latency,
		Warnings:      warnSort(warnings),
	}
}

func evaluateCandidate(rt *RuntimeOfferData, snapByID map[string]*OfferData, allowed map[string]bool, d *RequestData) *candidate {
	c := &candidate{rt: *rt}
	if !rt.Eligible {
		c.reason("not_eligible")
	}
	if len(allowed) > 0 && !allowed[rt.Provider] {
		c.reason("provider_disallowed")
	}
	if snap, ok := snapByID[rt.ID]; ok {
		c.snap = snap
		if rt.Revision != "" && snap.Revision != "" && rt.Revision != snap.Revision {
			// Discard the entire conflicting snapshot profile; runtime facts
			// may still admit the unscored candidate.
			c.snap = nil
			c.profileConflict = true
		}
	}
	c.checkAdmission(rt, d)
	c.estimate(rt, d)
	c.applyEstimateGate(d)
	c.boundCost(rt, d)
	c.checkLatency(d)
	return c
}

func (c *candidate) reason(r string) { c.reasons = append(c.reasons, r) }

func (c *candidate) warn(w string) { c.warnings = append(c.warnings, w) }

// checkAdmission applies capability, region, privacy, limit and freshness
// gates. Runtime region/privacy are authoritative; empty never falls back to
// snapshot labels.
func (c *candidate) checkAdmission(rt *RuntimeOfferData, d *RequestData) {
	// Capability union: runtime + snapshot positives, minus host negatives.
	caps := make(map[string]bool, len(rt.Capabilities)+8)
	for _, cap := range rt.Capabilities {
		caps[cap] = true
	}
	if c.snap != nil {
		for _, cap := range c.snap.Capabilities {
			caps[cap] = true
		}
	}
	for _, cap := range rt.UnsupportedCapabilities {
		delete(caps, cap)
	}
	for _, req := range requiredCapabilities(d) {
		if !caps[req] {
			c.reason("capability_missing:" + req)
		}
	}

	if d.Region != "" {
		if rt.Region == "" {
			c.reason("region_unknown")
		} else if rt.Region != d.Region {
			c.reason("region_mismatch")
		}
	}
	if d.Privacy != "" {
		if rt.Privacy == "" {
			c.reason("privacy_unknown")
		} else if rt.Privacy != d.Privacy {
			c.reason("privacy_mismatch")
		}
	}

	c.checkLimits(rt, d)

	if d.MaxAgeSeconds > 0 {
		switch {
		case c.snap == nil || c.snap.ConfirmedAt.IsZero():
			c.reason("freshness_unknown")
		default:
			// Millisecond comparison: a seconds bound near 2^53 overflows
			// time.Duration nanoseconds.
			boundMs := d.MaxAgeSeconds * 1000
			ageMs := d.EffectiveAt.UnixMilli() - c.snap.ConfirmedAt.UnixMilli()
			switch {
			case ageMs < 0:
				c.reason("freshness_future")
			case ageMs > boundMs:
				c.reason("freshness_exceeded")
			}
		}
	}
}

// requiredCapabilities unions explicit requirements with operation-derived
// capabilities.
func requiredCapabilities(d *RequestData) []string {
	req := make(map[string]bool, len(d.RequiredCapabilities)+4)
	for _, r := range d.RequiredCapabilities {
		req[r] = true
	}
	switch d.ImageOperation {
	case "none":
		req["text_input"] = true
		req["text_output"] = true
	case "vision":
		req["text_input"] = true
		req["text_output"] = true
		req["image_input"] = true
	case "generate":
		req["text_input"] = true
		req["image_output"] = true
		if d.OutputTokens > 0 {
			req["text_output"] = true // requested auxiliary text
		}
	case "edit":
		req["image_input"] = true
		req["image_output"] = true
		req["image_edit"] = true
		if d.InputTokens > 0 {
			req["text_input"] = true
		}
		if d.OutputTokens > 0 {
			req["text_output"] = true // requested auxiliary text
		}
	}
	if d.Mask {
		req["image_mask"] = true
	}
	if d.MaskAlpha {
		req["image_alpha"] = true
	}
	if d.ReferenceCount > 0 {
		req["image_reference"] = true
	}
	out := make([]string, 0, len(req))
	for r := range req {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// checkLimits overlays runtime limits per resource over snapshot limits and
// applies every requested resource bound. Unknown requested bounds reject.
func (c *candidate) checkLimits(rt *RuntimeOfferData, d *RequestData) {
	limits := make(map[string]int64, 8)
	if c.snap != nil {
		for _, l := range c.snap.Limits {
			limits[l.Resource] = l.Maximum
		}
	}
	for _, l := range rt.Limits {
		limits[l.Resource] = l.Maximum
	}
	check := func(res string, qty int64) {
		max, ok := limits[res]
		if !ok {
			c.reason("limit_unknown:" + res)
			return
		}
		if qty > max {
			c.reason("limit_exceeded:" + res)
		}
	}
	require := func(res string, qty int64) {
		if qty <= 0 {
			return
		}
		if _, own := limits[res]; own {
			check(res, qty)
			return
		}
		if _, ctx := limits["context_tokens"]; ctx && (res == "input_tokens" || res == "output_tokens") {
			return // covered by the combined context bound below
		}
		c.reason("limit_unknown:" + res)
	}
	require("input_tokens", d.InputTokens)
	require("output_tokens", d.OutputTokens)
	if max, ok := limits["context_tokens"]; ok && d.InputTokens+d.OutputTokens > 0 {
		if sum := d.InputTokens + d.OutputTokens; sum > 0 && sum > max {
			c.reason("limit_exceeded:context_tokens")
		}
	}
	if d.ImageOperation != "none" {
		check("image_width", d.ImageWidth)
		check("image_height", d.ImageHeight)
		check("image_bytes", d.ImageBytes)
		switch d.ImageOperation {
		case "vision":
			check("input_images", d.ImageCount)
		case "generate":
			check("output_images", d.ImageCount)
		case "edit":
			check("input_images", d.ImageCount)
			check("output_images", d.ImageCount)
		}
	}
	if d.ReferenceCount > 0 {
		check("references", d.ReferenceCount)
	}
}

// applyEstimateGate enforces the minimum-estimate and unscored policy.
// Unknown can enter solely when AllowUnscored and MinimumEstimate is zero;
// otherwise the rejection is estimate_unknown, never
// estimate_below_minimum: the unknown zero sentinel is not a low score.
func (c *candidate) applyEstimateGate(d *RequestData) {
	switch {
	case c.est.Known && c.est.Value < d.MinimumEstimate:
		c.reason("estimate_below_minimum")
	case !c.est.Known && !(d.AllowUnscored && d.MinimumEstimate == 0):
		c.reason("estimate_unknown")
	}
}

// estimate computes the heuristic estimate from task-matching signals.
func (c *candidate) estimate(rt *RuntimeOfferData, d *RequestData) {
	c.est = Estimate{Known: false, Value: 0, Basis: "unscored", EvidenceIDs: []string{}, Warnings: []string{}}
	if c.snap == nil {
		return
	}
	byEvidence := make(map[string]bool)
	revConflict, provisional := false, false
	var keptSignals []SignalData
	num, den := int64(0), int64(0)
	// Smallest signal ID per evidence/task wins dedup.
	sorted := append([]SignalData(nil), c.snap.Signals...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	for _, s := range sorted {
		if s.Task != d.Task && s.Task != "general" {
			continue
		}
		// Revision exclusion happens before deduplication.
		if s.Revision != "" && rt.Revision != "" && s.Revision != rt.Revision {
			revConflict = true
			continue
		}
		key := s.EvidenceID + "\x00" + s.Task
		if byEvidence[key] {
			continue
		}
		byEvidence[key] = true
		keptSignals = append(keptSignals, s)
		if c.snap.Revision == "" || s.Revision == "" || rt.Revision == "" {
			provisional = true
		}
		num += int64(s.Value) * int64(s.Weight)
		den += int64(s.Weight)
	}
	if len(keptSignals) == 0 {
		if revConflict {
			c.est.Warnings = warnSort([]string{"revision_conflict"})
		}
		return
	}
	c.est.Known = true
	c.est.Basis = "heuristic"
	if den > 0 {
		c.est.Value = int(num / den)
	}
	ev := make([]string, 0, len(keptSignals))
	for _, s := range keptSignals {
		ev = append(ev, s.EvidenceID)
		c.est.Warnings = append(c.est.Warnings, s.Warnings...)
	}
	if revConflict {
		c.est.Warnings = append(c.est.Warnings, "revision_conflict")
	}
	if provisional {
		c.est.Warnings = append(c.est.Warnings, "provisional_association")
	}
	c.est.EvidenceIDs = sortedCopy(ev)
	c.est.Warnings = warnSort(c.est.Warnings)
}

// boundCost computes the exact micro-USD upper bound over required priced
// usage. Unknown costs never rank as free.
func (c *candidate) boundCost(rt *RuntimeOfferData, d *RequestData) {
	prices := make(map[string]string, 6)
	if c.snap != nil {
		for _, p := range c.snap.Prices {
			prices[p.Component] = p.USDPerMillion
		}
	}
	for _, p := range rt.Prices {
		prices[p.Component] = p.USDPerMillion
	}
	usage := pricedUsage(d)
	missing := make([]string, 0, len(usage))
	total := new(big.Rat)
	comps := make([]string, 0, len(usage))
	for comp := range usage {
		comps = append(comps, comp)
	}
	sort.Strings(comps)
	for _, comp := range comps {
		qty := usage[comp]
		rate, ok := prices[comp]
		if !ok {
			missing = append(missing, comp)
			continue
		}
		r := new(big.Rat)
		if _, ok := r.SetString(rate); !ok {
			missing = append(missing, comp)
			continue
		}
		// Charge in micro-USD: rate USD per million units times quantity is
		// already rate*qty micro-USD.
		q := new(big.Rat).SetInt64(qty)
		total.Add(total, new(big.Rat).Mul(r, q))
	}
	if len(missing) > 0 {
		c.cost = CostEstimate{Known: false, UpperMicros: 0, Missing: sortedCopy(missing)}
	} else {
		// Charges round up once after summing, in exact micro-USD.
		micros := ceilRat(total)
		if micros.Cmp(big.NewInt(canonicalIntMax)) > 0 {
			c.cost = CostEstimate{Known: false, UpperMicros: 0, Missing: []string{}}
			c.warn("cost_unknown")
		} else {
			c.cost = CostEstimate{Known: true, UpperMicros: micros.Int64(), Missing: []string{}}
		}
	}
	switch {
	case !c.cost.Known && d.StrictCost:
		c.reason("cost_unknown")
	case !c.cost.Known:
		c.warn("cost_unknown")
	case d.BudgetKnown && c.cost.UpperMicros > d.BudgetMicros:
		c.reason("budget_exceeded")
	}
}

// pricedUsage enumerates required components and their quantities. Token
// quotes are always required; zero quantities do not erase them.
func pricedUsage(d *RequestData) map[string]int64 {
	u := map[string]int64{
		"input_tokens":  d.InputTokens,
		"output_tokens": d.OutputTokens,
	}
	switch d.ImageOperation {
	case "vision":
		u["input_images"] = d.ImageCount
	case "generate":
		u["output_images"] = d.ImageCount
	case "edit":
		u["input_images"] = d.ImageCount
		u["output_images"] = d.ImageCount
	}
	if d.ReferenceCount > 0 {
		u["references"] = d.ReferenceCount
	}
	if d.Searches > 0 {
		u["searches"] = d.Searches
	}
	return u
}

func ceilRat(r *big.Rat) *big.Int {
	q, m := new(big.Int).QuoRem(r.Num(), r.Denom(), new(big.Int))
	if r.Num().Sign() < 0 || m.Sign() == 0 {
		return q
	}
	return q.Add(q, big.NewInt(1))
}

func (c *candidate) checkLatency(d *RequestData) {
	// Latency comes from the snapshot profile only; 0 means unavailable.
	if c.snap != nil && c.snap.LatencyMillis > 0 {
		c.latency = c.snap.LatencyMillis
	}
	known := c.latency > 0
	if !known && d.StrictLatency {
		c.reason("latency_unknown")
	}
	// Non-strict excess or unknown is disclosed through the value and
	// ordering, not rejected.
	if known && d.StrictLatency && c.latency > d.DeadlineMillis {
		c.reason("deadline_exceeded")
	}
}

// sortCands ranks admissible candidates deterministically.
func sortCands(cs []*candidate, mode string) {
	less := func(a, b *candidate) bool {
		if mode == "performance" {
			if a.est.Known != b.est.Known {
				return a.est.Known
			}
			if a.est.Value != b.est.Value {
				return a.est.Value > b.est.Value
			}
			return costThenLatencyID(a, b)
		}
		return costThenLatencyID(a, b)
	}
	sort.SliceStable(cs, func(i, j int) bool { return less(cs[i], cs[j]) })
}

func costThenLatencyID(a, b *candidate) bool {
	if a.cost.Known != b.cost.Known {
		return a.cost.Known
	}
	if a.cost.Known && a.cost.UpperMicros != b.cost.UpperMicros {
		return a.cost.UpperMicros < b.cost.UpperMicros
	}
	if mode := a.est.Known; mode != b.est.Known {
		return mode
	}
	if a.est.Known && a.est.Value != b.est.Value {
		return a.est.Value > b.est.Value
	}
	if (a.latency > 0) != (b.latency > 0) {
		return a.latency > 0
	}
	if a.latency != b.latency {
		return a.latency < b.latency
	}
	return a.rt.ID < b.rt.ID
}

// warnSort returns the complete deduplicated, bytewise-sorted warning
// union. Derived output warning lists (estimate, selection, decision) are
// never truncated to the 64-entry per-signal input bound; only
// deduplication drops entries (educated-routing.md DTO bounds).
func warnSort(in []string) []string {
	return sortedCopy(in)
}

// sortedCopy returns a deduplicated, bytewise-sorted copy.
func sortedCopy(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
