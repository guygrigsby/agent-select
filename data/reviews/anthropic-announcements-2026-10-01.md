# Anthropic announcement acquisitions review

## Result

Round 1: FAIL (values all PASS; three metadata defects).
Round 2: PASS after verbatim-specified fixes; re-verified below.

## Round 1 verdict basis

All 113 cells (112 linked + 1 unresolved Mythos 5.1) reproduce byte-exactly
from the five retained captures: values, benchmark row labels, column
attributions, qualifiers, units, metric interpretations, footnotes, and the
claude-4 prose offsets. Full method and evidence at
/tmp/ar-phase8c/review.md (archived copy below).

## Round 1 defects and resolution

1. evidence_class misclassification, 35 cells: anthropic models in
   anthropic's own grids (older-generation comparison columns) were marked
   competitor_reported. The pages' embedded competitorModels fields list GPT
   columns only. Fixed: all 35 reclassified to provider_reported; the 21
   remaining competitor_reported cells are all openai/* columns.
2. Grid locator byte offsets pointed 308-554 bytes before the CSV start
   (caption region). Fixed: all 100 grid locators repointed to the actual
   CSV byte offsets (opus 557031, sonnet 352550, fable 479480).
3. Sonnet-page markers 1 and 4 have no on-page footnote definitions; glosses
   borrowed sibling-page prose. Fixed: four configuration fields now state
   the markers are undefined on-page, note the sibling opus xhigh-effort
   statement in tension, and record that GPT-6 Sol identity rests on the
   grid column header with the page's GPT-5.6 Sol substitution note.

No cell values, units or benchmark names changed. This review validates
source fidelity only: it does not approve comparison, calibration, rights,
current offer mappings or accepted_benchmarks promotion.
