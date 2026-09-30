package research

import (
	"fmt"
	"sort"
	"strings"
)

// CoverageRow is one coverage TSV row (§1).
type CoverageRow struct {
	ModelID                string
	Kind                   string
	Batches                string
	BenchmarkCoverage      string
	Measurements           int64
	UnresolvedMeasurements int64
	LocatedClaims          int64
	UnlocatedClaims        int64
}

// CoverageHeader is the TSV header line.
const CoverageHeader = "model_id\tkind\tbatches\tbenchmark_coverage\tmeasurements\tunresolved_measurements\tlocated_claims\tunlocated_claims"

// ComputeCoverage derives the coverage inventory from the assembled build
// facts (never the published DB). Column semantics per §1; coverage values
// mirror the schema 3 coverage view.
func ComputeCoverage(res *BuildResult, tables []*tableRows) []CoverageRow {
	var tModels, tResearch, tBench, tProp, tClaims, tUnloc *tableRows
	for _, t := range tables {
		switch t.name {
		case "catalog_models":
			tModels = t
		case "model_research":
			tResearch = t
		case "benchmark_measurements":
			tBench = t
		case "unresolved_identity_proposals":
			tProp = t
		case "claims":
			tClaims = t
		case "unlocated_claims":
			tUnloc = t
		}
	}

	batches := map[string][]string{}
	notFound := map[string]bool{}
	researched := map[string]bool{}
	for _, r := range tResearch.rows {
		mid, batch, status := r[0].(string), r[1].(string), r[2].(string)
		batches[mid] = append(batches[mid], batch)
		researched[mid] = true
		if status == "not_found_in_checked_sources" {
			notFound[mid] = true
		}
	}
	hasMeasurements := map[string]bool{}
	measureCount := map[string]int64{}
	for _, r := range tBench.rows {
		mid := r[1].(string)
		hasMeasurements[mid] = true
		measureCount[mid]++
	}
	unresCount := map[string]int64{}
	for _, r := range tProp.rows {
		mid := r[1].(string)
		unresCount[mid]++
	}
	claimCount := map[string]int64{}
	for _, r := range tClaims.rows {
		claimCount[r[1].(string)]++
	}
	unlocCount := map[string]int64{}
	for _, r := range tUnloc.rows {
		unlocCount[r[0].(string)]++
	}

	var rows []CoverageRow
	for _, r := range tModels.rows {
		mid, kind := r[0].(string), r[3].(string)
		cov := "out_of_research_scope"
		switch {
		case hasMeasurements[mid]:
			cov = "scores_found"
		case notFound[mid]:
			cov = "no_scores_in_checked_sources"
		case researched[mid]:
			cov = "unresearched"
		}
		bs := append([]string(nil), batches[mid]...)
		sort.Strings(bs)
		rows = append(rows, CoverageRow{
			ModelID:                mid,
			Kind:                   kind,
			Batches:                strings.Join(bs, ";"),
			BenchmarkCoverage:      cov,
			Measurements:           measureCount[mid],
			UnresolvedMeasurements: unresCount[mid],
			LocatedClaims:          claimCount[mid],
			UnlocatedClaims:        unlocCount[mid],
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ModelID < rows[j].ModelID })
	return rows
}

// RenderCoverageTSV renders the machine-readable TSV (LF, trailing newline).
func RenderCoverageTSV(rows []CoverageRow) []byte {
	var b strings.Builder
	b.WriteString(CoverageHeader)
	b.WriteByte('\n')
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\n",
			r.ModelID, r.Kind, r.Batches, r.BenchmarkCoverage,
			r.Measurements, r.UnresolvedMeasurements, r.LocatedClaims, r.UnlocatedClaims)
	}
	return []byte(b.String())
}

// RenderCoverageSummary renders the human summary for stderr.
func RenderCoverageSummary(rows []CoverageRow) string {
	buckets := map[string]int{}
	for _, r := range rows {
		buckets[r.BenchmarkCoverage]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "models: %d (", len(rows))
	var parts []string
	for _, k := range []string{"scores_found", "no_scores_in_checked_sources", "unresearched", "out_of_research_scope"} {
		parts = append(parts, fmt.Sprintf("%s %d", k, buckets[k]))
	}
	b.WriteString(strings.Join(parts, ", "))
	b.WriteString(")\n")
	return b.String()
}

// Coverage runs the coverage command (§1): full input validation, then the
// TSV derived from the same facts the export uses.
func Coverage(root string) ([]byte, string, error) {
	lk, err := acquireLock(root, false, LockTimeout)
	if err != nil {
		return nil, "", err
	}
	defer lk.Release()

	_, _, _, res, tables, err := loadInputs(root)
	if err != nil {
		return nil, "", err
	}
	rows := ComputeCoverage(res, tables)
	return RenderCoverageTSV(rows), RenderCoverageSummary(rows), nil
}
