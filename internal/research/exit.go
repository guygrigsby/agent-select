// Package research implements the local research artifact tool contract
// defined in docs/specs/research-artifact.md.
package research

import "fmt"

// Exit codes per §1 of the research artifact contract.
const (
	ExitOK        = 0
	ExitInvalid   = 1 // invalid inputs
	ExitStale     = 2 // stale review pins
	ExitInconsist = 3 // inconsistent export / staged output failure
	ExitIO        = 4 // IO failure, lock unavailable
	ExitUsage     = 64
)

// Error is a tool error carrying a contract exit code and a path/row locator.
// It never includes file contents, only paths, hashes and row identifiers.
type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

func errf(code int, format string, args ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// invalidf builds an exit-1 invalid-input error.
func invalidf(format string, args ...any) *Error {
	return errf(ExitInvalid, format, args...)
}

// stalef builds an exit-2 stale-review error.
func stalef(format string, args ...any) *Error {
	return errf(ExitStale, format, args...)
}

// inconsf builds an exit-3 inconsistent-export error.
func inconsf(format string, args ...any) *Error {
	return errf(ExitInconsist, format, args...)
}

// iof builds an exit-4 IO error.
func iof(format string, args ...any) *Error {
	return errf(ExitIO, format, args...)
}
