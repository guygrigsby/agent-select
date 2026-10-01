package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// daemonDSN reconstructs the DSN from a live pool.
func daemonDSN(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	return pool.Config().ConnString()
}

// TestRunCollectorLifecycle runs the daemon against a real Postgres: it
// migrates, reconciles, collects a seeded job against a live HTTP source
// and shuts down cleanly on context cancel.
func TestRunCollectorLifecycle(t *testing.T) {
	pool := startPostgres(t)
	dsn := daemonDSN(t, pool)

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n := atomic.AddInt32(&hits, 1); n > 1 {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Write([]byte(`{"daemon":true}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p := New(pool)
	now := time.Now().UTC()
	mustExec(t, pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ($1,$2,'provider_api','p1',3600,true)`, "dsrc", srv.URL)
	job, err := p.EnqueueJob(ctx, "dsrc", srv.URL, "fetch",
		hash32([]byte("d-input")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- RunCollectorWith(ctx, dsn, CollectorConfig{WorkerID: "daemon-test"},
			NewLocalTestFetcher(""))
	}()

	// Wait for collection evidence: a completion for the seeded job.
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM job_completions WHERE job_id = $1`, job.JobID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n == 1 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("ctx expired before collection completed")
		case err := <-done:
			t.Fatalf("daemon exited early: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
	}
	var obsN int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM observations o
		JOIN stage_output_links l ON l.observation_id = o.id
		WHERE l.job_id = $1 AND o.outcome = 'success'`, job.JobID).Scan(&obsN); err != nil {
		t.Fatal(err)
	}
	if obsN != 1 {
		t.Fatalf("linked success observations = %d", obsN)
	}

	// Graceful shutdown: cancel and expect a nil return.
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown returned %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("daemon did not exit on cancel")
	}
	if atomic.LoadInt32(&hits) < 1 {
		t.Fatal("source was never fetched")
	}
}

// TestLoadCollectorConfigMissingKeepsDefaults: absent config file is fine.
func TestLoadCollectorConfigMissingKeepsDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := LoadCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DSN != "" || cfg.WorkerID != "" {
		t.Fatalf("defaults clobbered: %+v", cfg)
	}
}

// TestResolveDSNPrecedence: flag over env over config.
func TestResolveDSNPrecedence(t *testing.T) {
	cfg := CollectorConfig{DSN: "cfg-dsn"}
	t.Setenv("AGENT_ROUTING_DSN", "env-dsn")
	if got, _ := ResolveDSN("flag-dsn", cfg); got != "flag-dsn" {
		t.Fatalf("flag: %q", got)
	}
	if got, _ := ResolveDSN("", cfg); got != "env-dsn" {
		t.Fatalf("env: %q", got)
	}
	t.Setenv("AGENT_ROUTING_DSN", "")
	if got, _ := ResolveDSN("", cfg); got != "cfg-dsn" {
		t.Fatalf("config: %q", got)
	}
	if _, err := ResolveDSN("", CollectorConfig{}); err == nil {
		t.Fatal("expected error when nothing resolves a DSN")
	}
}
