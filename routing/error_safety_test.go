package routing

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

const safetyMarker = "fake_task_secret_987654321"

func assertSafeError(t *testing.T, err error, code string) {
	t.Helper()
	e, ok := err.(*Error)
	if !ok || e.Code != code {
		t.Fatalf("want %s routing error, got %v", code, err)
	}
	if strings.Contains(e.Detail, safetyMarker) || strings.Contains(e.Error(), safetyMarker) {
		t.Error("error echoed caller/source content")
	}
	if len(e.Detail) > 512 || !utf8.ValidString(e.Detail) {
		t.Errorf("detail must be valid UTF-8 within 512 bytes: %d", len(e.Detail))
	}
}

func TestDecodeErrorsDoNotEchoArtifact(t *testing.T) {
	snap, err := NewSnapshot(fixtureSnapshotData())
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := EncodeSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}
	base := string(canonical)
	for name, payload := range map[string]string{
		"unknown member":      "{\"" + safetyMarker + "\":0," + base[1:],
		"duplicate member":    "{\"" + safetyMarker + "\":0,\"" + safetyMarker + "\":1," + base[1:],
		"invalid timestamp":   strings.Replace(base, fixtureTime.Format(millisFormat), safetyMarker, 1),
		"wrong integer type":  strings.Replace(base, "\"schema\":1", "\"schema\":\""+safetyMarker+"\"", 1),
		"huge unknown member": "{\"" + strings.Repeat(safetyMarker, 1000) + "\":0," + base[1:],
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeSnapshot([]byte(payload))
			assertSafeError(t, err, CodeInvalidSnapshot)
			engine := NewEngine(snap)
			assertSafeError(t, engine.Adopt([]byte(payload)), CodeInvalidSnapshot)
			if engine.ActiveHash() != snap.Hash() {
				t.Fatal("failed adoption replaced the active snapshot")
			}
		})
	}
	// A nested DTO decoder must not expose its unknown member either.
	var request RequestData
	err = json.Unmarshal([]byte("{\""+safetyMarker+"\":0}"), &request)
	assertSafeError(t, err, CodeInvalidSnapshot)
}

func TestSnapshotErrorsDoNotEchoValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		code   string
		mutate func(*SnapshotData)
	}{
		{"policy", CodeSchemaIncompatible, func(d *SnapshotData) { d.Policy = safetyMarker }},
		{"capability", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Capabilities = []string{safetyMarker} }},
		{"task", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Signals[0].Task = safetyMarker }},
		{"basis", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Signals[0].Basis = safetyMarker }},
		{"resource", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Limits[0].Resource = safetyMarker }},
		{"component", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Prices[0].Component = safetyMarker }},
		{"price", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Prices[0].USDPerMillion = safetyMarker }},
		{"revision", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Signals[0].Revision = safetyMarker }},
		{"warning", CodeInvalidSnapshot, func(d *SnapshotData) { d.Offers[0].Signals[0].Warnings = []string{safetyMarker + "!"} }},
		{"duplicate offer", CodeInvalidSnapshot, func(d *SnapshotData) {
			d.Offers[0].UpstreamID = safetyMarker
			d.Offers[0].ID = "prov/" + safetyMarker
			d.Offers[1] = d.Offers[0]
		}},
		{"duplicate input", CodeInvalidSnapshot, func(d *SnapshotData) { d.Inputs[0].ObservationID = safetyMarker; d.Inputs[1] = d.Inputs[0] }},
		{"duplicate signal", CodeInvalidSnapshot, func(d *SnapshotData) {
			d.Offers[0].Signals[0].ID = safetyMarker
			d.Offers[0].Signals[1] = d.Offers[0].Signals[0]
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := fixtureSnapshotData()
			tc.mutate(&d)
			_, err := NewSnapshot(d)
			assertSafeError(t, err, tc.code)
		})
	}
}

func TestRequestErrorsDoNotEchoValues(t *testing.T) {
	for name, mutate := range map[string]func(*RequestData){
		"task":               func(d *RequestData) { d.Task = safetyMarker },
		"mode":               func(d *RequestData) { d.Mode = safetyMarker },
		"operation":          func(d *RequestData) { d.ImageOperation = safetyMarker },
		"capability":         func(d *RequestData) { d.RequiredCapabilities = []string{safetyMarker} },
		"duplicate provider": func(d *RequestData) { d.AllowedProviders = []string{safetyMarker, safetyMarker} },
		"duplicate runtime": func(d *RequestData) {
			d.Runtime[0].UpstreamID = safetyMarker
			d.Runtime[0].ID = "prov/" + safetyMarker
			d.Runtime = append(d.Runtime, d.Runtime[0])
		},
	} {
		t.Run(name, func(t *testing.T) {
			d := fixtureRequestData()
			mutate(&d)
			_, err := NewRequest(d)
			assertSafeError(t, err, CodeInvalidRequest)
		})
	}
}

func TestRouteErrorDoesNotEchoRequestID(t *testing.T) {
	d := fixtureRequestData()
	d.ID = safetyMarker
	d.Runtime = nil
	decision, err := routeFixture(t, fixtureSnapshotData(), d)
	assertSafeError(t, err, CodeNoAdmissibleOffer)
	if decision.RequestID != safetyMarker {
		t.Fatal("decision lost caller correlation ID")
	}
}

func TestErrorDetailBound(t *testing.T) {
	err := newError(CodeInvalidRequest, "a%s", strings.Repeat("界", 400))
	assertSafeError(t, err, CodeInvalidRequest)
}
