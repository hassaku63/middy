package pattern

import (
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestStyleTensions (BL-125, ADR-0012 decision 4): a jazz example adds
// the 9th to its seventh chords and the 13th to its dominants, only when
// the key's scale has them and never a semitone above a chord tone; the
// numerals stay at the seventh, and the rhythm notes say what was added.
func TestStyleTensions(t *testing.T) {
	render := func(key music.Key, style string, steps ...music.Step) music.Timeline {
		t.Helper()
		tl, err := music.FromProgression(key, music.Progression(steps), 120)
		if err != nil {
			t.Fatal(err)
		}
		out, err := Render(tl, Choice{Accompaniment: "block", Style: style})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	classes := func(tl music.Timeline, at music.Tick) []music.PitchClass {
		var pcs []music.PitchClass
		for _, n := range tl.Tracks[0].Notes {
			if n.Start == at && !slices.Contains(pcs, n.Pitch.Class()) {
				pcs = append(pcs, n.Pitch.Class())
			}
		}
		slices.Sort(pcs)
		return pcs
	}
	c := music.Key{Tonic: music.C, Mode: music.Major}
	bar := music.Tick(4 * music.DefaultPPQ)
	jazz := render(c, "jazz",
		music.Step{Degree: 2, Seventh: true, Beats: 4},
		music.Step{Degree: 5, Seventh: true, Beats: 4},
		music.Step{Degree: 1, Seventh: true, Beats: 4},
	)
	want := [][]music.PitchClass{
		{music.C, music.D, music.E, music.F, music.A},          // Dm7 + 9 (E)
		{music.D, music.E, music.F, music.G, music.A, music.B}, // G7 + 9 (A) + 13 (E)
		{music.C, music.D, music.E, music.G, music.B},          // Cmaj7 + 9 (D)
	}
	for i, w := range want {
		if got := classes(jazz, music.Tick(i)*bar); !slices.Equal(got, w) {
			t.Errorf("chord %d: %v, want %v", i, got, w)
		}
	}
	if jazz.Harmony[1].RomanNumeral != "V7" {
		t.Errorf("the numeral stays at the seventh: %s", jazz.Harmony[1].RomanNumeral)
	}
	if s := jazz.Rhythm.Style; s == nil || !strings.Contains(s.Tensions, "9 度") {
		t.Errorf("the style says what it added: %+v", s)
	}

	// In A minor the harmonic-minor V7 (E7) has no diatonic 9th or 13th.
	minor := render(music.Key{Tonic: music.A, Mode: music.Minor}, "jazz",
		music.Step{Degree: 5, Seventh: true, Beats: 4, Quality: ptrQuality(music.Dominant7)},
	)
	if got := classes(minor, 0); len(got) != 4 {
		t.Errorf("E7 in A minor keeps its four tones: %v", got)
	}
	// Rock adds nothing.
	rock := render(c, "rock", music.Step{Degree: 1, Seventh: true, Beats: 4})
	if got := classes(rock, 0); len(got) != 4 || rock.Rhythm.Style.Tensions != "" {
		t.Errorf("rock keeps the plain chord: %v %q", got, rock.Rhythm.Style.Tensions)
	}
}

func ptrQuality(q music.ChordQuality) *music.ChordQuality { return &q }
