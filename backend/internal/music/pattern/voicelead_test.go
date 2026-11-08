package pattern

import (
	"slices"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func progTimeline(t *testing.T, steps ...music.Step) music.Timeline {
	t.Helper()
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, music.Progression(steps), 120)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

// chordsAt groups the chord track's notes above the bass by start tick.
func chordsAt(tl music.Timeline, minPitch music.Pitch) map[music.Tick][]music.Pitch {
	out := map[music.Tick][]music.Pitch{}
	for _, n := range tl.Tracks[0].Notes {
		if n.Pitch >= minPitch {
			out[n.Start] = append(out[n.Start], n.Pitch)
		}
	}
	for k := range out {
		slices.Sort(out[k])
	}
	return out
}

// TestVoiceLeading (BL-086): each chord is voiced from the one before with
// the least movement, so in ii7–V7–Imaj7 every upper voice moves by a
// semitone or a whole tone, and the first chord is the close root position.
func TestVoiceLeading(t *testing.T) {
	tl := progTimeline(t,
		music.Step{Degree: 2, Seventh: true, Beats: 4},
		music.Step{Degree: 5, Seventh: true, Beats: 4},
		music.Step{Degree: 1, Seventh: true, Beats: 4},
	)
	out, err := Render(tl, Choice{Accompaniment: "block"})
	if err != nil {
		t.Fatal(err)
	}
	chords := chordsAt(out, 52)
	bar := music.Tick(4 * music.DefaultPPQ)
	first := chords[0]
	if !slices.Equal(first, []music.Pitch{62, 65, 69, 72}) {
		t.Fatalf("ii7 starts in close root position, got %v", first)
	}
	prev := first
	for i := 1; i < 3; i++ {
		cur := chords[music.Tick(i)*bar]
		if len(cur) != len(prev) {
			t.Fatalf("chord %d has %d voices, want %d", i, len(cur), len(prev))
		}
		for v := range cur {
			if d := int(cur[v]) - int(prev[v]); d < -2 || d > 2 {
				t.Errorf("chord %d voice %d moves %d semitones (%v → %v)", i, v, d, prev, cur)
			}
		}
		prev = cur
	}
}

// TestVoiceLeadingStaysInRange (BL-086): a long cycle of fifths does not
// drift: the lowest chord tone stays between E3 and G4.
func TestVoiceLeadingStaysInRange(t *testing.T) {
	var steps []music.Step
	for range 4 {
		for _, d := range []int{1, 4, 7, 3, 6, 2, 5} {
			steps = append(steps, music.Step{Degree: d, Seventh: true, Beats: 4})
		}
	}
	out, err := Render(progTimeline(t, steps...), Choice{Accompaniment: "block"})
	if err != nil {
		t.Fatal(err)
	}
	for at, c := range chordsAt(out, 52) {
		if c[0] < 52 || c[0] > 67 {
			t.Errorf("chord at %d starts on %d, outside 52..67", at, c[0])
		}
	}
}

// TestWalkingBass (BL-086): the walking bass plays the root on the
// chord's first beat, chord tones after it, and enters the next root a
// semitone below on the chord's last beat, also when chords last two beats.
func TestWalkingBass(t *testing.T) {
	tl := progTimeline(t,
		music.Step{Degree: 2, Seventh: true, Beats: 2},
		music.Step{Degree: 5, Seventh: true, Beats: 2},
		music.Step{Degree: 1, Seventh: true, Beats: 4},
	)
	out, err := Render(tl, Choice{Accompaniment: "walking"})
	if err != nil {
		t.Fatal(err)
	}
	bass := map[music.Tick]music.Pitch{}
	for _, n := range out.Tracks[0].Notes {
		if n.Duration >= music.DefaultPPQ/2 { // the walk; the chord stabs are short
			bass[n.Start] = n.Pitch
		}
	}
	q := music.Tick(music.DefaultPPQ)
	want := map[music.Tick]music.Pitch{
		0: 50, q: 54, // D, then F# into G
		2 * q: 55, 3 * q: 47, // G, then B into C
		4 * q: 48, 5 * q: 52, 6 * q: 55, 7 * q: 49, // C E G, then C# into the D it loops back to
	}
	for at, p := range want {
		if bass[at] != p {
			t.Errorf("bass at %d = %d, want %d (all: %v)", at, bass[at], p, bass)
		}
	}
}
