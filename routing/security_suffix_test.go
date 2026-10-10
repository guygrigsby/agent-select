package routing

// Security regression cases for the one-value wire rule (review round 1,
// isolated probe at 2ad2ceb): json.Decoder.More() returns false for
// trailing '}' and ']' bytes, so strictCheck accepted a valid document
// followed by stray closers or by closers plus a second value. A hostile
// document must fail the strict one-value check without echoing any of
// its bytes into the diagnostic.

import (
	"strings"
	"testing"
)

// wireSuffixCanary is the hostile payload planted after a valid request
// document: its member name and value must never appear in any error.
const wireSuffixCanary = "credential"

// securityBaseWire is a minimal canonical-v1 request wire document that
// passes DecodeRequestWire and NewRequest on its own.
const securityBaseWire = `{"allowed_providers":[],"budget_known":true,"budget_micros":5000000,"deadline_millis":0,"effective_at":"2026-09-30T12:00:00.000Z","id":"r1","image_bytes":0,"image_count":0,"image_height":0,"image_operation":"none","image_width":0,"input_tokens":10,"mask":false,"mask_alpha":false,"max_age_seconds":0,"minimum_estimate":0,"mode":"cost","output_tokens":5,"privacy":"","reference_count":0,"region":"","required_capabilities":[],"runtime":[],"runtime_generation":"rg1","searches":0,"strict_cost":false,"strict_latency":false,"task":"code"}` + "\n"

// securityPoisonedSuffixes are byte sequences appended after a complete
// JSON value that a one-value strict check must reject. '}{'-style
// closers specifically defeat More(); the plain second value is the
// control for the already-rejected class.
var securityPoisonedSuffixes = []string{
	"}\n",
	"]\n",
	"} " + `{"credential":"secret"}` + "\n",
	" " + `{"credential":"secret"}` + "\n",
}

func TestStrictCheckRejectsStrayClosingSuffixes(t *testing.T) {
	base := `{"a":1}`
	if err := strictCheck([]byte(base + "\n")); err != nil {
		t.Fatalf("valid single document rejected: %v", err)
	}
	for _, suf := range securityPoisonedSuffixes {
		err := strictCheck([]byte(base + suf))
		if err == nil {
			t.Fatalf("suffix %q accepted after a complete value", suf)
		}
		if strings.Contains(err.Error(), wireSuffixCanary) || strings.Contains(err.Error(), "secret") {
			t.Fatalf("diagnostic echoes untrusted bytes for suffix %q: %v", suf, err)
		}
	}
}

func TestDecodeRequestWireRejectsPoisonedSuffixes(t *testing.T) {
	if rd, err := DecodeRequestWire([]byte(securityBaseWire)); err != nil {
		t.Fatalf("valid wire rejected: %v", err)
	} else if _, err := NewRequest(rd); err != nil {
		t.Fatalf("valid request not constructible: %v", err)
	}
	for _, suf := range securityPoisonedSuffixes {
		rd, err := DecodeRequestWire([]byte(securityBaseWire + suf))
		if err == nil {
			if _, nerr := NewRequest(rd); nerr == nil {
				t.Fatalf("poisoned wire document passed DecodeRequestWire and NewRequest (suffix %q)", suf)
			}
		}
		if err != nil && (strings.Contains(err.Error(), wireSuffixCanary) || strings.Contains(err.Error(), "secret")) {
			t.Fatalf("diagnostic echoes untrusted bytes for suffix %q: %v", suf, err)
		}
	}
}

// TestStrictCheckStillAcceptsCanonicalSnapshot pins the green path: the
// stricter EOF check must not reject real canonical snapshots or requests.
func TestStrictCheckStillAcceptsCanonicalSnapshot(t *testing.T) {
	if _, err := DecodeSnapshot([]byte(fixtureCanonicalBytes)); err != nil {
		t.Fatalf("valid snapshot artifact rejected: %v", err)
	}
}
