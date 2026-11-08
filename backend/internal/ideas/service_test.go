package ideas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/intent"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
	"github.com/hassaku63/middy/backend/internal/music"
)

const goodIntent = `{
  "summary": "雨の日曜の午後。静かで少し切ない。",
  "mood": ["静か", "切ない"],
  "candidates": [
    {"title": "窓辺の雨", "key": {"tonic": 9, "mode": "minor"}, "tempoBpm": 72,
     "progression": ["i", "VI", "III", "VII"], "beatsPerChord": 4, "reason": "短調で落ち着いた切なさ。"},
    {"title": "午後の光", "key": {"tonic": 5, "mode": "major"}, "tempoBpm": 80,
     "progression": ["Imaj7", "vi7", "ii7", "V7"], "beatsPerChord": 2, "reason": "七の和音で柔らかく。"}
  ]
}`

const badIntent = `{"summary": "x", "mood": [], "candidates": [
  {"title": "t", "key": {"tonic": 0, "mode": "major"}, "tempoBpm": 500, "progression": ["bIII"], "beatsPerChord": 4, "reason": "r"}
]}`

func user(text string) Message { return Message{Role: llm.RoleUser, Content: text} }

func TestImagineRendersEveryCandidate(t *testing.T) {
	client := fake.New(llm.Response{Text: goodIntent})
	res, err := New(client).Imagine(context.Background(), []Message{user("雨の日曜の午後")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if res.Reply != "雨の日曜の午後。静かで少し切ない。" || len(res.Mood) != 2 {
		t.Fatalf("unexpected reply/mood: %+v", res)
	}
	if res.Retry != "" {
		t.Errorf("no retry happened, but Retry = %q", res.Retry)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("got %d candidates", len(res.Candidates))
	}
	first := res.Candidates[0]
	if first.Title != "窓辺の雨" || first.Timeline.TempoBPM != 72 || len(first.Timeline.Harmony) != 4 {
		t.Fatalf("first candidate not rendered: %+v", first)
	}
	if first.Timeline.Harmony[0].RomanNumeral != "i" || first.Timeline.Harmony[0].Analysis == nil {
		t.Fatalf("timeline not analysed: %+v", first.Timeline.Harmony[0])
	}
	// 8 chords of 2 beats = 16 beats.
	if got := res.Candidates[1].Timeline.LengthTicks(); got != 8*480 {
		t.Fatalf("second candidate length = %d", got)
	}

	// The request carried the system prompt, the schema and the user text.
	reqs := client.Requests()
	if len(reqs) != 1 || reqs[0].System == "" || reqs[0].OutputSchema == nil || reqs[0].Messages[0].Content != "雨の日曜の午後" {
		t.Fatalf("unexpected request: %+v", reqs)
	}
}

func TestImagineRetriesOnceWithTheValidationErrors(t *testing.T) {
	client := fake.New(llm.Response{Text: badIntent}, llm.Response{Text: goodIntent})
	res, err := New(client).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if len(res.Candidates) != 2 {
		t.Fatalf("retry result not used: %+v", res)
	}
	if !strings.Contains(res.Retry, "bIII") {
		t.Errorf("the result must say why the first answer was not used: %q", res.Retry)
	}
	reqs := client.Requests()
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(reqs))
	}
	retry := reqs[1].Messages
	if len(retry) != 3 || retry[1].Role != llm.RoleAssistant || retry[1].Content != badIntent {
		t.Fatalf("retry must quote the bad answer as the assistant turn: %+v", retry)
	}
	if !strings.Contains(retry[2].Content, "tempoBpm") || !strings.Contains(retry[2].Content, "bIII") {
		t.Fatalf("retry must carry the validation errors: %q", retry[2].Content)
	}
}

func TestImagineGivesUpAfterTheRetry(t *testing.T) {
	client := fake.New(llm.Response{Text: badIntent}, llm.Response{Text: "not json at all"})
	_, err := New(client).Imagine(context.Background(), []Message{user("雨")})
	if !errors.Is(err, ErrInvalidOutput) {
		t.Fatalf("want ErrInvalidOutput, got %v", err)
	}
	if len(client.Requests()) != 2 {
		t.Fatalf("exactly one retry expected, got %d requests", len(client.Requests()))
	}
}

func TestImagineReportsProviderFailures(t *testing.T) {
	// Exhaust the script so the next call fails with ErrNoScriptedResponse.
	client := fake.New(llm.Response{Text: goodIntent})
	_, _ = client.Complete(context.Background(), llm.Request{})
	_, err := New(client).Imagine(context.Background(), []Message{user("x")})
	if !errors.Is(err, ErrLLMUnavailable) {
		t.Fatalf("want ErrLLMUnavailable, got %v", err)
	}
}

func TestImagineValidatesTheConversation(t *testing.T) {
	svc := New(fake.New(llm.Response{Text: goodIntent}))
	tests := []struct {
		name string
		msgs []Message
		want string
	}{
		{"empty", nil, "must not be empty"},
		{"ends with assistant", []Message{user("a"), {Role: llm.RoleAssistant, Content: "b"}}, "last message must be from the user"},
		{"bad role", []Message{{Role: "system", Content: "x"}}, "role must be user or assistant"},
		{"blank", []Message{user("   ")}, "content is empty"},
		{"too long", []Message{user(strings.Repeat("あ", MaxMessageRunes+1))}, "too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Imagine(context.Background(), tt.msgs)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want %q, got %v", tt.want, err)
			}
		})
	}
}

func TestUnscriptedFakeProducesRenderableSamples(t *testing.T) {
	svc := New(fake.New())
	for _, text := range []string{"雨の日曜の午後", "出発の朝、晴れ"} {
		res, err := svc.Imagine(context.Background(), []Message{user(text)})
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if len(res.Candidates) < 1 {
			t.Fatalf("%q: no candidates", text)
		}
	}
	rainy, _ := svc.Imagine(context.Background(), []Message{user("雨")})
	sunny, _ := svc.Imagine(context.Background(), []Message{user("晴れ")})
	if rainy.Candidates[0].Key.Mode == sunny.Candidates[0].Key.Mode {
		t.Fatal("samples should differ by mood keywords")
	}
}

func TestResultJSONShape(t *testing.T) {
	res, err := New(fake.New(llm.Response{Text: goodIntent})).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"reply"`, `"mood"`, `"candidates"`, `"title"`, `"timeline"`, `"romanNumeral"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON should contain %s", key)
		}
	}
}

func TestImagineAppliesCandidatePatterns(t *testing.T) {
	withPatterns := strings.Replace(goodIntent, `"beatsPerChord": 4, "reason": "短調で落ち着いた切なさ。"`,
		`"beatsPerChord": 4, "reason": "短調で落ち着いた切なさ。", "accompaniment": "arpeggio", "drums": "eight-beat"`, 1)
	res, err := New(fake.New(llm.Response{Text: withPatterns})).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	first := res.Candidates[0]
	if len(first.Timeline.Tracks) != 2 || first.Timeline.Tracks[1].Channel != 9 {
		t.Fatalf("patterns not applied: %+v", first.Timeline.Tracks)
	}
	if first.Timeline.Rhythm == nil || first.Timeline.Rhythm.Drums == nil {
		t.Fatal("rhythm explanation missing")
	}
	if first.Timeline.Harmony[0].Analysis == nil {
		t.Fatal("harmony analysis must survive pattern application")
	}
	// The second candidate has no patterns and stays block chords.
	if second := res.Candidates[1]; len(second.Timeline.Tracks) != 1 || second.Timeline.Rhythm != nil {
		t.Fatalf("unexpected patterns on second candidate: %+v", second.Timeline.Tracks)
	}
}

func TestSystemPromptListsTheCatalogue(t *testing.T) {
	p := SystemPrompt()
	for _, id := range []string{"block", "arpeggio", "strum", "eight-beat", "bossa"} {
		if !strings.Contains(p, "- "+id+":") {
			t.Errorf("prompt should list %s", id)
		}
	}
	if SystemPrompt() != p {
		t.Error("prompt must be stable across calls (caching)")
	}
}

func TestImagineSwingsWhenTheCandidateAsks(t *testing.T) {
	// The unscripted fake swings its first candidate on a swing word.
	res, err := New(fake.New()).Imagine(context.Background(), []Message{user("ジャズっぽく")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	first := res.Candidates[0]
	if first.Feel != "swing" || first.Timeline.Rhythm == nil || first.Timeline.Rhythm.Feel == nil {
		t.Fatalf("swing not applied: feel=%q rhythm=%+v", first.Feel, first.Timeline.Rhythm)
	}
	if !strings.Contains(SystemPrompt(), "swing") {
		t.Fatal("the prompt must offer the feels")
	}
}

func TestDiagnosticServiceAsksForAndKeepsTheUnconstrainedField(t *testing.T) {
	withNote := strings.Replace(goodIntent, `"summary": "雨の日曜の午後。静かで少し切ない。",`,
		`"summary": "雨の日曜の午後。静かで少し切ない。", "unconstrained": "本当は Dm9 と Bbmaj7 の浮遊感を使いたい",`, 1)
	client := fake.New(llm.Response{Text: withNote})
	res, err := NewDiagnostic(client).Imagine(context.Background(), []Message{user("雨")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(client.Requests()[0].OutputSchema), "unconstrained") {
		t.Fatal("the diagnostic service must send the diagnostic schema")
	}
	if res.Unconstrained != "本当は Dm9 と Bbmaj7 の浮遊感を使いたい" {
		t.Fatalf("unconstrained not kept: %q", res.Unconstrained)
	}
	// The ordinary service never asks.
	plain := fake.New(llm.Response{Text: goodIntent})
	if _, err := New(plain).Imagine(context.Background(), []Message{user("雨")}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(plain.Requests()[0].OutputSchema), "unconstrained") {
		t.Fatal("the production schema must not ask for the diagnostic field")
	}
}

// TestImagineRendersSixEight: the fake answers a rocking image in 6/8; the
// service renders it in dotted-quarter beats with the tempo of the beat
// (ADR-0017, ADR-0019), and the prompt lists the 6/8 catalogue.
func TestImagineRendersSixEight(t *testing.T) {
	res, err := New(fake.New()).Imagine(context.Background(), []Message{user("ゆりかごの子守唄")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	for _, c := range res.Candidates {
		tl := c.Timeline
		if tl.TimeSignature != music.SixEight || tl.TempoBPM != c.TempoBPM {
			t.Fatalf("%s: meter %v at %v BPM", c.Title, tl.TimeSignature, tl.TempoBPM)
		}
		want := music.Tick(c.BeatsPerChord) * 720 // dotted quarters
		if d := tl.Harmony[0].Duration; d != want {
			t.Errorf("%s: a chord of %d beats lasts %d ticks, want %d", c.Title, c.BeatsPerChord, d, want)
		}
	}
	p := SystemPrompt()
	for _, want := range []string{"6/8 で使える伴奏の型", "- rocking-6:", "- jig:", "付点 4 分"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}

// TestReasonSpeaksOnlyOfWhatSounds (#128): the reason is the listening
// guide to this example, so the prompt and the schema keep it to what the
// example plays; instruments it does not play, or a melody, stay out.
func TestReasonSpeaksOnlyOfWhatSounds(t *testing.T) {
	p := SystemPrompt()
	for _, want := range []string{"reason は、この例で実際に鳴るもの", "鳴らない楽器", "メロディ"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if !strings.Contains(string(intent.Schema), "この例で鳴るものだけ") {
		t.Error("the schema's reason should say it speaks only of what the example plays")
	}
}

// TestPromptSteersCelticAwayFromPopGrooves (BL-107, #131): the prompt lists
// the reel and bodhrán patterns and keeps the eight-beat kit off Celtic examples.
func TestPromptSteersCelticAwayFromPopGrooves(t *testing.T) {
	p := SystemPrompt()
	for _, want := range []string{"- reel:", "- bodhran:", "drums を bodhran", "ケルトの例には使わない", "- drone:", "accompaniment を drone"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	res, err := New(fake.New()).Imagine(context.Background(), []Message{user("アイルランドのパブのケルト音楽")})
	if err != nil {
		t.Fatal(err)
	}
	if c := res.Candidates[0]; c.Accompaniment != "reel" || c.Drums != "bodhran" {
		t.Errorf("the fake's Celtic reel should use the Celtic patterns: %s / %s", c.Accompaniment, c.Drums)
	}
}

// TestImagineRendersPerChordBeats (BL-109): a candidate's beats give each
// chord its own length; the prompt says how to write them.
func TestImagineRendersPerChordBeats(t *testing.T) {
	const held = `{"summary": "s", "mood": [], "candidates": [
	  {"title": "t", "key": {"tonic": 2, "mode": "dorian"}, "tempoBpm": 112,
	   "progression": ["i", "IV", "i", "VII"], "beatsPerChord": 4, "beats": [8, 4, 2, 2], "reason": "r"}]}`
	res, err := New(fake.New(llm.Response{Text: held})).Imagine(context.Background(), []Message{user("ケルト")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	var durations []music.Tick
	for _, h := range res.Candidates[0].Timeline.Harmony {
		durations = append(durations, h.Duration)
	}
	if got := fmt.Sprint(durations); got != "[3840 1920 960 960]" {
		t.Errorf("durations = %s, want two bars, a bar, two beats, two beats", got)
	}
	if !strings.Contains(SystemPrompt(), "beats に progression と同じ数の長さ") {
		t.Error("the prompt should explain beats")
	}
}

// TestPromptSteersJazzToTheRide (BL-081): the prompt lists the jazz
// patterns and asks for them, with swing, on a jazz image.
func TestPromptSteersJazzToTheRide(t *testing.T) {
	p := SystemPrompt()
	for _, want := range []string{"- comp:", "- jazz-ride:", "drums を jazz-ride", "8 ビートとボサノバのドラムはジャズには使わない", "- jazz-waltz:", "drums を jazz-waltz"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}

// TestFakeGraduationUsesSlashChords (BL-116): the fake's graduation answer
// renders its slash chords with the bass and the line explained.
func TestFakeGraduationUsesSlashChords(t *testing.T) {
	res, err := New(fake.New()).Imagine(context.Background(), []Message{user("卒業式の合唱曲")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	tl := res.Candidates[0].Timeline
	if h := tl.Harmony[1]; h.Symbol != "C/B" || h.Bass == nil {
		t.Errorf("second chord should be C/B with its bass: %+v", h)
	}
	if !strings.Contains(tl.Overview.Flow, "ベースが C → B → A → G → F → E → D と順に下がる") {
		t.Errorf("the flow should name the falling bass: %s", tl.Overview.Flow)
	}
	if !strings.Contains(res.Candidates[1].Timeline.Overview.Flow, "ライン・クリシェ") {
		t.Errorf("the second candidate is a line cliché: %s", res.Candidates[1].Timeline.Overview.Flow)
	}
}

// TestPromptListsTheStyles (BL-082): the prompt lists each style with the
// patterns that belong to it, generated from the catalogue.
func TestPromptListsTheStyles(t *testing.T) {
	p := SystemPrompt()
	for _, want := range []string{"最初にスタイルを選び", "- celtic: ケルト。", "ドラム: bodhran, jig, six-eight", "- jazz: ジャズ。"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}
