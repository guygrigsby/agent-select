package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BacklogTask is one acquisition/extraction task from
// data/research-inputs/backlog-tasks.json (§10.2).
type BacklogTask struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"` // batch_acquisition | source_acquisition
	Batches       []string `json:"batches"`
	EvidenceFiles []string `json:"evidence_files"`
	EvidenceURLs  []string `json:"evidence_urls"`
	DesiredFields []string `json:"desired_fields"`
	Closure       string   `json:"closure"`
}

type taskFile struct {
	Tasks []BacklogTask `json:"tasks"`
}

// LoadTasks reads and validates backlog-tasks.json. Validation is part of
// gate coverage: every model_research batch must be claimed by a task, every
// evidence file must be in the capture inventory (same seam as the batch
// findings gate), and every task must name desired fields and a closure
// condition.
func LoadTasks(root string, inv *Inventory, researchBatches []string) ([]BacklogTask, error) {
	p := filepath.Join(root, "data", "research-inputs", "backlog-tasks.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, iof("read %s: %v", p, err)
	}
	var tf taskFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		return nil, invalidf("%s: %v", rel(root, p), err)
	}
	if len(tf.Tasks) == 0 {
		return nil, invalidf("%s: no tasks", rel(root, p))
	}
	seen := map[string]bool{}
	claimed := map[string]bool{}
	for i, t := range tf.Tasks {
		if t.ID == "" {
			return nil, invalidf("%s task %d: empty id", rel(root, p), i+1)
		}
		if seen[t.ID] {
			return nil, invalidf("%s: duplicate task id %q", rel(root, p), t.ID)
		}
		seen[t.ID] = true
		if t.Kind != "batch_acquisition" && t.Kind != "source_acquisition" {
			return nil, invalidf("%s task %s: kind %q outside batch_acquisition|source_acquisition", rel(root, p), t.ID, t.Kind)
		}
		if (t.Kind == "batch_acquisition") != (len(t.Batches) > 0) {
			return nil, invalidf("%s task %s: batch_acquisition tasks must name batches and source tasks none", rel(root, p), t.ID)
		}
		for _, f := range t.EvidenceFiles {
			if inv.Get(f) == nil {
				return nil, invalidf("%s task %s: evidence file %s is not in the capture inventory (gate 5)", rel(root, p), t.ID, f)
			}
		}
		if len(t.DesiredFields) == 0 {
			return nil, invalidf("%s task %s: desired_fields empty", rel(root, p), t.ID)
		}
		if t.Closure == "" {
			return nil, invalidf("%s task %s: closure condition empty", rel(root, p), t.ID)
		}
		for _, b := range t.Batches {
			claimed[b] = true
		}
	}
	var missing []string
	for _, b := range researchBatches {
		if !claimed[b] {
			missing = append(missing, b)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, invalidf("%s: research batches without a backlog task: %s", rel(root, p), strings.Join(missing, ", "))
	}
	return tf.Tasks, nil
}

// BacklogRow is one per-model backlog TSV row (§10.1): one record per catalog
// model and applicable channel.
type BacklogRow struct {
	CoverageRow
	InScope             string // in_scope | out_of_scope
	ImageOutput         bool
	Offers              int
	PriceLeaves         int
	PriceUnresolvedUnit int
	LimitFacts          int
	ModalityFacts       int
	ParameterFacts      int
	AccessStates        string // sorted distinct source_access statuses for the model's batches
	Licenses            string // sorted distinct inventory licenses for the batch findings files
	PublicationPerm     string // sorted distinct publication permissions
}

// BacklogHeader is the backlog TSV header line.
const BacklogHeader = "model_id\tkind\tin_scope\tbatches\tbenchmark_coverage\tmeasurements\tunresolved_measurements\tlocated_claims\tunlocated_claims\timage_output\toffers\tprice_leaves\tprice_unresolved_unit\tlimit_facts\tmodality_facts\tparameter_facts\taccess_states\tlicenses\tpublication_permission"

const unresolvedUnit = "source_specific_unresolved"

// ComputeBacklog extends the coverage rows with pricing, capability, access
// and rights channels, assembled from the same tables that feed the export.
func ComputeBacklog(res *BuildResult, tables []*tableRows, inv *Inventory) []BacklogRow {
	cov := ComputeCoverage(res, tables)
	covIdx := map[string]int{}
	for i, r := range cov {
		covIdx[r.ModelID] = i
	}

	var tJoin, tPrice, tLimits, tMod, tParams, tAccess *tableRows
	for _, t := range tables {
		switch t.name {
		case "exact_catalog_joins":
			tJoin = t
		case "price_facts":
			tPrice = t
		case "catalog_limits":
			tLimits = t
		case "catalog_modalities":
			tMod = t
		case "catalog_parameters":
			tParams = t
		case "source_access":
			tAccess = t
		}
	}

	offer := map[string]string{} // model_id -> provider/upstream_id
	for _, r := range tJoin.rows {
		offer[r[0].(string)] = r[1].(string) + "/" + r[2].(string)
	}
	priceTot, priceUnres := map[string]int{}, map[string]int{}
	for _, r := range tPrice.rows {
		key := r[1].(string) + "/" + r[2].(string)
		priceTot[key]++
		if r[5].(string) == unresolvedUnit {
			priceUnres[key]++
		}
	}
	limitN, modN, paramN := map[string]int{}, map[string]int{}, map[string]int{}
	for _, r := range tLimits.rows {
		limitN[r[0].(string)]++
	}
	for _, r := range tMod.rows {
		modN[r[0].(string)]++
	}
	for _, r := range tParams.rows {
		paramN[r[0].(string)]++
	}
	imgOut := map[string]bool{}
	for _, r := range tMod.rows {
		if r[1].(string) == "output" && r[2].(string) == "image" {
			imgOut[r[0].(string)] = true
		}
	}
	batchAccess := map[string]map[string]bool{}
	for _, r := range tAccess.rows {
		b := r[0].(string)
		if batchAccess[b] == nil {
			batchAccess[b] = map[string]bool{}
		}
		batchAccess[b][r[3].(string)] = true
	}
	batchLicense := map[string]map[string]bool{}
	batchPerm := map[string]map[string]bool{}
	for b := range batchAccess {
		fp := "data/research/" + b + "/findings.json"
		if row := inv.Get(fp); row != nil {
			if batchLicense[b] == nil {
				batchLicense[b] = map[string]bool{}
				batchPerm[b] = map[string]bool{}
			}
			batchLicense[b][row.License] = true
			batchPerm[b][row.PublicationPermission] = true
		}
	}

	rows := make([]BacklogRow, len(cov))
	for i, c := range cov {
		mid := c.ModelID
		access := map[string]bool{}
		lic := map[string]bool{}
		perm := map[string]bool{}
		for _, b := range strings.Split(c.Batches, ";") {
			if b == "" {
				continue
			}
			for s := range batchAccess[b] {
				access[s] = true
			}
			for l := range batchLicense[b] {
				lic[l] = true
			}
			for p := range batchPerm[b] {
				perm[p] = true
			}
		}
		var accessL, licL, permL []string
		for s := range access {
			accessL = append(accessL, s)
		}
		for l := range lic {
			licL = append(licL, l)
		}
		for p := range perm {
			permL = append(permL, p)
		}
		sort.Strings(accessL)
		sort.Strings(licL)
		sort.Strings(permL)
		nOffers := 0
		if _, ok := offer[mid]; ok {
			nOffers = 1
		}
		rows[i] = BacklogRow{
			CoverageRow:         c,
			InScope:             "out_of_scope",
			ImageOutput:         imgOut[mid],
			Offers:              nOffers,
			PriceLeaves:         priceTot[offer[mid]],
			PriceUnresolvedUnit: priceUnres[offer[mid]],
			LimitFacts:          limitN[mid],
			ModalityFacts:       modN[mid],
			ParameterFacts:      paramN[mid],
			AccessStates:        strings.Join(accessL, ";"),
			Licenses:            strings.Join(licL, ";"),
			PublicationPerm:     strings.Join(permL, ";"),
		}
		if c.Batches != "" {
			rows[i].InScope = "in_scope"
		}
	}
	return rows
}

// RenderBacklogTSV renders the machine-readable backlog TSV (LF, trailing
// newline).
func RenderBacklogTSV(rows []BacklogRow) []byte {
	var b strings.Builder
	b.WriteString(BacklogHeader)
	b.WriteByte('\n')
	for _, r := range rows {
		img := "f"
		if r.ImageOutput {
			img = "t"
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
			r.ModelID, r.Kind, r.InScope, r.Batches, r.BenchmarkCoverage,
			r.Measurements, r.UnresolvedMeasurements, r.LocatedClaims, r.UnlocatedClaims,
			img, r.Offers, r.PriceLeaves, r.PriceUnresolvedUnit,
			r.LimitFacts, r.ModalityFacts, r.ParameterFacts,
			r.AccessStates, r.Licenses, r.PublicationPerm)
	}
	return []byte(b.String())
}

// backlogTotals aggregates the report's cross-model counts.
type backlogTotals struct {
	Models, InScope, OutOfScope     int
	ImageOutput                     int
	KindCounts                      map[string]int
	CoverageBuckets                 map[string]int
	UnlocatedByBatch                map[string]int
	UnlocatedTotal                  int
	PriceLeaves, PriceUnresolved    int
	UnresolvedMeasurements          int
	IdentityProposals               int
	ImageModels                     int
	BatchModels                     map[string]int
	OffersJoined, OffersUnjoined    int
	AccessStates                    map[string]int
	PermissionUnknown, PermissionOK int
}

func totalsFor(rows []BacklogRow, tables []*tableRows) backlogTotals {
	t := backlogTotals{KindCounts: map[string]int{}, CoverageBuckets: map[string]int{}, UnlocatedByBatch: map[string]int{}, BatchModels: map[string]int{}, AccessStates: map[string]int{}}
	var tUnres, tProp, tUnloc *tableRows
	for _, tb := range tables {
		switch tb.name {
		case "unresolved_measurements":
			tUnres = tb
		case "unresolved_identity_proposals":
			tProp = tb
		case "unlocated_claims":
			tUnloc = tb
		}
	}
	for _, r := range rows {
		t.Models++
		t.KindCounts[r.Kind]++
		t.CoverageBuckets[r.BenchmarkCoverage]++
		if r.InScope == "in_scope" {
			t.InScope++
			t.BatchModels[r.Batches]++
		} else {
			t.OutOfScope++
		}
		if r.ImageOutput {
			t.ImageOutput++
		}
		if r.Offers > 0 {
			t.OffersJoined++
		} else {
			t.OffersUnjoined++
		}
		t.PriceLeaves += r.PriceLeaves
		t.PriceUnresolved += r.PriceUnresolvedUnit
		if r.PublicationPerm == "unknown" {
			t.PermissionUnknown++
		} else if r.PublicationPerm != "" {
			t.PermissionOK++
		}
		for _, s := range strings.Split(r.AccessStates, ";") {
			if s != "" {
				t.AccessStates[s]++
			}
		}
		if r.ImageOutput && strings.Contains(r.Batches, "images") {
			t.ImageModels++
		}
	}
	t.UnresolvedMeasurements = len(tUnres.rows)
	t.IdentityProposals = len(tProp.rows)
	for _, r := range tUnloc.rows {
		t.UnlocatedByBatch[r[1].(string)]++
		t.UnlocatedTotal++
	}
	return t
}

// RenderBacklogReport renders the human backlog report from the same rows
// that feed the TSV plus the task file (§10.3). Deterministic: sorted maps,
// no timestamps.
func RenderBacklogReport(rows []BacklogRow, tables []*tableRows, tasks []BacklogTask) []byte {
	t := totalsFor(rows, tables)

	var tPrice, tLimits *tableRows
	for _, tb := range tables {
		switch tb.name {
		case "price_facts":
			tPrice = tb
		case "catalog_limits":
			tLimits = tb
		}
	}
	allPriceLeaves, allPriceUnresolved := 0, 0
	for _, r := range tPrice.rows {
		allPriceLeaves++
		if r[5].(string) == unresolvedUnit {
			allPriceUnresolved++
		}
	}
	imgLimits := map[string]bool{}
	for _, r := range tLimits.rows {
		if strings.Contains(r[1].(string), "image") {
			imgLimits[r[0].(string)] = true
		}
	}
	imgNoAccounting := 0
	for _, r := range rows {
		if r.ImageOutput && !imgLimits[r.ModelID] {
			imgNoAccounting++
		}
	}

	var b strings.Builder
	b.WriteString("# Research backlog\n\n")
	b.WriteString("Generated from the assembled build tables and `data/research-inputs/backlog-tasks.json` (never the published DB). Regenerate with `make research-backlog`; `agent-routing research backlog --check docs/research/backlog.md` gates drift in `make check`.\n\n")

	fmt.Fprintf(&b, "- catalog models: %d (in scope %d, out of scope %d)\n", t.Models, t.InScope, t.OutOfScope)
	var kinds []string
	for k := range t.KindCounts {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	var kc []string
	for _, k := range kinds {
		kc = append(kc, fmt.Sprintf("%s %d", k, t.KindCounts[k]))
	}
	fmt.Fprintf(&b, "- kinds: %s\n", strings.Join(kc, ", "))
	fmt.Fprintf(&b, "- image-output models: %d, all assigned to the images batch (%d image-kind, %d language-kind with image output; the two sets are gated equal)\n", t.ImageOutput, t.KindCounts["image"], t.ImageOutput-t.KindCounts["image"])
	fmt.Fprintf(&b, "- exact-ID price joins: %d models joined, %d unjoined\n", t.OffersJoined, t.OffersUnjoined)
	fmt.Fprintf(&b, "- benchmark coverage: scores_found %d, no_scores_in_checked_sources %d, unresearched %d, out_of_research_scope %d\n",
		t.CoverageBuckets["scores_found"], t.CoverageBuckets["no_scores_in_checked_sources"], t.CoverageBuckets["unresearched"], t.CoverageBuckets["out_of_research_scope"])

	b.WriteString("\n## Cross-model gaps\n\n")
	fmt.Fprintf(&b, "- unlocated claims: %d total (48 at the frozen baseline; the §2.4 URL rule moved 8 formerly located claims here)\n", t.UnlocatedTotal)
	var batches []string
	for b2 := range t.UnlocatedByBatch {
		batches = append(batches, b2)
	}
	sort.Strings(batches)
	for _, b2 := range batches {
		fmt.Fprintf(&b, "  - %s: %d\n", b2, t.UnlocatedByBatch[b2])
	}
	fmt.Fprintf(&b, "- price leaves: %d numeric with %d carrying %s units across catalog-joined offers; all provider offers carry %d numeric leaves with %d unresolved (unjoined offers are models.dev inventory beyond catalog scope)\n", t.PriceLeaves, t.PriceUnresolved, unresolvedUnit, allPriceLeaves, allPriceUnresolved)
	fmt.Fprintf(&b, "- alias/revision gaps: %d unresolved measurements, %d identity proposals recorded\n", t.UnresolvedMeasurements, t.IdentityProposals)
	fmt.Fprintf(&b, "- image-accounting gaps: %d of %d image-output models have no image-specific limit facts (format, mask, alpha, dimensions, count accounting)\n", imgNoAccounting, t.ImageOutput)
	fmt.Fprintf(&b, "- protocol metadata: cells missing evaluated-revision, harness or setting components stay unresolved rather than comparison-qualified; `accepted_benchmarks` remains empty\n")
	fmt.Fprintf(&b, "- rights: %d in-scope models have publication_permission unknown, %d have a decided permission\n", t.PermissionUnknown, t.PermissionOK)

	var as []string
	for s := range t.AccessStates {
		as = append(as, s)
	}
	sort.Strings(as)
	var asc []string
	for _, s := range as {
		asc = append(asc, fmt.Sprintf("%s %d", s, t.AccessStates[s]))
	}
	fmt.Fprintf(&b, "- access states over in-scope models: %s\n", strings.Join(asc, ", "))

	b.WriteString("\n## Acquisition and extraction tasks\n\n")
	b.WriteString("### Batch acquisition\n\n")
	b.WriteString("| task | batches | models | evidence files | closure condition |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, task := range tasks {
		if task.Kind != "batch_acquisition" {
			continue
		}
		n := batchModelCount(t.BatchModels, task.Batches)
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %s |\n",
			task.ID, strings.Join(task.Batches, ";"), n,
			taskFileList(task.EvidenceFiles), task.Closure)
	}
	b.WriteString("\n### Source acquisition\n\n")
	b.WriteString("| task | evidence files | desired fields | closure condition |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, task := range tasks {
		if task.Kind != "source_acquisition" {
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
			task.ID, taskFileList(task.EvidenceFiles),
			strings.Join(task.DesiredFields, "; "), task.Closure)
	}
	b.WriteString("\n## Per-model records\n\n")
	fmt.Fprintf(&b, "The machine-readable backlog is `agent-routing research backlog` (stdout TSV, %d rows, one per catalog model). Columns: %s\n", len(rows), strings.ReplaceAll(strings.Join(strings.Split(BacklogHeader, "\t")[1:], ", "), "_", " "))
	b.WriteString("\nEvery in-scope model has an explicit `benchmark_status` and channel assignment; out-of-scope kinds (embedding, evaluation, realtime, reranking, speech, transcription, video) are retained with their kind and no research status.\n")
	return []byte(b.String())
}

func containsBatch(joined, b string) bool {
	for _, s := range strings.Split(joined, ";") {
		if s == b {
			return true
		}
	}
	return false
}

func batchModelCount(batchModels map[string]int, batches []string) int {
	n := 0
	for joined, c := range batchModels {
		for _, b := range batches {
			if containsBatch(joined, b) {
				n += c
				break
			}
		}
	}
	return n
}

func taskFileList(fs []string) string {
	if len(fs) == 0 {
		return "(none captured yet)"
	}
	return strings.Join(fs, "; ")
}

// Backlog runs the backlog command: validate inputs, assemble rows from the
// same tables that feed the export, return TSV and report bytes.
func Backlog(root string) (tsv, report []byte, summary string, err error) {
	lk, err := acquireLock(root, false, LockTimeout)
	if err != nil {
		return nil, nil, "", err
	}
	defer lk.Release()

	_, inv, _, res, tables, err := loadInputs(root)
	if err != nil {
		return nil, nil, "", err
	}
	researchBatches := researchBatchNames(tables)
	tasks, err := LoadTasks(root, inv, researchBatches)
	if err != nil {
		return nil, nil, "", err
	}
	rows := ComputeBacklog(res, tables, inv)
	if err := backlogGates(rows, researchBatches); err != nil {
		return nil, nil, "", err
	}
	report = RenderBacklogReport(rows, tables, tasks)
	tsv = RenderBacklogTSV(rows)
	summary = fmt.Sprintf("backlog: %d models (%d in scope), %d tasks\n", len(rows), countInScope(rows), len(tasks))
	return tsv, report, summary, nil
}

func researchBatchNames(tables []*tableRows) []string {
	var tResearch *tableRows
	for _, t := range tables {
		if t.name == "model_research" {
			tResearch = t
		}
	}
	seen := map[string]bool{}
	var out []string
	for _, r := range tResearch.rows {
		b := r[1].(string)
		if !seen[b] {
			seen[b] = true
			out = append(out, b)
		}
	}
	sort.Strings(out)
	return out
}

// backlogGates enforce the phase exit criterion: no discovered in-scope
// model silently lacks a research status or channel assignment, and the
// images batch assignment stays in lockstep with image-output modality.
func backlogGates(rows []BacklogRow, researchBatches []string) error {
	batchSet := map[string]bool{}
	for _, b := range researchBatches {
		batchSet[b] = true
	}
	for _, r := range rows {
		if r.InScope == "in_scope" {
			if r.Batches == "" {
				return invalidf("backlog gate: in-scope model %s has no channel assignment", r.ModelID)
			}
			if r.BenchmarkCoverage == "" || r.BenchmarkCoverage == "out_of_research_scope" {
				return invalidf("backlog gate: in-scope model %s has no research status", r.ModelID)
			}
			for _, b := range strings.Split(r.Batches, ";") {
				if !batchSet[b] {
					return invalidf("backlog gate: model %s assigned to unknown batch %s", r.ModelID, b)
				}
			}
			if r.ImageOutput != containsBatch(r.Batches, "images") {
				return invalidf("backlog gate: model %s image-output %t disagrees with images-batch assignment", r.ModelID, r.ImageOutput)
			}
		} else if containsBatch(r.Batches, "images") {
			return invalidf("backlog gate: out-of-scope model %s assigned to the images batch", r.ModelID)
		}
	}
	return nil
}

// BacklogCheck regenerates the report and compares it against the committed
// file (drift gate).
func BacklogCheck(root, path string) error {
	_, report, _, err := Backlog(root)
	if err != nil {
		return err
	}
	p := filepath.Join(root, path)
	got, err := os.ReadFile(p)
	if err != nil {
		return iof("read %s: %v", p, err)
	}
	if !bytes.Equal(got, report) {
		return invalidf("%s is stale: regenerate with `make research-backlog`", rel(root, p))
	}
	return nil
}

func countInScope(rows []BacklogRow) int {
	n := 0
	for _, r := range rows {
		if r.InScope == "in_scope" {
			n++
		}
	}
	return n
}
