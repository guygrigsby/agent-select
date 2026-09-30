package routing

import (
	"bytes"
	"time"
)

// Snapshot is the immutable value projection of the RoutingSnapshot
// aggregate. It is only constructible in a valid state, via NewSnapshot or
// DecodeSnapshot.
type Snapshot struct {
	data SnapshotData
	hash string // lowercase hex SHA-256 over the canonical artifact bytes
	// byID maps offer ID to a pointer into data.Offers. It is built once
	// at construction and never mutated afterwards, so concurrent Route
	// calls may read it without synchronization. It is not part of the
	// encoding, the hash or any public API.
	byID map[string]*OfferData
}

// NewSnapshot validates the DTO, deep-copies all slices (nil slices become
// empty) and computes the artifact hash over the canonical encoding. The
// canonical encoding must fit the 64 MiB artifact bound.
func NewSnapshot(d SnapshotData) (*Snapshot, error) {
	if err := validateSnapshotData(d); err != nil {
		return nil, err
	}
	norm := copySnapshotData(d)
	enc, err := canonicalize(norm)
	if err != nil {
		return nil, err
	}
	if len(enc) > MaxArtifactBytes {
		return nil, invalidSnapshot("canonical artifact is %d bytes, at most %d allowed", len(enc), MaxArtifactBytes)
	}
	return &Snapshot{data: norm, hash: hashCanonical(enc), byID: buildOfferIndex(norm.Offers)}, nil
}

// buildOfferIndex indexes validated, normalized offers by ID. Offer IDs are
// unique (enforced by validateSnapshotData), so the map is total. The empty
// snapshot gets an empty, non-nil map.
func buildOfferIndex(offers []OfferData) map[string]*OfferData {
	idx := make(map[string]*OfferData, len(offers))
	for i := range offers {
		idx[offers[i].ID] = &offers[i]
	}
	return idx
}

// Data returns a deep copy of the snapshot data.
func (s *Snapshot) Data() SnapshotData {
	if s == nil {
		return SnapshotData{}
	}
	return copySnapshotData(s.data)
}

// Hash returns the validated artifact hash, lowercase hex SHA-256 over the
// canonical bytes.
func (s *Snapshot) Hash() string {
	if s == nil {
		return ""
	}
	return s.hash
}

// EncodeSnapshot returns the canonical RFC 8785 artifact bytes.
func EncodeSnapshot(s *Snapshot) ([]byte, error) {
	if s == nil || s.hash == "" {
		return nil, invalidSnapshot("snapshot is nil or not constructed")
	}
	enc, err := canonicalize(s.data)
	if err != nil {
		return nil, err
	}
	if len(enc) > MaxArtifactBytes {
		return nil, invalidSnapshot("canonical artifact is %d bytes, at most %d allowed", len(enc), MaxArtifactBytes)
	}
	if hashCanonical(enc) != s.hash {
		return nil, invalidSnapshot("snapshot state no longer matches its validated hash")
	}
	return enc, nil
}

// DecodeSnapshot validates bounded canonical bytes and returns the snapshot.
// Schema or policy mismatch reports schema_incompatible; every other defect
// reports invalid_snapshot.
func DecodeSnapshot(data []byte) (*Snapshot, error) {
	if len(data) > MaxArtifactBytes {
		return nil, invalidSnapshot("artifact is %d bytes, at most %d allowed", len(data), MaxArtifactBytes)
	}
	if len(data) == 0 {
		return nil, invalidSnapshot("artifact is empty")
	}
	if err := strictCheck(data); err != nil {
		return nil, err
	}
	var d SnapshotData
	if err := strictUnmarshal(data, &d); err != nil {
		return nil, err
	}
	if err := validateSnapshotData(d); err != nil {
		return nil, err
	}
	// Byte-for-byte replay: canonical bytes must round-trip exactly.
	enc, err := canonicalize(copySnapshotData(d))
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(enc, data) {
		return nil, invalidSnapshot("decoded value does not re-encode to the same canonical bytes")
	}
	norm := copySnapshotData(d)
	return &Snapshot{data: norm, hash: hashCanonical(enc), byID: buildOfferIndex(norm.Offers)}, nil
}

// copySnapshotData deep-copies all slices; nil slices become empty so
// encodings use [] rather than null.
func copySnapshotData(d SnapshotData) SnapshotData {
	c := d
	c.Inputs = make([]InputData, len(d.Inputs))
	copy(c.Inputs, d.Inputs)
	c.Offers = make([]OfferData, len(d.Offers))
	copy(c.Offers, d.Offers)
	for i := range c.Offers {
		o := &c.Offers[i]
		o.Capabilities = copyStrings(o.Capabilities)
		o.Limits = copyLimits(o.Limits)
		o.Prices = copyPrices(o.Prices)
		sig := make([]SignalData, len(o.Signals))
		copy(sig, o.Signals)
		o.Signals = sig
		for j := range o.Signals {
			o.Signals[j].Warnings = copyStrings(o.Signals[j].Warnings)
		}
	}
	return c
}

// creationTime returns CreatedAt for age computations (phase 11).
func (s *Snapshot) creationTime() time.Time { return s.data.CreatedAt }
