// Receipt grammar boundary tests for the security review finding
// data/reviews/receipt-boundary-pass1-2026-10-10.txt: the receipt document
// grammar must be closed (exact member casing, no missing or null required
// members, no alias/canonical duplicates) and known usage must carry an
// explicit value, so no zero is ever fabricated from absent evidence.
// Covers both grammar paths: the host boundary ParseReceipt and the direct
// validateReceipt port path. No database is needed.
package collector

import (
	"strings"
	"testing"
)

// omitValue marks a base member to leave out of the document entirely, so
// missing-member shapes can be expressed through the map builder.
type omitValue struct{}

// grammarEntry selects one usage element shape.
type grammarEntry int

const (
	grammarUnknown       grammarEntry = iota // known:false, no value
	grammarKnown                             // known:true with a value
	grammarKnownNoValue                      // known:true, value omitted
	grammarKnownNullValue                    // known:true, value:null
	grammarValueNull                         // known:false, value:null
)

// grammarUsage renders the seven-component usage array in spec order.
func grammarUsage(t *testing.T, entries map[string]grammarEntry) []map[string]any {
	t.Helper()
	var u []map[string]any
	for _, c := range usageComponents {
		switch entries[c] {
		case grammarKnown:
			u = append(u, map[string]any{"component": c, "known": true, "value": int64(5)})
		case grammarKnownNoValue:
			u = append(u, map[string]any{"component": c, "known": true})
		case grammarKnownNullValue:
			u = append(u, map[string]any{"component": c, "known": true, "value": nil})
		case grammarValueNull:
			u = append(u, map[string]any{"component": c, "known": false, "value": nil})
		default:
			u = append(u, map[string]any{"component": c, "known": false})
		}
	}
	return u
}

// grammarDoc builds a canonical valid cancelled receipt with the given
// overrides merged over the base members; omitValue{} deletes a member.
// A nil usage argument leaves the usage member out unless the overrides
// set it explicitly.
func grammarDoc(t *testing.T, overrides map[string]any, usage []map[string]any) []byte {
	t.Helper()
	doc := map[string]any{
		"receipt_id": "rcpt-g", "request_id": "req-g", "offer_id": "prov/up",
		"task": "code", "outcome": "cancelled", "validator_version": "",
	}
	for k, v := range overrides {
		if _, omit := v.(omitValue); omit {
			delete(doc, k)
			continue
		}
		doc[k] = v
	}
	if usage != nil {
		doc["usage"] = usage
	}
	return canonical(t, doc)
}

// splice replaces one exact fragment of a canonical document and asserts
// the result is still canonical, so any later rejection comes from the
// grammar fix and not the framing check. Replacement fragments carry keys
// in canonical (sorted) order; uppercase first letters sort before
// lowercase ones.
func splice(t *testing.T, base []byte, old, new string) []byte {
	t.Helper()
	if !strings.Contains(string(base), old) {
		t.Fatalf("fixture fragment %q not found in %s", old, base)
	}
	out := []byte(strings.Replace(string(base), old, new, 1))
	if !isCanonical(out) {
		t.Fatalf("spliced fixture is not canonical: %s", out)
	}
	return out
}

func TestParseReceiptGrammarRejections(t *testing.T) {
	base := grammarDoc(t, nil, grammarUsage(t, nil))
	aliasReceiptDoc := splice(t, base,
		`"offer_id":"prov/up","outcome":"cancelled","receipt_id":"rcpt-g"`,
		`"Receipt_ID":"rcpt-g","offer_id":"prov/up","outcome":"cancelled"`)
	aliasComponentDoc := splice(t, base,
		`{"component":"input_tokens","known":false}`,
		`{"Component":"input_tokens","known":false}`)
	aliasKnownDoc := splice(t, base,
		`{"component":"input_tokens","known":false}`,
		`{"Known":false,"component":"input_tokens"}`)
	aliasValueDoc := splice(t, base,
		`{"component":"input_tokens","known":false}`,
		`{"Value":5,"component":"input_tokens","known":true}`)
	// Duplicate alias plus canonical member: rejected whether the
	// canonicalizer collapses the duplicates or the decoder aliases them.
	dupDoc := []byte(strings.Replace(string(base), `"receipt_id":"rcpt-g"`,
		`"Receipt_ID":"rcpt-g","receipt_id":"rcpt-g"`, 1))
	missingKnownDoc := grammarDoc(t, nil, func() []map[string]any {
		u := grammarUsage(t, nil)
		for i := range u {
			if u[i]["component"] == "input_tokens" {
				delete(u[i], "known")
				u[i]["value"] = int64(5)
			}
		}
		return u
	}())
	nullKnownDoc := grammarDoc(t, nil, func() []map[string]any {
		u := grammarUsage(t, nil)
		for i := range u {
			if u[i]["component"] == "input_tokens" {
				u[i]["known"] = nil
				u[i]["value"] = int64(5)
			}
		}
		return u
	}())
	nullComponentDoc := grammarDoc(t, nil, func() []map[string]any {
		u := grammarUsage(t, nil)
		u[0]["component"] = nil
		return u
	}())
	nullElementDoc := grammarDoc(t, nil, func() []map[string]any {
		u := grammarUsage(t, nil)
		u[0] = nil
		return u
	}())

	cases := []struct {
		name string
		doc  []byte
	}{
		{"alias receipt_id key", aliasReceiptDoc},
		{"duplicate alias and canonical receipt_id", dupDoc},
		{"missing validator_version member", grammarDoc(t,
			map[string]any{"validator_version": omitValue{}}, grammarUsage(t, nil))},
		{"null validator_version", grammarDoc(t,
			map[string]any{"validator_version": nil}, grammarUsage(t, nil))},
		{"null task", grammarDoc(t, map[string]any{"task": nil}, grammarUsage(t, nil))},
		{"missing task member", grammarDoc(t,
			map[string]any{"task": omitValue{}}, grammarUsage(t, nil))},
		{"null receipt_id", grammarDoc(t, map[string]any{"receipt_id": nil}, grammarUsage(t, nil))},
		{"missing receipt_id member", grammarDoc(t,
			map[string]any{"receipt_id": omitValue{}}, grammarUsage(t, nil))},
		{"usage null", grammarDoc(t, map[string]any{"usage": nil}, nil)},
		{"missing usage member", grammarDoc(t,
			map[string]any{"usage": omitValue{}}, nil)},
		{"known usage without value", grammarDoc(t, nil,
			grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnownNoValue}))},
		{"known usage null value", grammarDoc(t, nil,
			grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnownNullValue}))},
		{"unknown usage null value", grammarDoc(t, nil,
			grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarValueNull}))},
		{"missing known in usage", missingKnownDoc},
		{"null known in usage", nullKnownDoc},
		{"null component in usage", nullComponentDoc},
		{"alias component key in usage", aliasComponentDoc},
		{"alias known key in usage", aliasKnownDoc},
		{"alias value key in usage", aliasValueDoc},
		{"null usage element", nullElementDoc},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseReceipt(tc.doc); err == nil {
				t.Fatal("invalid receipt document accepted at the host boundary")
			} else if codeOf(t, err) != CodeInvalidRequest {
				t.Fatalf("parse rejection code %s, want invalid_request", codeOf(t, err))
			}
		})
	}
}

func TestParseReceiptGrammarAccepts(t *testing.T) {
	// Known zero is an explicit measurement and must be accepted.
	zeroDoc := grammarDoc(t, nil,
		grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnown}))
	zeroDoc = splice(t, zeroDoc,
		`{"component":"input_tokens","known":true,"value":5}`,
		`{"component":"input_tokens","known":true,"value":0}`)
	params, err := ParseReceipt(zeroDoc)
	if err != nil {
		t.Fatalf("known zero value must be accepted: %v", err)
	}
	if string(params.ReceiptBytes) != string(zeroDoc) {
		t.Fatal("parse must preserve the exact input bytes")
	}
	for _, u := range params.Usage {
		if u.Component == "input_tokens" && (!u.Known || u.Value != 0) {
			t.Fatalf("known zero degraded: %+v", u)
		}
	}
	// Unknown without a value stays legal.
	if _, err := ParseReceipt(grammarDoc(t, nil, grammarUsage(t, nil))); err != nil {
		t.Fatalf("unknown usage without value must be accepted: %v", err)
	}
}

func TestParseReceiptSafeErrors(t *testing.T) {
	secret := "SECRET-RECEIPT-CONTENT-7f3a"
	doc := grammarDoc(t, map[string]any{"receipt_id": secret, "task": "not-a-token"},
		grammarUsage(t, nil))
	_, err := ParseReceipt(doc)
	if err == nil {
		t.Fatal("expected rejection")
	}
	if codeOf(t, err) != CodeInvalidRequest {
		t.Fatalf("code %s, want invalid_request", codeOf(t, err))
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("parse error leaked document content")
	}
	// A grammar-shape rejection must also stay content-free. The aliased
	// Task member sorts before the lowercase keys, so it moves to the front.
	noTask := splice(t, doc, `,"task":"not-a-token"`, "")
	aliasDoc := splice(t, noTask, `{"offer_id"`, `{"Task":"not-a-token","offer_id"`)
	if _, err := ParseReceipt(aliasDoc); err == nil {
		t.Fatal("expected rejection for aliased member")
	} else if strings.Contains(err.Error(), secret) {
		t.Fatal("grammar rejection leaked document content")
	}
}

func TestValidateReceiptPortGrammar(t *testing.T) {
	// Direct port path: params and bytes crafted the way a fabricated-zero
	// caller would produce them. No database is touched.
	portParams := func(docBytes []byte, usage []ReceiptUsage, outcome, validatorVersion string) RecordReceiptParams {
		return RecordReceiptParams{
			ReceiptID: "rcpt-g", RequestID: "req-g", OfferID: "prov/up",
			Task: "code", Outcome: outcome, ValidatorVersion: validatorVersion,
			Usage: usage, ReceiptBytes: docBytes, CollectorVersion: "cv",
		}
	}
	unvalidated := func(docBytes []byte, usage []ReceiptUsage) RecordReceiptParams {
		return portParams(docBytes, usage, "cancelled", "")
	}
	knownZeroUsage := allUsage(map[string]int64{"input_tokens": 0})
	// Fabricated zero: the document states known with no value while the
	// submitted fact claims known with value 0. Must be rejected, not
	// recorded as a zero measurement.
	fabricated := grammarDoc(t, nil,
		grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnownNoValue}))
	if _, err := validateReceipt(unvalidated(fabricated, knownZeroUsage), true); err == nil {
		t.Fatal("known usage without a document value must be rejected at the port")
	}
	// Null document value with the same fabricated zero: rejected.
	nullValue := grammarDoc(t, nil,
		grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnownNullValue}))
	if _, err := validateReceipt(unvalidated(nullValue, knownZeroUsage), true); err == nil {
		t.Fatal("known usage with a null document value must be rejected at the port")
	}
	// Aliased member name in the document: rejected even though the decoded
	// facts would match the submitted facts.
	base := grammarDoc(t, nil, grammarUsage(t, nil))
	aliasDoc := splice(t, base,
		`"offer_id":"prov/up","outcome":"cancelled","receipt_id":"rcpt-g"`,
		`"Receipt_Id":"rcpt-g","offer_id":"prov/up","outcome":"cancelled"`)
	if _, err := validateReceipt(unvalidated(aliasDoc, allUsage(nil)), true); err == nil {
		t.Fatal("aliased receipt_id member must be rejected at the port")
	}
	// Missing required member: an unvalidated receipt without the
	// validator_version member is not the same document as one stating it
	// empty; the stored bytes must not diverge from the grammar.
	missingValidator := grammarDoc(t, map[string]any{"validator_version": omitValue{}},
		grammarUsage(t, nil))
	if _, err := validateReceipt(unvalidated(missingValidator, allUsage(nil)), true); err == nil {
		t.Fatal("missing validator_version member must be rejected at the port")
	}
	// Null validator_version on an unvalidated receipt: rejected.
	nullValidator := grammarDoc(t, map[string]any{"validator_version": nil},
		grammarUsage(t, nil))
	if _, err := validateReceipt(unvalidated(nullValidator, allUsage(nil)), true); err == nil {
		t.Fatal("null validator_version must be rejected at the port")
	}
	// Missing usage member on the port path: rejected.
	missingUsage := grammarDoc(t, map[string]any{"usage": omitValue{}}, nil)
	if _, err := validateReceipt(unvalidated(missingUsage, allUsage(nil)), true); err == nil {
		t.Fatal("missing usage member must be rejected at the port")
	}
	// Acceptance: known zero and unknown without value both stay legal on
	// the port path.
	zeroDoc := grammarDoc(t, nil,
		grammarUsage(t, map[string]grammarEntry{"input_tokens": grammarKnown}))
	zeroDoc = splice(t, zeroDoc,
		`{"component":"input_tokens","known":true,"value":5}`,
		`{"component":"input_tokens","known":true,"value":0}`)
	if _, err := validateReceipt(unvalidated(zeroDoc, knownZeroUsage), true); err != nil {
		t.Fatalf("known zero must be accepted at the port: %v", err)
	}
	if _, err := validateReceipt(unvalidated(base, allUsage(nil)), true); err != nil {
		t.Fatalf("unknown usage without value must be accepted at the port: %v", err)
	}
	// Validated outcome with an explicit validator_version stays legal.
	validatedDoc := grammarDoc(t, map[string]any{
		"outcome": "validated_success", "validator_version": "tests-1",
	}, grammarUsage(t, nil))
	if _, err := validateReceipt(portParams(validatedDoc, allUsage(nil), "validated_success", "tests-1"), true); err != nil {
		t.Fatalf("validated receipt must be accepted at the port: %v", err)
	}
}
