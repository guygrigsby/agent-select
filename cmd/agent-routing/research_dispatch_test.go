// Top-level dispatch regression tests (§1 of docs/specs/research-artifact.md):
// the common --root/--json flags precede the research/collect subcommand and
// stay on the legacy exit table (usage 64), while adapter commands
// (route/snapshot/receipt) keep their own exit table and unknown bare
// subcommands stay on it (usage 2).
package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/guygrigsby/agent-select/internal/research"
)

// TestResearchCommonLeadingFlagsReachLegacyPath pins the pre-adapter
// invocation shape: leading --root/--json route to the legacy research path
// instead of the adapter usage exit. A missing root makes coverage fail with
// the legacy IO exit (4) and the research diagnostic, not the adapter usage
// line, so no real coverage fixtures are needed.
func TestResearchCommonLeadingFlagsReachLegacyPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	for _, args := range [][]string{
		{"--root", missing, "research", "coverage"},
		{"--json", "--root=" + missing, "research", "coverage"},
	} {
		stdout, stderr, exit := captureRun(t, args)
		if exit != research.ExitIO {
			t.Fatalf("args %v: exit = %d, want legacy IO exit %d; stderr = %q", args, exit, research.ExitIO, stderr)
		}
		if stdout != "" {
			t.Fatalf("args %v: failed coverage wrote stdout: %q", args, stdout)
		}
		if strings.HasPrefix(stderr, "usage: ") {
			t.Fatalf("args %v: adapter usage instead of the research diagnostic: %q", args, stderr)
		}
		if !strings.HasPrefix(stderr, "agent-routing: ") {
			t.Fatalf("args %v: want the research diagnostic, got: %q", args, stderr)
		}
	}
}

// TestUnknownBareSubcommandStaysOnAdapterExitTable: a bare unknown
// subcommand is a usage error on the adapter exit table (2), unchanged from
// the adapter dispatch behavior.
func TestUnknownBareSubcommandStaysOnAdapterExitTable(t *testing.T) {
	stdout, stderr, exit := captureRun(t, []string{"frobnicate"})
	if exit != adapterExitUsage {
		t.Fatalf("exit = %d, want %d", exit, adapterExitUsage)
	}
	if stdout != "" {
		t.Fatalf("nonzero exit wrote stdout: %q", stdout)
	}
	if !strings.HasPrefix(stderr, "usage: ") {
		t.Fatalf("want the adapter usage line, got: %q", stderr)
	}
}

// TestUnknownFlagExitTables: an unknown flag on the explicit research
// invocation (with or without other leading flags) is a legacy usage error
// (64); an unknown flag on an adapter command stays on the adapter table (2).
func TestUnknownFlagExitTables(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{"research subcommand flag", []string{"research", "--nope"}, research.ExitUsage},
		{"leading flag before research", []string{"--nope", "research", "coverage"}, research.ExitUsage},
		{"adapter command flag", []string{"route", "--bogus"}, adapterExitUsage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exit := captureRun(t, tc.args)
			if exit != tc.wantExit {
				t.Fatalf("exit = %d, want %d; stderr = %q", exit, tc.wantExit, stderr)
			}
		})
	}
}
