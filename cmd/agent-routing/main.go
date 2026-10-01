// Command agent-routing is the local research artifact tool (§1 of
// docs/specs/research-artifact.md). No network, no provider calls, no
// production mutations.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/guygrigsby/agent-routing/internal/collector"
	"github.com/guygrigsby/agent-routing/internal/research"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// Adapter commands (cli-adapter.md) dispatch before research flags:
	// they take no --root/--json and own their exit table.
	if len(args) == 0 || (args[0] != "research" && args[0] != "collect") {
		// Unknown subcommand is a usage error on the adapter exit table.
		return runAdapter(args)
	}
	fs := flag.NewFlagSet("agent-routing", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	jsonOut := fs.Bool("json", false, "emit a structured result object on stdout")
	if err := fs.Parse(args); err != nil {
		return research.ExitUsage
	}
	rest := fs.Args()
	if len(rest) >= 1 && rest[0] == "collect" {
		return runCollect(rest[1:], fs)
	}
	if len(rest) < 2 || rest[0] != "research" {
		usage()
		return research.ExitUsage
	}
	if rest[1] == "backlog" {
		return runBacklog(*root, fs)
	}
	if len(rest) != 2 {
		usage()
		return research.ExitUsage
	}

	var out any
	var err error
	var extra string
	switch rest[1] {
	case "check":
		out, err = research.Check(*root)
	case "rebuild":
		out, err = research.Rebuild(*root)
	case "coverage":
		var tsv []byte
		tsv, extra, err = research.Coverage(*root)
		out = tsv
	default:
		usage()
		return research.ExitUsage
	}
	if err != nil {
		code := research.ExitIO
		if e, ok := err.(*research.Error); ok {
			code = e.Code
		}
		fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
		return code
	}
	if extra != "" {
		fmt.Fprint(os.Stderr, extra)
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(struct {
			Command string `json:"command"`
			Result  any    `json:"result"`
			Driver  any    `json:"driver"`
		}{rest[1], jsonable(out), research.DriverSummary()}); err != nil {
			fmt.Fprintf(os.Stderr, "agent-routing: encode result: %v\n", err)
			return research.ExitIO
		}
		return research.ExitOK
	}
	switch v := out.(type) {
	case []byte:
		os.Stdout.Write(v)
	default:
		fmt.Println("ok")
	}
	return research.ExitOK
}

// jsonable converts raw TSV bytes to a JSON-safe wrapper.
func jsonable(v any) any {
	if b, ok := v.([]byte); ok {
		return map[string]any{"tsv": string(b)}
	}
	return v
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: agent-routing [--root dir] [--json] research {check|rebuild|coverage|backlog}")
	fmt.Fprintln(os.Stderr, "       agent-routing [--root dir] research backlog [--report PATH|--check PATH]")
	fmt.Fprintln(os.Stderr, "       agent-routing collect [--dsn DSN]")
}

// runCollect runs the collector daemon until SIGINT/SIGTERM: migrate,
// reconcile expired leases, then collect claimable work. The DSN resolves
// from --dsn, AGENT_ROUTING_DSN or config.toml (see internal/collector).
func runCollect(args []string, parent *flag.FlagSet) int {
	cfs := flag.NewFlagSet("collect", flag.ContinueOnError)
	cfs.SetOutput(os.Stderr)
	dsnFlag := cfs.String("dsn", "", "Postgres DSN (overrides env/config)")
	if err := cfs.Parse(args); err != nil {
		return research.ExitUsage
	}
	cfg, err := collector.LoadCollectorConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent-routing: config: %v\n", err)
		return research.ExitIO
	}
	dsn, err := collector.ResolveDSN(*dsnFlag, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
		return research.ExitUsage
	}
	ctx, cancel := collector.SignalContext()
	defer cancel()
	if err := collector.RunCollector(ctx, dsn, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
		return research.ExitIO
	}
	return research.ExitOK
}

// runBacklog handles the backlog subcommand's extra flags. Both modes run
// full validation and the backlog gates; --report writes the markdown
// report, --check compares it (drift gate), the default prints the TSV.
func runBacklog(root string, parent *flag.FlagSet) int {
	bfs := flag.NewFlagSet("backlog", flag.ContinueOnError)
	bfs.SetOutput(os.Stderr)
	report := bfs.String("report", "", "write the markdown report to this path (relative to --root)")
	check := bfs.String("check", "", "compare the markdown report at this path against a regeneration")
	if err := bfs.Parse(parent.Args()[2:]); err != nil {
		return research.ExitUsage
	}
	if *report != "" && *check != "" {
		fmt.Fprintln(os.Stderr, "agent-routing: --report and --check are mutually exclusive")
		return research.ExitUsage
	}
	tsv, rep, summary, err := research.Backlog(root)
	if err != nil {
		code := research.ExitIO
		if e, ok := err.(*research.Error); ok {
			code = e.Code
		}
		fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
		return code
	}
	switch {
	case *report != "":
		p := filepath.Join(root, *report)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
			return research.ExitIO
		}
		if err := os.WriteFile(p, rep, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
			return research.ExitIO
		}
		fmt.Fprint(os.Stderr, summary)
	case *check != "":
		if err := research.BacklogCheck(root, *check); err != nil {
			code := research.ExitIO
			if e, ok := err.(*research.Error); ok {
				code = e.Code
			}
			fmt.Fprintf(os.Stderr, "agent-routing: %v\n", err)
			return code
		}
		fmt.Fprint(os.Stderr, summary)
	default:
		fmt.Fprint(os.Stderr, summary)
		os.Stdout.Write(tsv)
	}
	return research.ExitOK
}
