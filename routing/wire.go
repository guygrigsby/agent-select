package routing

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"
)

// DecodeRequestWire decodes one canonical-v1 request document (the CLI and
// process-boundary transport form of RequestData): UTF-8 without BOM, LF
// newlines only, exactly one trailing newline, no duplicate keys, no
// unknown members, no trailing values. It performs wire framing only;
// every domain rule stays with NewRequest. The returned error is the
// domain invalid_request code; the transport maps it to its own
// wire-error code if its contract has one.
func DecodeRequestWire(data []byte) (RequestData, error) {
	if len(data) == 0 {
		return RequestData{}, invalidRequest("empty request document")
	}
	if !utf8.Valid(data) {
		return RequestData{}, invalidRequest("request document is not valid UTF-8")
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return RequestData{}, invalidRequest("request document carries a BOM")
	}
	if bytes.ContainsRune(data, '\r') {
		return RequestData{}, invalidRequest("request document contains CR bytes")
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		return RequestData{}, invalidRequest("request document lacks its trailing newline")
	}
	if bytes.HasSuffix(data, []byte("\n\n")) {
		return RequestData{}, invalidRequest("request document carries more than one trailing newline")
	}
	if err := strictCheck(data); err != nil {
		var de *Error
		if errors.As(err, &de) {
			return RequestData{}, de
		}
		return RequestData{}, invalidRequest("request document: %v", err)
	}
	var d RequestData
	if err := strictUnmarshal(data, &d); err != nil {
		var de *Error
		if errors.As(err, &de) {
			return RequestData{}, de
		}
		return RequestData{}, invalidRequest("request document: %v", err)
	}
	return d, nil
}

// CanonicalEncode returns the canonical-v1 encoding of a value: the pinned
// JSON canonicalization the delivery uses for snapshots, requests and
// decisions alike.
func CanonicalEncode(v any) ([]byte, error) {
	b, err := canonicalize(v)
	if err != nil {
		return nil, fmt.Errorf("canonical encode: %w", err)
	}
	return b, nil
}
