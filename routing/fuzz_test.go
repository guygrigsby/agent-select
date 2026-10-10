package routing

// Native Go fuzz targets for the two byte-boundary decoders,
// DecodeRequestWire and DecodeSnapshot. These are additive verification:
// no decoder is weakened or changed. Every seed corpus entry, valid and
// invalid, runs as an ordinary unit test on every `go test` invocation, so
// the default corpus is exercised automatically in `make test` and in the
// mandatory `go test -race ./routing` check. Bounded fuzzing sessions
// (Makefile `fuzz` target) run both targets concurrently with two workers each.
//
// Invariants exercised by the fuzz bodies:
//   - No panic on arbitrary input bytes (the fuzzing engine guarantees
//     this only by never finding one; the corpus and mutations are the
//     search).
//   - On rejection, DecodeRequestWire returns the zero RequestData and
//     DecodeSnapshot returns a nil snapshot.
//   - DecodeSnapshot rejections carry a closed routing error code
//     (invalid_snapshot or schema_incompatible).
//   - Accepted snapshot bytes round-trip byte-for-byte through
//     EncodeSnapshot and re-construct through NewSnapshot with the same
//     hash.
//   - A wire request that DecodeRequestWire accepts and NewRequest
//     accepts (wire framing and domain rules are deliberately separate
//     layers) re-encodes to canonical bytes that re-decode to an equal
//     request. Requests accepted by the wire layer but rejected by
//     NewRequest (for example a budget value without budget_known) are
//     legitimate: DecodeRequestWire performs framing only, and the fuzz
//     body asserts no round-trip for them.
//
// Request decoder error-code drift is tracked in ar-i2m. This target pins
// framing and round-trip invariants without redefining the error taxonomy.

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// FuzzDecodeRequestWire fuzzes the request wire decoder. Seeds include
// the minimal valid document from the security regression suite, a fully
// populated valid document (reused perf fixture, encoded through the real
// CanonicalEncode path), a valid document with an omitted optional member,
// and one rejection seed per framing class the decoder enforces.
func FuzzDecodeRequestWire(f *testing.F) {
	// Valid: minimal canonical request with trailing newline (owned by
	// security_suffix_test.go; reused so the corpus stays single-sourced).
	f.Add([]byte(securityBaseWire))

	// Valid: fully populated request with a runtime registry, produced by
	// the real canonical encoder so the seed is exactly what the library
	// emits.
	fullReq, err := NewRequest(perfRequestData(perfRuntimeOffers(3)))
	if err != nil {
		f.Fatalf("seed fixture not constructible: %v", err)
	}
	fullEnc, err := CanonicalEncode(fullReq.Data())
	if err != nil {
		f.Fatalf("seed fixture not encodable: %v", err)
	}
	f.Add(append(fullEnc, '\n'))

	// Valid: a member the DTO allows to be absent (empty string is the
	// field's unknown sentinel, and omission decodes to the same zero).
	f.Add([]byte(strings.Replace(securityBaseWire, `"region":"",`, "", 1)))

	// Invalid: one seed per rejection class.
	f.Add([]byte(""))                                                 // empty document
	f.Add(append([]byte{0xEF, 0xBB, 0xBF}, securityBaseWire...))      // BOM
	f.Add([]byte(strings.Replace(securityBaseWire, "\n", "\r\n", 1))) // CR bytes
	f.Add([]byte(strings.TrimSuffix(securityBaseWire, "\n")))         // missing trailing newline
	f.Add([]byte(securityBaseWire + "\n"))                            // double trailing newline
	f.Add([]byte(strings.Replace(securityBaseWire, `"task":"code"`,   // duplicate key
		`"task":"code","task":"code"`, 1)))
	f.Add([]byte(strings.Replace(securityBaseWire, `"task":"code"`, // unknown member
		`"task":"code","extra":1`, 1)))
	f.Add([]byte(securityBaseWire[:len(securityBaseWire)-1] + `{"id":"x"}` + "\n")) // trailing second value
	f.Add([]byte(strings.Replace(securityBaseWire, `"input_tokens":10`,             // float is not canonical
		`"input_tokens":10.5`, 1)))
	f.Add([]byte(strings.Replace(securityBaseWire, `"region":"",`, // null member
		`"region":null,`, 1)))
	f.Add([]byte(strings.Replace(securityBaseWire, // non-canonical time
		`"effective_at":"2026-09-30T12:00:00.000Z"`,
		`"effective_at":"2026-09-30T12:00:00Z"`, 1)))
	f.Add([]byte(securityBaseWire + "}\n")) // stray closer after the value
	f.Add([]byte("[]\n"))                   // wrong top-level type
	f.Add([]byte{0xFF, 0xFE, '\n'})         // invalid UTF-8

	f.Fuzz(func(t *testing.T, data []byte) {
		rd, err := DecodeRequestWire(data)
		if err != nil {
			if !reflect.DeepEqual(rd, RequestData{}) {
				t.Fatalf("DecodeRequestWire returned a non-zero value with an error")
			}
			return
		}
		// Accepted: wire framing only. The round-trip invariant applies
		// where the accepted value is also domain-constructible.
		req, nerr := NewRequest(rd)
		if nerr != nil {
			return
		}
		nd := req.Data()
		enc, cerr := CanonicalEncode(nd)
		if cerr != nil {
			t.Fatalf("accepted, constructible request does not re-encode: %v", cerr)
		}
		again, derr := DecodeRequestWire(append(enc, '\n'))
		if derr != nil {
			t.Fatalf("canonical re-encode of an accepted request failed wire re-decode: %v", derr)
		}
		req2, nerr2 := NewRequest(again)
		if nerr2 != nil {
			t.Fatalf("canonical re-encode of an accepted request is no longer constructible: %v", nerr2)
		}
		if !reflect.DeepEqual(nd, req2.Data()) {
			t.Fatalf("accepted request round-trip is not value-stable")
		}
	})
}

// FuzzDecodeSnapshot fuzzes the snapshot artifact decoder. Seeds include
// the hand-written canonical fixture (never produced by the library), the
// library's own canonical encoding of the shared fixture, and one
// rejection seed per class: bad JSON, empty input, schema mismatch,
// duplicate keys, non-canonical member order, floats, nulls, oversized
// integers, unknown members, trailing values and invalid UTF-8.
func FuzzDecodeSnapshot(f *testing.F) {
	// Valid: hand-written canonical bytes (fixture_test.go).
	f.Add([]byte(fixtureCanonicalBytes))

	// Valid: the library's own canonical encoding of the full fixture.
	fixtureSnap, err := NewSnapshot(fixtureSnapshotData())
	if err != nil {
		f.Fatalf("seed fixture not constructible: %v", err)
	}
	fixtureEnc, err := EncodeSnapshot(fixtureSnap)
	if err != nil {
		f.Fatalf("seed fixture not encodable: %v", err)
	}
	f.Add(fixtureEnc)

	// Invalid: one seed per rejection class.
	f.Add([]byte("nonsense"))                           // not JSON
	f.Add([]byte(""))                                   // empty artifact
	f.Add([]byte(`{"schema":2}`))                       // truncated artifact: missing required members
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // schema mismatch on an otherwise valid artifact
		`"schema":1`, `"schema":2`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // duplicate key
		`"schema":1`, `"schema":1,"schema":1`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // non-canonical member order
		`"created_at":"2026-09-30T12:00:00.000Z","distribution":"local"`,
		`"distribution":"local","created_at":"2026-09-30T12:00:00.000Z"`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // float is not canonical
		`"latency_millis":0`, `"latency_millis":0.5`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // null member
		`"region":""`, `"region":null`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // integer beyond the signed 53-bit range
		`"latency_millis":0`, `"latency_millis":9007199254740993`, 1)))
	f.Add([]byte(strings.Replace(fixtureCanonicalBytes, // unknown member
		`"schema":1`, `"schema":1,"zzz":1`, 1)))
	f.Add([]byte(fixtureCanonicalBytes + fixtureCanonicalBytes)) // trailing second value
	f.Add([]byte{0xFF, 0xFE})                                    // invalid UTF-8

	f.Fuzz(func(t *testing.T, data []byte) {
		s, err := DecodeSnapshot(data)
		if err != nil {
			if s != nil {
				t.Fatalf("DecodeSnapshot returned a snapshot with an error")
			}
			var re *Error
			if !errors.As(err, &re) {
				t.Fatalf("DecodeSnapshot error is not a routing Error: %T", err)
			}
			if re.Code != CodeInvalidSnapshot && re.Code != CodeSchemaIncompatible {
				t.Fatalf("DecodeSnapshot error has an open code %q", re.Code)
			}
			return
		}
		if s == nil {
			t.Fatal("DecodeSnapshot accepted input with a nil snapshot")
		}
		// Accepted artifacts must replay byte-for-byte (already enforced
		// inside DecodeSnapshot; re-asserted so the fuzzer hunts any path
		// that regresses it) and must re-construct to the same hash.
		enc, eerr := EncodeSnapshot(s)
		if eerr != nil {
			t.Fatalf("accepted snapshot does not re-encode: %v", eerr)
		}
		if !bytes.Equal(enc, data) {
			t.Fatal("accepted snapshot does not re-encode to its input bytes")
		}
		s2, nerr := NewSnapshot(s.Data())
		if nerr != nil {
			t.Fatalf("accepted snapshot data is not re-constructible: %v", nerr)
		}
		if s2.Hash() != s.Hash() {
			t.Fatalf("re-constructed snapshot hash %s differs from decoded hash %s", s2.Hash(), s.Hash())
		}
	})
}
