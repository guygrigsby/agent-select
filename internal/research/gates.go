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
	if err := RunGate3(root, db); err != nil {
		return err
	}
	if err := RunGate6(db); err != nil {
		return err
	}
	return RunGate7(db)
}

// RunGate6 is the §5 source-policy gate: policy URL coverage equals the
// distinct source_access URLs, every source_access batch has a policies
// file, evidence binds to terms_url through the inventory URL, allowed
// permissions carry qualifying evidence and redistribution=allowed carries
// an allowlisted license_name.
func RunGate6(db *sql.DB) error {
	type q struct {
		sql  string
		name string
	}
	qs := []q{
		{`SELECT a.url FROM (SELECT DISTINCT url FROM source_access) a LEFT JOIN source_policies p ON p.url=a.url WHERE p.url IS NULL LIMIT 1`, "captured url without a policy record"},
		{`SELECT p.url FROM source_policies p WHERE NOT EXISTS(SELECT 1 FROM source_access a WHERE a.url=p.url) LIMIT 1`, "policy record for a url absent from source_access"},
		{`SELECT DISTINCT a.batch FROM source_access a WHERE NOT EXISTS(SELECT 1 FROM source_policies p WHERE p.batch=a.batch) LIMIT 1`, "source_access batch without a policies file"},
		{`SELECT DISTINCT p.batch FROM source_policies p WHERE NOT EXISTS(SELECT 1 FROM source_access a WHERE a.batch=p.batch) LIMIT 1`, "policies file naming a batch absent from source_access"},
		{`SELECT p.url FROM source_policies p JOIN source_files f ON f.path=p.evidence_path WHERE f.url<>p.terms_url LIMIT 1`, "evidence path whose inventory url does not equal terms_url"},
		{`SELECT p.url FROM source_policies p WHERE (p.research='allowed' OR p.extraction='allowed' OR p.redistribution='allowed') AND p.evidence_outcome<>'success' AND (p.license_text='unknown' OR length(p.license_text)<20) LIMIT 1`, "allowed permission without qualifying evidence"},
		{`SELECT p.url FROM source_policies p WHERE p.redistribution='allowed' AND lower(trim(p.license_name)) NOT IN ('mit','modified-mit','kimi-k3','openmdw-1.1','apache-2.0','0bsd','bsd-2-clause','bsd-3-clause','cc-by-4.0','cc-by-sa-4.0','cdla-permissive-1.0','cdla-permissive-2.0') LIMIT 1`, "redistribution=allowed with license_name outside the express-grant allowlist"},
	}
	for _, one := range qs {
		var v string
		err := db.QueryRow(one.sql).Scan(&v)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return inconsf("source-policy gate (%s): %v", one.name, err)
		}
		return inconsf("source-policy gate (%s): %s", one.name, v)
	}
	return nil
}

// componentListSQL renders the common component vocabulary as a quoted SQL
// IN list, generated from commonComponents so the gate cannot drift from the
// derivation.
func componentListSQL() string {
	quoted := make([]string, len(commonComponents))
	for i, c := range commonComponents {
		quoted[i] = "'" + c + "'"
	}
	return strings.Join(quoted, ",")
}

// RunGate7 is the §12.6 qualification gate: derived qualification tables
// cover their inputs exactly and match recomputation from the source rows.
func RunGate7(db *sql.DB) error {
	type q struct {
		sql  string
		name string
	}
	qs := []q{
		{`SELECT m.id FROM catalog_models m LEFT JOIN offer_binding_qualifications b ON b.model_id=m.id WHERE b.model_id IS NULL LIMIT 1`, "catalog model without an offer binding qualification"},
		{`SELECT b.model_id FROM offer_binding_qualifications b WHERE NOT EXISTS(SELECT 1 FROM catalog_models m WHERE m.id=b.model_id) LIMIT 1`, "offer binding qualification for a model absent from the catalog"},
		{`SELECT j.model_id FROM exact_catalog_joins j JOIN offer_binding_qualifications b ON b.model_id=j.model_id WHERE b.binding<>'exact_join' OR b.provider<>j.provider OR b.upstream_id<>j.upstream_id LIMIT 1`, "exact_catalog_joins model whose offer binding is not the matching exact_join row"},
		{`SELECT b.model_id FROM offer_binding_qualifications b WHERE b.binding='unjoined' AND (b.provider<>'' OR b.upstream_id<>'') LIMIT 1`, "unjoined offer binding with a non-empty provider or upstream_id"},
		{`SELECT b.model_id FROM offer_binding_qualifications b WHERE b.binding='exact_join' AND NOT EXISTS(SELECT 1 FROM exact_catalog_joins j WHERE j.model_id=b.model_id AND j.provider=b.provider AND j.upstream_id=b.upstream_id) LIMIT 1`, "exact_join offer binding that does not reproduce its exact_catalog_joins row"},
		{`SELECT m.id FROM benchmark_measurements m WHERE (SELECT COUNT(*) FROM measurement_settings s WHERE s.scope='linked' AND s.measurement_id=m.id)<>12 LIMIT 1`, "linked measurement without exactly twelve settings rows"},
		{`SELECT u.id FROM unresolved_measurements u WHERE (SELECT COUNT(*) FROM measurement_settings s WHERE s.scope='unresolved' AND s.measurement_id=u.id)<>12 LIMIT 1`, "unresolved measurement without exactly twelve settings rows"},
		{`SELECT s.scope||' '||s.measurement_id FROM measurement_settings s WHERE s.scope='linked' AND NOT EXISTS(SELECT 1 FROM benchmark_measurements m WHERE m.id=s.measurement_id) LIMIT 1`, "settings row referencing a linked measurement that does not exist"},
		{`SELECT s.scope||' '||s.measurement_id FROM measurement_settings s WHERE s.scope='unresolved' AND NOT EXISTS(SELECT 1 FROM unresolved_measurements u WHERE u.id=s.measurement_id) LIMIT 1`, "settings row referencing an unresolved measurement that does not exist"},
		{`SELECT s.scope||' '||s.measurement_id||' '||s.component FROM measurement_settings s WHERE s.knowledge='unknown' AND s.value_or_reason<>'` + unknownReason + `' LIMIT 1`, "unknown settings row without the fixed reason"},
		{`SELECT scope||' '||measurement_id||' '||component FROM measurement_settings WHERE component NOT IN (` + componentListSQL() + `) LIMIT 1`, "settings row with a component outside the common critical components"},
		{`SELECT l.model_id||' '||l.name FROM catalog_limits l LEFT JOIN limit_qualifications q ON q.model_id=l.model_id AND q.name=l.name WHERE q.model_id IS NULL OR q.value<>l.value LIMIT 1`, "catalog limit without a matching limit qualification carrying the same value"},
		{`SELECT q.model_id||' '||q.name FROM limit_qualifications q WHERE NOT EXISTS(SELECT 1 FROM catalog_limits l WHERE l.model_id=q.model_id AND l.name=q.name) LIMIT 1`, "limit qualification without a catalog limit row"},
	}
	for _, one := range qs {
		var v string
		err := db.QueryRow(one.sql).Scan(&v)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return inconsf("qualification gate (%s): %v", one.name, err)
		}
		return inconsf("qualification gate (%s): %s", one.name, v)
	}
	// Known settings values must occur in the measurement's configuration
	// (or equal benchmark_version for benchmark_release).
	knownValue := `
SELECT s.scope, s.measurement_id, s.component, s.value_or_reason, m.configuration, m.benchmark_version
FROM measurement_settings s
JOIN benchmark_measurements m ON m.id=s.measurement_id AND s.scope='linked'
WHERE s.knowledge='known'
UNION ALL
SELECT s.scope, s.measurement_id, s.component, s.value_or_reason, u.configuration, u.benchmark_version
FROM measurement_settings s
JOIN unresolved_measurements u ON u.id=s.measurement_id AND s.scope='unresolved'
WHERE s.knowledge='known'`
	rows, err := db.Query(knownValue)
	if err != nil {
		return inconsf("qualification gate (known settings scan): %v", err)
	}
	for rows.Next() {
		var scope, comp, val, config, version string
		var id int64
		if err := rows.Scan(&scope, &id, &comp, &val, &config, &version); err != nil {
			return inconsf("qualification gate (known settings scan): %v", err)
		}
		ok := comp == "benchmark_release" && val == version
		if !ok && strings.Contains(strings.ToLower(config), strings.ToLower(val)) {
			ok = true
		}
		if !ok {
			return inconsf("qualification gate (known settings value absent from configuration): %s %d %s", scope, id, comp)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return inconsf("qualification gate (known settings scan): %v", err)
	}
	rows.Close()
	// Comparison qualifications must cover every measurement and equal
	// recomputation from the settings rows.
	type cmp struct {
		scope    string
		id       int64
		eligible int
		blocking string
	}
	stored := map[[2]any]cmp{}
	crows, err := db.Query(`SELECT scope, measurement_id, eligible, blocking_components FROM comparison_qualifications`)
	if err != nil {
		return inconsf("qualification gate (comparison scan): %v", err)
	}
	for crows.Next() {
		var c cmp
		if err := crows.Scan(&c.scope, &c.id, &c.eligible, &c.blocking); err != nil {
			crows.Close()
			return inconsf("qualification gate (comparison scan): %v", err)
		}
		stored[[2]any{c.scope, c.id}] = c
	}
	crows.Close()
	recompute := func(scope string) error {
		tbl := "benchmark_measurements"
		if scope == "unresolved" {
			tbl = "unresolved_measurements"
		}
		r, err := db.Query(`SELECT id, benchmark_version, configuration FROM ` + tbl)
		if err != nil {
			return inconsf("qualification gate (comparison recompute): %v", err)
		}
		for r.Next() {
			var id int64
			var version, config string
			if err := r.Scan(&id, &version, &config); err != nil {
				return inconsf("qualification gate (comparison recompute): %v", err)
			}
			settings := deriveSettings(config, version)
			blocking := ""
			for _, s := range settings {
				if s.knowledge == "unknown" {
					if blocking != "" {
						blocking += ","
					}
					blocking += s.component
				}
			}
			eligible := 0
			if scope == "linked" && blocking == "" {
				eligible = 1
			}
			c, ok := stored[[2]any{scope, id}]
			if !ok {
				return inconsf("qualification gate (comparison qualification missing): %s %d", scope, id)
			}
			if c.eligible != eligible || c.blocking != blocking {
				return inconsf("qualification gate (comparison qualification mismatch): %s %d", scope, id)
			}
			delete(stored, [2]any{scope, id})
		}
		if err := r.Err(); err != nil {
			return inconsf("qualification gate (comparison recompute): %v", err)
		}
		r.Close()
		return nil
	}
	if err := recompute("linked"); err != nil {
		return err
	}
	if err := recompute("unresolved"); err != nil {
		return err
	}
	for k := range stored {
		return inconsf("qualification gate (comparison qualification without a measurement): %v %v", k[0], k[1])
	}
	// Price dimension interpretations must equal recomputation per distinct pair
	// (pairs are collected first: a single-connection database cannot run the
	// per-pair lookup while the pair scan is still open).
	prows, err := db.Query(`SELECT DISTINCT unit, field_path FROM price_facts`)
	if err != nil {
		return inconsf("qualification gate (price pair scan): %v", err)
	}
	var pairs [][2]string
	for prows.Next() {
		var unit, path string
		if err := prows.Scan(&unit, &path); err != nil {
			prows.Close()
			return inconsf("qualification gate (price pair scan): %v", err)
		}
		pairs = append(pairs, [2]string{unit, path})
	}
	if err := prows.Err(); err != nil {
		prows.Close()
		return inconsf("qualification gate (price pair scan): %v", err)
	}
	prows.Close()
	for _, pair := range pairs {
		unit, path := pair[0], pair[1]
		class, basis := classifyPriceDimension(unit, path)
		var gotClass, gotBasis string
		err := db.QueryRow(`SELECT dimension_class, basis FROM price_dimension_interpretations WHERE unit=? AND field_path=?`, unit, path).Scan(&gotClass, &gotBasis)
		if err == sql.ErrNoRows {
			return inconsf("qualification gate (price pair without an interpretation): %s %s", unit, path)
		}
		if err != nil {
			return inconsf("qualification gate (price pair scan): %v", err)
		}
		if gotClass != class || gotBasis != basis {
			return inconsf("qualification gate (price interpretation mismatch): %s %s", unit, path)
		}
	}
	var extra int
	if err := db.QueryRow(`SELECT COUNT(*) FROM price_dimension_interpretations i WHERE NOT EXISTS(SELECT 1 FROM price_facts f WHERE f.unit=i.unit AND f.field_path=i.field_path)`).Scan(&extra); err != nil {
		return inconsf("qualification gate (price interpretation scan): %v", err)
	}
	if extra > 0 {
		return inconsf("qualification gate (price interpretation without a price fact pair): %d rows", extra)
	}
	return nil
}
