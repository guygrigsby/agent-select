# Current core library review

Do not accept the blanket PASS on `2ad2cebf26faa1c056fe98751bd3e4ce599106b4`. Preserve the [original review](core-library-current-2026-10-10.txt) as a new review, not recovered historical signoff.

Parent reran the eight [reviewer probes](core-library-probes-2026-10-10.txt) under the race detector: [result](core-library-probes-result-2026-10-10.txt). Those probes passed. Their freshness probe tests age zero and 1.5 seconds against a one-second bound, not inclusive equality.

Parent's [public wire probe](core-library-boundary-probe-2026-10-10.txt) contradicts the blanket trailing-value claim: [result](core-library-boundary-result-2026-10-10.txt). On the same baseline, `DecodeRequestWire` followed by `NewRequest` accepts each of `}\n`, `]\n` and `} {"credential":"secret"}\n` appended after a valid request object.

Reproduce in a disposable worktree at the baseline: copy the probe into `routing/zz_parent_boundary_test.go`, then run `go test ./routing -run '^TestParentBaselineWireSuffixRejected$' -count=1`. Expect three failures on that baseline. Current EOF remediation is separate and requires its own review.

This review alone does not justify closing `ar-0h0` or phase 11. Error-detail secrecy, new performance tests and fuzz coverage were excluded from this review. Reviewer resumption was unavailable, so no corrected reviewer verdict is claimed.
