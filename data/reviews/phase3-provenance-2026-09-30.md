# Phase 3 provenance review — 2026-09-30

Reviewer: fresh-context pass (z-ai/glm-5.3 via openrouter), adversarial brief at
`/tmp/ar-phase3/review-base.md`, raw report at `/tmp/ar-phase3/review-1.md`.
Scope: the phase 3 provenance artifact set — `data/research-inputs/capture-inventory.tsv`
(340 rows), `data/research-inputs/fetch-logs/` (22 logs), nine corrected
`data/research/*/findings.json`, `data/reviews/decisions/*.v1.json` (7),
`data/research-inputs/build.json`, and the phase 3 spec amendments to
`docs/specs/research-artifact.md`.

## Verdict: PASS (2 defects found, both fixed before commit; 6 notes, 4 actioned)

## Basis

- Inventory: 340 rows, 10 columns, path-sorted; sha256+bytes verified against
  disk for all 340 rows including the 9 corrected findings files; 306
  captured_remote + 2 header_snapshot + 32 local_derived; no remote-origin row
  with empty url; all 32 local_derived urls empty; exactly 77 empty
  `recovery_basis` rows, each with url inherited unchanged from the frozen
  manifest.
- Fetch logs: `### N` command structure intact in all 22 logs; credential scan
  clean against token/key/Bearer/API-key patterns; all 222
  `fetch-log:<file> command <N>` basis references resolved to existing
  commands whose fetched URL matches the evidence path and the inventory url
  (221 parsed, 1 hand-verified: `images/sources/bytedance_seedream_5.html` ←
  images.log command 12).
- Findings diffs (HEAD vs disk, structural walk): exactly 68 `sources[].url`
  provenance repairs (anthropic 13, coverage-followup 7, google 2, images 1,
  mistral-meta 1, moonshot-minimax 5, openai 4, other-text 27, qwen 8),
  4 `benchmark_status` unresearched→found plus 4 paired note rewrites
  (moonshot-minimax: `minimax/minimax-m2`, `moonshotai/kimi-k2`, `kimi-k2.5`,
  `kimi-k3`), 1 status repair `not_found`→`retrieved`
  (`data/research/benchmark-sources/_evalplus_results.json`) plus the paired
  purpose correction fixed after round 1. Zero measurement, value, claim or
  gap changes; moonshot-minimax benchmark rows byte-identical to the frozen
  review. All 158 `sources[].url` values across all batches agree with the
  inventory after repair.
- EvalPlus: `_evalplus_results.json` parses as a JSON object with exactly 125
  reported model labels. The four other not_found JSON captures are genuine
  API error envelopes (GitHub `{"message":"Not Found",...,"status":"404"}` ×3,
  Hugging Face `{"error":"Invalid username or password."}`); the spec gate-3
  predicate over all 26 failure-status rows yields exactly 1 repair + 4 honest
  envelopes.
- Decisions: all 6 `published_measurements` v1 pin sets equal the frozen
  `evidence_review_inputs` rows for their batch; pins match `git show HEAD:`
  hashes; `catalog-sources.v1` pins match current disk.
- build.json: expected 395/394/302/31659/12 match the database; batch_formats
  keys equal the 15 batch dirs.
- Spec amendments match artifact behavior (control-file lane, gate-3 envelope
  exception, delta counts).

## Round 1 findings and dispositions

1. DEFECT (fixed): stale EvalPlus `purpose` claimed 404 HTML over the JSON
   scoreboard; corrected in `coverage-followup/findings.json` in the same
   reviewed class as the note rewrites.
2. DEFECT (fixed): §7 delta undercounted URL changes as 227; actual differing
   rows are 256 = 227 unknown/pointer (159 recovered + 32 to-empty + 36
   pointer text) + 29 fetch-log-backed page-vs-raw corrections. Spec amended.
3. NOTE (actioned): `recovery_basis` semantics for original-url rows pinned in
   §2.1 (basis optional for inherited urls, e.g. header corroboration).
4. NOTE (actioned): fetch-log basis grammar pinned in §2.1
   (`fetch-log:<file> command <N> url+filename match`), enabling phase 4
   machine validation.
5. NOTE (accepted): gate 4 reads v1 pins as stale until v2 decisions land;
   expected mid-phase sequencing, v2 records written with this report.
6. NOTE (accepted): three fetch logs contain python urllib fetches, not curl;
   spec wording already says fetch-command records.
7. NOTE (actioned): the 9 findings `source_files` sha/bytes/observed_at deltas
   added to the §7 delta list.
8. NOTE (accepted): phase-3 inputs are covered by the integrity story once
   this commit lands; committed together with this report.

## Bound

This review validates capture provenance, byte identities and the honesty of
the corrections. It does not accept benchmarks for routing (`accepted_benchmarks`
stays empty), does not establish revision identity for unresolved labels, and
does not grant publication rights.
