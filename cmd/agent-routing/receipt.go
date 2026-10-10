// receipt record: the host reporter's outcome-capture process boundary
// (cli-adapter.md V2). Reads one canonical receipt document, calls the
// record_receipt port, prints the observation correlation line.
package main

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guygrigsby/agent-routing/internal/collector"
	"github.com/guygrigsby/agent-routing/routing"
)

func runReceipt(args []string) int {
	if len(args) < 2 || args[1] != "record" {
		adapterUsage()
		return adapterExitUsage
	}
	const usageLine = "agent-routing receipt record --receipt PATH [--dsn DSN]"
	fs := newAdapterFlagSet("receipt record", usageLine)
	receiptPath := fs.String("receipt", "", "canonical-v1 receipt document path")
	dsnFlag := fs.String("dsn", "", "Postgres DSN (overrides env/config)")
	if err := fs.Parse(args[2:]); err != nil {
		return adapterExitUsage
	}
	if fs.NArg() > 0 || *receiptPath == "" {
		fs.Usage()
		return adapterExitUsage
	}

	data, err := os.ReadFile(*receiptPath)
	if err != nil {
		return adapterFail("invalid_request", "receipt document read failed", adapterExitUsage)
	}
	params, perr := collector.ParseReceipt(data)
	if perr != nil {
		return adapterFail("invalid_request", "receipt document is invalid", adapterExitUsage)
	}
	cfg, cerr := collector.LoadCollectorConfig()
	if cerr != nil {
		// Raw config errors (parse text, paths, maybe a config DSN line)
		// never cross the boundary.
		return adapterFail("internal_error", "collector configuration could not be read", adapterExitInternal)
	}
	dsn, derr := collector.ResolveDSN(*dsnFlag, cfg)
	if derr != nil {
		return adapterFail("invalid_request", derr.Error(), adapterExitUsage)
	}
	collectorVersion := cfg.CollectorVersion
	if collectorVersion == "" {
		collectorVersion = "collector-0.1"
	}
	params.RecordedAt = time.Now().UTC()
	params.CollectorVersion = collectorVersion

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Split parse from connect so each failure gets a fixed safe detail:
	// raw pgx errors echo connection-string material and never cross
	// the boundary.
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return adapterFail("snapshot_unavailable", "postgres DSN is invalid", adapterExitSnapshot)
	}
	pool, perr2 := pgxpool.NewWithConfig(ctx, poolConfig)
	if perr2 != nil {
		return adapterFail("snapshot_unavailable", "postgres connection failed", adapterExitSnapshot)
	}
	defer pool.Close()
	res, rerr := collector.New(pool).RecordReceipt(ctx, params)
	if rerr != nil {
		return portFailure(rerr)
	}
	out, aerr := routing.CanonicalEncode(struct {
		ReceiptID     string `json:"receipt_id"`
		ObservationID string `json:"observation_id"`
		UsageHash     string `json:"usage_hash"`
	}{res.ReceiptID, res.ObservationID, hex.EncodeToString(res.UsageHash)})
	if aerr != nil {
		// Fixed label: a raw encoding error is internal detail and never
		// crosses the boundary.
		return adapterFail("internal_error", "acknowledgement encoding failed", adapterExitInternal)
	}
	// The write error is checked: a lost acknowledgement must not report
	// success (pass-1 review finding). Persistence may already have
	// committed, so the caller replays the same document and gets the same
	// stored correlation; the port is idempotent by receipt ID plus body
	// hash. Exit 5 tells the host the outcome is unknown, never confirmed.
	if _, werr := os.Stdout.Write(append(out, '\n')); werr != nil {
		return adapterFail("internal_error", "acknowledgement write failed", adapterExitInternal)
	}
	return adapterExitOK
}

// portFailure preserves closed port codes without echoing any error detail.
func portFailure(err error) int {
	var ce *collector.Error
	if !errors.As(err, &ce) {
		return adapterFail("internal_error", "record_receipt failed", adapterExitInternal)
	}
	switch ce.Code {
	case collector.CodeTemporarilyUnavailable:
		return adapterFail(string(ce.Code), "collector storage is temporarily unavailable", adapterExitSnapshot)
	case collector.CodeNotFound:
		return adapterFail(string(ce.Code), "host_local source is not registered", adapterExitSnapshot)
	case collector.CodeConflict:
		return adapterFail(string(ce.Code), "receipt ID conflicts with existing bytes", adapterExitSnapshot)
	case collector.CodeInvalidRequest, collector.CodeInvalidEvidence,
		collector.CodeUnauthorized, collector.CodeForbidden, collector.CodeStaleLease,
		collector.CodeSourceBlocked, collector.CodeIdentityUnresolved:
		return adapterFail(string(ce.Code), "receipt rejected at the collector port", adapterExitSnapshot)
	default:
		return adapterFail("internal_error", "record_receipt failed", adapterExitInternal)
	}
}
