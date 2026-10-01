# Phase 8 acquisition-half adversarial review

Repo: `/Users/guygrigsby/projects/agent-routing` (uncommitted Phase 8 work: 12 policy JSON files, 161 records, 73 terms captures, gate 6, schema 4).

Spec pinned before review: `MD5 (docs/specs/research-artifact.md) = 40f73816611c4e734cfc499750efc817`
Spec re-pinned after review: **identical** (`40f73816611c4e734cfc499750efc817`). No spec drift.

Gate: `bin/agent-routing research check` → `ok`, exit 0. `make check` (build, go test ./..., go vet, research check, backlog) → exit 0.

## VERDICT: FAIL

One blocker (Area G): the spec §2.5 enumerated redistribution allowlist omits `openmdw-1.1` while gate SQL includes it and 1 record depends on it. Fix is a one-token spec amendment. Everything else verified clean against bytes.

## Per-area results

### A. Allowlist license bytes — PASS
- `modified-mit` (moonshot-minimax/kimi-k2-instruct): verbatim "publish, distribute, sublicense, and/or sell" grant in retained bytes; 100M MAU / $20M branding conditions restrict scale of use, not the grant.
- `kimi-k3` (kimi-k3-license.txt): express grant "to deal in the Software without restriction ... publish, distribute, sublicense"; $20M MaaS condition with internal-use exemption (§4a: "internal use of the Software, defined as any use that does not make the Software, its outputs, or its underlying capabilities available to third parties") — byte-verified.
- `openmdw-1.1` (arcee-trinity): "deal in the Model Materials without restriction" with notice-retention condition — byte-verified.
- Conditions do not negate research-repo use.

### B. 9 Class A URL normalizations — PASS
Inventory `url == terms_url` for **all** 161 records (stronger than the 9 required); normalized URLs resolve to the inventoried capture; no orphan evidence paths.

### C. Class B `_shared` curl captures — PASS
11 `_shared` files inventoried; `observed_at == mtime` for all; 14 records cite `_shared` evidence with `retrieved_at == mtime` (the handoff brief's "15" was off-by-one; disk says 14). Note: D verification shows `_shared/mrl-0.1.html` does contain the Ministral-8B quote text — no retrieval shortfall.

### D. 62 worker-fetched terms files — PASS with notes
- sha256 recomputed and matching; timestamps match `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`; closed key sets.
- license_text fidelity: 114 records carry quotes. Initial fragment matcher flagged 14; full resolution below — **no fabricated license text found**:
  - 7 records (kimi-k3, minimax-m3, llama3_1/3_3/4 MODEL_CARD, Llama-3.1-70B HF card, Ministral-8B MRL) are **spliced multi-section excerpts**: fragments split on `...`, blank lines, and literal `\n` are each verbatim in retained bytes (fold: curly→straight quotes, em-dash, nbsp; entity-decoded; terminal elision punctuation only divergence, e.g. kimi-k3 quote ends `parties.` vs file `parties;`).
  - 6 records quote text from a document other than `evidence.path`, but the substantiating license bytes ARE retained and joined via other corpus rows: simple-evals (README quote; evidence holds full license), Mistral Large-3/Nemo/Small-3.1 (model-card quote; evidence holds `apache-2.0.txt`), ImageReward (badge-JSON quote absent from bytes; evidence holds full Apache-2.0 text in rawLines), stepfun (quote is the entity-decoded `cardData` JSON, which IS in the HTML bytes as `&quot;`-escaped).
  - 1 weakest record: **deepseek-v3** — quoted MIT grant text is not in any retained bytes for that record; evidence is the GitHub API JSON (license metadata only, `key: mit`). Decision (`license_name=mit`) rests on retained metadata and is sound; the quote is verbatim of the upstream LICENSE, unretained. Quality item, within spec letter (see below).
- Spec §2.5 says only "`license_text` is a verbatim excerpt of at most 65 536 UTF-8 bytes or the literal `unknown`" — unanchored, does not require the excerpt come from `evidence.path`. All findings are therefore quality notes, not contract violations. Recommend a bd item: prefer quoting `evidence.path` bytes; re-quote deepseek-v3 and ImageReward from retained bytes.

### E. Dedup + 161-URL coverage — PASS
161 records = 161 distinct URLs = 161 distinct `source_access` URLs (set equality both directions); 0 duplicate URLs; 12 policy files ↔ 12 source_access batches exactly; `source_policies` has 161 rows, all fields equal JSON (expires_at = checked_at + 90d). All 161 evidence paths inventoried (416 inventory rows, +64 all under `data/research/terms/`); the 352 pre-existing rows changed only by ms-truncation/zero-fill of already-floored instants.

### F. Timestamp format/mtimes — PASS
Format regex holds across all records and inventory rows; `observed_at`/`retrieved_at` == file mtimes (5/5 spot checks); date-only→midnight legacy rows are the same instant, not rewrites.

### G. Gate 6 SQL vs spec §2.5 — **FAIL (blocker)**
- Spec lines 329–333 enumerate 11 licenses (`mit, modified-mit, kimi-k3, apache-2.0, 0bsd, bsd-2-clause, bsd-3-clause, cc-by-4.0, cc-by-sa-4.0, cdla-permissive-1.0, cdla-permissive-2.0`); gate SQL (`internal/research/gates.go:271`) enumerates 12 — adding `openmdw-1.1`. Spec prose (line 334) admits `openmdw-1.1` as a reviewed fact, and `data/research-inputs/policies/arcee-trinity` carries redistribution=allowed on it, but the normative enumeration omits it.
- **Fix (exact)**: `docs/specs/research-artifact.md` line 333 — extend the enumeration: `` `cdla-permissive-1.0`, `cdla-permissive-2.0`, `openmdw-1.1`. `` (or insert `openmdw-1.1` after `kimi-k3`). Then re-pin md5 in the Phase 8 handoff.
- Minor gate notes (non-blocking, worth a bd sweep): duplicate URL surfaces as a PK-insert iof failure at gate 6 exit 3 instead of a clean load-time exit-1 diagnostic; `evidence_path` FK declared but `PRAGMA foreign_keys` never enabled (mitigated by load-time `CheckReadPath`); unknown-batch-name checked at gate-6 exit 3 while spec lists it as load-time exit 1.

### H. publication_permission mapping — PASS
Terms-inventory rows' `publication_permission` == the record's redistribution decision; no path serves a conflicting decision; non-terms rows all `unknown`; no `allowed`+non-success record lacks the ≥20-char attestation in `license_text`.

## Required fixes (blockers)

1. **Spec enumeration omits `openmdw-1.1`** — `docs/specs/research-artifact.md:333`. Add `` `openmdw-1.1` `` to the §2.5 enumerated redistribution allowlist (prose already admits it as a reviewed fact; gate SQL and the arcee-trinity record already include it). One-token edit, spec-version bump per repo convention.
2. **Dead drift-trap map** — `internal/research/policies.go:42-50`. `redistributionLicenses` has zero references outside its definition and is missing `modified-mit`, `kimi-k3`, `openmdw-1.1` relative to gate SQL. Delete the comment + var (10 lines).

## Non-blocking notes

- license_text sourcing: 7 spliced-excerpt quotes (disclose elisions), 6 quote-outside-evidence cases, deepseek-v3 weakest — bd item to re-anchor quotes in `evidence.path` bytes.
- Gate 6 exit-code nuances (duplicate-URL iof, FK pragma, unknown-batch timing) — bd item.
- `_shared` citation count is 14, not the 15 claimed in the task brief (disk verified).

## Gate status

- `bin/agent-routing research check`: exit 0.
- `make check`: exit 0 (build, vet, tests, check, backlog 395 models / 302 in scope / 28 tasks).
- Spec md5 before = after = `40f73816611c4e734cfc499750efc817`.
- Exit code of this review: 0 (verdict FAIL is a content finding; no infra failure).

---

## Resolution (author, 2026-10-01)

Both blockers fixed as specified:
1. `openmdw-1.1` added to the §2.5 enumerated allowlist (spec line 332); changelog amended.
2. Dead `redistributionLicenses` map deleted from internal/research/policies.go.

Verified: grep enumeration matches gate SQL token set; `go build`, `go test ./internal/research/`, `make check` all green. Non-blocking notes tracked as `ar-orn.4` (license_text quote re-anchoring) and `ar-orn.5` (gate 6 exit-code nuances).
