package research

// SchemaVersion is the research schema version this tool carries (§7).
// The DDL below is the schema 3 reconstruction, byte-faithful to the frozen
// corpus (all tables, views, indexes, checks).
const SchemaVersion = 3

// SchemaDDL creates every schema 3 object in declaration order (§4:
// tables are inserted in schema order after creation).
const SchemaDDL = `CREATE TABLE metadata(key TEXT PRIMARY KEY,value TEXT NOT NULL);
CREATE TABLE source_files(path TEXT PRIMARY KEY,url TEXT NOT NULL,sha256 TEXT NOT NULL CHECK(length(sha256)=64),bytes INTEGER NOT NULL CHECK(bytes>=0),observed_at TEXT NOT NULL,timestamp_basis TEXT NOT NULL,license TEXT NOT NULL,publication_permission TEXT NOT NULL CHECK(publication_permission IN ('unknown','allowed','denied')));
CREATE UNIQUE INDEX source_hash_identity ON source_files(path,sha256);
CREATE TABLE catalog_models(id TEXT PRIMARY KEY,name TEXT NOT NULL,owner TEXT NOT NULL,kind TEXT NOT NULL,description TEXT NOT NULL,source_path TEXT NOT NULL REFERENCES source_files(path));
CREATE TABLE catalog_modalities(model_id TEXT NOT NULL REFERENCES catalog_models(id),direction TEXT NOT NULL CHECK(direction IN ('input','output')),modality TEXT NOT NULL,PRIMARY KEY(model_id,direction,modality));
CREATE TABLE catalog_limits(model_id TEXT NOT NULL REFERENCES catalog_models(id),name TEXT NOT NULL,value INTEGER NOT NULL,PRIMARY KEY(model_id,name));
CREATE TABLE catalog_parameters(model_id TEXT NOT NULL REFERENCES catalog_models(id),parameter TEXT NOT NULL,PRIMARY KEY(model_id,parameter));
CREATE TABLE catalog_tags(model_id TEXT NOT NULL REFERENCES catalog_models(id),tag TEXT NOT NULL,PRIMARY KEY(model_id,tag));
CREATE TABLE provider_offers(provider TEXT NOT NULL,upstream_id TEXT NOT NULL,name TEXT NOT NULL,source_path TEXT NOT NULL REFERENCES source_files(path),PRIMARY KEY(provider,upstream_id));
CREATE TABLE canonical_labels(provider TEXT NOT NULL,upstream_id TEXT NOT NULL,label TEXT NOT NULL,PRIMARY KEY(provider,upstream_id),FOREIGN KEY(provider,upstream_id) REFERENCES provider_offers(provider,upstream_id));
CREATE TABLE exact_catalog_joins(model_id TEXT PRIMARY KEY REFERENCES catalog_models(id),provider TEXT NOT NULL,upstream_id TEXT NOT NULL,FOREIGN KEY(provider,upstream_id) REFERENCES provider_offers(provider,upstream_id));
CREATE TABLE price_facts(source_path TEXT NOT NULL REFERENCES source_files(path),provider TEXT NOT NULL,upstream_id TEXT NOT NULL,field_path TEXT NOT NULL,raw_decimal TEXT NOT NULL,unit TEXT NOT NULL,PRIMARY KEY(source_path,provider,upstream_id,field_path));
CREATE TABLE pricing_disagreements(model_id TEXT NOT NULL REFERENCES catalog_models(id),component TEXT NOT NULL,vercel_per_million TEXT NOT NULL,models_dev_per_million TEXT NOT NULL,PRIMARY KEY(model_id,component));
CREATE TABLE review_results(batch TEXT PRIMARY KEY,grade TEXT NOT NULL CHECK(grade IN ('accepted','quarantined')),reason TEXT NOT NULL);
CREATE TABLE model_research(model_id TEXT NOT NULL REFERENCES catalog_models(id),batch TEXT NOT NULL,benchmark_status TEXT NOT NULL CHECK(benchmark_status IN ('found','not_found_in_checked_sources','unresearched')),notes TEXT NOT NULL,PRIMARY KEY(model_id,batch));
CREATE TABLE claims(id INTEGER PRIMARY KEY,model_id TEXT NOT NULL REFERENCES catalog_models(id),batch TEXT NOT NULL,polarity TEXT NOT NULL CHECK(polarity IN ('strength','weakness')),basis TEXT NOT NULL,text TEXT NOT NULL,source_url TEXT NOT NULL CHECK(length(source_url)>0),source_path TEXT NOT NULL REFERENCES source_files(path),locator TEXT NOT NULL);
CREATE TABLE unlocated_claims(model_id TEXT NOT NULL REFERENCES catalog_models(id),batch TEXT NOT NULL,polarity TEXT NOT NULL,basis TEXT NOT NULL,text TEXT NOT NULL,source_url TEXT NOT NULL,locator TEXT NOT NULL,reason TEXT NOT NULL);
CREATE TABLE benchmark_measurements(id INTEGER PRIMARY KEY,model_id TEXT NOT NULL REFERENCES catalog_models(id),batch TEXT NOT NULL,benchmark TEXT NOT NULL,benchmark_version TEXT NOT NULL,metric TEXT NOT NULL,raw_value TEXT NOT NULL,unit TEXT NOT NULL,configuration TEXT NOT NULL,source_url TEXT NOT NULL,source_path TEXT NOT NULL REFERENCES source_files(path),locator TEXT NOT NULL,evidence_class TEXT NOT NULL,identity_basis TEXT NOT NULL);
CREATE TABLE unresolved_measurements(id INTEGER PRIMARY KEY,batch TEXT NOT NULL REFERENCES review_results(batch),reported_model TEXT NOT NULL,benchmark TEXT NOT NULL,benchmark_version TEXT NOT NULL,metric TEXT NOT NULL,raw_value TEXT NOT NULL,unit TEXT NOT NULL,configuration TEXT NOT NULL,source_url TEXT NOT NULL,source_path TEXT NOT NULL REFERENCES source_files(path),locator TEXT NOT NULL,evidence_class TEXT NOT NULL,identity_status TEXT NOT NULL CHECK(identity_status='unresolved'));
CREATE TABLE unresolved_identity_proposals(measurement_id INTEGER NOT NULL REFERENCES unresolved_measurements(id),model_id TEXT NOT NULL REFERENCES catalog_models(id),PRIMARY KEY(measurement_id,model_id));
CREATE TABLE source_access(batch TEXT NOT NULL,url TEXT NOT NULL,path TEXT NOT NULL,status TEXT NOT NULL,purpose TEXT NOT NULL,license TEXT NOT NULL,PRIMARY KEY(batch,url,path));
CREATE TABLE research_gaps(batch TEXT NOT NULL,ordinal INTEGER NOT NULL,detail TEXT NOT NULL,PRIMARY KEY(batch,ordinal));
CREATE INDEX benchmark_model ON benchmark_measurements(model_id,benchmark);
CREATE INDEX unresolved_label ON unresolved_measurements(reported_model,benchmark);
CREATE INDEX price_offer ON price_facts(provider,upstream_id);
CREATE INDEX claim_model ON claims(model_id,polarity);
CREATE TABLE evidence_reviews(batch TEXT PRIMARY KEY REFERENCES review_results(batch),scope TEXT NOT NULL CHECK(scope='published_measurements'),state TEXT NOT NULL CHECK(state='source_validated'),review_path TEXT NOT NULL REFERENCES source_files(path));
CREATE TABLE evidence_review_inputs(batch TEXT NOT NULL REFERENCES evidence_reviews(batch),path TEXT NOT NULL,sha256 TEXT NOT NULL CHECK(length(sha256)=64),PRIMARY KEY(batch,path),FOREIGN KEY(path,sha256) REFERENCES source_files(path,sha256));
CREATE VIEW source_validated_benchmarks AS SELECT b.* FROM benchmark_measurements b JOIN evidence_reviews r ON b.batch=r.batch WHERE r.state='source_validated';
CREATE VIEW source_validated_unresolved AS SELECT b.* FROM unresolved_measurements b JOIN evidence_reviews r ON b.batch=r.batch WHERE r.state='source_validated';
CREATE VIEW accepted_benchmarks AS SELECT b.* FROM benchmark_measurements b JOIN review_results r ON b.batch=r.batch WHERE r.grade='accepted';
CREATE VIEW coverage AS SELECT m.id,m.kind,CASE WHEN EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=m.id) THEN 'scores_found' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id AND r.benchmark_status='not_found_in_checked_sources') THEN 'no_scores_in_checked_sources' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id) THEN 'unresearched' ELSE 'out_of_research_scope' END benchmark_coverage FROM catalog_models m;`
