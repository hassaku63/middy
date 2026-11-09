package ideas

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
)

// marusa builds an answer with one C-major candidate on the Marunouchi
// Sadistic loop (#272) and the given reason and reply.
func marusa(t *testing.T, summary, reason string, progression ...string) string {
	t.Helper()
	return answer(t, map[string]any{"tonic": 0, "mode": "major"}, summary, reason, progression...)
}

func answer(t *testing.T, key map[string]any, summary, reason string, progression ...string) string {
	t.Helper()
	if len(progression) == 0 {
		progression = []string{"IVmaj7", "V7/vi", "vi7", "I7", "IVmaj7", "V7/vi", "vi7", "subV7/vi"}
	}
	b, err := json.Marshal(map[string]any{
		"summary": summary,
		"mood":    []string{"都会的"},
		"candidates": []map[string]any{{
			"title": "丸サ", "key": key, "tempoBpm": 98,
			"progression": progression, "beatsPerChord": 4, "reason": reason,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The answer of #272: a chord named wrongly and a motion that does not
// happen are both retried, quoting the facts; a corrected answer is used.
func TestClaimsTheChordsContradictAreRetried(t *testing.T) {
	wrong := marusa(t, "丸サ進行のイメージですね。",
		"同じ丸サ進行をスウィングさせた版です。最後の subV7/vi（F7）が半音下の vi7 へ滑り込み、ループの頭へ洒落た緊張感を与えます。")
	right := marusa(t, "丸サ進行のイメージですね。",
		"最後の subV7/vi（Bb7）は、本来は半音下の vi7 へ向かう裏コードですが、ここではループの頭の IVmaj7 へ戻って緊張を残します。")
	client := fake.New(llm.Response{Text: wrong}, llm.Response{Text: right})
	res, err := New(client).Imagine(context.Background(), []Message{user("丸の内サディスティック風の進行")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	for _, want := range []string{"subV7/vi を F7 と書いている", "この調の subV7/vi は Bb7", "subV7/vi が vi7 へ進む", "次は IVmaj7"} {
		if !strings.Contains(res.Retry, want) {
			t.Errorf("retry reason lacks %q: %q", want, res.Retry)
		}
	}
	retry := client.Requests()[1].Messages
	if got := retry[len(retry)-1].Content; !strings.Contains(got, "F7") || !strings.Contains(got, "IVmaj7") {
		t.Errorf("the retry does not quote the problems: %q", got)
	}
	if res.Candidates[0].Reason != "最後の subV7/vi（Bb7）は、本来は半音下の vi7 へ向かう裏コードですが、ここではループの頭の IVmaj7 へ戻って緊張を残します。" {
		t.Errorf("the corrected answer is not used: %q", res.Candidates[0].Reason)
	}
}

// Claims the chords bear out, and wording the check cannot read, pass on
// the first answer.
func TestClaimsTheChordsBearOutAreNotRetried(t *testing.T) {
	tests := []struct {
		name, summary, reason string
		progression           []string
	}{
		{"named and resolving", "丸サ進行ですね。",
			"V7 を V7/vi（E7）に替えて vi7 へ強く引き込み、I7（C7）でブルージーに IVmaj7 へ戻る版です。", nil},
		{"through another chord", "明るい循環。", "G ミクソリディアンの I–VII–IV–I。VII（F）から IV を経て I へ戻ります。", nil},
		{"the loop back to the first chord", "丸サ。", "最後の I7 から頭の IVmaj7 へ戻ります。",
			[]string{"IVmaj7", "V7/vi", "vi7", "I7"}},
		{"a contrast with the usual motion", "丸サ。", "V7/vi は通常 vi へ解決するが、ここでは IVmaj7 へ進まない。", nil},
		{"a numeral of another candidate", "丸サ。", "候補 1 の V7 の代わりに V7/vi を使い、vi7 に解決します。", nil},
		{"letters inside words", "Vm7 と I7 の Swing。", "Imaj7 ではなく I7 を使い、Very lively に。", nil},
		{"a numeral the engine cannot read", "丸サ進行は IVmaj7–III7–vi7–I7 と書かれることもあります。", "III7（E7）は V7/vi と同じ和音で、vi7 へ解決します。", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := marusa(t, tt.summary, tt.reason, tt.progression...)
			if tt.name == "through another chord" {
				text = answer(t, map[string]any{"tonic": 7, "mode": "mixolydian"}, tt.summary, tt.reason, "I", "VII", "IV", "I")
			}
			res, err := New(fake.New(llm.Response{Text: text})).Imagine(context.Background(), []Message{user("x")})
			if err != nil {
				t.Fatalf("Imagine: %v", err)
			}
			if res.Retry != "" {
				t.Errorf("retried: %q", res.Retry)
			}
		})
	}
}

// A run of numerals named name by name in one bracket is checked one to
// one (the evaluation of #283 read "I–V–vi–IV（D–A–Bm–G）" as IV named D).
func TestClaimsNamingARunAreCheckedNameByName(t *testing.T) {
	d := map[string]any{"tonic": 2, "mode": "major"}
	am := map[string]any{"tonic": 9, "mode": "minor"}
	f := map[string]any{"tonic": 5, "mode": "major"}
	tests := []struct {
		name        string
		key         map[string]any
		reason      string
		progression []string
		retry       string // "" for none
	}{
		{"a run named right", d, "王道の I–V–vi–IV（D–A–Bm–G）を 8 ビートで回します。", []string{"I", "V", "vi", "IV"}, ""},
		{"arrows and sevenths", d, "Imaj7→vi7→ii7→V7（Dmaj7→Bm7→Em7→A7）の循環です。", []string{"Imaj7", "vi7", "ii7", "V7"}, ""},
		{"a minor run", am, "i–VII–VI–V（Am–G–F–E）のアンダルシア終止です。", []string{"i", "VII", "VI", "V"}, ""},
		{"a run named wrong", d, "I–V–vi–IV（D–A–Bm–F）を回します。", []string{"I", "V", "vi", "IV"}, "IV を F と書いているが、この調の IV は G"},
		{"more names than numerals", d, "I–V（D–A–Bm）で始まります。", []string{"I", "V", "vi", "IV"}, ""},
		{"a name and words", d, "IV（G、明るい下属和音）へ広がります。", []string{"I", "V", "vi", "IV"}, ""},
		{"names split by a slash", f, "V–V/V（C/G）で盛り上げます。", []string{"I", "V", "V/V", "V"}, ""},
		{"a slash chord", d, "I/3（D/F#）でベースが上がります。", []string{"I", "I/3", "IV", "V"}, ""},
		{"what a chord belongs to", d, "V7（D の V7）で戻ります。", []string{"I", "IV", "V7", "I"}, ""},
		{"a slash chord named wrong", d, "IV（E/G#）へ進みます。", []string{"I", "V", "vi", "IV"}, "IV を E/G# と書いているが"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := answer(t, tt.key, "x。", tt.reason, tt.progression...)
			right := answer(t, tt.key, "x。", "素直な進行です。", tt.progression...)
			res, err := New(fake.New(llm.Response{Text: text}, llm.Response{Text: right})).Imagine(context.Background(), []Message{user("x")})
			if err != nil {
				t.Fatalf("Imagine: %v", err)
			}
			if tt.retry == "" && res.Retry != "" {
				t.Errorf("retried: %q", res.Retry)
			}
			if tt.retry != "" && !strings.Contains(res.Retry, tt.retry) {
				t.Errorf("retry = %q, want %q", res.Retry, tt.retry)
			}
		})
	}
}

// The reply speaks of the image: a named chord must fit some candidate.
// When the retried answer still contradicts its chords, the turn is not
// failed: the contradicted sentences are dropped and the rest is kept.
func TestClaimsStillContradictedAfterTheRetryAreDropped(t *testing.T) {
	text := marusa(t, "都会的な進行ですね。V7/vi（D7）で vi へ引き寄せます。",
		"丸サ進行の形です。最後の subV7/vi が vi7 へ滑り込みます。ループの頭へ戻ります。")
	res, err := New(fake.New(llm.Response{Text: text}, llm.Response{Text: text})).Imagine(context.Background(), []Message{user("x")})
	if err != nil {
		t.Fatalf("Imagine: %v", err)
	}
	if res.Reply != "都会的な進行ですね。" {
		t.Errorf("reply = %q", res.Reply)
	}
	if got := res.Candidates[0].Reason; got != "丸サ進行の形です。ループの頭へ戻ります。" {
		t.Errorf("reason = %q", got)
	}
	if !strings.Contains(res.Retry, "summary: V7/vi を D7 と書いているが、この調の V7/vi は E7") || !strings.Contains(res.Retry, "dropped") {
		t.Errorf("retry = %q", res.Retry)
	}
	if len(res.Candidates[0].Timeline.Harmony) != 8 {
		t.Errorf("the music is kept: %d chords", len(res.Candidates[0].Timeline.Harmony))
	}
}

func TestLabelsInSkipsWordsAndKeepsNumerals(t *testing.T) {
	var got []string
	for _, m := range labelsIn("Vm7 と Ivy、IVmaj7→V7/vi→vi7、subV7/vi（Bb7）、bVII と V7sus4") {
		got = append(got, m.label)
	}
	want := []string{"IVmaj7", "V7/vi", "vi7", "subV7/vi", "bVII", "V7sus4"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("labelsIn = %v, want %v", got, want)
	}
}
