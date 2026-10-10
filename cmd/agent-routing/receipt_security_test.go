package main

import (
	"strings"
	"testing"

	"github.com/guygrigsby/agent-select/internal/collector"
)

func TestReceiptPortFailureNeverEchoesDetails(t *testing.T) {
	for _, code := range []collector.Code{
		collector.CodeConflict, collector.CodeNotFound,
		collector.CodeInvalidRequest, collector.CodeInvalidEvidence,
		collector.CodeTemporarilyUnavailable,
	} {
		t.Run(string(code), func(t *testing.T) {
			stdout, stderr, exit := captureExit(t, func() int {
				return portFailure(&collector.Error{Code: code, Detail: receiptSecret + "\nFORGED_LOG_LINE"})
			})
			if exit != adapterExitSnapshot || stdout != "" {
				t.Fatalf("exit %d stdout %q", exit, stdout)
			}
			if strings.Contains(stderr, receiptSecret) || strings.Count(stderr, "\n") != 1 {
				t.Fatalf("receipt port detail leaked or injected a line: %q", stderr)
			}
			if !strings.HasPrefix(stderr, "agent-routing: "+string(code)+": ") {
				t.Fatalf("port error code lost: %q", stderr)
			}
		})
	}
}
