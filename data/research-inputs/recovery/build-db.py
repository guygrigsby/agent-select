import json,sqlite3,pathlib,hashlib,datetime,collections,re,ast,csv
from decimal import Decimal
root=pathlib.Path('/Users/guygrigsby/projects/agent-routing');db=root/'data/research.sqlite'
if db.exists():db.unlink()
c=sqlite3.connect(db);c.execute('PRAGMA foreign_keys=ON');c.executescript('''
CREATE TABLE metadata(key TEXT PRIMARY KEY,value TEXT NOT NULL);
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
CREATE VIEW coverage AS SELECT m.id,m.kind,CASE WHEN EXISTS(SELECT 1 FROM benchmark_measurements b WHERE b.model_id=m.id) THEN 'scores_found' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id AND r.benchmark_status='not_found_in_checked_sources') THEN 'no_scores_in_checked_sources' WHEN EXISTS(SELECT 1 FROM model_research r WHERE r.model_id=m.id) THEN 'unresearched' ELSE 'out_of_research_scope' END benchmark_coverage FROM catalog_models m;
''')
urlmap={'data/raw/vercel-models.json':'https://ai-gateway.vercel.sh/v1/models','data/raw/vercel.headers':'https://ai-gateway.vercel.sh/v1/models','data/raw/models-dev.json':'https://models.dev/api.json','data/raw/models-dev.headers':'https://models.dev/api.json'}
licenses={}
# Research source rows are explicit provenance. Baseline-derived files are not primary downloads.
for f in (root/'data/research').glob('*/findings.json'):
 d=json.loads(f.read_text())
 for s in d.get('sources',[]):
  p=s.get('file','');
  if (root/p).is_file():urlmap[p]=s['url'];licenses[p]=s.get('license','unknown')
 for b in d.get('benchmarks',[])+d.get('measurements',[]):urlmap[b['source_file']]=b['source_url']
# Recover literal filename/URL tuples from fetching commands without executing them.
for log in pathlib.Path('/tmp/agent-routing-research').glob('*.log'):
 batch=log.stem.split('-extract')[0].split('-mid')[0]
 if batch.startswith(('fix-','review-')):continue
 base=root/'data/research'/batch
 for line in log.read_text(errors='replace').splitlines():
  if not line.startswith('TOOL bash '):continue
  try:cmd=json.loads(line[len('TOOL bash '):])['command']
  except Exception:continue
  for a,b in re.findall(r'''["']([^"'\n]+)["']\s*,\s*["']([^"'\n]+)["']''',cmd):
   u,p=(a,b) if a.startswith('https://') else (b,a)
   if not u.startswith('https://') or not re.search(r'\.(html|md|json|pdf|csv|yml)$',p):continue
   paths=list(base.rglob(pathlib.Path(p).name)) if base.exists() else []
   if len(paths)==1:urlmap.setdefault(str(paths[0].relative_to(root)),u)
rows=[]
for p in sorted((root/'data').rglob('*')):
 if not p.is_file() or p==db or p.name in {'source-manifest.tsv','source-hashes.sha256'}:continue
 rel=str(p.relative_to(root));content=p.read_bytes();stamp=datetime.datetime.fromtimestamp(p.stat().st_mtime,datetime.timezone.utc).isoformat()
 row=(rel,urlmap.get(rel,'unknown'),hashlib.sha256(content).hexdigest(),len(content),stamp,'local_file_mtime_not_http_retrieval',licenses.get(rel,'unknown'),'unknown')
 c.execute('INSERT INTO source_files VALUES(?,?,?,?,?,?,?,?)',row);rows.append(row)
with (root/'data/source-manifest.tsv').open('w') as f:
 w=csv.writer(f,delimiter='\t',lineterminator='\n');w.writerow(['path','url','sha256','bytes','observed_at','timestamp_basis','license','publication_permission']);w.writerows(rows)
v=json.loads((root/'data/raw/vercel-models.json').read_text())['data'];m=json.loads((root/'data/raw/models-dev.json').read_text())
def prices(obj,path=''):
 if isinstance(obj,dict):
  for k,x in obj.items():yield from prices(x,path+'/'+k)
 elif isinstance(obj,list):
  for i,x in enumerate(obj):yield from prices(x,path+'/'+str(i))
 elif not isinstance(obj,bool):
  try:Decimal(str(obj))
  except Exception:return
  yield path,str(obj)
for x in v:
 c.execute('INSERT INTO catalog_models VALUES(?,?,?,?,?,?)',(x['id'],x['name'],x['owned_by'],x['type'],x['description'],'data/raw/vercel-models.json'))
 for di,mods in x['modalities'].items():
  for mod in mods:c.execute('INSERT INTO catalog_modalities VALUES(?,?,?)',(x['id'],di,mod))
 for key in ['context_window','max_tokens']:
  if key in x:c.execute('INSERT INTO catalog_limits VALUES(?,?,?)',(x['id'],key,x[key]))
 for k in x.get('supported_parameters',[]):c.execute('INSERT OR IGNORE INTO catalog_parameters VALUES(?,?)',(x['id'],k))
 for k in x.get('tags',[]):c.execute('INSERT OR IGNORE INTO catalog_tags VALUES(?,?)',(x['id'],k))
 for path,num in prices(x['pricing']):
  leaf=path.split('/')[-1]; unit='USD_per_token' if leaf in {'input','output','input_cache_read','input_cache_write','audio_input_token_cost','audio_output_token_cost'} else 'USD_per_image' if leaf=='image' else 'source_specific_unresolved'
  c.execute('INSERT INTO price_facts VALUES(?,?,?,?,?,?)',('data/raw/vercel-models.json','vercel',x['id'],path,num,unit))
for provider,p in m.items():
 for uid,x in p.get('models',{}).items():
  c.execute('INSERT INTO provider_offers VALUES(?,?,?,?)',(provider,uid,x.get('name',uid),'data/raw/models-dev.json'))
  if x.get('canonical_model_id'):c.execute('INSERT INTO canonical_labels VALUES(?,?,?)',(provider,uid,x['canonical_model_id']))
  for path,num in prices(x.get('cost',{})):c.execute('INSERT INTO price_facts VALUES(?,?,?,?,?,?)',('data/raw/models-dev.json',provider,uid,path,num,'USD_per_million_tokens' if path.split('/')[-1] in {'input','output','cache_read','cache_write'} else 'source_specific_unresolved'))
for x in v:
 y=m['vercel']['models'].get(x['id'])
 if y is None:continue
 c.execute('INSERT INTO exact_catalog_joins VALUES(?,?,?)',(x['id'],'vercel',x['id']))
 for vk,mk in [('input','input'),('output','output'),('input_cache_read','cache_read'),('input_cache_write','cache_write')]:
  if vk in x['pricing'] and mk in y.get('cost',{}):
   a=Decimal(str(x['pricing'][vk]))*1000000;b=Decimal(str(y['cost'][mk]))
   if a!=b:c.execute('INSERT INTO pricing_disagreements VALUES(?,?,?,?)',(x['id'],vk,str(a),str(b)))
for batch in sorted(p.name for p in (root/'data/research').iterdir() if p.is_dir()):
 c.execute('INSERT INTO review_results VALUES(?,?,?)',(batch,'accepted' if batch=='catalog-sources' else 'quarantined','Catalog comparison passed actual-source review' if batch=='catalog-sources' else 'Provisional research evidence. Composite review failed or batch not separately approved; no routing admission.'))
for f in sorted((root/'data/research').glob('*/findings.json')):
 d=json.loads(f.read_text());batch=f.parent.name
 for r in d.get('models',[]):
  c.execute('INSERT INTO model_research VALUES(?,?,?,?)',(r['id'],batch,r['benchmark_status'],r.get('notes','')))
  for pol in ['strengths','weaknesses']:
   for a in r.get(pol,[]):
    source=a['source_file'];url=a['source_url'];loc=a['locator']
    if not (root/source).is_file() and (f.parent/source).is_file():source=str((f.parent/source).relative_to(root))
    if a['basis']=='catalog_claim':source='data/raw/vercel-models.json';url='https://ai-gateway.vercel.sh/v1/models';loc='data entry id='+r['id']+'; '+loc
    if not url:
     url = 'https://models.dev/api.json' if source=='data/raw/models-dev.json' else 'https://ai-gateway.vercel.sh/v1/models' if source.endswith('assignment.json') or source=='data/raw/vercel-models.json' else 'unknown'
    if not source or not (root/source).is_file():
     c.execute('INSERT INTO unlocated_claims VALUES(?,?,?,?,?,?,?,?)',(r['id'],batch,('strength' if pol=='strengths' else 'weakness'),a['basis'],a['text'],url,loc,'No saved evidence file supplied; quarantined'))
     continue
    c.execute('INSERT INTO claims(model_id,batch,polarity,basis,text,source_url,source_path,locator) VALUES(?,?,?,?,?,?,?,?)',(r['id'],batch,('strength' if pol=='strengths' else 'weakness'),a['basis'],a['text'],url,source,loc))
 for b in d.get('benchmarks',[]):
  Decimal(b['value'])
  if not (root/b['source_file']).is_file() and (f.parent/b['source_file']).is_file():b['source_file']=str((f.parent/b['source_file']).relative_to(root))
  assert (root/b['source_file']).is_file()
  c.execute('INSERT INTO benchmark_measurements(model_id,batch,benchmark,benchmark_version,metric,raw_value,unit,configuration,source_url,source_path,locator,evidence_class,identity_basis) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)',tuple([b['model_id'],batch]+[b[k] for k in ['benchmark','benchmark_version','metric','value','unit','configuration','source_url','source_file','locator','evidence_class','identity_basis']]))
 for b in d.get('measurements',[]):
  Decimal(b['value']);cur=c.execute('INSERT INTO unresolved_measurements(batch,reported_model,benchmark,benchmark_version,metric,raw_value,unit,configuration,source_url,source_path,locator,evidence_class,identity_status) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)',(batch,)+tuple(b[k] for k in ['reported_model','benchmark','benchmark_version','metric','value','unit','configuration','source_url','source_file','locator','evidence_class','identity_status']))
  for mid in b.get('proposed_offer_ids',[]):c.execute('INSERT INTO unresolved_identity_proposals VALUES(?,?)',(cur.lastrowid,mid))
 for s in d.get('sources',[]):c.execute('INSERT OR REPLACE INTO source_access VALUES(?,?,?,?,?,?)',(batch,s['url'],s.get('file',''),s['status'],s.get('purpose',''),s.get('license','unknown')))
 for i,gap in enumerate(d.get('gaps',[])):c.execute('INSERT INTO research_gaps VALUES(?,?,?)',(batch,i,gap))
for batch,review in [('other-text','data/reviews/source-evidence-2026-09-30.md'),('unresolved','data/reviews/source-evidence-2026-09-30.md'),('mistral-meta','data/reviews/source-evidence-corrections-2026-09-30.md'),('coverage-followup','data/reviews/source-evidence-corrections-2026-09-30.md'),('google','data/reviews/google-moonshot-source-validation.md'),('moonshot-minimax','data/reviews/google-moonshot-source-validation.md')]:
 c.execute('INSERT INTO evidence_reviews VALUES(?,?,?,?)',(batch,'published_measurements','source_validated',review))
for batch,review in c.execute('SELECT batch,review_path FROM evidence_reviews').fetchall():
 paths={review,'data/research/'+batch+'/findings.json'}
 paths.update(r[0] for r in c.execute('SELECT source_path FROM benchmark_measurements WHERE batch=? UNION SELECT source_path FROM unresolved_measurements WHERE batch=?',(batch,batch)))
 for path in sorted(paths):
  sha=c.execute('SELECT sha256 FROM source_files WHERE path=?',(path,)).fetchone()[0]
  c.execute('INSERT INTO evidence_review_inputs VALUES(?,?,?)',(batch,path,sha))
for key,value in {'artifact_version':'0.1.2','schema_version':'3','phase':'research_only_not_routing_authority','baseline_url':'https://ai-gateway.vercel.sh/v1/models','pricing_crossref_url':'https://models.dev/api.json','collection_date_utc':'2026-09-30','publication_permission':'unknown','expected_catalog_models':str(len(v)),'expected_exact_joins':'394','expected_text_and_image_research_models':'302'}.items():c.execute('INSERT INTO metadata VALUES(?,?)',(key,value))
c.commit();print('integrity',c.execute('PRAGMA integrity_check').fetchall());print('foreign_keys',c.execute('PRAGMA foreign_key_check').fetchall());
for table in ['catalog_models','provider_offers','exact_catalog_joins','claims','benchmark_measurements','unresolved_measurements','price_facts','pricing_disagreements','model_research','source_files']:print(table,c.execute('SELECT COUNT(*) FROM '+table).fetchone()[0])
print('coverage',c.execute('SELECT benchmark_coverage,count(*) FROM coverage GROUP BY 1').fetchall());c.close()
