// Package routing implements the educated routing contract of
// docs/specs/educated-routing.md: an offline, pure, immutable-snapshot
// routing domain library. No network, database, inference or event surface
// exists here by contract.
package routing

import (
	"fmt"
	"unicode/utf8"
)

// Closed error codes (educated-routing.md "Errors have ...").
const (
	CodeInvalidRequest      = "invalid_request"
	CodeInvalidSnapshot     = "invalid_snapshot"
	CodeSchemaIncompatible  = "schema_incompatible"
	CodeSnapshotUnavailable = "snapshot_unavailable"
	CodeNoAdmissibleOffer   = "no_admissible_offer"
)

// Error is the routing error shape. Detail is safe: it never includes raw
// source bytes, credentials or task content.
type Error struct {
	Code   string
	Detail string
}

func (e *Error) Error() string { return e.Code + ": " + e.Detail }

func newError(code, format string, args ...any) *Error {
	detail := fmt.Sprintf(format, args...)
	if len(detail) > 512 {
		detail = detail[:512]
		for !utf8.ValidString(detail) {
			detail = detail[:len(detail)-1]
		}
	}
	return &Error{Code: code, Detail: detail}
}

// invalidRequest reports a caller bug in a request or constructor argument.
func invalidRequest(format string, args ...any) *Error {
	return newError(CodeInvalidRequest, format, args...)
}

// invalidSnapshot reports malformed snapshot bytes or invariants.
func invalidSnapshot(format string, args ...any) *Error {
	return newError(CodeInvalidSnapshot, format, args...)
}

// schemaIncompatible reports unsupported schema or policy versions.
func schemaIncompatible(format string, args ...any) *Error {
	return newError(CodeSchemaIncompatible, format, args...)
}
