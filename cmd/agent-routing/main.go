// Command agent-routing is the local research artifact tool (§1 of
// docs/specs/research-artifact.md). No network, no provider calls, no
// production mutations.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/guygrigsby/agent-routing/internal/research"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("agent-routing", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	jsonOut := fs.Bool("json", false, "emit a structured result object on stdout")
	if err := fs.Parse(args); err != nil {
		return research.ExitUsage
	}
	rest := fs.Args()
	if len(rest) != 2 || rest[0] != "research" {
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
	fmt.Fprintln(os.Stderr, "usage: agent-routing [--root dir] [--json] research {check|rebuild|coverage}")
}
