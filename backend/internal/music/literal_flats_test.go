package music

import (
	"strings"
	"testing"
)

// #109: outside the major key's borrowed set, a flat that names one of the
// key's own chords by its distance from the tonic is accepted and respelled;
// a flat that names a chord outside the key still fails.
func TestFlatsNamingTheKeysOwnChordAreRespelled(t *testing.T) {
	ok := []struct {
		key   Key
		label string
		want  string
		chord Chord
	}{
		{Key{G, Mixolydian}, "bVII", "VII", Chord{Root: F, Quality: MajorTriad}},
		{Key{A, Mixolydian}, "bVII", "VII", Chord{Root: G, Quality: MajorTriad}},
		{Key{E, Minor}, "bVI", "VI", Chord{Root: C, Quality: MajorTriad}},
		{Key{D, Minor}, "bVI", "VI", Chord{Root: As, Quality: MajorTriad}},
		{Key{G, Minor}, "bVImaj7", "VImaj7", Chord{Root: Ds, Quality: Major7}},
		{Key{C, Minor}, "bVII7", "VII7", Chord{Root: As, Quality: Dominant7}},
		{Key{C, Dorian}, "bIII", "III", Chord{Root: Ds, Quality: MajorTriad}},
	}
	for _, tt := range ok {
		got, err := CanonicalLabel(tt.key, tt.label)
		if err != nil || got != tt.want {
			t.Errorf("%s %q: CanonicalLabel = %q, %v; want %q", tt.key, tt.label, got, err, tt.want)
			continue
		}
		step, err := StepFromLabel(tt.key, tt.label, 4)
		if err != nil {
			t.Fatal(err)
		}
		tl, err := FromProgression(tt.key, Progression{step}, 100)
		if err != nil {
			t.Fatal(err)
		}
		if tl.Harmony[0].Chord != tt.chord {
			t.Errorf("%s %q sounds as %v, want %v", tt.key, tt.label, tl.Harmony[0].Chord, tt.chord)
		}
	}
	bad := []struct {
		key   Key
		label string
	}{
		{Key{E, Minor}, "bV"},         // Bb: not in E minor, not borrowed
		{Key{C, Mixolydian}, "bVII7"}, // Bb7: mixolydian's VII7 is Bbmaj7
		{Key{C, Major}, "bIII"},       // Eb: not in C major, not in the borrowed set
		{Key{C, Dorian}, "bVI"},       // Ab: dorian has A
	}
	for _, tt := range bad {
		if _, err := CanonicalLabel(tt.key, tt.label); err == nil || !strings.Contains(err.Error(), "not available") {
			t.Errorf("%s %q: want \"not available\", got %v", tt.key, tt.label, err)
		}
	}
}
