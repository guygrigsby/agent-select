# Routing error security remediation

Independent read-only review of committed baseline `2ad2ceb` returned FAIL. Original phase 11 security signoff was unavailable. This is a new review, not a reconstruction of a historical verdict.

Confirmed in an isolated archive under `/tmp/routing-security-gate.zccwZ5`:

- `routing/canonical.go` used `Decoder.More()` instead of checking EOF after the root value. Valid request JSON followed by a stray closing delimiter or another document passed `DecodeRequestWire` and `NewRequest`. CLI probe exited 0 and emitted a decision instead of rejecting the wire.
- `cmd/agent-routing/adapter.go` forwarded raw snapshot-store errors. A regular-file store path with an embedded newline leaked its basename and produced two stderr lines.

Probes: `go test ./routing -count=1 -run '^TestSecurityGateMalformedSuffix$' -v` failed all three rejection cases. `go test ./cmd/agent-routing -count=1 -run '^TestSecurityGate' -v` confirmed both the exit-0 malformed request and multiline path disclosure.

Existing focused safety, decoder, validator, CLI exit and adoption-retention tests passed. `go test -race ./routing -count=1 -run '^TestAdoptRace$'` passed. No adoption fail-open was confirmed. Temporary probes were removed from the isolated archive; no main-tree changes came from this review.

Status: remediation pending. Security approval remains withheld until the committed regression tests, fixes and a fresh review pass.
