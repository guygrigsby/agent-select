package research

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Gate kinds for gate 1.
const (
	gateBool      = "bool"       // query returns 1 when the invariant holds
	gateCountZero = "count_zero" // query returns a count of violations
	gateIntegrity = "integrity"  // PRAGMA integrity_check returns "ok"
	gateFK        = "fk"         // PRAGMA foreign_key_check returns no rows
	gateJSONBan   = "json_ban"   // count of JSON/JSONB columns, must be 0
)

// sqlGate is one gate-1 assertion over the built database. Queries port the
// existing SQL gates verbatim (§5.1).
type sqlGate struct {
	kind  string
	query string
	msg   string
}

var gate1Queries = []sqlGate{
	{gateIntegrity, "PRAGMA integrity_check", "Database integrity check failed"},
	{gateFK, "PRAGMA foreign_key_check", "Foreign key check failed"},
	{gateBool, "SELECT COUNT(*)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_catalog_models') FROM catalog_models", "Catalog export count differs from captured metadata"},
	{gateBool, "SELECT COUNT(*)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_exact_joins') FROM exact_catalog_joins", "Exact-ID pricing join count differs from captured metadata"},
	{gateBool, "SELECT COUNT(DISTINCT model_id)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_text_and_image_research_models') FROM model_research", "Text/image model research coverage is incomplete"},
	{gateCountZero, "SELECT COUNT(*) FROM catalog_models m WHERE (m.kind='language' OR EXISTS(SELECT 1 FROM catalog_modalities d WHERE d.model_id=m.id AND d.direction='output' AND d.modality='image')) AND NOT EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id)", "A text/image model has no explicit research status"},
	{gateCountZero, "SELECT COUNT(*) FROM model_research r WHERE r.benchmark_status='found' AND NOT EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=r.model_id AND b.batch=r.batch)", "Found benchmark status has no source measurement"},
	{gateCountZero, "SELECT COUNT(*) FROM model_research r WHERE r.notes LIKE '%No exact-model numeric benchmark was located%' AND EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=r.model_id AND b.batch=r.batch)", "Research notes contradict recorded benchmark rows"},
	{gateBool, "SELECT (SELECT COUNT(*) FROM benchmark_measurements)=(SELECT COUNT(*) FROM source_validated_benchmarks) AND (SELECT COUNT(*) FROM unresolved_measurements)=(SELECT COUNT(*) FROM source_validated_unresolved)", "Stored score lacks source review"},
	{gateCountZero, "WITH needed(batch,path) AS (SELECT batch,source_path FROM benchmark_measurements UNION SELECT batch,source_path FROM unresolved_measurements UNION SELECT batch,'data/research/'||batch||'/findings.json' FROM evidence_reviews UNION SELECT batch,review_path FROM evidence_reviews) SELECT COUNT(*) FROM needed n JOIN evidence_reviews r ON r.batch=n.batch LEFT JOIN evidence_review_inputs i ON i.batch=n.batch AND i.path=n.path WHERE i.path IS NULL", "Source review has an unpinned input"},
	{gateCountZero, "SELECT COUNT(*) FROM accepted_benchmarks", "Benchmark acceptance changed without a review decision"},
	{gateJSONBan, "SELECT COUNT(*) FROM pragma_table_list t JOIN pragma_table_info(t.name) p WHERE upper(p.type) IN ('JSON','JSONB')", "Research schema must use relational facts, not JSON columns"},
}

// RunGate1 evaluates every §5.1 assertion.
func RunGate1(db *sql.DB) error {
	for _, g := range gate1Queries {
		rows, err := db.Query(g.query)
		if err != nil {
			return inconsf("gate query failed: %s: %v", g.msg, err)
		}
		n := 0
		for rows.Next() {
			n++
			var v any
			if err := rows.Scan(&v); err != nil {
				rows.Close()
				return inconsf("gate scan failed: %s: %v", g.msg, err)
			}
			if !gateRowOK(g.kind, v) {
				rows.Close()
				return inconsf("%s", g.msg)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return inconsf("gate iteration failed: %s: %v", g.msg, err)
		}
		if g.kind == gateFK && n > 0 {
			return inconsf("%s", g.msg)
		}
		if g.kind == gateIntegrity && n != 1 {
			return inconsf("%s", g.msg)
		}
	}
	return nil
}

func gateRowOK(kind string, v any) bool {
	switch kind {
	case gateIntegrity:
		switch t := v.(type) {
		case string:
			return t == "ok"
		case []byte:
			return string(t) == "ok"
		}
		return false
	case gateFK:
		return false // any row is a violation
	case gateBool:
		i, ok := v.(int64)
		return ok && i == 1
	case gateCountZero, gateJSONBan:
		i, ok := v.(int64)
		return ok && i == 0
	}
	return false
}

// RunGate2 enforces the inverse status rule (§5.2): a model with measurements
// for batch B must not be benchmark_status='unresearched' for B.
func RunGate2(db *sql.DB) error {
	const q = `SELECT r.model_id, r.batch FROM model_research r JOIN benchmark_measurements b ON b.model_id=r.model_id AND b.batch=r.batch WHERE r.benchmark_status='unresearched'`
	rows, err := db.Query(q)
	if err != nil {
		return inconsf("inverse gate query: %v", err)
	}
	defer rows.Close()
	var violations []string
	for rows.Next() {
		var mid, batch string
		if err := rows.Scan(&mid, &batch); err != nil {
			return inconsf("inverse gate scan: %v", err)
		}
		violations = append(violations, fmt.Sprintf("%s/%s", mid, batch))
	}
	if err := rows.Err(); err != nil {
		return inconsf("inverse gate iteration: %v", err)
	}
	if len(violations) > 0 {
		return inconsf("model has measurements but is marked unresearched for the same batch: %s", strings.Join(violations, ", "))
	}
	return nil
}

// RunGate3 enforces the access-status vocabulary against capture bytes
// (§5.3), reading .json/.csv evidence files under root.
func RunGate3(root string, db *sql.DB) error {
	rows, err := db.Query(`SELECT batch, url, path, status FROM source_access WHERE status IN ('blocked','blocked_403','blocked_404','blocked_non_score','blocked_redirect','not_found','redirect_or_not_found')`)
	if err != nil {
		return inconsf("access-status gate query: %v", err)
	}
	type sa struct{ batch, url, path, status string }
	var check []sa
	for rows.Next() {
		var s sa
		if err := rows.Scan(&s.batch, &s.url, &s.path, &s.status); err != nil {
			rows.Close()
			return inconsf("access-status gate scan: %v", err)
		}
		check = append(check, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return inconsf("access-status gate iteration: %v", err)
	}
	for _, s := range check {
		if s.path == "" || !(strings.HasSuffix(s.path, ".json") || strings.HasSuffix(s.path, ".csv")) {
			continue // bound: only .json/.csv captures are byte-checkable
		}
		if !ValidEvidencePath(s.path) {
			return inconsf("source_access batch %s url %s: path %q violates path rules", s.batch, s.url, s.path)
		}
		raw, err := readBounded(filepath.Join(root, filepath.FromSlash(s.path)))
		if err != nil {
			if os.IsNotExist(err) {
				continue // nothing on disk to contradict the failure status
			}
			return iof("access-status gate read %s: %v", s.path, err)
		}
		if parsesAsSuccessFormat(s.path, raw) {
			return inconsf("source_access batch %s url %s: status %q contradicts %s bytes that parse as the success format", s.batch, s.url, s.status, s.path)
		}
	}
	return nil
}

// readBounded reads a file enforcing the per-file size bound.
func readBounded(p string) ([]byte, error) {
	const perFile = 64 << 20
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if fi.Size() > perFile {
		return nil, invalidf("%s: file exceeds 64 MiB bound (%d bytes)", p, fi.Size())
	}
	return os.ReadFile(p)
}

// parsesAsSuccessFormat reports whether bytes parse as the success format
// implied by the extension, with the API-error-envelope exception (§5.3).
func parsesAsSuccessFormat(path string, raw []byte) bool {
	if strings.HasSuffix(path, ".json") {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return false
		}
		obj, ok := v.(map[string]any)
		if ok && isAPIErrorEnvelope(obj) {
			return false
		}
		return true
	}
	// .csv: multi-column header plus at least one data row
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) < 2 {
		return false
	}
	header := strings.Split(lines[0], ",")
	return len(header) >= 2 && strings.TrimSpace(lines[1]) != ""
}

// isAPIErrorEnvelope matches a JSON object whose keys are a subset of the
// envelope set carrying an HTTP error status or non-empty error string.
func isAPIErrorEnvelope(obj map[string]any) bool {
	for k := range obj {
		if !envelopeKeys[k] {
			return false
		}
	}
	if s, ok := obj["error"].(string); ok && s != "" {
		return true
	}
	if s, ok := obj["status"].(string); ok && httpErrorStatus(s) {
		return true
	}
	if f, ok := obj["status"].(float64); ok && f >= 400 {
		return true
	}
	if s, ok := obj["message"].(string); ok && strings.Contains(strings.ToLower(s), "invalid username or password") {
		return true
	}
	return false
}

// envelopeKeys is the API error envelope key set (§5.3).
var envelopeKeys = map[string]bool{
	"message": true, "documentation_url": true, "status": true, "error": true,
}

func httpErrorStatus(s string) bool {
	switch s {
	case "404", "403", "401", "400", "410", "429", "500", "502", "503":
		return true
	}
	return strings.HasPrefix(s, "4") || strings.HasPrefix(s, "5")
}

// RunGatesDB runs gates 1-3 against an open database.
func RunGatesDB(root string, db *sql.DB) error {
	if err := RunGate1(db); err != nil {
		return err
	}
	if err := RunGate2(db); err != nil {
		return err
	}
	return RunGate3(root, db)
}
