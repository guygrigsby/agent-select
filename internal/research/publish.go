package research

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LockTimeout bounds lock acquisition (exit 4 past it).
const LockTimeout = 30 * time.Second

// journalEntry is one output's publication record (§3 step 6).
type journalEntry struct {
	Staged    string `json:"staged"`
	Dest      string `json:"dest"`
	NewSHA256 string `json:"new_sha256"`
	OldSHA256 string `json:"old_sha256"`
}

type journalFile struct {
	Entries []journalEntry `json:"entries"`
}

// backupManifestEntry records one backed-up output's old hash.
type backupManifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type backupManifest struct {
	Files []backupManifestEntry `json:"files"`
}

// loadInputs loads and validates every durable input (gates 4/5 pre-staging
// included) and returns the fully assembled build.
func loadInputs(root string) (*Config, *Inventory, *Ledger, *BuildResult, []*tableRows, error) {
	cfg, err := LoadConfig(root)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if cfg.SchemaVersion != SchemaVersion {
		return nil, nil, nil, nil, nil, inconsf("build.json schema_version %d but tool carries DDL schema %d; refusing to rebuild", cfg.SchemaVersion, SchemaVersion)
	}
	if err := CheckDriverPin(cfg); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	inv, err := LoadInventory(root)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := inv.VerifyDisk(root); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	led, err := LoadDecisions(root)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := led.CheckReviewPathsInInventory(inv); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	// batches carrying measurement rows, for the pin gate's scope binding
	batchesWithMeasurements := map[string]bool{}
	batchDirs, err := listBatchDirs(root)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	adapterOf := cfg.AdapterOf()
	for _, batch := range batchDirs {
		if adapterOf[batch] == "" {
			return nil, nil, nil, nil, nil, invalidf("build.json: batch directory data/research/%s has no batch_formats mapping", batch)
		}
		// Gate 5: the findings file is a read evidence path and must be
		// inventoried like every other read path.
		fp := "data/research/" + batch + "/findings.json"
		if inv.Get(fp) == nil {
			return nil, nil, nil, nil, nil, invalidf("%s: batch findings file is not in the capture inventory (gate 5)", fp)
		}
		b, err := LoadBatch(root, batch, adapterOf[batch])
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if b.HasMeasurements() {
			batchesWithMeasurements[batch] = true
		}
	}
	if err := led.CheckPinGate(inv, adapterOf, batchesWithMeasurements); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	res, tables, err := assemble(root, cfg, inv, led)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return cfg, inv, led, res, tables, nil
}

// checkExpected compares derived counts against build.json expectations
// (§2.2: mismatch is exit 3).
func checkExpected(cfg *Config, res *BuildResult) error {
	if res.CatalogModels != cfg.Expected.CatalogModels {
		return inconsf("derived catalog_models %d but build.json expects %d", res.CatalogModels, cfg.Expected.CatalogModels)
	}
	if res.ExactJoins != cfg.Expected.ExactCatalogJoins {
		return inconsf("derived exact_catalog_joins %d but build.json expects %d", res.ExactJoins, cfg.Expected.ExactCatalogJoins)
	}
	if res.ModelResearch != cfg.Expected.ModelResearch {
		return inconsf("derived model_research %d but build.json expects %d", res.ModelResearch, cfg.Expected.ModelResearch)
	}
	if res.PriceFacts != cfg.Expected.PriceFacts {
		return inconsf("derived price_facts %d but build.json expects %d", res.PriceFacts, cfg.Expected.PriceFacts)
	}
	if res.PricingDisagreements != cfg.Expected.PricingDisagreements {
		return inconsf("derived pricing_disagreements %d but build.json expects %d", res.PricingDisagreements, cfg.Expected.PricingDisagreements)
	}
	return nil
}

// stageGeneration builds the complete artifact set into the staging
// directory and validates it (§3 step 4-5). Returns the staged hashes.
func stageGeneration(root string, cfg *Config, inv *Inventory, tables []*tableRows) (map[string]string, error) {
	stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
	if err := os.RemoveAll(stg); err != nil {
		return nil, iof("clean staging: %v", err)
	}
	if err := os.MkdirAll(stg, 0o755); err != nil {
		return nil, iof("mkdir staging: %v", err)
	}
	dbPath := filepath.Join(stg, StagedDB)
	tsvPath := filepath.Join(stg, StagedManifest)
	hashPath := filepath.Join(stg, StagedHashes)

	if err := writeDB(dbPath, tables); err != nil {
		return nil, err
	}
	tsv := BuildManifestTSV(inv)
	if err := writeSynced(tsvPath, tsv); err != nil {
		return nil, err
	}
	manifest, _, err := BuildHashManifest(root, inv, dbPath, tsvPath)
	if err != nil {
		return nil, err
	}
	if err := writeSynced(hashPath, manifest); err != nil {
		return nil, err
	}
	if err := syncDir(stg); err != nil {
		return nil, err
	}
	return map[string]string{
		ArtifactDB:       mustHash(dbPath),
		ArtifactManifest: mustHash(tsvPath),
		ArtifactHashes:   mustHash(hashPath),
	}, nil
}

// validateStaged opens the staged DB and runs gates 1-3 plus determinism
// (a second full build must byte-match).
func validateStaged(root string, cfg *Config, inv *Inventory, res *BuildResult, tables []*tableRows) error {
	distinctModels := map[string]bool{}
	for _, t := range tables {
		if t.name != "model_research" {
			continue
		}
		for _, r := range t.rows {
			distinctModels[r[0].(string)] = true
		}
	}
	res2 := *res
	res2.ModelResearch = len(distinctModels)
	if err := checkExpected(cfg, &res2); err != nil {
		return err
	}
	stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
	dbPath := filepath.Join(stg, StagedDB)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return iof("open staged db: %v", err)
	}
	defer db.Close()
	if err := RunGatesDB(root, db); err != nil {
		return err
	}
	if err := db.Close(); err != nil {
		return iof("close staged db: %v", err)
	}
	// repeat-build determinism: second build must produce identical bytes
	detPath := filepath.Join(stg, ".determinism.db")
	if err := writeDB(detPath, tables); err != nil {
		return err
	}
	h1, h2 := mustHash(dbPath), mustHash(detPath)
	if h1 != h2 {
		return inconsf("determinism: repeat build produced different DB bytes (%s vs %s)", h1[:8], h2[:8])
	}
	if err := os.Remove(detPath); err != nil {
		return iof("remove determinism probe: %v", err)
	}
	return nil
}

// writeSynced writes bytes to p and fsyncs the file.
func writeSynced(p string, b []byte) error {
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return iof("write %s: %v", p, err)
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return iof("write %s: %v", p, err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return iof("fsync %s: %v", p, err)
	}
	return f.Close()
}

func syncDir(p string) error {
	d, err := os.Open(p)
	if err != nil {
		return iof("open dir %s: %v", p, err)
	}
	if err := d.Sync(); err != nil {
		d.Close()
		return iof("fsync dir %s: %v", p, err)
	}
	return d.Close()
}

func mustHash(p string) string {
	h, err := HashFile(p)
	if err != nil {
		panic(err)
	}
	return h
}

// readBackupManifest loads data/.research-backup/manifest.json if present.
func readBackupManifest(root string) (*backupManifest, bool, error) {
	p := filepath.Join(root, filepath.FromSlash(BackupManifest))
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, iof("read backup manifest: %v", err)
	}
	var bm backupManifest
	if err := json.Unmarshal(raw, &bm); err != nil {
		return nil, true, iof("parse backup manifest: %v", err)
	}
	return &bm, true, nil
}

// backupGeneration copies the current outputs into the backup area and
// records their hashes (§3 step 3). Absent outputs produce an empty set.
func backupGeneration(root string) (*backupManifest, error) {
	bd := filepath.Join(root, filepath.FromSlash(BackupDir))
	if err := os.RemoveAll(bd); err != nil {
		return nil, iof("clean backup dir: %v", err)
	}
	bm := &backupManifest{}
	for _, name := range OutputNames {
		src := filepath.Join(root, filepath.FromSlash(name))
		h, err := HashFile(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, iof("hash %s for backup: %v", name, err)
		}
		if err := os.MkdirAll(bd, 0o755); err != nil {
			return nil, iof("mkdir backup: %v", err)
		}
		dst := filepath.Join(bd, filepath.Base(name))
		raw, err := os.ReadFile(src)
		if err != nil {
			return nil, iof("read %s for backup: %v", name, err)
		}
		if err := writeSynced(dst, raw); err != nil {
			return nil, err
		}
		bm.Files = append(bm.Files, backupManifestEntry{Path: name, SHA256: h})
	}
	raw, err := json.MarshalIndent(bm, "", "  ")
	if err != nil {
		return nil, iof("marshal backup manifest: %v", err)
	}
	if err := os.MkdirAll(bd, 0o755); err != nil {
		return nil, iof("mkdir backup: %v", err)
	}
	if err := writeSynced(filepath.Join(bd, "manifest.json"), append(raw, '\n')); err != nil {
		return nil, err
	}
	return bm, nil
}

// verifyBackup restores-time check: every backup file must hash to its
// recorded old hash.
func verifyBackup(root string, bm *backupManifest) error {
	for _, e := range bm.Files {
		p := filepath.Join(root, filepath.FromSlash(BackupDir), filepath.Base(e.Path))
		h, err := HashFile(p)
		if err != nil {
			return inconsf("backup verification: %s unreadable: %v", e.Path, err)
		}
		if h != e.SHA256 {
			return inconsf("backup verification: %s hash %s but manifest pins %s", e.Path, h[:8], e.SHA256[:8])
		}
	}
	return nil
}

// recoverPublication runs §3.1 before staging. Returns true if a recovery
// action was taken.
func recoverPublication(root string) (bool, error) {
	jp := filepath.Join(root, filepath.FromSlash(JournalPath))
	raw, err := os.ReadFile(jp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, iof("read journal: %v", err)
	}
	var j journalFile
	if jerr := json.Unmarshal(raw, &j); jerr != nil {
		// torn journal: crash happened before fsync, before any rename.
		bm, _, err := readBackupManifest(root)
		if err != nil {
			return false, err
		}
		if bm != nil {
			if err := verifyBackup(root, bm); err != nil {
				return false, err
			}
			for _, e := range bm.Files {
				p := filepath.Join(root, filepath.FromSlash(e.Path))
				h, herr := HashFile(p)
				if herr != nil && !os.IsNotExist(herr) {
					return false, iof("hash %s during torn-journal recovery: %v", e.Path, herr)
				}
				if herr == nil && h != e.SHA256 {
					return false, inconsf("torn journal with published %s diverging from backup manifest: %s vs %s", e.Path, h[:8], e.SHA256[:8])
				}
			}
		} else {
			// no backup manifest: outputs must not exist (first generation)
			for _, name := range OutputNames {
				if _, serr := os.Stat(filepath.Join(root, filepath.FromSlash(name))); serr == nil {
					return false, inconsf("torn journal but %s exists with no backup manifest", name)
				}
			}
		}
		if err := os.Remove(jp); err != nil {
			return false, iof("remove torn journal: %v", err)
		}
		stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
		if err := os.RemoveAll(stg); err != nil {
			return false, iof("clean staging after torn journal: %v", err)
		}
		return true, nil
	}

	corrupt := false
	for _, e := range j.Entries {
		dest := filepath.Join(root, filepath.FromSlash(e.Dest))
		dh, derr := HashFile(dest)
		if derr == nil && dh == e.NewSHA256 {
			continue // complete
		}
		staged := filepath.Join(root, filepath.FromSlash(e.Staged))
		sh, serr := HashFile(staged)
		if (derr == nil && dh != e.NewSHA256) || serr != nil || sh != e.NewSHA256 {
			corrupt = true
			break
		}
		if derr != nil && os.IsNotExist(derr) {
			if err := os.Rename(staged, dest); err != nil {
				return false, iof("recovery rename %s: %v", e.Dest, err)
			}
			continue
		}
		corrupt = true
		break
	}

	if corrupt {
		bm, exists, err := readBackupManifest(root)
		if err != nil {
			return false, err
		}
		if exists && len(bm.Files) > 0 {
			if err := verifyBackup(root, bm); err != nil {
				return false, err
			}
			for _, e := range bm.Files {
				src := filepath.Join(root, filepath.FromSlash(BackupDir), filepath.Base(e.Path))
				dst := filepath.Join(root, filepath.FromSlash(e.Path))
				raw, rerr := os.ReadFile(src)
				if rerr != nil {
					return false, iof("rollback read %s: %v", e.Path, rerr)
				}
				if err := writeSynced(dst, raw); err != nil {
					return false, err
				}
			}
		} else {
			for _, e := range j.Entries {
				p := filepath.Join(root, filepath.FromSlash(e.Dest))
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					return false, iof("rollback remove %s: %v", e.Dest, err)
				}
			}
		}
		stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
		if err := os.RemoveAll(stg); err != nil {
			return false, iof("clean staging after rollback: %v", err)
		}
	} else {
		// all entries complete: validate destinations one last time
		for _, e := range j.Entries {
			dh, derr := HashFile(filepath.Join(root, filepath.FromSlash(e.Dest)))
			if derr != nil || dh != e.NewSHA256 {
				return false, inconsf("recovery: %s does not hash to the journaled new value", e.Dest)
			}
		}
	}
	if err := os.Remove(jp); err != nil {
		return false, iof("clear journal: %v", err)
	}
	if err := syncDir(filepath.Join(root, "data")); err != nil {
		return false, err
	}
	return true, nil
}

// publish runs steps 6-8: journal, rename, validate in place, clear.
func publish(root string, stagedHashes map[string]string, bm *backupManifest) error {
	oldHash := map[string]string{}
	for _, e := range bm.Files {
		oldHash[e.Path] = e.SHA256
	}
	j := journalFile{}
	for _, name := range OutputNames {
		j.Entries = append(j.Entries, journalEntry{
			Staged:    BuildStagingDir + "/" + filepath.Base(name),
			Dest:      name,
			NewSHA256: stagedHashes[name],
			OldSHA256: oldHash[name],
		})
	}
	raw, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return iof("marshal journal: %v", err)
	}
	if err := writeSynced(filepath.Join(root, filepath.FromSlash(JournalPath)), append(raw, '\n')); err != nil {
		return err
	}
	// rename in manifest, TSV, DB order
	for _, e := range j.Entries {
		src := filepath.Join(root, filepath.FromSlash(e.Staged))
		dst := filepath.Join(root, filepath.FromSlash(e.Dest))
		if err := os.Rename(src, dst); err != nil {
			return iof("publish rename %s: %v", e.Dest, err)
		}
	}
	// validate in place
	for _, e := range j.Entries {
		h, err := HashFile(filepath.Join(root, filepath.FromSlash(e.Dest)))
		if err != nil || h != e.NewSHA256 {
			return inconsf("post-publication validation failed for %s; journal left in place for recovery", e.Dest)
		}
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(JournalPath))); err != nil {
		return iof("clear journal: %v", err)
	}
	if err := syncDir(filepath.Join(root, "data")); err != nil {
		return err
	}
	// prune staging leftovers (determinism probe etc.)
	stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
	if err := os.RemoveAll(stg); err != nil {
		return iof("prune staging: %v", err)
	}
	return nil
}

// RebuildResult is the structured outcome for --json.
type RebuildResult struct {
	Result       string
	Noop         bool
	OutputHashes map[string]string
	Build        *BuildResult
}

// Rebuild is the research rebuild command (§1, §3).
func Rebuild(root string) (*RebuildResult, error) {
	lk, err := acquireLock(root, true, LockTimeout)
	if err != nil {
		return nil, err
	}
	defer lk.Release()

	if _, err := recoverPublication(root); err != nil {
		return nil, err
	}
	bm, err := backupGeneration(root)
	if err != nil {
		return nil, err
	}
	cfg, inv, _, res, tables, err := loadInputs(root)
	if err != nil {
		return nil, err
	}
	stagedHashes, err := stageGeneration(root, cfg, inv, tables)
	if err != nil {
		return nil, err
	}
	if err := validateStaged(root, cfg, inv, res, tables); err != nil {
		return nil, err
	}
	// no-op detection: published set already identical
	noop := true
	published := map[string]string{}
	for _, name := range OutputNames {
		h, herr := HashFile(filepath.Join(root, filepath.FromSlash(name)))
		if herr != nil || h != stagedHashes[name] {
			noop = false
			break
		}
		published[name] = h
	}
	if noop {
		stg := filepath.Join(root, filepath.FromSlash(BuildStagingDir))
		if err := os.RemoveAll(stg); err != nil {
			return nil, iof("prune staging after no-op: %v", err)
		}
		return &RebuildResult{Result: "ok", Noop: true, OutputHashes: published, Build: res}, nil
	}
	if err := publish(root, stagedHashes, bm); err != nil {
		return nil, err
	}
	return &RebuildResult{Result: "ok", Noop: false, OutputHashes: stagedHashes, Build: res}, nil
}
