package routing

import (
	"encoding/json"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Snapshot schema and policy pins (educated-routing.md SnapshotData).
const (
	SnapshotSchema = 1
	SnapshotPolicy = "educated-v1"
)

// Snapshot bounds.
const (
	MaxOffers          = 100000
	MaxSignalsPerOffer = 1024
	MaxArtifactBytes   = 64 << 20 // 64 MiB
	maxIDBytes         = 512
	maxWarnings        = 64
)

var warningPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// SnapshotData is the boundary DTO for a routing snapshot. It is not a valid
// domain object until NewSnapshot accepts it.
type SnapshotData struct {
	Schema          int         `json:"schema"`
	CreatedAt       time.Time   `json:"created_at"`
	KnowledgeCutoff time.Time   `json:"knowledge_cutoff"`
	Policy          string      `json:"policy"`
	Distribution    string      `json:"distribution"`
	Inputs          []InputData `json:"inputs"`
	Offers          []OfferData `json:"offers"`
}

// InputData names one observation, its source and the authorizing policy
// version. All fields are nonempty; inputs are unique by observation and
// sorted by observation/source/policy.
type InputData struct {
	ObservationID string `json:"observation_id"`
	SourceID      string `json:"source_id"`
	PolicyVersion string `json:"policy_version"`
}

// OfferData is one snapshot offer profile. No credentials.
type OfferData struct {
	ID            string       `json:"id"`
	Provider      string       `json:"provider"`
	UpstreamID    string       `json:"upstream_id"`
	ModelID       string       `json:"model_id"`
	Revision      string       `json:"revision"`
	Capabilities  []string     `json:"capabilities"`
	Limits        []LimitData  `json:"limits"`
	Prices        []PriceData  `json:"prices"`
	Signals       []SignalData `json:"signals"`
	LatencyMillis int64        `json:"latency_millis"`
	Region        string       `json:"region"`
	Privacy       string       `json:"privacy"`
	ConfirmedAt   time.Time    `json:"confirmed_at,omitempty"`
}

// LimitData is one positive resource bound. An absent resource means
// unknown.
type LimitData struct {
	Resource string `json:"resource"`
	Maximum  int64  `json:"maximum"`
}

// PriceData is one exact-decimal USD-per-million quote for a component,
// conservatively covering all applicable tiers.
type PriceData struct {
	Component     string `json:"component"`
	USDPerMillion string `json:"usd_per_million"`
}

// SignalData is one derived heuristic input, never a rewritten raw score.
type SignalData struct {
	ID         string   `json:"id"`
	Task       string   `json:"task"`
	Value      int      `json:"value"`
	Weight     int      `json:"weight"`
	Basis      string   `json:"basis"`
	EvidenceID string   `json:"evidence_id"`
	Revision   string   `json:"revision"`
	Warnings   []string `json:"warnings"`
}

// MarshalJSON serializes times in the pinned canonical format; a zero
// ConfirmedAt is omitted via the shadow field.
func (d SnapshotData) MarshalJSON() ([]byte, error) {
	type alias SnapshotData
	type out struct {
		alias
		CreatedAt       json.RawMessage `json:"created_at"`
		KnowledgeCutoff json.RawMessage `json:"knowledge_cutoff"`
	}
	ca, err := encodeMillisJSON(d.CreatedAt)
	if err != nil {
		return nil, invalidRequest("created_at: %v", err)
	}
	kc, err := encodeMillisJSON(d.KnowledgeCutoff)
	if err != nil {
		return nil, invalidRequest("knowledge_cutoff: %v", err)
	}
	return json.Marshal(out{alias(d), ca, kc})
}

// UnmarshalJSON accepts only the pinned canonical time format.
func (d *SnapshotData) UnmarshalJSON(b []byte) error {
	type alias SnapshotData
	type out struct {
		alias
		CreatedAt       json.RawMessage `json:"created_at"`
		KnowledgeCutoff json.RawMessage `json:"knowledge_cutoff"`
	}
	var o out
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&o); err != nil {
		return decodeFailure("snapshot member decode failed", err)
	}
	ca, err := parseMillisJSON(o.CreatedAt, "created_at")
	if err != nil {
		return err
	}
	kc, err := parseMillisJSON(o.KnowledgeCutoff, "knowledge_cutoff")
	if err != nil {
		return err
	}
	d.Schema = o.Schema
	d.CreatedAt = ca
	d.KnowledgeCutoff = kc
	d.Policy = o.Policy
	d.Distribution = o.Distribution
	d.Inputs = o.Inputs
	d.Offers = o.Offers
	return nil
}

// MarshalJSON serializes ConfirmedAt in the pinned format, omitting zero.
func (o OfferData) MarshalJSON() ([]byte, error) {
	type alias OfferData
	type out struct {
		alias
		ConfirmedAt json.RawMessage `json:"confirmed_at,omitempty"`
	}
	var ca json.RawMessage
	if !o.ConfirmedAt.IsZero() {
		var err error
		ca, err = encodeMillisJSON(o.ConfirmedAt)
		if err != nil {
			return nil, invalidRequest("confirmed_at: %v", err)
		}
	}
	return json.Marshal(out{alias(o), ca})
}

// UnmarshalJSON accepts only the pinned canonical time format.
func (o *OfferData) UnmarshalJSON(b []byte) error {
	type alias OfferData
	type out struct {
		alias
		ConfirmedAt json.RawMessage `json:"confirmed_at"`
	}
	var v out
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return decodeFailure("offer member decode failed", err)
	}
	if len(v.ConfirmedAt) == 0 {
		o.ID = v.ID
		o.Provider = v.Provider
		o.UpstreamID = v.UpstreamID
		o.ModelID = v.ModelID
		o.Revision = v.Revision
		o.Capabilities = v.Capabilities
		o.Limits = v.Limits
		o.Prices = v.Prices
		o.Signals = v.Signals
		o.LatencyMillis = v.LatencyMillis
		o.Region = v.Region
		o.Privacy = v.Privacy
		return nil
	}
	ca, err := parseMillisJSON(v.ConfirmedAt, "confirmed_at")
	if err != nil {
		return err
	}
	*o = OfferData(v.alias)
	o.ConfirmedAt = ca
	return nil
}

// validateSnapshotData enforces every SnapshotData invariant without
// mutating its input. Schema/policy mismatches report schema_incompatible;
// every other field violation reports invalid_snapshot.
func validateSnapshotData(d SnapshotData) error {
	if d.Schema != SnapshotSchema {
		return schemaIncompatible("schema %d is not %d", d.Schema, SnapshotSchema)
	}
	if d.Policy != SnapshotPolicy {
		return schemaIncompatible("unsupported snapshot policy")
	}
	if err := validateSnapshotFields(d); err != nil {
		// Field violations from shared helpers carry invalid_request; on the
		// snapshot path they are snapshot defects.
		if re, ok := err.(*Error); ok && re.Code == CodeInvalidRequest {
			re.Code = CodeInvalidSnapshot
		}
		return err
	}
	return nil
}

func validateSnapshotFields(d SnapshotData) error {
	if !distributionTokens[d.Distribution] {
		return invalidRequest("distribution must be local or distributable")
	}
	if err := checkMillisTime(d.CreatedAt, "created_at", invalidRequest); err != nil {
		return err
	}
	if err := checkMillisTime(d.KnowledgeCutoff, "knowledge_cutoff", invalidRequest); err != nil {
		return err
	}
	if d.KnowledgeCutoff.After(d.CreatedAt) {
		return invalidRequest("knowledge_cutoff cannot exceed created_at")
	}
	if len(d.Offers) > MaxOffers {
		return invalidRequest("snapshot has %d offers, at most %d allowed", len(d.Offers), MaxOffers)
	}
	if err := validateInputs(d.Inputs); err != nil {
		return err
	}
	seen := make(map[string]bool, len(d.Offers))
	for i := range d.Offers {
		o := &d.Offers[i]
		if err := validateOffer(o, d.KnowledgeCutoff); err != nil {
			return invalidRequest("offer %d: %v", i, err)
		}
		if seen[o.ID] {
			return invalidRequest("duplicate offer id")
		}
		seen[o.ID] = true
	}
	return nil
}

func validateInputs(in []InputData) error {
	seen := make(map[string]bool, len(in))
	for i := range in {
		inp := &in[i]
		if err := checkIDPart(inp.ObservationID, "input observation id"); err != nil {
			return err
		}
		if err := checkIDPart(inp.SourceID, "input source id"); err != nil {
			return err
		}
		if err := checkIDPart(inp.PolicyVersion, "input policy version"); err != nil {
			return err
		}
		if seen[inp.ObservationID] {
			return invalidRequest("duplicate input observation")
		}
		seen[inp.ObservationID] = true
	}
	if !sort.SliceIsSorted(in, func(a, b int) bool {
		ia, ib := in[a], in[b]
		if ia.ObservationID != ib.ObservationID {
			return ia.ObservationID < ib.ObservationID
		}
		if ia.SourceID != ib.SourceID {
			return ia.SourceID < ib.SourceID
		}
		return ia.PolicyVersion < ib.PolicyVersion
	}) {
		return invalidRequest("inputs must be sorted by observation, source, policy")
	}
	return nil
}

func validateOffer(o *OfferData, cutoff time.Time) error {
	if err := checkIDPart(o.Provider, "provider"); err != nil {
		return err
	}
	if err := checkIDPart(o.UpstreamID, "upstream id"); err != nil {
		return err
	}
	if o.ID != o.Provider+"/"+o.UpstreamID {
		return invalidRequest("id must equal provider/upstream")
	}
	for _, fs := range []struct {
		v    string
		what string
	}{{o.ModelID, "model id"}, {o.Revision, "offer revision"}, {o.Region, "offer region"}, {o.Privacy, "offer privacy"}} {
		if err := checkFreeForm(fs.v, fs.what); err != nil {
			return err
		}
	}
	if err := checkTokenList(o.Capabilities, capabilityTokens, "capability"); err != nil {
		return err
	}
	if err := validateLimits(o.Limits); err != nil {
		return err
	}
	if err := validatePrices(o.Prices); err != nil {
		return err
	}
	if len(o.Signals) > MaxSignalsPerOffer {
		return invalidRequest("%d signals exceeds the %d per-offer bound", len(o.Signals), MaxSignalsPerOffer)
	}
	seenSignals := make(map[string]bool, len(o.Signals))
	for i := range o.Signals {
		s := &o.Signals[i]
		if err := validateSignal(s, o.Revision); err != nil {
			return invalidRequest("signal %d: %v", i, err)
		}
		if seenSignals[s.ID] {
			return invalidRequest("duplicate signal id")
		}
		seenSignals[s.ID] = true
	}
	if o.LatencyMillis < 0 {
		return invalidRequest("latency millis is negative")
	}
	if o.LatencyMillis > canonicalIntMax {
		return invalidRequest("latency millis exceeds the canonical integer range")
	}
	if !o.ConfirmedAt.IsZero() && o.ConfirmedAt.After(cutoff) {
		return invalidRequest("confirmed_at cannot exceed knowledge_cutoff")
	}
	return nil
}

func validateLimits(ls []LimitData) error {
	seen := make(map[string]bool, len(ls))
	for i := range ls {
		l := &ls[i]
		if !resourceTokens[l.Resource] {
			return invalidRequest("unknown limit resource")
		}
		if seen[l.Resource] {
			return invalidRequest("duplicate limit resource")
		}
		seen[l.Resource] = true
		if l.Maximum <= 0 {
			return invalidRequest("limit maximum must be positive")
		}
		if l.Maximum > canonicalIntMax {
			return invalidRequest("limit maximum exceeds the canonical integer range")
		}
	}
	return nil
}

func validatePrices(ps []PriceData) error {
	seen := make(map[string]bool, len(ps))
	for i := range ps {
		p := &ps[i]
		if !priceComponentTokens[p.Component] {
			return invalidRequest("unknown price component")
		}
		if seen[p.Component] {
			return invalidRequest("duplicate price component")
		}
		seen[p.Component] = true
		if len(p.USDPerMillion) > 64 {
			return invalidRequest("price decimal exceeds 64 characters")
		}
		if !decimalPattern.MatchString(p.USDPerMillion) {
			return invalidRequest("price does not match the decimal syntax")
		}
		if err := checkDecimalBounds(p.USDPerMillion); err != nil {
			return err
		}
	}
	return nil
}

// checkDecimalBounds enforces absolute exponent at most 18 and at most 18
// fractional digits after exponent expansion. Both counts derive from the
// decimal string itself; rational normalization would shrink the
// denominator (gcd reduction) and undercount digits.
func checkDecimalBounds(s string) error {
	if _, ok := new(big.Rat).SetString(s); !ok {
		return invalidRequest("price is not an exact decimal")
	}
	frac := 0
	if i := strings.IndexByte(s, '.'); i >= 0 {
		frac = strings.IndexAny(s[i:], "eE")
		if frac < 0 {
			frac = len(s[i:])
		}
		frac-- // exclude the '.' itself
	}
	exp := int64(0)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		var err error
		exp, err = parseExponent(s[i+1:])
		if err != nil {
			return invalidRequest("price has an invalid exponent")
		}
		if exp > 18 || exp < -18 {
			return invalidRequest("price exponent exceeds the ±18 bound")
		}
	}
	// Fractional digits after expansion: shifting the decimal point by exp.
	digits := int64(frac) - exp
	if digits > 18 {
		return invalidRequest("price has more than 18 fractional digits after expansion")
	}
	return nil
}

func validateSignal(s *SignalData, offerRevision string) error {
	if err := checkIDPart(s.ID, "signal id"); err != nil {
		return err
	}
	if !taskTokens[s.Task] {
		return invalidRequest("unknown signal task")
	}
	if s.Value < 0 || s.Value > 10000 {
		return invalidRequest("signal value %d outside 0..10000", s.Value)
	}
	if s.Weight < 1 || s.Weight > 10000 {
		return invalidRequest("signal weight %d outside 1..10000", s.Weight)
	}
	if !basisTokens[s.Basis] {
		return invalidRequest("unknown signal basis")
	}
	if err := checkIDPart(s.EvidenceID, "signal evidence id"); err != nil {
		return err
	}
	if err := checkFreeForm(s.Revision, "signal revision"); err != nil {
		return err
	}
	if s.Revision != "" && offerRevision != "" && s.Revision != offerRevision {
		return invalidRequest("signal revision conflicts with offer revision")
	}
	if len(s.Warnings) > maxWarnings {
		return invalidRequest("more than %d warnings", maxWarnings)
	}
	for _, w := range s.Warnings {
		if !warningPattern.MatchString(w) {
			return invalidRequest("warning is not a bounded token")
		}
	}
	return nil
}

// checkTokenList requires a closed vocabulary with no duplicates.
func checkTokenList(list []string, vocab map[string]bool, what string) error {
	seen := make(map[string]bool, len(list))
	for _, t := range list {
		if !vocab[t] {
			return invalidRequest("unknown %s token", what)
		}
		if seen[t] {
			return invalidRequest("duplicate %s token", what)
		}
		seen[t] = true
	}
	return nil
}

// checkIDPart validates an identifier: nonempty, valid UTF-8, at most 512
// UTF-8 bytes, no surrounding whitespace or control characters.
func checkIDPart(s, what string) error {
	if s == "" {
		return invalidRequest("%s is empty", what)
	}
	return checkFreeForm(s, what)
}

// checkFreeForm validates a bounded free-form string. Empty is allowed
// (the field's unknown sentinel).
func checkFreeForm(s, what string) error {
	if !utf8.ValidString(s) {
		return invalidRequest("%s is not valid UTF-8", what)
	}
	if len(s) > maxIDBytes {
		return invalidRequest("%s exceeds %d UTF-8 bytes", what, maxIDBytes)
	}
	if s != strings.TrimSpace(s) {
		return invalidRequest("%s has surrounding whitespace", what)
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return invalidRequest("%s contains control characters", what)
		}
	}
	return nil
}

func parseExponent(s string) (int64, error) {
	var n int64
	var neg bool
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if s == "" {
		return 0, invalidRequest("empty exponent")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, invalidRequest("invalid exponent digit")
		}
		n = n*10 + int64(c-'0')
		if n > 1<<30 {
			return 0, invalidRequest("exponent too large")
		}
	}
	if neg {
		n = -n
	}
	return n, nil
}
