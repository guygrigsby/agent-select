package collector

import "fmt"

// Code is the closed storage error taxonomy of delivery-storage.md. Route
// raises none of these by design; the routing codes are owned by the routing
// package.
type Code string

const (
	CodeInvalidRequest         Code = "invalid_request"
	CodeUnauthorized           Code = "unauthorized"
	CodeForbidden              Code = "forbidden"
	CodeNotFound               Code = "not_found"
	CodeStaleLease             Code = "stale_lease"
	CodeConflict               Code = "conflict"
	CodeSourceBlocked          Code = "source_blocked"
	CodeInvalidEvidence        Code = "invalid_evidence"
	CodeIdentityUnresolved     Code = "identity_unresolved"
	CodeTemporarilyUnavailable Code = "temporarily_unavailable"
)

// Error is a storage port error: code plus safe detail. No source bytes or
// credentials in detail text.
type Error struct {
	Code   Code
	Detail string
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }

func errf(c Code, format string, args ...any) *Error {
	return &Error{Code: c, Detail: fmt.Sprintf(format, args...)}
}

// storageErr coerces an unexpected driver error into the taxonomy.
func storageErr(err error) *Error {
	return &Error{Code: CodeTemporarilyUnavailable, Detail: err.Error()}
}
