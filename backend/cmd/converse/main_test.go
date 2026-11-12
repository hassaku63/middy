package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/intent"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
	"github.com/hassaku63/middy/backend/internal/music"
)

// TestMeasure (#453): bars, sus4, a sounding ninth and a claimed one.
func TestMeasure(t *testing.T) {
	f := music.Chord{Root: music.F, Quality: music.MajorTriad}
	tl := music.Timeline{
		PPQ:           480,
		TimeSignature: music.FourFour,
		Harmony: []music.HarmonyEvent{
			{Start: 0, Duration: 1920, Chord: f},
			{Start: 1920, Duration: 1920 + 960, Chord: f},
		},
		Tracks: []music.Track{{Name: "chords", Notes: []music.NoteEvent{
			{Pitch: 65, Start: 0, Duration: 1920},
			{Pitch: 67, Start: 0, Duration: 960},    // G, the add9, inside the first chord
			{Pitch: 67, Start: 1680, Duration: 480}, // pushed across the line: not the first chord's
		}}},
	}
	c := ideas.Candidate{Candidate: intent.Candidate{Progression: []string{"I", "Isus4"}, Reason: "9 度で浮遊感"}, Timeline: tl}
	m := measure(c, "")
	if m.Bars != "2.50" || !m.Sus4 || m.Tensioned != 1 || m.Chords != 2 || !m.Claims {
		t.Fatalf("got %+v", m)
	}
}

// TestGistMatchesTheApp (BL-310): the same line history.test.ts expects.
func TestGistMatchesTheApp(t *testing.T) {
	c := ideas.Candidate{Candidate: intent.Candidate{
		Title: "窓辺の雨", Key: music.Key{Tonic: music.A, Mode: music.Minor}, TempoBPM: 72,
		Progression: []string{"i", "VI", "III", "VII"}, BeatsPerChord: 4, Beats: []int{4, 4, 2, 2}, Bars: 3,
		Style: "pop", Accompaniment: "arpeggio", Drums: "ballad", Tensions: "9",
	}}
	want := "- 窓辺の雨: A minor, 4/4, 72 BPM, 3 bars, i:4 VI:4 III:2 VII:2, style pop, accompaniment arpeggio, drums ballad, tensions 9"
	if got := gist(c); got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// TestReadScript (BL-314): a bare array of messages still reads, a script
// object carries its expectations, and a bad tempo direction is refused.
func TestReadScript(t *testing.T) {
	s, err := readScript([]byte(`["雨", "明るく"]`))
	if err != nil || len(s.Steps) != 2 || s.Steps[1].Say != "明るく" || s.Steps[1].Expect != nil {
		t.Fatalf("plain: %+v, %v", s, err)
	}
	s, err = readScript([]byte(`{"name": "n", "turns": [{"say": "3 拍子に", "expect": {"meter": "3/4", "tempo": "down"}}]}`))
	if err != nil || s.Name != "n" || s.Steps[0].Expect.Meter != "3/4" {
		t.Fatalf("script: %+v, %v", s, err)
	}
	for _, bad := range []string{
		`{"turns": [{"say": "x", "expect": {"tempo": "faster"}}]}`,
		`{"turns": []}`,
		`{"turns": [{"say": " "}]}`,
	} {
		if _, err := readScript([]byte(bad)); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

// TestScenariosRead: every script in scenarios/ reads (BL-314).
func TestScenariosRead(t *testing.T) {
	paths, err := filepath.Glob("scenarios/*.json")
	if err != nil || len(paths) < 4 {
		t.Fatalf("scenarios: %v, %v", paths, err)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p) //nolint:gosec // the test's own glob of scenarios/
		if err != nil {
			t.Fatal(err)
		}
		s, err := readScript(data)
		if err != nil || s.Name == "" {
			t.Errorf("%s: %+v, %v", p, s, err)
		}
	}
}

// TestCheck (BL-314): each expectation is marked for one example, in order.
func TestCheck(t *testing.T) {
	waltz := music.TimeSignature{Numerator: 3, Denominator: 4}
	c := ideas.Candidate{
		Candidate: intent.Candidate{
			Key: music.Key{Tonic: music.C, Mode: music.Major}, TempoBPM: 90,
			Progression: []string{"I", "Vsus4", "V"}, Style: "bossa",
		},
		Timeline: music.Timeline{PPQ: 480, TimeSignature: waltz, Harmony: []music.HarmonyEvent{
			{Start: 0, Duration: 1440 * 8},
		}},
	}
	tests := []struct {
		name string
		e    *Expect
		prev float64
		want []string
	}{
		{"nothing to check", nil, 0, nil},
		{"all met", &Expect{Meter: "3/4", Bars: 8, Modes: []string{"major"}, Tempo: "down", Has: []string{"sus4", "bossa"}}, 100,
			[]string{"✓ meter 3/4", "✓ 8 bars", "✓ major", "✓ tempo down", "✓ sus4", "✓ bossa"}},
		{"none met", &Expect{Meter: "6/8", Bars: 4, Modes: []string{"minor", "dorian"}, Tempo: "up", Has: []string{"jazz"}}, 100,
			[]string{"✗ meter 6/8", "✗ 4 bars", "✗ minor/dorian", "✗ tempo up", "✗ jazz"}},
		{"no previous tempo to compare", &Expect{Tempo: "same"}, 0, nil},
		{"the same tempo", &Expect{Tempo: "same"}, 90, []string{"✓ tempo same"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := check(tt.e, c, tt.prev); !slices.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCollectAndReport (BL-314): a script runs against a server the way
// the app talks to it, each turn is reported as it comes back, and the
// report checks the expectations against the previous answer's tempo.
// A run without expectations keeps the table as it was.
func TestCollectAndReport(t *testing.T) {
	tempos := []float64{100, 80}
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []message `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		// The history grows with each turn: user, assistant, user, ...
		if len(body.Messages) != 2*calls+1 {
			t.Errorf("turn %d sent %d messages", calls+1, len(body.Messages))
		}
		res := ideas.Result{Reply: "r", Candidates: []ideas.Candidate{{
			Candidate: intent.Candidate{Title: "t", Key: music.Key{Mode: music.Minor}, TempoBPM: tempos[calls], Progression: []string{"i"}},
			Timeline:  music.Timeline{PPQ: 480, TimeSignature: music.FourFour, Harmony: []music.HarmonyEvent{{Duration: 1920}}},
		}}}
		calls++
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer srv.Close()

	var log strings.Builder
	steps := []Step{{Say: "雨"}, {Say: "ゆっくり", Expect: &Expect{Tempo: "down", Modes: []string{"major"}}}}
	turns := conversation(srv.Client(), srv.URL, steps, progress(&log, 1))
	if !strings.Contains(log.String(), "run 1, turn 1/2: ok") || !strings.Contains(log.String(), "run 1, turn 2/2: ok") {
		t.Errorf("progress: %q", log.String())
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "run-1.json")
	data, err := json.Marshal(turns)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := report(&out, []string{path}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "| expected |") || !strings.Contains(out.String(), "| ✗ major, ✓ tempo down |") {
		t.Errorf("report:\n%s", out.String())
	}

	// Without expectations, the table has no column for them.
	plain := []Turn{{Sent: "雨", Answer: turns[0].Answer}}
	data, err = json.Marshal(plain)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := report(&out, []string{path}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "expected") {
		t.Errorf("a run without expectations grew a column:\n%s", out.String())
	}
}

// TestRunName: runs of several scripts, kept in a directory each, are told
// apart in the report (BL-314).
func TestRunName(t *testing.T) {
	one := []string{"out/run-1.json", "out/run-2.json"}
	if got := runName(one, 1); got != "run-2" {
		t.Errorf("one directory: %q", got)
	}
	several := []string{"out/length/run-1.json", "out/genre/run-1.json"}
	if got := runName(several, 1); got != "genre/run-1" {
		t.Errorf("several directories: %q", got)
	}
}

// TestJudgeInput (BL-315): the reviewer reads each turn's words, the reply
// and every example as it sounds, but not the script's expectations.
func TestJudgeInput(t *testing.T) {
	res := ideas.Result{Reply: "3 拍子にしました。", Candidates: []ideas.Candidate{{
		Candidate: intent.Candidate{Title: "海辺のワルツ", Key: music.Key{Tonic: music.C, Mode: music.Major}, TempoBPM: 90,
			Style: "pop", Reason: "3 拍子で揺れる。"},
		Timeline: music.Timeline{PPQ: 480, TimeSignature: music.TimeSignature{Numerator: 3, Denominator: 4}, Harmony: []music.HarmonyEvent{
			{Duration: 1440, RomanNumeral: "I", Symbol: "C"},
		}},
	}}}
	turns := []Turn{
		{Sent: "海辺", Answer: &res},
		{Sent: "3 拍子にして", Expect: &Expect{Meter: "3/4"}, Answer: &res},
		{Sent: "もっと", Error: "status 502"},
	}
	got := judgeInput(turns)
	for _, want := range []string{"## 2 回目\nユーザー: 3 拍子にして", "返事: 3 拍子にしました。", "候補 1: 海辺のワルツ（C major、3/4、90 BPM、1 小節）", "和音: I（C）", "型: スタイル pop", "理由: 3 拍子で揺れる。", "## 3 回目\nユーザー: もっと\n（答えなし）"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "expect") || strings.Contains(got, "期待") {
		t.Errorf("the expectations leaked:\n%s", got)
	}
}

// TestJudge (BL-315): the reviewer's findings are kept, an unknown
// category is dropped, and a failed call reads as not reviewed.
func TestJudge(t *testing.T) {
	answer := `{"findings": [
		{"category": 1, "turn": 2, "candidate": 0, "quote": "3 拍子にして", "why": "どの候補も 4/4 のまま。"},
		{"category": 9, "turn": 1, "candidate": 1, "quote": "x", "why": "y"}
	]}`
	client := fake.New(llm.Response{Text: answer}, llm.Response{Text: "not json"})
	ok := judge(context.Background(), client, "stacking/run-1", []Turn{{Sent: "x"}})
	if ok.Err != "" || len(ok.Findings) != 1 || ok.Findings[0].Turn != 2 {
		t.Fatalf("got %+v", ok)
	}
	if req := client.Requests()[0]; !strings.Contains(string(req.OutputSchema), "findings") || req.System == "" {
		t.Errorf("request: %+v", req)
	}
	bad := judge(context.Background(), client, "genre/run-1", []Turn{{Sent: "x"}})
	if bad.Err == "" {
		t.Fatal("an unreadable answer must read as not reviewed")
	}
	report := judgeReport([]Verdict{ok, bad})
	for _, want := range []string{
		"- 査読した会話: 1（査読できなかったもの: 1）",
		"反映されていない 1、頼んでいないのに変わった 0",
		"| stacking/run-1 | 2 | 返事 | 反映されていない | 3 拍子にして：どの候補も 4/4 のまま。 |",
		"| genre/run-1 | | | 査読できず |",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("missing %q in:\n%s", want, report)
		}
	}
}
