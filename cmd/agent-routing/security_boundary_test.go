package main

// Security boundary regression cases for the CLI adapter (review round 1,
// isolated probe at 2ad2ceb): (1) a valid request document followed by
// stray closers or closers plus a hostile value must exit 2 with a one-line
// invalid_request_wire diagnostic that echoes none of the request bytes;
// (2) snapshot adopt must never pass raw storage errors through to stderr,
// so a store path carrying an embedded-newline marker produces exactly one
// safe line instead of a forged multi-line log.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// adoptMarker is the hostile basename planted in store paths: its embedded
// newline forges a second log line when a raw path crosses the boundary,
// and its credential-looking bytes must never appear in a diagnostic.
const adoptMarker = "credential_secret\nFORGED_LOG_LINE"

// routePoisonedSuffixes mirrors routing's securityPoisonedSuffixes: byte
// sequences appended after a complete wire request that the strict
// one-value rule must reject at the CLI boundary.
var routePoisonedSuffixes = []string{
	"}\n",
	"]\n",
	"} " + `{"credential":"secret"}` + "\n",
	" " + `{"credential":"secret"}` + "\n",
}

func TestRouteRejectsPoisonedWireSuffixes(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "snapshot.bin")
	writeFileT(t, art, adapterArtifact(t))
	base := strings.TrimSuffix(string(requestWire(t, adapterRequest("code", nil))), "\n")
	for _, suf := range routePoisonedSuffixes {
		req := writeReq(t, dir, "poison", base+suf)
		stdout, stderr, exit := captureRun(t, []string{"route", "--snapshot", art, "--request", req})
		if exit != adapterExitUsage {
			t.Fatalf("suffix %q: exit = %d, want %d (stdout %q stderr %q)", suf, exit, adapterExitUsage, stdout, stderr)
		}
		if stdout != "" {
			t.Fatalf("suffix %q: nonzero exit wrote stdout: %q", suf, stdout)
		}
		lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
		if len(lines) != 1 {
			t.Fatalf("suffix %q: stderr must be exactly one line: %q", suf, stderr)
		}
		if !strings.HasPrefix(lines[0], "agent-routing: invalid_request_wire: ") {
			t.Fatalf("suffix %q: stderr code form: %q", suf, lines[0])
		}
		if strings.Contains(stderr, "credential") || strings.Contains(stderr, "secret") {
			t.Fatalf("suffix %q: diagnostic echoes untrusted request bytes: %q", suf, stderr)
		}
	}
}

func TestSnapshotAdoptStoreErrorBoundary(t *testing.T) {
	dir := t.TempDir()
	art := filepath.Join(dir, "artifact.bin")
	writeFileT(t, art, adapterArtifact(t))

	// A regular file where the store directory should be: the raw mkdir
	// failure carries the caller-supplied path, embedded newline included.
	storeFile := filepath.Join(dir, adoptMarker)
	writeFileT(t, storeFile, []byte("not a directory"))

	stdout, stderr, exit := captureRun(t, []string{"snapshot", "adopt", "--artifact", art, "--store", storeFile})
	if exit != adapterExitSnapshot {
		t.Fatalf("exit = %d, want %d (stdout %q stderr %q)", exit, adapterExitSnapshot, stdout, stderr)
	}
	assertSafeOneLine(t, stderr, stdout, "snapshot_unavailable")
}

func TestSnapshotAdoptStageAdoptErrorsSafe(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-based failure injection needs a non-root user")
	}
	dir := t.TempDir()
	art := filepath.Join(dir, "artifact.bin")
	writeFileT(t, art, adapterArtifact(t))
	markerDir := filepath.Join(dir, adoptMarker)

	// Stage fails: the snapshots subdirectory is read-only, so the temp-file
	// write carries the store path in the raw error.
	if err := os.MkdirAll(filepath.Join(markerDir, "snapshots"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(markerDir, "snapshots"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(markerDir, "snapshots"), 0o755) })
	stdout, stderr, exit := captureRun(t, []string{"snapshot", "adopt", "--artifact", art, "--store", markerDir})
	if exit != adapterExitSnapshot {
		t.Fatalf("stage case: exit = %d, want %d (stdout %q stderr %q)", exit, adapterExitSnapshot, stdout, stderr)
	}
	assertSafeOneLine(t, stderr, stdout, "snapshot_unavailable")
	os.Chmod(filepath.Join(markerDir, "snapshots"), 0o755)

	// Adopt fails: staging succeeds, but the store root is read-only so the
	// CURRENT swap carries the store path in the raw error.
	if err := os.Chmod(markerDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(markerDir, 0o755) })
	stdout, stderr, exit = captureRun(t, []string{"snapshot", "adopt", "--artifact", art, "--store", markerDir})
	if exit != adapterExitSnapshot {
		t.Fatalf("adopt case: exit = %d, want %d (stdout %q stderr %q)", exit, adapterExitSnapshot, stdout, stderr)
	}
	assertSafeOneLine(t, stderr, stdout, "snapshot_unavailable")
}

// assertSafeOneLine pins the boundary rule for every nonzero exit: exactly
// one stderr line in the code form, no stdout, and none of the hostile
// marker bytes.
func assertSafeOneLine(t *testing.T, stderr, stdout, code string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("nonzero exit wrote stdout: %q", stdout)
	}
	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("stderr must be exactly one line: %q", stderr)
	}
	if !strings.HasPrefix(lines[0], "agent-routing: "+code+": ") {
		t.Fatalf("stderr code form: %q", lines[0])
	}
	if strings.Contains(stderr, "FORGED_LOG_LINE") || strings.Contains(stderr, "credential_secret") {
		t.Fatalf("diagnostic echoes caller-supplied path bytes: %q", stderr)
	}
}
