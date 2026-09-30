package routing

import (
	"strings"
	"testing"
)

func TestNestedUnknownMembersRejected(t *testing.T) {
	// Unknown member inside a nested offer object: OfferData.UnmarshalJSON
	// runs its own decoder, which must also reject unknown members.
	b := strings.Replace(fixtureCanonicalBytes,
		`"capabilities":["text_input"]`, `"capabilities":["text_input"],"surprise":1`, 1)
	if _, err := DecodeSnapshot([]byte(b)); err == nil {
		t.Fatal("unknown member inside offer accepted")
	}
	// Unknown member inside a nested limit object.
	b = strings.Replace(fixtureCanonicalBytes,
		`"limits":[]`, `"limits":[{"maximum":5,"resource":"input_tokens","x":1}]`, 1)
	if _, err := DecodeSnapshot([]byte(b)); err == nil {
		t.Fatal("unknown member inside limit accepted")
	}
}

func TestRequestUnknownMembersRejected(t *testing.T) {
	// A request DTO decoded through strictUnmarshal must reject unknown
	// members at the top level and inside runtime offers.
	base := `{"allowed_providers":[],"budget_known":true,"budget_micros":5000000,"deadline_millis":0,"effective_at":"2026-09-30T12:00:00.000Z","id":"r1","image_bytes":0,"image_count":0,"image_height":0,"image_operation":"none","image_width":0,"input_tokens":10,"mask":false,"mask_alpha":false,"max_age_seconds":0,"minimum_estimate":0,"mode":"cost","output_tokens":5,"privacy":"","reference_count":0,"region":"","required_capabilities":[],"runtime":[{"capabilities":[],"eligible":true,"id":"p/m","limits":[],"prices":[],"privacy":"","provider":"p","region":"","revision":"","unsupported_capabilities":[],"upstream_id":"m"}],"runtime_generation":"rg1","searches":0,"strict_cost":false,"strict_latency":false,"task":"code"}`

	var d RequestData
	if err := strictUnmarshal([]byte(base), &d); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if _, err := NewRequest(d); err != nil {
		t.Fatalf("valid request not constructible: %v", err)
	}

	bad := strings.Replace(base, `"task":"code"`, `"task":"code","extra":1`, 1)
	if err := strictUnmarshal([]byte(bad), &d); err == nil {
		t.Fatal("unknown top-level request member accepted")
	}
	bad = strings.Replace(base, `"upstream_id":"m"}]`, `"upstream_id":"m","extra":1}]`, 1)
	if err := strictUnmarshal([]byte(bad), &d); err == nil {
		t.Fatal("unknown runtime member accepted")
	}
	bad = strings.Replace(base, `"effective_at":"2026-09-30T12:00:00.000Z"`, `"effective_at":"2026-09-30T12:00:00Z"`, 1)
	if err := strictUnmarshal([]byte(bad), &d); err == nil {
		t.Fatal("non-canonical request time accepted")
	}
}
