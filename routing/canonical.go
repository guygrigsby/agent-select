package routing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

// canonicalIntMin/Max are the signed 53-bit bounds (RFC 8785 number
// serialization follows ECMAScript IEEE 754; within this range the JCS number
// algorithm reduces to plain decimal digits).
const (
	canonicalIntMin = -9007199254740991
	canonicalIntMax = 9007199254740991
)

// millisFormat is the pinned canonical time layout, including ".000" for
// whole seconds.
const millisFormat = "2006-01-02T15:04:05.000Z"

// checkMillisTime validates that t is nonzero, UTC, millisecond-aligned
// and within years 1..9999. errf selects the caller's error code.
func checkMillisTime(t time.Time, what string, errf func(string, ...any) *Error) error {
	if t.IsZero() {
		return errf("%s is required and nonzero", what)
	}
	if _, offset := t.Zone(); offset != 0 {
		return errf("%s is not UTC", what)
	}
	if t.Nanosecond()%int(time.Millisecond) != 0 {
		return errf("%s is not millisecond-aligned", what)
	}
	if t.Year() < 1 || t.Year() > 9999 {
		return errf("%s year is outside 1..9999", what)
	}
	return nil
}

// encodeMillisJSON renders a nonzero UTC millisecond-aligned time as the
// pinned quoted canonical string.
func encodeMillisJSON(t time.Time) (json.RawMessage, error) {
	if err := checkMillisTime(t, "time", invalidSnapshot); err != nil {
		return nil, err
	}
	return json.RawMessage(strconv.Quote(t.UTC().Format(millisFormat))), nil
}

// parseMillisJSON accepts only the pinned canonical time format.
func parseMillisJSON(b []byte, what string) (time.Time, error) {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return time.Time{}, invalidSnapshot("%s is not a string", what)
	}
	t, err := time.Parse(millisFormat, s)
	if err != nil {
		return time.Time{}, invalidSnapshot("%s does not match the canonical format", what)
	}
	return t, nil
}

// canonicalize encodes v to JSON and transforms it to RFC 8785 form.
func canonicalize(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, invalidSnapshot("encoding failed: invalid snapshot fields")
	}
	out, err := jsoncanonicalizer.Transform(raw)
	if err != nil {
		return nil, invalidSnapshot("canonicalization failed: invalid snapshot encoding")
	}
	return out, nil
}

// hashCanonical returns the lowercase hex SHA-256 over the exact bytes.
func hashCanonical(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// strictCheck validates raw canonical-v1 JSON bytes before struct decoding:
// no duplicate keys, no trailing values, no nulls, no floats, and integers
// within the signed 53-bit range. Unknown members are rejected during struct
// decode by DisallowUnknownFields.
func strictCheck(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := walkValue(dec); err != nil {
		return err
	}
	// Exactly one complete JSON value. More() is not an EOF test: it also
	// returns false for trailing '}' and ']' bytes, so a document followed
	// by a stray closer or a closer plus a second value would pass. Pull
	// the next token instead and require the stdlib io.EOF: anything else
	// is trailing data, rejected by a fixed label without echoing it.
	if _, err := dec.Token(); err != io.EOF {
		return invalidSnapshot("trailing values after the JSON document")
	}
	return nil
}

func walkValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return invalidSnapshot("invalid JSON")
	}
	return walkToken(dec, tok)
}

func walkToken(dec *json.Decoder, tok json.Token) error {
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			seen := make(map[string]bool)
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return invalidSnapshot("invalid JSON object")
				}
				key, ok := keyTok.(string)
				if !ok {
					return invalidSnapshot("object key is not a string")
				}
				if seen[key] {
					return invalidSnapshot("duplicate JSON key")
				}
				seen[key] = true
				if err := walkValue(dec); err != nil {
					return err
				}
			}
			end, err := dec.Token()
			if err != nil || end != json.Delim('}') {
				return invalidSnapshot("unterminated JSON object")
			}
		case '[':
			for dec.More() {
				if err := walkValue(dec); err != nil {
					return err
				}
			}
			end, err := dec.Token()
			if err != nil || end != json.Delim(']') {
				return invalidSnapshot("unterminated JSON array")
			}
		default:
			return invalidSnapshot("unexpected delimiter")
		}
	case json.Number:
		if err := checkInt53(v); err != nil {
			return err
		}
	case nil:
		return invalidSnapshot("null is not a canonical-v1 value; omit the member instead")
	}
	return nil
}

func checkInt53(n json.Number) error {
	if i, err := n.Int64(); err == nil {
		if i < canonicalIntMin || i > canonicalIntMax {
			return invalidSnapshot("integer exceeds the signed 53-bit canonical range")
		}
		return nil
	}
	if _, ok := new(big.Rat).SetString(n.String()); ok {
		return invalidSnapshot("float is not a canonical-v1 integer")
	}
	return invalidSnapshot("number is not a canonical-v1 integer")
}

// strictUnmarshal validates raw bytes with strictCheck, then decodes into v
// rejecting unknown members.
func strictUnmarshal(data []byte, v any) error {
	if err := strictCheck(data); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return decodeFailure("decode failed", err)
	}
	return nil
}

// Preserve diagnostics produced by our fixed DTO parsers, never raw decoder
// messages that can contain untrusted keys or values. Context is a fixed label.
func decodeFailure(context string, err error) *Error {
	if e, ok := err.(*Error); ok {
		return invalidSnapshot("%s: %s", context, e.Detail)
	}
	return invalidSnapshot("%s: unknown or invalid member", context)
}
