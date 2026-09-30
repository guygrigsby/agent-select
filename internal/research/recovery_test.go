package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Recovery tests construct every journal boundary state §3.1 defines and
// assert the recovery action. This is state-construction testing of the
// crash-recovery state machine: constructing the exact on-disk state a
// SIGKILL at each boundary leaves is equivalent to (and more deterministic
// than) killing a live process mid-publication.

func mkFile(t *testing.T, root, rel, content string) (path string, hash string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := HashFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return p, h
}

func journalFor(t *testing.T, entries []journalEntry) []byte {
	t.Helper()
	raw, err := json.Marshal(journalFile{Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// outputsFor writes fresh "new" outputs plus staged copies, returning
// journal entries where OldSHA256 is "old" (unverifiable in this harness).
func outputsFor(t *testing.T, root, tag string) []journalEntry {
	t.Helper()
	var entries []journalEntry
	for _, name := range OutputNames {
		_, h := mkFile(t, root, name, "new-"+tag+"-"+name)
		staged := BuildStagingDir + "/" + filepath.Base(name)
		mkFile(t, root, staged, "new-"+tag+"-"+name)
		entries = append(entries, journalEntry{
			Staged: staged, Dest: name, NewSHA256: h, OldSHA256: strings.Repeat("0", 64),
		})
	}
	return entries
}

func writeJournal(t *testing.T, root string, raw []byte) {
	t.Helper()
	mkFile(t, root, JournalPath, string(raw))
}

func TestRecoverNoJournal(t *testing.T) {
	root := t.TempDir()
	did, err := recoverPublication(root)
	if err != nil || did {
		t.Fatalf("no journal: did=%v err=%v", did, err)
	}
}

func TestRecoverCompleteJournal(t *testing.T) {
	root := t.TempDir()
	entries := outputsFor(t, root, "complete")
	writeJournal(t, root, journalFor(t, entries))
	did, err := recoverPublication(root)
	if err != nil || !did {
		t.Fatalf("recover: did=%v err=%v", did, err)
	}
	if _, err := os.Stat(filepath.Join(root, JournalPath)); !os.IsNotExist(err) {
		t.Fatal("journal not cleared")
	}
	for _, e := range entries {
		if h, _ := HashFile(filepath.Join(root, e.Dest)); h != e.NewSHA256 {
			t.Fatalf("%s disturbed on complete-journal recovery", e.Dest)
		}
	}
}

func TestRecoverRenameNeeded(t *testing.T) {
	root := t.TempDir()
	entries := outputsFor(t, root, "rename")
	// simulate crash between renames: drop the first destination
	victim := filepath.Join(root, filepath.FromSlash(entries[0].Dest))
	os.Remove(victim)
	writeJournal(t, root, journalFor(t, entries))
	did, err := recoverPublication(root)
	if err != nil || !did {
		t.Fatalf("recover: did=%v err=%v", did, err)
	}
	for _, e := range entries {
		h, herr := HashFile(filepath.Join(root, filepath.FromSlash(e.Dest)))
		if herr != nil || h != e.NewSHA256 {
			t.Fatalf("%s not restored from staging", e.Dest)
		}
	}
	if _, err := os.Stat(filepath.Join(root, JournalPath)); !os.IsNotExist(err) {
		t.Fatal("journal not cleared")
	}
}

func TestRecoverCorruptRollbackFromBackup(t *testing.T) {
	root := t.TempDir()
	// previous generation outputs + verified backup
	var backup backupManifest
	for _, name := range OutputNames {
		p, h := mkFile(t, root, name, "old-"+name)
		_ = p
		mkFile(t, root, BackupDir+"/"+filepath.Base(name), "old-"+name)
		backup.Files = append(backup.Files, backupManifestEntry{Path: name, SHA256: h})
	}
	raw, _ := json.Marshal(backup)
	mkFile(t, root, BackupManifest, string(raw))

	// journal claims new hashes, but a rename half-applied garbage and
	// staging is gone
	entries := outputsFor(t, root, "corrupt")
	os.Remove(filepath.Join(root, filepath.FromSlash(BuildStagingDir)))
	os.RemoveAll(filepath.Join(root, filepath.FromSlash(BuildStagingDir)))
	mkFile(t, root, entries[0].Dest, "garbage-partial")
	writeJournal(t, root, journalFor(t, entries))

	did, err := recoverPublication(root)
	if err != nil || !did {
		t.Fatalf("recover: did=%v err=%v", did, err)
	}
	for _, e := range backup.Files {
		h, herr := HashFile(filepath.Join(root, filepath.FromSlash(e.Path)))
		if herr != nil || h != e.SHA256 {
			t.Fatalf("rollback did not restore %s to its backed-up bytes", e.Path)
		}
	}
	if _, err := os.Stat(filepath.Join(root, JournalPath)); !os.IsNotExist(err) {
		t.Fatal("journal not cleared after rollback")
	}
}

func TestRecoverCorruptFirstGenerationDeletesPartials(t *testing.T) {
	root := t.TempDir()
	entries := outputsFor(t, root, "firstgen")
	// no backup manifest, staging vanished, one partial dest exists
	os.RemoveAll(filepath.Join(root, filepath.FromSlash(BuildStagingDir)))
	mkFile(t, root, entries[0].Dest, "partial-bytes")
	writeJournal(t, root, journalFor(t, entries))
	did, err := recoverPublication(root)
	if err != nil || !did {
		t.Fatalf("recover: did=%v err=%v", did, err)
	}
	for _, e := range entries {
		if _, serr := os.Stat(filepath.Join(root, filepath.FromSlash(e.Dest))); serr == nil {
			t.Fatalf("%s survived first-generation rollback", e.Dest)
		}
	}
}

func TestRecoverTornJournalClearsWhenConsistent(t *testing.T) {
	root := t.TempDir()
	// crash before journal fsync: garbage journal bytes
	var backup backupManifest
	for _, name := range OutputNames {
		_, h := mkFile(t, root, name, "old-"+name)
		mkFile(t, root, BackupDir+"/"+filepath.Base(name), "old-"+name)
		backup.Files = append(backup.Files, backupManifestEntry{Path: name, SHA256: h})
	}
	raw, _ := json.Marshal(backup)
	mkFile(t, root, BackupManifest, string(raw))
	writeJournal(t, root, []byte("{\"entries\": [ {\"staged\": \"data/.resea"))

	did, err := recoverPublication(root)
	if err != nil || !did {
		t.Fatalf("recover: did=%v err=%v", did, err)
	}
	if _, serr := os.Stat(filepath.Join(root, JournalPath)); serr == nil {
		t.Fatal("torn journal not cleared")
	}
}

func TestRecoverTornJournalFirstGenerationRequiresNoOutputs(t *testing.T) {
	root := t.TempDir()
	writeJournal(t, root, []byte("not json"))
	mkFile(t, root, ArtifactDB, "stray")
	_, err := recoverPublication(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("torn journal with stray outputs must exit 3, got %v", err)
	}
}

func TestRecoverTornJournalRejectsDivergenceFromBackup(t *testing.T) {
	root := t.TempDir()
	var backup backupManifest
	for _, name := range OutputNames {
		_, h := mkFile(t, root, name, "old-"+name)
		mkFile(t, root, BackupDir+"/"+filepath.Base(name), "old-"+name)
		backup.Files = append(backup.Files, backupManifestEntry{Path: name, SHA256: h})
	}
	raw, _ := json.Marshal(backup)
	mkFile(t, root, BackupManifest, string(raw))
	// published bytes drifted after backup
	mkFile(t, root, ArtifactDB, "drifted")
	writeJournal(t, root, []byte("garbage"))
	_, err := recoverPublication(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("diverging published bytes must exit 3, got %v", err)
	}
}

// TestCheckReadOnlyUnderJournal pins check's exit-3 contract while a
// journal exists.
func TestCheckReadOnlyUnderJournal(t *testing.T) {
	root := copyCorpus(t)
	if _, err := Rebuild(root); err != nil {
		t.Fatalf("baseline rebuild: %v", err)
	}
	writeJournal(t, root, journalFor(t, outputsFor(t, root, "check")))
	_, err := Check(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("check under journal must exit 3, got %v", err)
	}
	// rebuild recovers and returns to a healthy state
	if _, err := Rebuild(root); err != nil {
		t.Fatalf("rebuild after journal: %v", err)
	}
	if _, err := Check(root); err != nil {
		t.Fatalf("check after recovery: %v", err)
	}
}

// TestRecoverBackupVerificationFailure covers the §3.1 unrecoverable path:
// backup bytes failing verification is exit 3, never a publish on top of a
// mixed generation.
func TestRecoverBackupVerificationFailure(t *testing.T) {
	root := t.TempDir()
	var backup backupManifest
	for _, name := range OutputNames {
		_, h := mkFile(t, root, name, "old-"+name)
		mkFile(t, root, BackupDir+"/"+filepath.Base(name), "old-"+name)
		backup.Files = append(backup.Files, backupManifestEntry{Path: name, SHA256: h})
	}
	raw, _ := json.Marshal(backup)
	mkFile(t, root, BackupManifest, string(raw))
	// corrupt one backup copy so verification fails
	mkFile(t, root, BackupDir+"/"+filepath.Base(OutputNames[0]), "corrupted-backup")
	// force the corrupt-generation path: journal names hashes nothing matches
	writeJournal(t, root, journalFor(t, []journalEntry{{
		Staged: BuildStagingDir + "/gone", Dest: ArtifactDB,
		NewSHA256: strings.Repeat("c", 64), OldSHA256: strings.Repeat("0", 64),
	}}))
	_, err := recoverPublication(root)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitInconsist {
		t.Fatalf("backup verification failure must exit 3, got %v", err)
	}
	if _, serr := os.Stat(filepath.Join(root, JournalPath)); serr != nil {
		t.Fatal("journal must survive an unrecoverable rollback for forensics")
	}
}

// TestLockTimeoutExit4 pins the exit-4 lock contract.
func TestLockTimeoutExit4(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "data"), 0o755)
	held, err := acquireLock(root, true, time.Second)
	if err != nil {
		t.Fatalf("hold lock: %v", err)
	}
	defer held.Release()
	_, err = acquireLock(root, false, 150*time.Millisecond)
	ce, ok := err.(*Error)
	if !ok || ce.Code != ExitIO {
		t.Fatalf("lock timeout must exit 4, got %v", err)
	}
}
