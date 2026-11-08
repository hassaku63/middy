package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func postJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body)) //nolint:gosec // test-local URL
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestFromProgressionHappyPath(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"},
		"tempoBpm": 90,
		"progression": [{"numeral": "I"}, {"numeral": "V", "beats": 2}, {"numeral": "vi", "beats": 2}, {"numeral": "IV"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	tl := body.Timeline
	if err := tl.Validate(); err != nil {
		t.Fatalf("returned timeline invalid: %v", err)
	}
	if tl.TempoBPM != 90 || tl.Key != (music.Key{Tonic: music.C, Mode: music.Major}) {
		t.Fatalf("unexpected header %+v", tl)
	}
	var numerals, cadences []string
	for _, h := range tl.Harmony {
		numerals = append(numerals, h.RomanNumeral)
		if h.Analysis == nil {
			t.Fatalf("harmony %s has no analysis", h.RomanNumeral)
		}
		cadences = append(cadences, string(h.Analysis.Cadence))
	}
	if got := strings.Join(numerals, " "); got != "I V vi IV" {
		t.Errorf("numerals = %q", got)
	}
	if got := strings.Join(cadences, ","); got != ",half,deceptive," {
		t.Errorf("cadences = %q", got)
	}
	// Default beats (4) + 2 + 2 + 4 = 12 beats.
	if got := tl.LengthTicks(); got != 12*music.DefaultPPQ {
		t.Errorf("LengthTicks = %d, want %d", got, 12*music.DefaultPPQ)
	}
}

func TestFromProgressionErrors(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		name     string
		body     string
		wantCode string
		wantMsg  string
	}{
		{"malformed json", `{"key":`, "invalid_json", "invalid JSON body"},
		{"unknown field", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[{"numeral":"I"}],"tempo":1}`, "invalid_json", "unknown field"},
		{"unknown mode", `{"key":{"tonic":0,"mode":"lydian"},"tempoBpm":90,"progression":[{"numeral":"I"}]}`, "invalid_json", "lydian"},
		{"empty progression", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[]}`, "invalid_request", "at least one chord"},
		{"bad numerals reported together", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[{"numeral":"I"},{"numeral":"VIII"},{"numeral":"X"}]}`, "invalid_request", "progression[1]"},
		{"zero beats", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[{"numeral":"I","beats":0}]}`, "invalid_request", "beats must be >= 1"},
		{"zero tempo", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":0,"progression":[{"numeral":"I"}]}`, "invalid_request", "tempoBpm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", tt.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			var body errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Error.Code != tt.wantCode || !strings.Contains(body.Error.Message, tt.wantMsg) {
				t.Fatalf("error = %+v, want code %q containing %q", body.Error, tt.wantCode, tt.wantMsg)
			}
			if tt.name == "bad numerals reported together" && !strings.Contains(body.Error.Message, "progression[2]") {
				t.Fatalf("both bad labels should be reported: %q", body.Error.Message)
			}
		})
	}
}

func TestFromProgressionWithPatterns(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 100,
		"progression": [{"numeral": "I"}, {"numeral": "V"}],
		"accompaniment": "arpeggio", "drums": "eight-beat"
	}`)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	tl := body.Timeline
	if len(tl.Tracks) != 2 || tl.Tracks[1].Channel != 9 {
		t.Fatalf("expected chords + drums, got %+v", tl.Tracks)
	}
	if n := len(tl.Tracks[0].Notes); n != 16 { // 8 eighths per bar, 2 bars
		t.Fatalf("arpeggio notes = %d, want 16", n)
	}
	if tl.Harmony[1].Analysis == nil {
		t.Fatal("patterns must keep the analysis")
	}

	// The same request swung: the "and" of beat 1 moves from 240 to 320.
	swung := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 100,
		"progression": [{"numeral": "I"}, {"numeral": "V"}],
		"accompaniment": "arpeggio", "drums": "eight-beat", "feel": "swing"
	}`)
	var sw timelineResponse
	if err := json.NewDecoder(swung.Body).Decode(&sw); err != nil {
		t.Fatal(err)
	}
	if sw.Timeline.Tracks[0].Notes[1].Start != 320 || sw.Timeline.Rhythm.Feel == nil {
		t.Fatalf("swing not applied: second note at %d, feel %+v", sw.Timeline.Tracks[0].Notes[1].Start, sw.Timeline.Rhythm.Feel)
	}

	// Unknown ids are the caller's mistake.
	bad := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 100,
		"progression": [{"numeral": "I"}], "drums": "techno"
	}`)
	if bad.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown pattern status = %d", bad.StatusCode)
	}
	var e errorResponse
	_ = json.NewDecoder(bad.Body).Decode(&e)
	if e.Error.Code != "invalid_request" || !strings.Contains(e.Error.Message, "techno") {
		t.Fatalf("error = %+v", e.Error)
	}
}

func TestPatternsCatalogue(t *testing.T) {
	srv := newTestServer(t)
	resp := get(t, srv.URL+"/api/v1/patterns")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body patternsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Accompaniments) < 5 || len(body.Drums) < 3 || body.Accompaniments[0].ID != "block" {
		t.Fatalf("unexpected catalogue: %d accompaniments, %d drums", len(body.Accompaniments), len(body.Drums))
	}
	if len(body.Feels) != 2 || body.Feels[0].ID != "straight" {
		t.Fatalf("feels missing from the catalogue: %+v", body.Feels)
	}
	for _, a := range body.Accompaniments {
		if a.Name == "" || a.Description == "" {
			t.Errorf("%s needs name and description for the UI", a.ID)
		}
	}
}

// TestFromProgressionInSixEight: a numeral without beats lasts one bar of
// the meter, which is two dotted-quarter beats (1440 ticks) in 6/8.
func TestFromProgressionInSixEight(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"},
		"tempoBpm": 60,
		"timeSignature": {"numerator": 6, "denominator": 8},
		"accompaniment": "rocking-6",
		"progression": [{"numeral": "I"}, {"numeral": "IV", "beats": 1}, {"numeral": "V", "beats": 1}, {"numeral": "I"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, b)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var durations []music.Tick
	for _, h := range body.Timeline.Harmony {
		durations = append(durations, h.Duration)
	}
	if got := fmt.Sprint(durations); got != "[1440 720 720 1440]" {
		t.Errorf("durations = %s, want one bar, two half bars, one bar", got)
	}
	swing := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 60,
		"timeSignature": {"numerator": 6, "denominator": 8},
		"accompaniment": "rocking-6", "feel": "swing",
		"progression": [{"numeral": "I"}]
	}`)
	if swing.StatusCode != http.StatusBadRequest {
		t.Errorf("swing in 6/8 should be a bad request, got %d", swing.StatusCode)
	}
}
