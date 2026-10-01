// Package-level daemon wiring for the collector: TOML config at
// <UserConfigDir>/agent-routing/config.toml, env and flag overrides, and the
// long-running loop (migrate, reconcile, collect) with signal-driven
// graceful shutdown. Modeled on the rookery/perch daemon pattern.
package collector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CollectorConfig is the daemon's config.toml shape. All keys optional.
type CollectorConfig struct {
	DSN              string `toml:"dsn"`
	WorkerID         string `toml:"worker_id"`
	CollectorVersion string `toml:"collector_version"`
	LeaseSeconds     int    `toml:"lease_seconds"`
	MinHostMillis    int    `toml:"min_host_interval_millis"`
	MaxInFlight      int    `toml:"max_in_flight"`
	ScanLimit        int    `toml:"scan_limit"`
	PollIdleSeconds  int    `toml:"poll_idle_seconds"`
	UserAgent        string `toml:"user_agent"`
	ReconcileSeconds int    `toml:"reconcile_seconds"`

	// Successor pipeline (delivery-storage V8). The publisher loop runs
	// only when a validator version is configured.
	SnapshotChannel          string `toml:"snapshot_channel"`
	SnapshotValidatorVersion string `toml:"snapshot_validator_version"`
	SnapshotPolicy           string `toml:"snapshot_projection_policy_version"`
	SnapshotDistribution     string `toml:"snapshot_distribution"`
	StoreDir                 string `toml:"store_dir"`
	StoreKeepGenerations     int    `toml:"store_keep_generations"`
}

// ResolveDSN picks the DSN: flag wins, then AGENT_ROUTING_DSN, then config.
func ResolveDSN(flagVal string, cfg CollectorConfig) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if v := os.Getenv("AGENT_ROUTING_DSN"); v != "" {
		return v, nil
	}
	if cfg.DSN != "" {
		return cfg.DSN, nil
	}
	return "", errors.New("no collector DSN: pass --dsn, set AGENT_ROUTING_DSN or set dsn in config.toml")
}

// LoadCollectorConfig decodes <UserConfigDir>/agent-routing/config.toml
// into cfg. A missing file keeps the caller's defaults; malformed TOML is
// an error.
func LoadCollectorConfig() (CollectorConfig, error) {
	var cfg CollectorConfig
	dir, err := os.UserConfigDir()
	if err != nil {
		return cfg, nil // no config dir: defaults only
	}
	_, err = toml.DecodeFile(filepath.Join(dir, "agent-routing", "config.toml"), &cfg)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	return cfg, err
}

// WorkerConfigFrom maps the TOML shape onto the worker configuration,
// applying defaults for unset fields.
func (c CollectorConfig) WorkerConfigFrom() WorkerConfig {
	wc := WorkerConfig{
		WorkerID:         c.WorkerID,
		CollectorVersion: c.CollectorVersion,
	}
	if c.LeaseSeconds > 0 {
		wc.LeaseDuration = time.Duration(c.LeaseSeconds) * time.Second
	}
	if c.MinHostMillis > 0 {
		wc.MinHostInterval = time.Duration(c.MinHostMillis) * time.Millisecond
	}
	if c.MaxInFlight > 0 {
		wc.MaxInFlight = c.MaxInFlight
	}
	if c.ScanLimit > 0 {
		wc.ScanLimit = c.ScanLimit
	}
	if c.PollIdleSeconds > 0 {
		wc.PollIdle = time.Duration(c.PollIdleSeconds) * time.Second
	}
	return wc
}

// SuccessorConfigFrom maps the TOML shape onto the publisher loop
// configuration, applying defaults. ok reports whether the loop is enabled
// (a validator version is required).
func (c CollectorConfig) SuccessorConfigFrom() (SuccessorConfig, string, int, bool, error) {
	if c.SnapshotValidatorVersion == "" {
		return SuccessorConfig{}, "", 0, false, nil
	}
	channel := c.SnapshotChannel
	if channel == "" {
		channel = "local"
	}
	dist := c.SnapshotDistribution
	if dist == "" {
		dist = "local"
	}
	keep := c.StoreKeepGenerations
	if keep == 0 {
		keep = 2
	}
	root := c.StoreDir
	if root == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return SuccessorConfig{}, "", 0, false, err
		}
		root = filepath.Join(base, "agent-routing", "store")
	}
	return SuccessorConfig{
		Channel:               channel,
		ValidatorVersion:      c.SnapshotValidatorVersion,
		BootstrapPolicy:       c.SnapshotPolicy,
		BootstrapDistribution: dist,
	}, root, keep, true, nil
}

// SignalContext returns a context cancelled on SIGINT/SIGTERM.
func SignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// RunCollector is the daemon body: connect, migrate, reconcile expired
// leases, then collect until ctx is done. Periodic reconciliation adopts
// leases of crashed workers between startup sweeps. It fetches over HTTPS
// with private-network access blocked; tests inject their own fetcher via
// RunCollectorWith.
func RunCollector(ctx context.Context, dsn string, cfg CollectorConfig) error {
	return RunCollectorWith(ctx, dsn, cfg, NewHTTPFetcher(cfg.UserAgent))
}

// RunCollectorWith is RunCollector over an injected fetcher.
func RunCollectorWith(ctx context.Context, dsn string, cfg CollectorConfig, f Fetcher) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("collector: connect: %w", err)
	}
	defer pool.Close()
	p := New(pool)
	if err := Migrate(ctx, pool); err != nil {
		return fmt.Errorf("collector: migrate: %w", err)
	}
	if adopted, err := ReconcileExpiredLeases(ctx, p, time.Now().UTC(), 1000); err != nil {
		return fmt.Errorf("collector: startup reconciliation: %w", err)
	} else {
		fmt.Fprintf(os.Stderr, "collector: startup reconciliation adopted %d expired leases\n", adopted)
	}
	wc := cfg.WorkerConfigFrom()
	if wc.CollectorVersion == "" {
		wc.CollectorVersion = "collector-0.1"
	}
	w := NewWorker(p, f, wc)

	errc := make(chan error, 3)
	go func() { errc <- w.Run(ctx) }()
	if sc, root, keep, ok, err := cfg.SuccessorConfigFrom(); ok {
		if err != nil {
			return fmt.Errorf("collector: successor configuration: %w", err)
		}
		store, err := NewArtifactStore(root, keep)
		if err != nil {
			return fmt.Errorf("collector: artifact store: %w", err)
		}
		go func() { errc <- RunSuccessorLoop(ctx, p, store, sc) }()
	}
	reconcileEvery := time.Minute
	if cfg.ReconcileSeconds > 0 {
		reconcileEvery = time.Duration(cfg.ReconcileSeconds) * time.Second
	}
	go func() {
		t := time.NewTicker(reconcileEvery)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				errc <- nil
				return
			case <-t.C:
				if _, err := ReconcileExpiredLeases(ctx, p, time.Now().UTC(), 1000); err != nil && ctx.Err() == nil {
					fmt.Fprintf(os.Stderr, "collector: reconciliation: %v\n", err)
				}
			}
		}
	}()
	return <-errc
}
