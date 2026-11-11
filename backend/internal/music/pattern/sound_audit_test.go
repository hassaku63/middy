package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// soundAuditTimeline builds a C major example from labels of the given
// beats each.
func soundAuditTimeline(t *testing.T, beats int, labels ...string) music.Timeline {
	t.Helper()
	return soundAuditTimelineIn(t, music.FourFour, beats, labels...)
}

func soundAuditTimelineIn(t *testing.T, ts music.TimeSignature, beats int, labels ...string) music.Timeline {
	t.Helper()
	key := music.Key{Tonic: music.C, Mode: music.Major}
	var prog music.Progression
	for _, l := range labels {
		s, err := music.StepFromLabel(key, l, beats)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := music.FromProgressionIn(ts, key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

// classesOver is the pitch classes the chord track sounds over chord i.
func classesOver(out music.Timeline, i int) map[music.PitchClass]bool {
	h := out.Harmony[i]
	got := map[music.PitchClass]bool{}
	for _, n := range out.Tracks[0].Notes {
		if n.Start >= h.Start && n.Start < h.Start+h.Duration {
			got[n.Pitch.Class()] = true
		}
	}
	return got
}

// TestArpeggioSpellsTheSeventh (#404 S1): an arpeggio plays a seventh
// chord's seventh, so the V7 it names sounds its tritone; a short chord
// with a tension keeps the seventh, and the notes do not claim a tension
// that did not sound.
func TestArpeggioSpellsTheSeventh(t *testing.T) {
	for acc, ts := range map[string]music.TimeSignature{"arpeggio": music.FourFour, "arpeggio-3": music.ThreeFour, "arpeggio-6": music.SixEight} {
		bar := ts.BeatsPerBar()
		tl := soundAuditTimelineIn(t, ts, bar, "Imaj7", "vi7", "ii7", "V7")
		out, err := Render(tl, Choice{Accompaniment: acc})
		if err != nil {
			t.Fatal(err)
		}
		if e := out.Rhythm.Accompaniment.Explanation; !strings.Contains(e, "7 度を弾いて") {
			t.Errorf("%s: the notes do not say the seventh replaces the octave: %s", acc, e)
		}
		for i, h := range out.Harmony {
			got := classesOver(out, i)
			for _, pc := range h.Chord.PitchClasses() {
				if !got[pc.Normalize()] {
					t.Errorf("%s over %s lacks %v: %v", acc, h.RomanNumeral, pc, got)
				}
			}
		}
	}

	short := soundAuditTimeline(t, 2, "Imaj7", "ii7", "V7", "Imaj7")
	out, err := Render(short, Choice{Accompaniment: "arpeggio", Tensions: TensionsNinths})
	if err != nil {
		t.Fatal(err)
	}
	if got := classesOver(out, 2); !got[music.F] || !got[music.B] {
		t.Errorf("a two-beat V7 lacks its tritone: %v", got)
	}
	if rt := out.Rhythm.Tensions; rt != nil && strings.Contains(rt.Explanation, "足し") {
		t.Errorf("no tension sounded, yet the notes say one did: %+v", rt)
	}
}

// TestTripletGridKeepsItsTriplets (#404 S3): a grid in triplets already
// swings; a shuffle leaves its strokes on the triplets, where the
// shuffled eighths of the chords fall too.
func TestTripletGridKeepsItsTriplets(t *testing.T) {
	tl := soundAuditTimeline(t, 4, "I7", "IV7")
	g := grid(3, map[string]string{"kick": "x.. x.. x.. x..", "hat": "x.x x.x x.x x.x"})
	out, err := Render(tl, Choice{Accompaniment: "pulse", Grid: &g, Feel: FeelShuffle})
	if err != nil {
		t.Fatal(err)
	}
	var drums music.Track
	for _, tr := range out.Tracks {
		if tr.Channel == 9 {
			drums = tr
		}
	}
	for _, n := range drums.Notes {
		if n.Start%160 != 0 {
			t.Errorf("a triplet stroke moved off the triplets: %d", n.Start)
		}
	}
	for _, n := range out.Tracks[0].Notes {
		if pos := n.Start % 480; pos != 0 && pos != 320 {
			t.Errorf("a shuffled chord hit is off the triplets: %d", n.Start)
		}
	}
}

// TestShortChordsSound (#404 S2): a chord shorter than the pattern still
// sounds all its notes, struck together on its first beat, and the notes
// say so; a chord the pattern covers is not struck twice.
func TestShortChordsSound(t *testing.T) {
	tl := soundAuditTimeline(t, 1, "Imaj7", "vi7", "ii7", "V7")
	for _, acc := range []string{"ballad", "arpeggio", "pulse", "strum"} {
		out, err := Render(tl, Choice{Accompaniment: acc})
		if err != nil {
			t.Fatal(err)
		}
		for i, h := range out.Harmony {
			got := classesOver(out, i)
			for _, pc := range h.Chord.PitchClasses() {
				if !got[pc.Normalize()] {
					t.Errorf("%s over a one-beat %s lacks %v: %v", acc, h.RomanNumeral, pc, got)
				}
			}
		}
	}
	out, err := Render(tl, Choice{Accompaniment: "ballad"})
	if err != nil {
		t.Fatal(err)
	}
	if e := out.Rhythm.Accompaniment.Explanation; !strings.Contains(e, "型より短い和音") {
		t.Errorf("the notes do not say the short chords were struck: %s", e)
	}
	full, err := Render(soundAuditTimeline(t, 4, "Imaj7", "V7"), Choice{Accompaniment: "ballad"})
	if err != nil {
		t.Fatal(err)
	}
	if e := full.Rhythm.Accompaniment.Explanation; strings.Contains(e, "型より短い和音") {
		t.Errorf("whole-bar chords are said to be short: %s", e)
	}
}

// TestWalkKeepsWalking (#404 S7): over a two-bar chord the walking bass
// goes up the chord and back down instead of repeating its top note, and
// stays under the chords (C4).
func TestWalkKeepsWalking(t *testing.T) {
	tl := soundAuditTimeline(t, 8, "Imaj7", "ii7")
	for i := range tl.Harmony {
		line := walkLine(tl.Harmony, i, 8)
		for k, p := range line {
			if p >= 60 {
				t.Errorf("the walk under %s reaches the chords: %v", tl.Harmony[i].RomanNumeral, line)
			}
			if k >= 2 && p == line[k-1] && p == line[k-2] {
				t.Errorf("the walk under %s repeats a note three times: %v", tl.Harmony[i].RomanNumeral, line)
			}
		}
	}
}

// TestWaltzFillSaysWhatItPlays (#404 S9): 3/4's fill has two strokes,
// the snare and the low tom, and its notes do not say it goes round the
// toms; 4/4's fill does go round them and says so.
func TestWaltzFillSaysWhatItPlays(t *testing.T) {
	waltz := soundAuditTimelineIn(t, music.ThreeFour, 3, "I", "IV", "V", "I")
	out, err := Render(waltz, Choice{Accompaniment: "waltz", Drums: "waltz"})
	if err != nil {
		t.Fatal(err)
	}
	if e := out.Rhythm.Drums.Explanation; strings.Contains(e, "高い方から低い方へ回す") || !strings.Contains(e, "スネアから低いタムへ") {
		t.Errorf("3/4 fill text does not match its two strokes: %s", e)
	}
	rock := soundAuditTimeline(t, 4, "I", "IV", "V", "I")
	out, err = Render(rock, Choice{Accompaniment: "pulse", Drums: "eight-beat"})
	if err != nil {
		t.Fatal(err)
	}
	if e := out.Rhythm.Drums.Explanation; !strings.Contains(e, "高い方から低い方へ回す") {
		t.Errorf("4/4 fill text lost its tom round: %s", e)
	}
}
