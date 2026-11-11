package evals

import (
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/music"
)

func chord(root music.PitchClass, q music.ChordQuality, symbol string) music.HarmonyEvent {
	return music.HarmonyEvent{Chord: music.Chord{Root: root, Quality: q}, Symbol: symbol}
}

func slash(root music.PitchClass, q music.ChordQuality, bass music.PitchClass, symbol string) music.HarmonyEvent {
	h := chord(root, q, symbol)
	h.Bass = &bass
	return h
}

// TestUnplayedNames (#331): a chord name the reason claims and the
// candidate never plays is found; names it plays, key names, bare letters
// outside a run, notes of a bass line and hedged sentences are not.
func TestUnplayedNames(t *testing.T) {
	bossa := []music.HarmonyEvent{
		chord(music.D, music.Major7, "Dmaj7"), chord(music.F, music.Dominant7, "F7"),
		chord(music.E, music.Minor7, "Em7"), chord(music.A, music.Dominant7, "A7"),
	}
	andalusian := []music.HarmonyEvent{
		chord(music.A, music.MinorTriad, "Am"), chord(music.G, music.MajorTriad, "G"),
		chord(music.F, music.MajorTriad, "F"), chord(music.E, music.MinorTriad, "Em"),
	}
	line := []music.HarmonyEvent{
		chord(music.G, music.MajorTriad, "G"), slash(music.D, music.MajorTriad, music.Fs, "D/F#"),
		chord(music.E, music.MinorTriad, "Em"), chord(music.C, music.MajorTriad, "C"),
	}
	pop := []music.HarmonyEvent{
		chord(music.C, music.MajorTriad, "C"), chord(music.G, music.Dominant7, "G7"),
		chord(music.F, music.Major7, "Fmaj7"),
	}
	cases := []struct {
		name    string
		reason  string
		harmony []music.HarmonyEvent
		want    []string
	}{
		{"a wrong substitute", "Dmaj7 から Eb7（subV7/ii）が半音で Em7 へ下りる。", bossa, []string{"Eb7"}},
		{"names it plays", "Dmaj7 から F7 を経て Em7 へ、A7 で戻る。", bossa, nil},
		{"a sus2 it plays (BL-309)", "Csus2 の開いた響きから C へ。", []music.HarmonyEvent{chord(music.C, music.Sus2, "Csus2"), chord(music.C, music.MajorTriad, "C")}, nil},
		{"an enharmonic spelling of a chord it plays", "Em7 の前で D#7 が鳴る。", []music.HarmonyEvent{chord(music.Ds, music.Dominant7, "Eb7"), chord(music.E, music.Minor7, "Em7")}, nil},
		{"a major for a minor in a run", "Am–G–F–E と下りるアンダルシア終止。", andalusian, []string{"E"}},
		{"a key name", "A minor の切なさと、E♭ メジャーの明るさ。", andalusian, nil},
		{"a bare letter outside a run", "A の音が高く鳴り、サビの B パートへ。", andalusian, nil},
		{"a hedged sentence", "Bb7 ではなく Ab7 が入る。", bossa, nil},
		{"a bass line", "ベースが A→G#→G と下がる。", andalusian, nil},
		{"a slash bass it does not play", "G から G/F# を経て Em へ下りる。", line, []string{"G/F#"}},
		{"a slash bass it plays", "G から D/F# を経て Em へ下りる。", line, nil},
		{"a tension or a plain name for a seventh", "Fmaj9 の柔らかさと、C–G–C の明快さ。", pop, nil},
		{"inside an ASCII word", "BPM は 72、Cadd9 で始まる。", pop, nil},
		{"once each", "C7 から C7 へ。", pop, []string{"C7"}},
	}
	for _, c := range cases {
		if got := unplayedNames(c.reason, c.harmony); !slices.Equal(got, c.want) {
			t.Errorf("%s: %q gave %q, want %q", c.name, c.reason, got, c.want)
		}
	}
}

// TestReportListsUnplayedNames (#331): the report lists, per case and
// candidate, the names a reason claims and the chords that sound, and
// says how many; it is reported, not judged.
func TestReportListsUnplayedNames(t *testing.T) {
	res := result()
	res.Candidates[0].Reason = "C から A7 を経て F へ。"
	res.Candidates[0].Timeline.Harmony = []music.HarmonyEvent{
		chord(music.C, music.MajorTriad, "C"), chord(music.D, music.Dominant7, "D7"), chord(music.F, music.MajorTriad, "F"),
	}
	res.Candidates = append(res.Candidates, ideas.Candidate{})
	v := Judge(Case{ID: "c", Mode: "either"}, res)
	got := Report([]Verdict{v})
	for _, want := range []string{"件数: 1（報告だけで", "| c | 1: A7（鳴るのは C D7 F） |"} {
		if !strings.Contains(got, want) {
			t.Errorf("report lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(strings.Join(v.Problems, ";"), "A7") {
		t.Errorf("unplayed names must not fail the case: %v", v.Problems)
	}
}
