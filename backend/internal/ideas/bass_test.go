package ideas

import (
	"context"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/intent"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
	"github.com/hassaku63/middy/backend/internal/music"
)

var cMajor = map[string]any{"tonic": 0, "mode": "major"}

// graduation walks the bass down from C under a C-major loop.
var graduation = []string{"I", "V/7", "vi", "I/5", "IV", "I/3", "ii", "V"}

func rendered(t *testing.T, key map[string]any, reason string, progression ...string) Result {
	t.Helper()
	in, err := intent.Parse([]byte(answer(t, key, "例です。", reason, progression...)))
	if err != nil {
		t.Fatal(err)
	}
	res, err := render(in)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBassRunsAreCheckedAgainstWhatIsPlayed(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		progression  []string
		problem      string // "" when the reason stands
	}{
		{"a walk down the slash basses", "ベースが C→B→A→G と 1 音ずつ下がります。", graduation, ""},
		{"a run across the loop's end", "低音は E→D→G→C と戻ってきます。", graduation, ""},
		{"a held bass moves once", "ベースは C→A→F→G と動きます。", []string{"I", "I", "vi", "IV", "V"}, ""},
		{"enharmonic spelling", "ベースが C→Bb→A と下がる。", []string{"I", "bVII", "vi"}, ""},
		{"a descent that is not there", "ベースが F→E→D→C と下がり、切なさを出します。", []string{"I", "IV", "V", "I"},
			"ベースを F→E→D→C と書いているが、この候補のベースは C→F→G→C"},
		{"chord names are not a bass run", "ベースの上で F→Em→Dm→C と和音が動きます。", []string{"I", "IV", "V"}, ""},
		{"a sentence not about the bass", "メロディが F→E→D→C と下がります。", []string{"I", "IV", "V"}, ""},
		{"a contrast", "ベースが本来の F→E→D→C ではなく、跳躍します。", []string{"I", "IV", "V"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			problems := checkClaims(rendered(t, cMajor, tc.reason, tc.progression...))
			var got string
			if len(problems) > 0 {
				got = problems[0].message
			}
			if got != tc.problem {
				t.Errorf("problem = %q, want %q", got, tc.problem)
			}
		})
	}
}

// A walking bass's passing notes count: the lowest notes the
// accompaniment plays bear out a run between the chord roots.
func TestBassRunsMayFollowTheLowestNotes(t *testing.T) {
	c := rendered(t, cMajor, "例です。", "I", "vi").Candidates[0]
	c.Timeline.Tracks = []music.Track{{Name: music.ChordTrackName, Notes: []music.NoteEvent{
		{Pitch: 48, Velocity: 80, Start: 0, Duration: 480}, {Pitch: 64, Velocity: 80, Start: 0, Duration: 960},
		{Pitch: 47, Velocity: 80, Start: 960, Duration: 480},
		{Pitch: 45, Velocity: 80, Start: 1920, Duration: 480},
	}}}
	if got := bassProblems("ベースが C→B→A と歩きます。", c); len(got) != 0 {
		t.Errorf("problems = %q, want none", got)
	}
	if got := bassProblems("ベースが C→D→E と歩きます。", c); len(got) != 1 {
		t.Errorf("problems = %q, want one", got)
	}
}

func TestAnInventedBassRunIsRetried(t *testing.T) {
	wrong := answer(t, cMajor, "卒業の歌ですね。", "ベースが F→E→D→C と下がります。", "I", "IV", "V", "I")
	right := answer(t, cMajor, "卒業の歌ですね。", "ベースが C→B→A→G と下がります。", graduation...)
	client := fake.New(llm.Response{Text: wrong}, llm.Response{Text: right})
	res, err := New(client).Imagine(context.Background(), []Message{user("卒業式の歌")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if !strings.Contains(res.Retry, "この候補のベースは C→F→G→C") {
		t.Errorf("retry reason = %q", res.Retry)
	}
	if res.Candidates[0].Reason != "ベースが C→B→A→G と下がります。" {
		t.Errorf("reason = %q", res.Candidates[0].Reason)
	}
}
