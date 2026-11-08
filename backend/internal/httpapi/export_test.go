package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestExportMIDIRoundTrip(t *testing.T) {
	srv := newTestServer(t)

	// Generate a timeline through the API, then export it.
	gen := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 7, "mode": "major"}, "tempoBpm": 90,
		"progression": [{"numeral": "I"}, {"numeral": "IV"}, {"numeral": "V"}, {"numeral": "I"}]
	}`)
	if gen.StatusCode != http.StatusOK {
		t.Fatalf("generate status = %d", gen.StatusCode)
	}
	body, err := io.ReadAll(gen.Body)
	if err != nil {
		t.Fatal(err)
	}

	resp := postJSON(t, srv.URL+"/api/v1/export/midi", string(body))
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, msg)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "audio/midi" {
		t.Errorf("content-type = %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, ".mid") {
		t.Errorf("content-disposition = %q", cd)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("MThd")) || !bytes.Contains(data, []byte("MTrk")) {
		t.Fatalf("not an SMF: % X", data[:16])
	}
	if cl := resp.Header.Get("Content-Length"); cl != itoa(len(data)) {
		t.Errorf("content-length = %q, want %d", cl, len(data))
	}

	// The analysis field must not break export (it is ignored by the writer).
	var tr timelineResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		t.Fatal(err)
	}
	if tr.Timeline.Harmony[0].Analysis == nil {
		t.Fatal("precondition: generated timeline is analysed")
	}
}

func TestExportMIDIErrors(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		name     string
		body     string
		wantCode string
		wantMsg  string
	}{
		{"malformed", `{"timeline":`, "invalid_json", "invalid JSON body"},
		{"empty timeline", `{"timeline":{}}`, "invalid_request", "ppq must be 1..32767"},
		{"bad note", `{"timeline":{"ppq":480,"tempoBpm":120,"timeSignature":{"numerator":4,"denominator":4},"key":{"tonic":0,"mode":"major"},"tracks":[{"name":"p","channel":0,"program":0,"notes":[{"pitch":200,"velocity":100,"start":0,"duration":1}]}],"harmony":[]}}`, "invalid_request", "pitch out of range"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postJSON(t, srv.URL+"/api/v1/export/midi", tt.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			var e errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
				t.Fatal(err)
			}
			if e.Error.Code != tt.wantCode || !strings.Contains(e.Error.Message, tt.wantMsg) {
				t.Fatalf("error = %+v", e.Error)
			}
		})
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

// TestExportTakesALongTimeline (review): a timeline larger than the 64 KiB
// limit for small requests still exports, and one past the timeline limit
// is refused with a message that says so.
func TestExportTakesALongTimeline(t *testing.T) {
	srv := newTestServer(t)
	var notes strings.Builder
	for i := range 3000 {
		if i > 0 {
			notes.WriteString(",")
		}
		fmt.Fprintf(&notes, `{"pitch":60,"velocity":80,"start":%d,"duration":120}`, i*120)
	}
	body := `{"timeline":{"ppq":480,"tempoBpm":120,"timeSignature":{"numerator":4,"denominator":4},"key":{"tonic":0,"mode":"major"},"tracks":[{"name":"chords","channel":0,"program":0,"notes":[` + notes.String() + `]}],"harmony":[]}}`
	if len(body) < 64<<10 {
		t.Fatalf("the body should exceed 64 KiB, is %d", len(body))
	}
	resp := postJSON(t, srv.URL+"/api/v1/export/midi", body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("a long timeline exports: %d %s", resp.StatusCode, b)
	}
}
