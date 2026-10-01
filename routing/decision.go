package routing

// Decision DTOs (educated-routing.md "complete data shapes"). Route arrives
// with the admission and ordering engine; the wire shapes are pinned here.

// Estimate is a relative heuristic estimate, never a probability.
type Estimate struct {
	Known       bool     `json:"known"`
	Value       int      `json:"value"`
	Basis       string   `json:"basis"`
	EvidenceIDs []string `json:"evidence_ids"`
	Warnings    []string `json:"warnings"`
}

// CostEstimate is a conservative micro-USD upper bound over required priced
// usage. Unknown costs do not rank as free.
type CostEstimate struct {
	Known       bool     `json:"known"`
	UpperMicros int64    `json:"upper_micros"`
	Missing     []string `json:"missing"`
}

// Selection is one ranked admissible candidate.
type Selection struct {
	OfferID       string       `json:"offer_id"`
	Provider      string       `json:"provider"`
	UpstreamID    string       `json:"upstream_id"`
	Estimate      Estimate     `json:"estimate"`
	Cost          CostEstimate `json:"cost"`
	LatencyMillis int64        `json:"latency_millis"`
	Warnings      []string     `json:"warnings"`
}

// Rejection records complete stable reason codes for one candidate.
type Rejection struct {
	OfferID string   `json:"offer_id"`
	Reasons []string `json:"reasons"`
}

// Decision is the complete Route result. Selected nil means no admissible
// candidate; the error then carries code no_admissible_offer alongside this
// decision.
type Decision struct {
	RequestID         string      `json:"request_id"`
	Selected          *Selection  `json:"selected,omitempty"`
	Alternatives      []Selection `json:"alternatives"`
	Rejections        []Rejection `json:"rejections"`
	SnapshotHash      string      `json:"snapshot_hash"`
	Policy            string      `json:"policy"`
	RuntimeGeneration string      `json:"runtime_generation"`
	Warnings          []string    `json:"warnings"`
}
