package routing

import (
	"strings"
	"testing"
)

// The fixture bytes are hand-written, independent of the encoder. Decoding
// must succeed and re-encode to identical bytes.
func TestDecodeIndependentFixture(t *testing.T) {
	s, err := DecodeSnapshot([]byte(fixtureCanonicalBytes))
	if err != nil {
		t.Fatalf("independent fixture rejected: %v", err)
	}
	enc, err := EncodeSnapshot(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(enc) != fixtureCanonicalBytes {
		t.Fatalf("re-encode differs:\n got %s\nwant %s", enc, fixtureCanonicalBytes)
	}
	if s.Hash() != hashCanonical(enc) {
		t.Fatal("hash is not SHA-256 over canonical bytes")
	}
	if len(s.Data().Offers) != 1 || s.Data().Offers[0].ID != "p/m" {
		t.Fatalf("decoded data wrong: %+v", s.Data())
	}
}

func TestDecodeRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"duplicate keys", `{"created_at":"2026-09-30T12:00:00.000Z","created_at":"2026-09-30T12:00:00.000Z","distribution":"local","inputs":[],"knowledge_cutoff":"2026-09-30T12:00:00.000Z","offers":[],"policy":"educated-v1","schema":1}`, "duplicate"},
		{"duplicate nested keys", strings.Replace(fixtureCanonicalBytes, `"latency_millis":0,"limits":[]`, `"latency_millis":0,"latency_millis":0,"limits":[]`, 1), "duplicate"},
		{"unknown member", strings.Replace(fixtureCanonicalBytes, `"schema":1}`, `"schema":1,"extra":1}`, 1), "unknown"},
		{"trailing value", fixtureCanonicalBytes + ` 42`, "trailing"},
		{"trailing object", fixtureCanonicalBytes + `{"a":1}`, "trailing"},
		{"null value", strings.Replace(fixtureCanonicalBytes, `"inputs":[]`, `"inputs":null`, 1), "null"},
		{"float value", strings.Replace(fixtureCanonicalBytes, `"latency_millis":0`, `"latency_millis":0.5`, 1), "float"},
		{"exponent integer", strings.Replace(fixtureCanonicalBytes, `"latency_millis":0`, `"latency_millis":1e2`, 1), "not a canonical-v1 integer"},
		{"over 53-bit", strings.Replace(fixtureCanonicalBytes, `"latency_millis":0`, `"latency_millis":9007199254740992`, 1), "53-bit"},
		{"negative latency invariant", strings.Replace(fixtureCanonicalBytes, `"latency_millis":0`, `"latency_millis":-9007199254740991`, 1), "negative"},
		{"time missing millis", strings.Replace(fixtureCanonicalBytes, `"2026-09-30T12:00:00.000Z"`, `"2026-09-30T12:00:00Z"`, 1), "canonical format"},
		{"time micros", strings.Replace(fixtureCanonicalBytes, `"2026-09-30T12:00:00.000Z"`, `"2026-09-30T12:00:00.000500Z"`, 1), "canonical format"},
		{"time offset", strings.Replace(fixtureCanonicalBytes, `"2026-09-30T12:00:00.000Z"`, `"2026-09-30T12:00:00.000+01:00"`, 1), "canonical format"},
		{"time lowercase z", strings.Replace(fixtureCanonicalBytes, `"2026-09-30T12:00:00.000Z"`, `"2026-09-30T12:00:00.000z"`, 1), "canonical format"},
		{"time date only", strings.Replace(fixtureCanonicalBytes, `"2026-09-30T12:00:00.000Z"`, `"2026-09-30"`, 1), "canonical format"},
		{"truncated", fixtureCanonicalBytes[:len(fixtureCanonicalBytes)-5], ""},
		{"empty object", `{}`, "decode failed"},
		{"array root", `[]`, "decode failed"},
		{"string root", `"x"`, "decode failed"},
		{"not json", `nonsense`, "invalid JSON"},
	}
	for _, tc := range cases {
		_, err := DecodeSnapshot([]byte(tc.in))
		if err == nil {
			t.Errorf("%s: accepted", tc.name)
			continue
		}
		if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", tc.name, err.Error(), tc.want)
		}
	}
}

func TestDecodeSchemaIncompatible(t *testing.T) {
	// Schema 2 parses as strict JSON but reports schema_incompatible.
	b := strings.Replace(fixtureCanonicalBytes, `"schema":1}`, `"schema":2}`, 1)
	_, err := DecodeSnapshot([]byte(b))
	if err == nil {
		t.Fatal("schema 2 accepted")
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeSchemaIncompatible {
		t.Fatalf("want schema_incompatible, got %v", err)
	}
	// Wrong policy is also an incompatibility.
	b = strings.Replace(fixtureCanonicalBytes, `"policy":"educated-v1"`, `"policy":"educated-v2"`, 1)
	_, err = DecodeSnapshot([]byte(b))
	if err == nil {
		t.Fatal("policy educated-v2 accepted")
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeSchemaIncompatible {
		t.Fatalf("want schema_incompatible, got %v", err)
	}
}

func TestDecodeRejectsNonCanonicalOrdering(t *testing.T) {
	// Same members, different (valid JSON) key order: not canonical bytes.
	b := `{"schema":1,"policy":"educated-v1","distribution":"local","created_at":"2026-09-30T12:00:00.000Z","knowledge_cutoff":"2026-09-30T12:00:00.000Z","inputs":[],"offers":[]}`
	_, err := DecodeSnapshot([]byte(b))
	if err == nil {
		t.Fatal("non-canonical key order accepted")
	}
}

func TestDecodeRejectsOversizedArtifact(t *testing.T) {
	big := make([]byte, MaxArtifactBytes+1)
	if _, err := DecodeSnapshot(big); err == nil {
		t.Fatal("oversized artifact accepted")
	} else if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
		t.Fatalf("want invalid_snapshot, got %v", err)
	}
}

func TestDecodeRejectsInvariantDefect(t *testing.T) {
	// Strictly canonical JSON whose values violate an invariant:
	// duplicate offer ids. The error must surface as invalid_snapshot, not
	// invalid_request.
	b := `{"created_at":"2026-09-30T12:00:00.000Z","distribution":"local","inputs":[],"knowledge_cutoff":"2026-09-30T12:00:00.000Z","offers":[{"capabilities":[],"id":"p/m","latency_millis":0,"limits":[],"model_id":"","prices":[],"privacy":"","provider":"p","region":"","revision":"","signals":[],"upstream_id":"m"},{"capabilities":[],"id":"p/m","latency_millis":0,"limits":[],"model_id":"","prices":[],"privacy":"","provider":"p","region":"","revision":"","signals":[],"upstream_id":"m"}],"policy":"educated-v1","schema":1}`
	_, err := DecodeSnapshot([]byte(b))
	if err == nil {
		t.Fatal("duplicate offers accepted")
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
		t.Fatalf("want invalid_snapshot, got %v", err)
	}
}

func TestCorruptBytesRejected(t *testing.T) {
	orig, err := EncodeSnapshot(mustSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	for i := 10; i < len(orig); i += 97 {
		corrupt := append([]byte(nil), orig...)
		corrupt[i] ^= 0x20
		s, err := DecodeSnapshot(corrupt)
		if err != nil {
			continue // rejected: fine
		}
		// A flip inside an unconstrained string can yield different but
		// self-consistent canonical bytes. The guarantee is then that the
		// decoded snapshot re-encodes to exactly those bytes (no silent
		// normalization) and hashes them.
		re, err := EncodeSnapshot(s)
		if err != nil {
			t.Fatalf("corruption at byte %d: decoding succeeded but re-encode failed: %v", i, err)
		}
		if string(re) != string(corrupt) {
			t.Fatalf("corruption at byte %d: decoded value re-encodes differently", i)
		}
		if s.Hash() != hashCanonical(corrupt) {
			t.Fatalf("corruption at byte %d: hash does not match decoded bytes", i)
		}
	}
}

func mustSnapshot(t *testing.T) *Snapshot {
	t.Helper()
	s, err := NewSnapshot(fixtureSnapshotData())
	if err != nil {
		t.Fatal(err)
	}
	return s
}
