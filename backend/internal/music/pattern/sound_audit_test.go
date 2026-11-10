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
