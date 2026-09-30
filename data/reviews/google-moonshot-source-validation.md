# Source review

## Result

- Google: **PASS** (30/30 source cells)
- Moonshot/MiniMax: **PASS** (12/12 source cells)

All reviewed values, labels, rows/columns, metrics, units, benchmark versions, and recorded configuration footnotes match the saved source HTML/READMEs.

## Finding

The Google worker audit incorrectly called records 1, 2, and 11 invalid (and similarly questioned record 8) because it searched raw HTML for `evalResults`. `hf-gemma-4-26b.html` contains HTML-escaped JSON `evalResults` in `data-props` (four occurrences). After HTML unescaping/parsing, these are genuine source fields; the records should be retained and their evalResults locators are not fabricated.

“No think mode specified” remains unknown, not no-think. The Moonshot/MiniMax records correctly preserve exact harness, effort, tools, sampling/average counts, pass metrics, and benchmark versions where the saved sources state them; unstated details remain unknown.

No source-row edits are required. This review does not approve comparison/calibration/rights, current Vercel bindings, or exact immutable evaluated revisions.
