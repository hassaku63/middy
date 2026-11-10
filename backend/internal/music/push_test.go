package music

import (
	"strings"
	"testing"
)

// TestPushedChordsArriveEarly (BL-085): a pushed chord sounds an eighth
// or a sixteenth before its beat and is tied over; the chord before it
// stops there. The harmony keeps the chord on its beat, so the bars still
// count the same, and the explanation names the push.
func TestPushedChordsArriveEarly(t *testing.T) {
	key := Key{Tonic: C, Mode: Major}
	prog := Progression{
		{Degree: 2, Seventh: true, Beats: 2},
		{Degree: 5, Seventh: true, Beats: 2, Push: PushEighth},
		{Degree: 1, Beats: 4, Push: PushSixteenth},
	}
	tl, err := FromProgression(key, prog, 120)
	if err != nil {
		t.Fatal(err)
	}
	starts := map[Tick][]Pitch{}
	ends := map[Tick]bool{}
	for _, n := range tl.Tracks[0].Notes {
		starts[n.Start] = append(starts[n.Start], n.Pitch)
		ends[n.Start+n.Duration] = true
	}
	// V7 on the "and" of beat 2 (960 − 240), I a sixteenth before bar 2.
	for _, at := range []Tick{0, 720, 1800} {
		if len(starts[at]) == 0 {
			t.Errorf("no chord starts at %d; starts: %v", at, keys(starts))
		}
	}
	if !ends[720] || !ends[1800] || !ends[3840] {
		t.Errorf("chords should end where the next is pushed in: ends %v", ends)
	}
	if h := tl.Harmony[1]; h.Start != 960 || h.Push != PushEighth {
		t.Errorf("harmony keeps V7 on beat 3 with its push, got start %d push %q", h.Start, h.Push)
	}

	an, err := Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range map[int]string{1: "8 分音符ひとつ早く（2 拍目の裏で）入る", 2: "16 分音符ひとつ早く入る"} {
		if got := an.Harmony[i].Analysis.Explanation; !strings.Contains(got, want) || !strings.Contains(got, "押し") {
			t.Errorf("chord %d: want %q, got %q", i, want, got)
		}
	}
	if got := an.Harmony[0].Analysis.Explanation; strings.Contains(got, "押し") {
		t.Errorf("an unpushed chord says nothing of pushes: %q", got)
	}
}

func TestPushValidation(t *testing.T) {
	tests := []struct {
		name string
		prog Progression
		want string
	}{
		{"first chord", Progression{{Degree: 1, Beats: 4, Push: PushEighth}}, "the first chord has nothing to push into"},
		{"unknown", Progression{{Degree: 1, Beats: 4}, {Degree: 5, Beats: 4, Push: "quarter"}}, `push must be "eighth" or "sixteenth"`},
	}
	for _, tt := range tests {
		err := tt.prog.Validate()
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: want %q, got %v", tt.name, tt.want, err)
		}
	}
}

func keys(m map[Tick][]Pitch) []Tick {
	var out []Tick
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestPushTextReadsAsItsOwnSentence: the push sentence follows the others
// after a space, and its numeral is set off like theirs ("IV は").
func TestPushTextReadsAsItsOwnSentence(t *testing.T) {
	k := Key{C, Major}
	var prog Progression
	for _, l := range []string{"I", "V", "vi", "IV"} {
		s, err := StepFromLabel(k, l, 4)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	prog[3].Push = PushEighth
	tl, err := FromProgression(k, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	if tl, err = Analyze(tl); err != nil {
		t.Fatal(err)
	}
	if e := tl.Harmony[3].Analysis.Explanation; !strings.Contains(e, "。 IV は拍の頭を待たず") {
		t.Errorf("push sentence:\n%s", e)
	}
}
