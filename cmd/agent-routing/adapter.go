// CLI adapter commands (docs/specs/cli-adapter.md): route, snapshot check,
// snapshot adopt. Thin over the public routing API: no ranking, policy,
// storage or collection logic; one process, one decision, exit.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/guygrigsby/agent-routing/internal/collector"
	"github.com/guygrigsby/agent-routing/routing"
)

// CLI adapter exit codes (closed table, cli-adapter.md "Errors and exit
// codes").
const (
	adapterExitOK       = 0 // route completed with a selection
	adapterExitUsage    = 2 // usage error, invalid_request, invalid_request_wire
	adapterExitNoOffer  = 3 // route completed with no admissible offer
	adapterExitSnapshot = 4 // invalid_snapshot, schema_incompatible, snapshot_unavailable
	adapterExitInternal = 5
)

// adapterKeepGenerations matches the daemon's store retention default.
const adapterKeepGenerations = 2

// runAdapter dispatches the adapter subcommands. Everything else stays with
// the research tool dispatch in run.
func runAdapter(args []string) int {
	if len(args) == 0 {
		adapterUsage()
		return adapterExitUsage
	}
	switch args[0] {
	case "route":
		return runRoute(args[1:])
	case "snapshot":
		return runSnapshot(args[1:])
	default:
		adapterUsage()
		return adapterExitUsage
	}
}

// adapterUsage prints exactly one usage line: every nonzero exit writes
// exactly one stderr line (cli-adapter.md "Errors and exit codes").
func adapterUsage() {
	fmt.Fprint(os.Stderr, "usage: agent-routing route --snapshot PATH --request PATH | snapshot check --artifact PATH | snapshot adopt --artifact PATH --store PATH | research {check|rebuild|coverage|backlog}\n")
}

func newAdapterFlagSet(name, usageLine string) *flag.FlagSet {
	fs := flag.NewFlagSet("agent-routing "+name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, "usage: "+usageLine) }
	return fs
}

// adapterFail writes the one-line stderr form: agent-routing: <code>:
// <safe detail>. Never writes stdout.
func adapterFail(code, detail string, exit int) int {
	fmt.Fprintf(os.Stderr, "agent-routing: %s: %s\n", code, detail)
	return exit
}

// routingCodeExit maps a routing error to the adapter exit table by domain
// code; anything unexpected is exit 5.
func routingCodeExit(code string) int {
	switch code {
	case routing.CodeInvalidSnapshot, routing.CodeSchemaIncompatible, routing.CodeSnapshotUnavailable:
		return adapterExitSnapshot
	case routing.CodeInvalidRequest:
		return adapterExitUsage
	default:
		return adapterExitInternal
	}
}

// readArtifact reads and decodes a local artifact file. Read failures are
// snapshot_unavailable; decode failures keep their domain code.
func readArtifact(path string) ([]byte, *routing.Snapshot, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, adapterFail(routing.CodeSnapshotUnavailable, "artifact read failed", adapterExitSnapshot)
	}
	snap, derr := routing.DecodeSnapshot(data)
	if derr != nil {
		var re *routing.Error
		if errors.As(derr, &re) {
			return nil, nil, adapterFail(re.Code, re.Detail, routingCodeExit(re.Code))
		}
		// Fixed label: the raw error never crosses the boundary.
		return nil, nil, adapterFail(routing.CodeInvalidSnapshot, "decode failed", adapterExitSnapshot)
	}
	return data, snap, -1
}

func runRoute(args []string) int {
	const usageLine = "agent-routing route --snapshot PATH --request PATH (- reads the request from stdin)"
	fs := newAdapterFlagSet("route", usageLine)
	snapshotPath := fs.String("snapshot", "", "local artifact path")
	requestPath := fs.String("request", "", "request document path, or - for stdin")
	if err := fs.Parse(args); err != nil {
		return adapterExitUsage
	}
	if fs.NArg() > 0 || *snapshotPath == "" || *requestPath == "" {
		fs.Usage()
		return adapterExitUsage
	}

	data, snap, exit := readArtifact(*snapshotPath)
	if exit >= 0 {
		return exit
	}
	engine := routing.NewEngine(snap)
	if err := engine.Adopt(data); err != nil {
		var re *routing.Error
		if errors.As(err, &re) {
			return adapterFail(re.Code, re.Detail, routingCodeExit(re.Code))
		}
		return adapterFail("internal_error", "engine adoption failed", adapterExitInternal)
	}

	var wire []byte
	var rerr error
	if *requestPath == "-" {
		wire, rerr = io.ReadAll(os.Stdin)
	} else {
		wire, rerr = os.ReadFile(*requestPath)
	}
	if rerr != nil {
		return adapterFail("invalid_request_wire", "request document read failed", adapterExitUsage)
	}
	rd, werr := routing.DecodeRequestWire(wire)
	if werr != nil {
		var re *routing.Error
		if errors.As(werr, &re) {
			return adapterFail("invalid_request_wire", re.Detail, adapterExitUsage)
		}
		// Fixed label: raw decoder errors carry untrusted bytes.
		return adapterFail("invalid_request_wire", "request document is invalid", adapterExitUsage)
	}
	req, nerr := routing.NewRequest(rd)
	if nerr != nil {
		var re *routing.Error
		if errors.As(nerr, &re) {
			return adapterFail(re.Code, re.Detail, routingCodeExit(re.Code))
		}
		return adapterFail(routing.CodeInvalidRequest, "request failed domain validation", adapterExitUsage)
	}

	decision, err := engine.Route(req)
	enc, cerr := routing.CanonicalEncode(decision)
	if cerr != nil {
		return adapterFail("internal_error", "decision encoding failed", adapterExitInternal)
	}
	os.Stdout.Write(append(enc, '\n'))
	if err != nil {
		var re *routing.Error
		if errors.As(err, &re) && re.Code == routing.CodeNoAdmissibleOffer {
			return adapterExitNoOffer
		}
		if errors.As(err, &re) {
			return adapterFail(re.Code, re.Detail, routingCodeExit(re.Code))
		}
		return adapterFail("internal_error", "routing failed", adapterExitInternal)
	}
	return adapterExitOK
}

func runSnapshot(args []string) int {
	if len(args) == 0 {
		adapterUsage()
		return adapterExitUsage
	}
	switch args[0] {
	case "check":
		return runSnapshotCheck(args[1:])
	case "adopt":
		return runSnapshotAdopt(args[1:])
	default:
		adapterUsage()
		return adapterExitUsage
	}
}

func runSnapshotCheck(args []string) int {
	const usageLine = "agent-routing snapshot check --artifact PATH"
	fs := newAdapterFlagSet("snapshot check", usageLine)
	artifactPath := fs.String("artifact", "", "local artifact path")
	if err := fs.Parse(args); err != nil {
		return adapterExitUsage
	}
	if fs.NArg() > 0 || *artifactPath == "" {
		fs.Usage()
		return adapterExitUsage
	}
	_, snap, exit := readArtifact(*artifactPath)
	if exit >= 0 {
		return exit
	}
	fmt.Println(snap.Hash())
	return adapterExitOK
}

func runSnapshotAdopt(args []string) int {
	const usageLine = "agent-routing snapshot adopt --artifact PATH --store PATH"
	fs := newAdapterFlagSet("snapshot adopt", usageLine)
	artifactPath := fs.String("artifact", "", "local artifact path")
	storePath := fs.String("store", "", "local store directory")
	if err := fs.Parse(args); err != nil {
		return adapterExitUsage
	}
	if fs.NArg() > 0 || *artifactPath == "" || *storePath == "" {
		fs.Usage()
		return adapterExitUsage
	}
	data, snap, exit := readArtifact(*artifactPath)
	if exit >= 0 {
		return exit
	}
	hash := snap.Hash()
	// Fixed labels below: raw storage errors carry caller-supplied paths
	// (embedded newlines forge log lines), so only fixed diagnostics cross
	// the stderr boundary. Exit codes stay per the cli-adapter.md table.
	store, serr := collector.NewArtifactStore(*storePath, adapterKeepGenerations)
	if serr != nil {
		return adapterFail(routing.CodeSnapshotUnavailable, "store directory unusable", adapterExitSnapshot)
	}
	if err := store.Stage(hash, data); err != nil {
		return adapterFail(routing.CodeSnapshotUnavailable, "staging failed", adapterExitSnapshot)
	}
	if err := store.Adopt(hash); err != nil {
		return adapterFail(routing.CodeSnapshotUnavailable, "adoption failed, prior generation retained", adapterExitSnapshot)
	}
	if err := store.Prune(); err != nil {
		return adapterFail(routing.CodeSnapshotUnavailable, "adoption succeeded but pruning failed", adapterExitSnapshot)
	}
	fmt.Println(hash)
	return adapterExitOK
}
