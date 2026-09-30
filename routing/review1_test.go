package routing

import (
	"strings"
	"testing"
	"time"
)

// Review round 1 regressions: error-code mapping, free-form bounds, year
// bounds, encoded-artifact bound on construction, request integer bounds.

func TestNewSnapshotErrorCodes(t *testing.T) {
	// Schema/policy mismatch reports schema_incompatible.
	d := fixtureSnapshotData()
	d.Schema = 2
	_, err := NewSnapshot(d)
	if e, ok := err.(*Error); !ok || e.Code != CodeSchemaIncompatible {
		t.Fatalf("schema: want schema_incompatible, got %v", err)
	}
	d = fixtureSnapshotData()
	d.Policy = "other"
	_, err = NewSnapshot(d)
	if e, ok := err.(*Error); !ok || e.Code != CodeSchemaIncompatible {
		t.Fatalf("policy: want schema_incompatible, got %v", err)
	}
	// Field violations report invalid_snapshot.
	d = fixtureSnapshotData()
	d.Offers[0].LatencyMillis = -1
	_, err = NewSnapshot(d)
	if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
		t.Fatalf("field: want invalid_snapshot, got %v", err)
	}
	// Decode path keeps the same mapping.
	enc, err := EncodeSnapshot(mustSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeSnapshot(enc); err != nil {
		t.Fatalf("valid artifact rejected: %v", err)
	}
	// EncodeSnapshot state errors report invalid_snapshot.
	if _, err := EncodeSnapshot(nil); err == nil {
		t.Fatal("nil encoded")
	} else if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
		t.Fatalf("nil: want invalid_snapshot, got %v", err)
	}
}

func TestFreeFormBounds(t *testing.T) {
	bad := func(mutate func(*SnapshotData)) {
		t.Helper()
		d := fixtureSnapshotData()
		mutate(&d)
		if _, err := NewSnapshot(d); err == nil {
			t.Fatal("accepted")
		} else if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
			t.Fatalf("want invalid_snapshot, got %v", err)
		}
	}
	bad(func(d *SnapshotData) { d.Offers[0].Region = strings.Repeat("r", 513) })
	bad(func(d *SnapshotData) { d.Offers[0].ModelID = strings.Repeat("m", 513) })
	bad(func(d *SnapshotData) { d.Offers[0].Revision = strings.Repeat("v", 513) })
	bad(func(d *SnapshotData) { d.Offers[0].Privacy = "no\x01control" })
	bad(func(d *SnapshotData) { d.Offers[0].Region = " padded " })
	bad(func(d *SnapshotData) { d.Offers[0].ModelID = "\xff\xfe" }) // invalid UTF-8
	bad(func(d *SnapshotData) { d.Inputs[0].ObservationID = strings.Repeat("o", 513) })
	bad(func(d *SnapshotData) { d.Inputs[0].SourceID = "src\x7f" })
	bad(func(d *SnapshotData) { d.Inputs[0].PolicyVersion = " pv" })
	bad(func(d *SnapshotData) { d.Offers[0].Signals[0].EvidenceID = strings.Repeat("e", 513) })
	bad(func(d *SnapshotData) { d.Offers[0].Signals[0].Revision = "rev\x02" })
	// Boundary: exactly 512 bytes is accepted.
	d := fixtureSnapshotData()
	d.Offers[0].Region = strings.Repeat("r", 512)
	if _, err := NewSnapshot(d); err != nil {
		t.Fatalf("512-byte region rejected: %v", err)
	}
	// Composite offer ID: 512-byte provider + 512-byte upstream = 1025 bytes,
	// within the composite bound.
	d = fixtureSnapshotData()
	p, u := strings.Repeat("a", 512), strings.Repeat("b", 512)
	d.Offers[1].Provider, d.Offers[1].UpstreamID, d.Offers[1].ID = p, u, p+"/"+u
	if _, err := NewSnapshot(d); err != nil {
		t.Fatalf("1025-byte composite id rejected: %v", err)
	}
}

func TestRequestFreeFormBounds(t *testing.T) {
	bad := func(mutate func(*RequestData)) {
		t.Helper()
		d := fixtureRequestData()
		mutate(&d)
		if _, err := NewRequest(d); err == nil {
			t.Fatal("accepted")
		}
	}
	bad(func(d *RequestData) { d.Region = strings.Repeat("r", 513) })
	bad(func(d *RequestData) { d.Privacy = "p\x01" })
	bad(func(d *RequestData) { d.Runtime[0].Revision = strings.Repeat("v", 513) })
	bad(func(d *RequestData) { d.Runtime[0].Region = " r" })
	bad(func(d *RequestData) { d.Runtime[0].Privacy = "\xff" })
}

func TestYearBounds(t *testing.T) {
	for _, y := range []int{12000, 0, -500} {
		d := fixtureSnapshotData()
		d.CreatedAt = time.Date(y, 9, 30, 12, 0, 0, 0, time.UTC)
		if _, err := NewSnapshot(d); err == nil {
			t.Fatalf("year %d accepted", y)
		}
		rd := fixtureRequestData()
		rd.EffectiveAt = time.Date(y, 9, 30, 12, 0, 0, 0, time.UTC)
		if _, err := NewRequest(rd); err == nil {
			t.Fatalf("request year %d accepted", y)
		}
	}
	for _, y := range []int{1, 9999} {
		d := fixtureSnapshotData()
		d.CreatedAt = time.Date(y, 9, 30, 12, 0, 0, 0, time.UTC)
		d.KnowledgeCutoff = d.CreatedAt
		d.Offers[0].ConfirmedAt = time.Time{} // year-1 cutoff precedes fixture confirmation
		if _, err := NewSnapshot(d); err != nil {
			t.Fatalf("year %d rejected: %v", y, err)
		}
	}
}

func TestNewSnapshotRejectsOversizeArtifact(t *testing.T) {
	if testing.Short() {
		t.Skip("heavy artifact test")
	}
	ts := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	obs := strings.Repeat("o", 512)
	src := strings.Repeat("s", 512)
	pv := strings.Repeat("p", 512)
	in := make([]InputData, 0, 45000)
	for i := range 45000 {
		in = append(in, InputData{
			ObservationID: obs + "-" + pad6(i), SourceID: src, PolicyVersion: pv,
		})
	}
	d := SnapshotData{Schema: 1, CreatedAt: ts, KnowledgeCutoff: ts,
		Policy: "educated-v1", Distribution: "local", Inputs: in}
	_, err := NewSnapshot(d)
	if err == nil {
		t.Fatal("oversize artifact accepted by NewSnapshot")
	}
	if e, ok := err.(*Error); !ok || e.Code != CodeInvalidSnapshot {
		t.Fatalf("want invalid_snapshot, got %v", err)
	}
}

func pad6(i int) string {
	s := func(n int) string {
		if n == 0 {
			return "0"
		}
		var b []byte
		for n > 0 {
			b = append([]byte{byte('0' + n%10)}, b...)
			n /= 10
		}
		return string(b)
	}(i)
	for len(s) < 6 {
		s = "0" + s
	}
	return s
}

func TestRequestIntegerUpperBounds(t *testing.T) {
	fields := []func(*RequestData){
		func(d *RequestData) { d.InputTokens = canonicalIntMax + 1 },
		func(d *RequestData) { d.OutputTokens = canonicalIntMax + 1 },
		func(d *RequestData) { d.ImageBytes = canonicalIntMax + 1 },
		func(d *RequestData) { d.Searches = canonicalIntMax + 1 },
		func(d *RequestData) { d.BudgetKnown = true; d.BudgetMicros = canonicalIntMax + 1 },
		func(d *RequestData) { d.DeadlineMillis = canonicalIntMax + 1 },
		func(d *RequestData) { d.MaxAgeSeconds = canonicalIntMax + 1 },
	}
	for i, f := range fields {
		d := fixtureRequestData()
		f(&d)
		if _, err := NewRequest(d); err == nil {
			t.Fatalf("field %d over 53-bit accepted", i)
		}
	}
}

func TestC1ControlRejection(t *testing.T) {
	// Unicode C1 controls (U+0080-U+009F) are control characters and must
	// be rejected even though they are multi-byte UTF-8 and not ASCII.
	for _, c := range []rune{'\u0080', '\u0081', '\u0090', '\u009f'} {
		d := fixtureSnapshotData()
		d.Offers[0].Region = "r" + string(c) + "g"
		if _, err := NewSnapshot(d); err == nil {
			t.Fatalf("C1 control %q accepted in region", c)
		}
		rd := fixtureRequestData()
		rd.Privacy = string(c)
		if _, err := NewRequest(rd); err == nil {
			t.Fatalf("C1 control %q accepted in request privacy", c)
		}
	}
}
