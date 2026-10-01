package collector

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// canonicalize encodes v as JSON and transforms it to RFC 8785 (JCS) form,
// the canonical-v1 byte encoding pinned by delivery-storage.md.
func canonicalize(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsoncanonicalizer.Transform(raw)
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
// members are rejected and the input must already be canonical.
func strictDecode(data []byte, v any) error {
	if !isCanonical(data) {
		return errf(CodeInvalidEvidence, "payload bytes are not canonical-v1")
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
