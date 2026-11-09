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
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
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
		{"push on the first chord", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[{"numeral":"I","push":"eighth"}]}`, "invalid_request", "nothing to push into"},
		{"unsupported meter", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"timeSignature":{"numerator":0,"denominator":4},"progression":[{"numeral":"I"}]}`, "invalid_request", "meter 0/4 is not supported"},
		{"a label the key would change", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":90,"progression":[{"numeral":"Vmaj7"}]}`, "invalid_request", "Vmaj7"},
		{"zero tempo", `{"key":{"tonic":0,"mode":"major"},"tempoBpm":0,"progression":[{"numeral":"I"}]}`, "invalid_request", "tempoBpm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", tt.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			var body problem
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Code != tt.wantCode || !strings.Contains(body.Detail, tt.wantMsg) {
				t.Fatalf("error = %+v, want code %q containing %q", body, tt.wantCode, tt.wantMsg)
			}
			if tt.name == "bad numerals reported together" && !strings.Contains(body.Detail, "progression[2]") {
				t.Fatalf("both bad labels should be reported: %q", body.Detail)
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

	// The same request swung: at 100 BPM the default ratio is 2.6 : 1
	// (BL-084), so the "and" of beat 1 moves from 240 to 347.
	swung := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 100,
		"progression": [{"numeral": "I"}, {"numeral": "V"}],
		"accompaniment": "arpeggio", "drums": "eight-beat", "feel": "swing"
	}`)
	var sw timelineResponse
	if err := json.NewDecoder(swung.Body).Decode(&sw); err != nil {
		t.Fatal(err)
	}
	if sw.Timeline.Tracks[0].Notes[1].Start != 347 || sw.Timeline.Rhythm.Feel == nil || sw.Timeline.Rhythm.Feel.Ratio != 2.6 {
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
	var e problem
	_ = json.NewDecoder(bad.Body).Decode(&e)
	if e.Code != "invalid_request" || !strings.Contains(e.Detail, "techno") {
		t.Fatalf("error = %+v", e)
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
	// The styles come with the catalogue so the card can name them (BL-119).
	if len(body.Styles) != 6 || body.Styles[5].ID != "celtic" || body.Styles[5].Name != "ケルト" {
		t.Fatalf("styles missing from the catalogue: %+v", body.Styles)
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

// TestFromProgressionKeepsTheStyle (BL-110): a style puts a keyboard
// pattern on its instrument and is recorded on the timeline, so a key
// change can send it back.
func TestFromProgressionKeepsTheStyle(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 2, "mode": "dorian"}, "tempoBpm": 100,
		"accompaniment": "arpeggio", "style": "celtic",
		"progression": [{"numeral": "i"}, {"numeral": "VII"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, b)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	tl := body.Timeline
	if tl.Tracks[0].Program != 46 || tl.Rhythm == nil || tl.Rhythm.Style == nil || tl.Rhythm.Style.Instrument != "ハープ" {
		t.Errorf("celtic arpeggios should be on the harp: program %d, rhythm %+v", tl.Tracks[0].Program, tl.Rhythm)
	}
	bad := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 2, "mode": "dorian"}, "tempoBpm": 100, "style": "polka",
		"progression": [{"numeral": "i"}]
	}`)
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("an unknown style should be a bad request, got %d", bad.StatusCode)
	}
}

// TestFromProgressionPushes (BL-085): a pushed chord is struck early and
// keeps its push, so a key change can send it back.
func TestFromProgressionPushes(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 120,
		"progression": [{"numeral": "ii7", "beats": 2}, {"numeral": "V7", "beats": 2, "push": "eighth"}, {"numeral": "Imaj7"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, b)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	h := body.Timeline.Harmony[1]
	if h.Push != music.PushEighth || h.Start != 960 || !strings.Contains(h.Analysis.Explanation, "押し") {
		t.Errorf("V7 should keep its beat and its push and be explained: %+v", h)
	}
}

// TestFromProgressionDrumGrid (BL-123): drums written as a grid render,
// come back on the timeline so a key change can send them again, and a
// bad grid is a bad request naming every problem.
func TestFromProgressionDrumGrid(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 120,
		"drumGrid": {"division": 2, "voices": {"kick": "x... x.x.", "snare": "..x. ..x."}},
		"progression": [{"numeral": "I"}, {"numeral": "V"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, b)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	r := body.Timeline.Rhythm
	if r == nil || r.Drums == nil || r.Drums.ID != "grid" || r.DrumGrid == nil || r.DrumGrid.Voices["kick"] != "x... x.x." {
		t.Fatalf("the timeline should keep the grid: %+v", r)
	}
	if !strings.Contains(r.Drums.Explanation, "バックビート") {
		t.Errorf("the grid is explained: %s", r.Drums.Explanation)
	}

	bad := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"}, "tempoBpm": 120,
		"drumGrid": {"division": 2, "voices": {"kick": "x...", "cowbell": "x......."}},
		"progression": [{"numeral": "I"}]
	}`)
	var e problem
	if err := json.NewDecoder(bad.Body).Decode(&e); err != nil {
		t.Fatal(err)
	}
	if bad.StatusCode != http.StatusBadRequest || !strings.Contains(e.Detail, `voices["kick"]`) || !strings.Contains(e.Detail, "cowbell") {
		t.Errorf("a bad grid should be a 400 naming both problems: %d %+v", bad.StatusCode, e)
	}
}

// TestFromProgressionTakesATritoneSubstitute (BL-128): subV7/x goes through
// the API, is spelled in the key and analysed as a substitute, and keeps
// its label for a rebuild in another key.
func TestFromProgressionTakesATritoneSubstitute(t *testing.T) {
	srv := newTestServer(t)
	resp := postJSON(t, srv.URL+"/api/v1/timelines/from-progression", `{
		"key": {"tonic": 0, "mode": "major"},
		"tempoBpm": 120,
		"progression": [{"numeral": "ii7"}, {"numeral": "subV7/I"}, {"numeral": "Imaj7"}]
	}`)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	var body timelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	sub := body.Timeline.Harmony[1]
	if sub.RomanNumeral != "subV7/I" || sub.Symbol != "Db7" || sub.Analysis == nil || !sub.Analysis.Substitute || sub.Analysis.AppliedTo != 1 {
		t.Fatalf("subV7/I = %+v %+v", sub, sub.Analysis)
	}
}
