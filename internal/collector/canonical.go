package collector

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// maxSafeInteger is the signed 53-bit ceiling of canonical-v1 integers.
const maxSafeInteger = int64(1)<<53 - 1

// canonicalize encodes v as JSON and transforms it to RFC 8785 (JCS) form,
// the canonical-v1 byte encoding pinned by delivery-storage.md. Integers
// outside the signed 53-bit range are rejected here: JCS number serialization
// would silently round them through float64, so the hash would pin mutated
// facts.
func canonicalize(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := checkNumberBounds(raw); err != nil {
		return nil, err
	}
	return jsoncanonicalizer.Transform(raw)
}

// checkNumberBounds walks raw JSON and rejects integer literals outside the
// signed 53-bit range and non-representable numbers. Every path that accepts
// caller-supplied bytes or marshals caller-supplied values into canonical-v1
// runs this first.
func checkNumberBounds(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return errf(CodeInvalidEvidence, "payload is not valid JSON: %v", err)
	}
	var walk func(v any) error
	walk = func(v any) error {
		switch t := v.(type) {
		case json.Number:
			s := t.String()
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				if n > maxSafeInteger || n < -maxSafeInteger {
					return errf(CodeInvalidEvidence, "integer %s exceeds the signed 53-bit range", s)
				}
				return nil
			}
			f, err := t.Float64()
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return errf(CodeInvalidEvidence, "number %s is not representable as a double", s)
			}
			return nil
		case map[string]any:
			for _, vv := range t {
				if err := walk(vv); err != nil {
					return err
				}
			}
		case []any:
			for _, vv := range t {
				if err := walk(vv); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(doc)
}

// canonicalHash is SHA-256 over the exact canonical bytes.
func canonicalHash(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

// hexHash is the lowercase hex form of canonicalHash.
func hexHash(b []byte) string { return hex.EncodeToString(canonicalHash(b)) }

// isCanonical reports whether b is already in canonical-v1 form: the JCS
// transform is idempotent, so a re-encode mismatch means non-canonical input.
func isCanonical(b []byte) bool {
	out, err := jsoncanonicalizer.Transform(b)
	return err == nil && bytes.Equal(out, b)
}

// strictDecode decodes canonical bytes into a closed struct schema: unknown
// members are rejected, the input must already be canonical and integer
// literals must stay inside the signed 53-bit range.
func strictDecode(data []byte, v any) error {
	if !isCanonical(data) {
		return errf(CodeInvalidEvidence, "payload bytes are not canonical-v1")
	}
	if err := checkNumberBounds(data); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errf(CodeInvalidEvidence, "payload schema mismatch: %v", err)
	}
	if dec.More() {
		return errf(CodeInvalidEvidence, "trailing values after payload document")
	}
	return nil
}

// decodeHash32 parses a 64-character hex string into 32 bytes.
func decodeHash32(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 32 {
		return nil, errf(CodeInvalidEvidence, "hash %q is not 64 hex characters", s)
	}
	return b, nil
}
