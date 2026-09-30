package routing

import (
	"encoding/json"
	"strings"
	"time"
)

// RequestData is the boundary DTO for one routing request.
type RequestData struct {
	ID                   string             `json:"id"`
	Task                 string             `json:"task"`
	Mode                 string             `json:"mode"`
	EffectiveAt          time.Time          `json:"effective_at"`
	RuntimeGeneration    string             `json:"runtime_generation"`
	Runtime              []RuntimeOfferData `json:"runtime"`
	RequiredCapabilities []string           `json:"required_capabilities"`
	InputTokens          int64              `json:"input_tokens"`
	OutputTokens         int64              `json:"output_tokens"`
	ImageOperation       string             `json:"image_operation"`
	ImageCount           int64              `json:"image_count"`
	ImageWidth           int64              `json:"image_width"`
	ImageHeight          int64              `json:"image_height"`
	ImageBytes           int64              `json:"image_bytes"`
	ReferenceCount       int64              `json:"reference_count"`
	Mask                 bool               `json:"mask"`
	MaskAlpha            bool               `json:"mask_alpha"`
	Searches             int64              `json:"searches"`
	BudgetKnown          bool               `json:"budget_known"`
	BudgetMicros         int64              `json:"budget_micros"`
	StrictCost           bool               `json:"strict_cost"`
	DeadlineMillis       int64              `json:"deadline_millis"`
	StrictLatency        bool               `json:"strict_latency"`
	MaxAgeSeconds        int64              `json:"max_age_seconds"`
	AllowedProviders     []string           `json:"allowed_providers"`
	Region               string             `json:"region"`
	Privacy              string             `json:"privacy"`
	AllowUnscored        bool               `json:"allow_unscored"`
	MinimumEstimate      int                `json:"minimum_estimate"`
}

// RuntimeOfferData is one live host offer in the request runtime registry.
// Host assertions are trusted local facts, never public research approval.
type RuntimeOfferData struct {
	ID                      string      `json:"id"`
	Provider                string      `json:"provider"`
	UpstreamID              string      `json:"upstream_id"`
	Revision                string      `json:"revision"`
	Eligible                bool        `json:"eligible"`
	Capabilities            []string    `json:"capabilities"`
	UnsupportedCapabilities []string    `json:"unsupported_capabilities"`
	Limits                  []LimitData `json:"limits"`
	Prices                  []PriceData `json:"prices"`
	Region                  string      `json:"region"`
	Privacy                 string      `json:"privacy"`
}

// MarshalJSON serializes EffectiveAt in the pinned canonical format.
func (d RequestData) MarshalJSON() ([]byte, error) {
	type alias RequestData
	type out struct {
		alias
		EffectiveAt json.RawMessage `json:"effective_at"`
	}
	at, err := encodeMillisJSON(d.EffectiveAt)
	if err != nil {
		return nil, invalidRequest("effective_at: %v", err)
	}
	return json.Marshal(out{alias(d), at})
}

// UnmarshalJSON accepts only the pinned canonical time format.
func (d *RequestData) UnmarshalJSON(b []byte) error {
	type alias RequestData
	type out struct {
		alias
		EffectiveAt json.RawMessage `json:"effective_at"`
	}
	var v out
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return decodeFailure("request member decode failed", err)
	}
	at, err := parseMillisJSON(v.EffectiveAt, "effective_at")
	if err != nil {
		return err
	}
	*d = RequestData(v.alias)
	d.EffectiveAt = at
	return nil
}

// NewRequest validates RequestData and returns the immutable Request value
// object. A zero-value Request is rejected by Route with invalid_request.
func NewRequest(d RequestData) (Request, error) {
	if err := validateRequestData(d); err != nil {
		return Request{}, err
	}
	return Request{data: copyRequestData(d), validated: true}, nil
}

// Request is an immutable validated routing request.
type Request struct {
	data      RequestData
	validated bool
}

// Data returns a deep copy of the validated DTO.
func (r Request) Data() RequestData { return copyRequestData(r.data) }

func validateRequestData(d RequestData) error {
	if err := checkIDPart(d.ID, "request id"); err != nil {
		return err
	}
	if !taskTokens[d.Task] {
		return invalidRequest("unknown task")
	}
	if !modeTokens[d.Mode] {
		return invalidRequest("unknown mode")
	}
	if err := checkMillisTime(d.EffectiveAt, "effective_at", invalidRequest); err != nil {
		return err
	}
	if err := checkIDPart(d.RuntimeGeneration, "runtime generation"); err != nil {
		return err
	}
	if err := checkTokenList(d.RequiredCapabilities, capabilityTokens, "required capability"); err != nil {
		return err
	}
	if err := validateRuntime(d.Runtime); err != nil {
		return err
	}
	if err := validateProviderList(d.AllowedProviders); err != nil {
		return err
	}
	if err := checkFreeForm(d.Region, "request region"); err != nil {
		return err
	}
	if err := checkFreeForm(d.Privacy, "request privacy"); err != nil {
		return err
	}

	negatives := []struct {
		name string
		v    int64
	}{
		{"input tokens", d.InputTokens},
		{"output tokens", d.OutputTokens},
		{"image count", d.ImageCount},
		{"image width", d.ImageWidth},
		{"image height", d.ImageHeight},
		{"image bytes", d.ImageBytes},
		{"reference count", d.ReferenceCount},
		{"searches", d.Searches},
		{"budget micros", d.BudgetMicros},
		{"deadline millis", d.DeadlineMillis},
		{"max age seconds", d.MaxAgeSeconds},
	}
	for _, n := range negatives {
		if n.v < 0 {
			return invalidRequest("%s cannot be negative", n.name)
		}
		if n.v > canonicalIntMax {
			return invalidRequest("%s exceeds the canonical integer range", n.name)
		}
	}
	if !d.BudgetKnown && d.BudgetMicros != 0 {
		return invalidRequest("budget value requires budget_known")
	}
	if d.StrictCost && !d.BudgetKnown {
		return invalidRequest("strict cost requires a known budget")
	}
	if d.DeadlineMillis == 0 && d.StrictLatency {
		return invalidRequest("deadline 0 means no deadline and requires strict_latency false")
	}
	if d.StrictLatency && d.DeadlineMillis <= 0 {
		return invalidRequest("strict latency requires a positive deadline")
	}
	if d.MinimumEstimate < 0 || d.MinimumEstimate > 10000 {
		return invalidRequest("minimum estimate outside 0..10000")
	}
	if err := validateImageOperation(d); err != nil {
		return err
	}
	return nil
}

func validateImageOperation(d RequestData) error {
	if !imageOperationTokens[d.ImageOperation] {
		return invalidRequest("unknown image operation")
	}
	if d.ImageOperation == "none" {
		if d.ImageCount != 0 || d.ImageWidth != 0 || d.ImageHeight != 0 || d.ImageBytes != 0 {
			return invalidRequest("image operation none requires all image quantities zero")
		}
		if d.Mask || d.MaskAlpha {
			return invalidRequest("image operation none requires mask flags false")
		}
		return nil
	}
	if d.ImageCount <= 0 || d.ImageWidth <= 0 || d.ImageHeight <= 0 || d.ImageBytes <= 0 {
		return invalidRequest("image operation requires positive count, width, height and bytes")
	}
	if d.Mask && d.ImageOperation != "edit" {
		return invalidRequest("mask requires the edit operation")
	}
	if d.MaskAlpha && !d.Mask {
		return invalidRequest("mask alpha requires mask")
	}
	return nil
}

func validateRuntime(rs []RuntimeOfferData) error {
	seen := make(map[string]bool, len(rs))
	for i := range rs {
		r := &rs[i]
		if err := checkIDPart(r.Provider, "runtime provider"); err != nil {
			return err
		}
		if err := checkIDPart(r.UpstreamID, "runtime upstream id"); err != nil {
			return err
		}
		if r.ID != r.Provider+"/"+r.UpstreamID {
			return invalidRequest("runtime id must equal provider/upstream")
		}
		if seen[r.ID] {
			return invalidRequest("duplicate runtime offer")
		}
		seen[r.ID] = true
		if err := checkTokenList(r.Capabilities, capabilityTokens, "runtime capability"); err != nil {
			return err
		}
		if err := checkTokenList(r.UnsupportedCapabilities, capabilityTokens, "runtime unsupported capability"); err != nil {
			return err
		}
		for _, fs := range []struct {
			v    string
			what string
		}{{r.Revision, "runtime revision"}, {r.Region, "runtime region"}, {r.Privacy, "runtime privacy"}} {
			if err := checkFreeForm(fs.v, fs.what); err != nil {
				return err
			}
		}
		for _, c := range r.UnsupportedCapabilities {
			for _, p := range r.Capabilities {
				if c == p {
					return invalidRequest("capability appears in both runtime capability lists")
				}
			}
		}
		if err := validateLimits(r.Limits); err != nil {
			return err
		}
		if err := validatePrices(r.Prices); err != nil {
			return err
		}
	}
	return nil
}

func validateProviderList(list []string) error {
	seen := make(map[string]bool, len(list))
	for _, s := range list {
		if err := checkIDPart(s, "allowed provider"); err != nil {
			return err
		}
		if seen[s] {
			return invalidRequest("duplicate allowed provider")
		}
		seen[s] = true
	}
	return nil
}

// copyRequestData deep-copies all slices so the value object stays immutable.
func copyRequestData(d RequestData) RequestData {
	c := d
	c.RequiredCapabilities = copyStrings(d.RequiredCapabilities)
	c.AllowedProviders = copyStrings(d.AllowedProviders)
	c.Runtime = make([]RuntimeOfferData, len(d.Runtime))
	copy(c.Runtime, d.Runtime)
	for i := range c.Runtime {
		c.Runtime[i].Capabilities = copyStrings(d.Runtime[i].Capabilities)
		c.Runtime[i].UnsupportedCapabilities = copyStrings(d.Runtime[i].UnsupportedCapabilities)
		c.Runtime[i].Limits = copyLimits(d.Runtime[i].Limits)
		c.Runtime[i].Prices = copyPrices(d.Runtime[i].Prices)
	}
	return c
}

func copyStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

func copyLimits(l []LimitData) []LimitData {
	if l == nil {
		return []LimitData{}
	}
	out := make([]LimitData, len(l))
	copy(out, l)
	return out
}

func copyPrices(p []PriceData) []PriceData {
	if p == nil {
		return []PriceData{}
	}
	out := make([]PriceData, len(p))
	copy(out, p)
	return out
}
