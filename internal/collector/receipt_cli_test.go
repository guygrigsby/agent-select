// Automated real-executable/Postgres coverage for the `receipt record`
// host-reporter process boundary (cli-adapter.md V2). Every case builds the
// real cmd/agent-routing binary, invokes it as a real child process and
// checks the real Postgres 16 container started by startPostgres, so the
// whole gate (AR_REQUIRE_POSTGRES=1 go test ./...) exercises the executable
// and the database, not the in-process functions.
//
// Recorded verification gap: a mid-transaction database failure (the
// transaction dies after partial writes) is explicitly unverified here.
// Exercising it needs a fault-injection framework that does not exist and
// is out of scope; the boundary review records the gap instead.
package collector

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// cliReceiptSecret is the canary planted in hostile DSNs and receipt
// documents: it must never appear in a diagnostic.
const cliReceiptSecret = "SUP3RSECRET-CLI-DO-NOT-LEAK-77b2"

// buildReceiptCLI compiles the real adapter binary from the current source
// tree, so a red test proves the shipped executable misbehaves and a green
// test proves the shipped executable is fixed.
func buildReceiptCLI(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	bin := filepath.Join(t.TempDir(), "agent-routing")
	build := exec.Command("go", "build", "-o", bin, "./cmd/agent-routing")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build cmd/agent-routing: %v\n%s", err, out)
	}
	return bin
}

// cliRun is one child-process environment: hermetic HOME (config.toml),
// AGENT_ROUTING_DSN and an optional raw stdout descriptor for the
// read-only-stdout regression.
type cliRun struct {
	home   string
	dsnEnv string
	stdout io.Writer
}

// cliConfigPath mirrors os.UserConfigDir for the child's HOME.
func cliConfigPath(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "agent-routing", "config.toml")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "agent-routing", "config.toml")
	default:
		return filepath.Join(home, ".config", "agent-routing", "config.toml")
	}
}

func cliWriteConfig(t *testing.T, home, contents string) {
	t.Helper()
	path := cliConfigPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func cliWriteFile(t *testing.T, dir, name string, contents []byte) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// cliCanonical encodes with the same canonicalize the binary's strict
// decoders demand.
func cliCanonical(t *testing.T, v any) []byte {
	t.Helper()
	b, err := canonicalize(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// cliUsageUnknown and cliUsageKnown build the seven closed usage members.
func cliUsageKnown(component string, value int64) map[string]any {
	return map[string]any{"component": component, "known": true, "value": value}
}

func cliUsageUnknown(component string) map[string]any {
	return map[string]any{"component": component, "known": false}
}

// cliReceiptDoc builds the canonical-v1 encoding of one closed-grammar
// receipt document with overrides for the case under test.
func cliReceiptDoc(t *testing.T, receiptID string, overrides map[string]any) []byte {
	t.Helper()
	doc := map[string]any{
		"receipt_id": receiptID, "request_id": "req-" + receiptID,
		"offer_id": "prov/up", "task": "code", "outcome": "cancelled",
		"validator_version": "",
		"usage": []map[string]any{
			cliUsageUnknown("input_tokens"),
			cliUsageUnknown("output_tokens"),
			cliUsageUnknown("input_images"),
			cliUsageUnknown("output_images"),
			cliUsageUnknown("references"),
			cliUsageUnknown("searches"),
			cliUsageUnknown("elapsed_millis"),
		},
	}
	for k, v := range overrides {
		doc[k] = v
	}
	return cliCanonical(t, doc)
}

// cliBadDSN carries the canary secret and points at a refused port, so any
// code path that selects it fails with temporarily_unavailable instead of
// succeeding.
func cliBadDSN() string {
	return "postgres://node:" + cliReceiptSecret + "@127.0.0.1:1/none?sslmode=disable"
}

// cliProc builds one child process without running it. stdout/stderr
// buffers are returned for inspection after Start/Wait (concurrent cases).
func cliProc(bin string, r cliRun, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	cmd := exec.Command(bin, args...)
	var outBuf, errBuf bytes.Buffer
	if r.stdout != nil {
		// A raw *os.File is inherited by the child as its stdout
		// descriptor; an O_RDONLY file makes the child's ack write fail.
		cmd.Stdout = r.stdout
	} else {
		cmd.Stdout = &outBuf
	}
	cmd.Stderr = &errBuf
	env := make([]string, 0, len(os.Environ())+3)
	for _, kv := range os.Environ() {
		switch {
		case strings.HasPrefix(kv, "HOME="),
			strings.HasPrefix(kv, "AGENT_ROUTING_DSN="),
			strings.HasPrefix(kv, "XDG_CONFIG_HOME="):
			// Replaced below: hermetic config and DSN per run.
		default:
			env = append(env, kv)
		}
	}
	env = append(env, "HOME="+r.home, "AGENT_ROUTING_DSN="+r.dsnEnv)
	cmd.Env = env
	return cmd, &outBuf, &errBuf
}

// cliRunProc runs one child process and returns stdout, stderr, exit code.
func cliRunProc(t *testing.T, bin string, r cliRun, args ...string) (string, string, int) {
	t.Helper()
	cmd, outBuf, errBuf := cliProc(bin, r, args...)
	exit := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("receipt CLI run: %v", err)
		}
		exit = ee.ExitCode()
	}
	return outBuf.String(), errBuf.String(), exit
}

// cliWantOneLine checks the closed stderr form: exactly one line, the
// expected prefix, and no leaked secret.
func cliWantOneLine(t *testing.T, stderr, prefix string) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("stderr not exactly one line: %q", stderr)
	}
	if !strings.HasPrefix(lines[0], prefix) {
		t.Fatalf("stderr prefix %q want %q", lines[0], prefix)
	}
	if strings.Contains(stderr, cliReceiptSecret) {
		t.Fatalf("secret leaked in diagnostic: %q", stderr)
	}
}

// cliCorrelation parses the exit-0 stdout correlation line.
func cliCorrelation(t *testing.T, stdout string) (receiptID, observationID, usageHash string) {
	t.Helper()
	if !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("stdout not exactly one LF-terminated line: %q", stdout)
	}
	var line struct {
		ReceiptID     string `json:"receipt_id"`
		ObservationID string `json:"observation_id"`
		UsageHash     string `json:"usage_hash"`
	}
	if err := json.Unmarshal([]byte(stdout), &line); err != nil {
		t.Fatalf("correlation line is not the canonical object: %v", err)
	}
	return line.ReceiptID, line.ObservationID, line.UsageHash
}

// TestReceiptCLIProcessPath drives the real binary against real Postgres
// through the full receipt record process path.
func TestReceiptCLIProcessPath(t *testing.T) {
	pool := startPostgres(t)
	bin := buildReceiptCLI(t)
	dsn := pool.Config().ConnString()
	ctx := context.Background()

	t.Run("missing host_local source fails safely", func(t *testing.T) {
		id := "rcpt-nf-" + rid(t)
		path := cliWriteFile(t, t.TempDir(), "r.json", cliReceiptDoc(t, id, nil))
		stdout, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 4 {
			t.Fatalf("exit = %d, want 4 (not_found), stderr %q", exit, stderr)
		}
		if stdout != "" {
			t.Fatalf("stdout not empty: %q", stdout)
		}
		cliWantOneLine(t, stderr, "agent-routing: not_found: host_local source is not registered")
	})

	// Register the host_local source with the existing source SQL pattern.
	mustExec(t, pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ('s-host','https://s-host','host_local','p1',3600,true)`)

	t.Run("record persists byte-exact document and correlation", func(t *testing.T) {
		id := "rcpt-main-" + rid(t)
		doc := cliReceiptDoc(t, id, map[string]any{
			"outcome":           "validated_success",
			"validator_version": "mech-1",
			"usage": []map[string]any{
				cliUsageKnown("input_tokens", 100),
				cliUsageKnown("output_tokens", 0), // known zero is not unknown
				cliUsageKnown("input_images", 2),
				cliUsageUnknown("output_images"),
				cliUsageUnknown("references"),
				cliUsageUnknown("searches"),
				cliUsageUnknown("elapsed_millis"),
			},
		})
		path := cliWriteFile(t, t.TempDir(), "r.json", doc)
		stdout, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("exit = %d, want 0, stderr %q", exit, stderr)
		}
		if stderr != "" {
			t.Fatalf("exit 0 wrote stderr: %q", stderr)
		}
		gotID, obsID, usageHash := cliCorrelation(t, stdout)
		if gotID != id {
			t.Fatalf("correlation receipt_id %q want %q", gotID, id)
		}
		// Independent oracle: usage_hash is the canonical hash of the seven
		// members sorted bytewise, known members carrying value.
		wantDoc := []map[string]any{
			{"component": "elapsed_millis"},
			{"component": "input_images", "value": int64(2)},
			{"component": "input_tokens", "value": int64(100)},
			{"component": "output_images"},
			{"component": "output_tokens", "value": int64(0)},
			{"component": "references"},
			{"component": "searches"},
		}
		if usageHash != hex.EncodeToString(canonicalHash(cliCanonical(t, wantDoc))) {
			t.Fatalf("usage_hash %q does not match the canonical seven-member document", usageHash)
		}

		// The stored document is byte-exact, never re-serialized.
		var stored []byte
		if err := pool.QueryRow(ctx,
			`SELECT bytes FROM content_blobs WHERE hash = $1`, canonicalHash(doc)).Scan(&stored); err != nil {
			t.Fatalf("stored body blob: %v", err)
		}
		if !bytes.Equal(stored, doc) {
			t.Fatalf("stored document is not byte-exact: got %q want %q", stored, doc)
		}

		var status int
		var outcome, version, sourceID, resourceURL string
		var bodyHash []byte
		if err := pool.QueryRow(ctx, `
			SELECT http_status, outcome, collector_version, source_id, resource_url, body_hash
			FROM observations WHERE id = $1`, obsID).
			Scan(&status, &outcome, &version, &sourceID, &resourceURL, &bodyHash); err != nil {
			t.Fatalf("stored observation: %v", err)
		}
		if status != 0 || outcome != "success" || version != "collector-0.1" ||
			sourceID != "s-host" || resourceURL != "host-local://receipts/"+id ||
			string(bodyHash) != string(canonicalHash(doc)) {
			t.Fatalf("stored observation mismatch: status=%d outcome=%s version=%s source=%s url=%s",
				status, outcome, version, sourceID, resourceURL)
		}
		var reqID, offerID, task, factOutcome string
		if err := pool.QueryRow(ctx, `
			SELECT request_id, offer_id, task, outcome FROM outcome_facts
			WHERE receipt_id = $1 AND observation_id = $2`, id, obsID).
			Scan(&reqID, &offerID, &task, &factOutcome); err != nil {
			t.Fatalf("stored outcome fact: %v", err)
		}
		if reqID != "req-"+id || offerID != "prov/up" || task != "code" || factOutcome != "validated_success" {
			t.Fatalf("stored outcome fact mismatch: %q %q %q %q", reqID, offerID, task, factOutcome)
		}
		var validator string
		if err := pool.QueryRow(ctx,
			`SELECT validator_version FROM validated_outcomes WHERE receipt_id = $1`, id).
			Scan(&validator); err != nil {
			t.Fatalf("validated outcome not persisted: %v", err)
		}
		if validator != "mech-1" {
			t.Fatalf("validator_version %q want mech-1", validator)
		}
		// Known usage carries exact values, zero included; unknown carries
		// no value row at all.
		rows, err := pool.Query(ctx,
			`SELECT component, value FROM outcome_usage_known WHERE receipt_id = $1`, id)
		if err != nil {
			t.Fatal(err)
		}
		known := map[string]int64{}
		for rows.Next() {
			var c string
			var v int64
			if err := rows.Scan(&c, &v); err != nil {
				t.Fatal(err)
			}
			known[c] = v
		}
		rows.Close()
		if len(known) != 3 || known["input_tokens"] != 100 ||
			known["output_tokens"] != 0 || known["input_images"] != 2 {
			t.Fatalf("known usage rows mismatch: %v", known)
		}
		rows, err = pool.Query(ctx,
			`SELECT component FROM outcome_usage_unknown WHERE receipt_id = $1`, id)
		if err != nil {
			t.Fatal(err)
		}
		unknown := map[string]bool{}
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				t.Fatal(err)
			}
			unknown[c] = true
		}
		rows.Close()
		if len(unknown) != 4 || !unknown["references"] || !unknown["output_images"] ||
			!unknown["searches"] || !unknown["elapsed_millis"] {
			t.Fatalf("unknown usage rows mismatch: %v", unknown)
		}
		if unknown["output_tokens"] {
			t.Fatal("known-zero output_tokens was stored as unknown")
		}
		if _, ok := known["references"]; ok {
			t.Fatal("unknown references was stored as known")
		}
		var events int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM outbox_events WHERE aggregate_id = $1 AND event_name = 'ReceiptRecorded'`, id).
			Scan(&events); err != nil {
			t.Fatal(err)
		}
		if events != 1 {
			t.Fatalf("ReceiptRecorded events = %d, want 1", events)
		}
	})

	t.Run("identical replay returns same correlation", func(t *testing.T) {
		id := "rcpt-replay-" + rid(t)
		doc := cliReceiptDoc(t, id, nil)
		path := cliWriteFile(t, t.TempDir(), "r.json", doc)
		first, _, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("first submission exit = %d, stdout %q", exit, first)
		}
		second, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("replay exit = %d, stderr %q", exit, stderr)
		}
		if first != second {
			t.Fatalf("replay correlation differs: %q vs %q", second, first)
		}
		var observations int
		if err := pool.QueryRow(ctx, `
			SELECT count(*) FROM observations o JOIN outcome_facts f ON o.id = f.observation_id
			WHERE f.receipt_id = $1`, id).Scan(&observations); err != nil {
			t.Fatal(err)
		}
		if observations != 1 {
			t.Fatalf("observations for receipt = %d, want 1", observations)
		}
	})

	t.Run("changed bytes conflict is safe and stores nothing", func(t *testing.T) {
		id := "rcpt-conflict-" + rid(t)
		doc := cliReceiptDoc(t, id, nil)
		path := cliWriteFile(t, t.TempDir(), "r.json", doc)
		first, _, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("first submission exit = %d, stdout %q", exit, first)
		}
		changed := cliReceiptDoc(t, id, map[string]any{
			"outcome": "unvalidated_completion",
			"usage": []map[string]any{
				cliUsageKnown("input_tokens", 7),
				cliUsageUnknown("output_tokens"),
				cliUsageUnknown("input_images"),
				cliUsageUnknown("output_images"),
				cliUsageUnknown("references"),
				cliUsageUnknown("searches"),
				cliUsageUnknown("elapsed_millis"),
			},
		})
		changedPath := cliWriteFile(t, t.TempDir(), "r.json", changed)
		stdout, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", changedPath, "--dsn", dsn)
		if exit != 4 {
			t.Fatalf("exit = %d, want 4 (conflict), stderr %q", exit, stderr)
		}
		if stdout != "" {
			t.Fatalf("stdout not empty: %q", stdout)
		}
		cliWantOneLine(t, stderr, "agent-routing: conflict: receipt ID conflicts with existing bytes")
		// The conflicting bytes were not stored and the original document
		// is still byte-exact.
		var changedBlobs, observations int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM content_blobs WHERE hash = $1`, canonicalHash(changed)).
			Scan(&changedBlobs); err != nil {
			t.Fatal(err)
		}
		if changedBlobs != 0 {
			t.Fatal("conflicting document bytes were stored")
		}
		if err := pool.QueryRow(ctx, `
			SELECT count(*) FROM observations o JOIN outcome_facts f ON o.id = f.observation_id
			WHERE f.receipt_id = $1`, id).Scan(&observations); err != nil {
			t.Fatal(err)
		}
		if observations != 1 {
			t.Fatalf("observations for receipt = %d, want 1", observations)
		}
		var stored []byte
		if err := pool.QueryRow(ctx,
			`SELECT bytes FROM content_blobs WHERE hash = $1`, canonicalHash(doc)).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(stored, doc) {
			t.Fatal("original document changed after conflict")
		}
	})

	t.Run("read-only stdout fails the process safely", func(t *testing.T) {
		id := "rcpt-ro-" + rid(t)
		doc := cliReceiptDoc(t, id, nil)
		path := cliWriteFile(t, t.TempDir(), "r.json", doc)
		// A read-only stdout descriptor: the ack write must fail and the
		// process must report internal_error exit 5, not exit 0 with no
		// acknowledgement (the pass-1 review finding).
		roPath := cliWriteFile(t, t.TempDir(), "stdout", []byte("ignored"))
		ro, err := os.Open(roPath)
		if err != nil {
			t.Fatal(err)
		}
		defer ro.Close()
		_, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir(), stdout: ro},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 5 {
			t.Fatalf("exit = %d, want 5 (internal_error) on stdout write failure", exit)
		}
		cliWantOneLine(t, stderr, "agent-routing: internal_error: ")
		// Persistence may already have committed before the ack write: the
		// receipt exists and a replay returns the same stored correlation.
		// This proves commit-not-rollback, not a discarded submission.
		var storedObsID string
		if err := pool.QueryRow(ctx,
			`SELECT observation_id FROM outcome_facts WHERE receipt_id = $1`, id).
			Scan(&storedObsID); err != nil {
			t.Fatalf("receipt not persisted despite reported failure: %v", err)
		}
		stdout, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("replay exit = %d, stderr %q", exit, stderr)
		}
		_, replayObsID, _ := cliCorrelation(t, stdout)
		if replayObsID != storedObsID {
			t.Fatalf("replay observation %q want the stored %q", replayObsID, storedObsID)
		}
	})

	t.Run("malformed documents exit 2 before config and DSN", func(t *testing.T) {
		// Hostile canaries: malformed config.toml, secret-bearing env DSN
		// and secret-bearing flag DSN. If the document were parsed after
		// config or DSN handling, these would surface as exit 5 or 4
		// instead of exit 2.
		badHome := t.TempDir()
		cliWriteConfig(t, badHome,
			"dsn = \""+cliBadDSN()+"\"\n[unclosed\n")
		dir := t.TempDir()
		docs := map[string][]byte{
			"unknown member": cliReceiptDoc(t, "rcpt-bad", map[string]any{"task_prompt": cliReceiptSecret}),
			"duplicate usage component": cliReceiptDoc(t, "rcpt-bad", map[string]any{
				"usage": []map[string]any{
					cliUsageUnknown("input_tokens"),
					cliUsageUnknown("input_tokens"),
					cliUsageUnknown("input_images"),
					cliUsageUnknown("output_images"),
					cliUsageUnknown("references"),
					cliUsageUnknown("searches"),
					cliUsageUnknown("elapsed_millis"),
				},
			}),
			"validated outcome without validator_version": cliReceiptDoc(t, "rcpt-bad",
				map[string]any{"outcome": "validated_success"}),
			"non-canonical bytes": []byte(`{"receipt_id":"rcpt-bad","request_id":"r","offer_id":"o",` +
				`"task":"code","outcome":"cancelled","validator_version":"","usage":[]}`),
		}
		for name, doc := range docs {
			t.Run(name, func(t *testing.T) {
				path := cliWriteFile(t, dir, "r.json", doc)
				stdout, stderr, exit := cliRunProc(t, bin,
					cliRun{home: badHome, dsnEnv: cliBadDSN()},
					"receipt", "record", "--receipt", path, "--dsn", cliBadDSN())
				if exit != 2 {
					t.Fatalf("exit = %d, want 2 (invalid_request), stderr %q", exit, stderr)
				}
				if stdout != "" {
					t.Fatalf("stdout not empty: %q", stdout)
				}
				cliWantOneLine(t, stderr, "agent-routing: invalid_request: receipt document is invalid")
			})
		}
	})

	t.Run("DSN precedence and configured collector version", func(t *testing.T) {
		// Control: the bad DSN genuinely fails when selected.
		id := "rcpt-ctl-" + rid(t)
		path := cliWriteFile(t, t.TempDir(), "r.json", cliReceiptDoc(t, id, nil))
		_, stderr, exit := cliRunProc(t, bin, cliRun{home: t.TempDir(), dsnEnv: cliBadDSN()},
			"receipt", "record", "--receipt", path)
		if exit != 4 {
			t.Fatalf("control exit = %d, want 4 (temporarily_unavailable), stderr %q", exit, stderr)
		}
		cliWantOneLine(t, stderr, "agent-routing: temporarily_unavailable: ")

		// Config DSN wins when flag and env are absent; collector_version
		// comes from the same config.
		cfgHome := t.TempDir()
		cliWriteConfig(t, cfgHome, "dsn = \""+dsn+"\"\ncollector_version = \"collector-cli-cfg\"\n")
		id = "rcpt-cfg-" + rid(t)
		path = cliWriteFile(t, t.TempDir(), "r.json", cliReceiptDoc(t, id, nil))
		stdout, stderr, exit := cliRunProc(t, bin, cliRun{home: cfgHome},
			"receipt", "record", "--receipt", path)
		if exit != 0 {
			t.Fatalf("config DSN exit = %d, stderr %q", exit, stderr)
		}
		gotID, obsID, _ := cliCorrelation(t, stdout)
		if gotID != id {
			t.Fatalf("correlation receipt_id %q want %q", gotID, id)
		}
		var version string
		if err := pool.QueryRow(ctx,
			`SELECT collector_version FROM observations WHERE id = $1`, obsID).Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != "collector-cli-cfg" {
			t.Fatalf("collector_version %q want collector-cli-cfg", version)
		}

		// Env DSN beats a bad config DSN.
		envHome := t.TempDir()
		cliWriteConfig(t, envHome, "dsn = \""+cliBadDSN()+"\"\n")
		id = "rcpt-env-" + rid(t)
		path = cliWriteFile(t, t.TempDir(), "r.json", cliReceiptDoc(t, id, nil))
		stdout, stderr, exit = cliRunProc(t, bin, cliRun{home: envHome, dsnEnv: dsn},
			"receipt", "record", "--receipt", path)
		if exit != 0 {
			t.Fatalf("env DSN exit = %d, stderr %q", exit, stderr)
		}
		cliCorrelation(t, stdout)

		// Flag DSN beats bad env and bad config.
		flagHome := t.TempDir()
		cliWriteConfig(t, flagHome, "dsn = \""+cliBadDSN()+"\"\n")
		id = "rcpt-flag-" + rid(t)
		path = cliWriteFile(t, t.TempDir(), "r.json", cliReceiptDoc(t, id, nil))
		stdout, stderr, exit = cliRunProc(t, bin, cliRun{home: flagHome, dsnEnv: cliBadDSN()},
			"receipt", "record", "--receipt", path, "--dsn", dsn)
		if exit != 0 {
			t.Fatalf("flag DSN exit = %d, stderr %q", exit, stderr)
		}
		cliCorrelation(t, stdout)
	})

	t.Run("concurrent identical submissions", func(t *testing.T) {
		// Two identical real processes racing on the same receipt bytes:
		// both must succeed idempotently with the same correlation, per
		// cli-adapter.md's replay contract. Any genuine failure here is a
		// real finding to report, not an assertion to weaken.
		for i := range 5 {
			id := fmt.Sprintf("rcpt-conc-%d-%s", i, rid(t))
			doc := cliReceiptDoc(t, id, nil)
			path := cliWriteFile(t, t.TempDir(), "r.json", doc)
			cmd1, out1, err1 := cliProc(bin, cliRun{home: t.TempDir()},
				"receipt", "record", "--receipt", path, "--dsn", dsn)
			cmd2, out2, err2 := cliProc(bin, cliRun{home: t.TempDir()},
				"receipt", "record", "--receipt", path, "--dsn", dsn)
			for _, cmd := range []*exec.Cmd{cmd1, cmd2} {
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
			}
			exit1, exit2 := 0, 0
			var werr error
			if err := cmd1.Wait(); err != nil {
				var ee *exec.ExitError
				if !errors.As(err, &ee) {
					werr = err
				} else {
					exit1 = ee.ExitCode()
				}
			}
			if err := cmd2.Wait(); err != nil {
				var ee *exec.ExitError
				if !errors.As(err, &ee) {
					werr = err
				} else {
					exit2 = ee.ExitCode()
				}
			}
			if werr != nil {
				t.Fatal(werr)
			}
			if exit1 != 0 || exit2 != 0 {
				t.Errorf("pair %d: concurrent identical submissions exited %d and %d (both want 0); stderr %q %q",
					i, exit1, exit2, err1.String(), err2.String())
				continue
			}
			if out1.String() != out2.String() {
				t.Errorf("pair %d: concurrent correlations differ: %q vs %q", i, out1.String(), out2.String())
			}
			var observations int
			if err := pool.QueryRow(ctx, `
				SELECT count(*) FROM observations o JOIN outcome_facts f ON o.id = f.observation_id
				WHERE f.receipt_id = $1`, id).Scan(&observations); err != nil {
				t.Fatal(err)
			}
			if observations != 1 {
				t.Errorf("pair %d: observations for receipt = %d, want 1", i, observations)
			}
		}
	})
}
