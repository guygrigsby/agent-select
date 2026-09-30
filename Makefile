.DEFAULT_GOAL := build
DB := data/research.sqlite

.PHONY: build test check lint install redeploy

build: check

check: test lint
	@shasum -a 256 -c data/source-hashes.sha256 >/dev/null || { echo 'Evidence changed. Rebuild the research export and hash manifest together after source review.'; exit 1; }

# Research artifact gates only. No router or collector implementation exists.
test:
	@test "$$(sqlite3 $(DB) 'PRAGMA integrity_check;')" = ok || { echo 'Research database integrity failed.'; exit 1; }
	@test -z "$$(sqlite3 $(DB) 'PRAGMA foreign_key_check;')" || { echo 'Research foreign keys failed.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_catalog_models') FROM catalog_models;")" = 1 || { echo 'Catalog export count differs from captured metadata.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_exact_joins') FROM exact_catalog_joins;")" = 1 || { echo 'Exact-ID pricing join count differs from captured metadata.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(DISTINCT model_id)=(SELECT CAST(value AS INTEGER) FROM metadata WHERE key='expected_text_and_image_research_models') FROM model_research;")" = 1 || { echo 'Text/image model research coverage is incomplete.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*) FROM catalog_models m WHERE (m.kind='language' OR EXISTS(SELECT 1 FROM catalog_modalities d WHERE d.model_id=m.id AND d.direction='output' AND d.modality='image')) AND NOT EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id);")" = 0 || { echo 'A text/image model has no explicit research status.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*) FROM model_research r WHERE r.benchmark_status='found' AND NOT EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=r.model_id AND b.batch=r.batch);")" = 0 || { echo 'Found benchmark status has no source measurement. Fix the batch findings and rebuild the export.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*) FROM model_research r WHERE r.notes LIKE '%No exact-model numeric benchmark was located%' AND EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=r.model_id AND b.batch=r.batch);")" = 0 || { echo 'Research notes contradict recorded benchmark rows. Fix the batch findings and rebuild the export.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT (SELECT COUNT(*) FROM benchmark_measurements)=(SELECT COUNT(*) FROM source_validated_benchmarks) AND (SELECT COUNT(*) FROM unresolved_measurements)=(SELECT COUNT(*) FROM source_validated_unresolved);")" = 1 || { echo 'Stored score lacks source review. Review the affected findings and source cells before updating evidence_reviews.'; exit 1; }
	@test "$$(sqlite3 $(DB) "WITH needed(batch,path) AS (SELECT batch,source_path FROM benchmark_measurements UNION SELECT batch,source_path FROM unresolved_measurements UNION SELECT batch,'data/research/'||batch||'/findings.json' FROM evidence_reviews UNION SELECT batch,review_path FROM evidence_reviews) SELECT COUNT(*) FROM needed n JOIN evidence_reviews r ON r.batch=n.batch LEFT JOIN evidence_review_inputs i ON i.batch=n.batch AND i.path=n.path WHERE i.path IS NULL;")" = 0 || { echo 'Source review has an unpinned input. Record the reviewed findings, report and source-file hashes in evidence_review_inputs.'; exit 1; }
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*) FROM accepted_benchmarks;")" = 0 || { echo 'Benchmark acceptance changed. Update the review ledger and acceptance gate only after evidence review.'; exit 1; }

lint:
	@test "$$(sqlite3 $(DB) "SELECT COUNT(*) FROM pragma_table_list t JOIN pragma_table_info(t.name) p WHERE upper(p.type) IN ('JSON','JSONB');")" = 0 || { echo 'Research schema must use relational facts, not JSON columns.'; exit 1; }

install redeploy:
	@echo 'Research phase only. Router and collector are not implemented.'
	@exit 1
