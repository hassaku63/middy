package music

import "testing"

// The label a timeline shows for each chord must read back as the same
// chord in the same key. Changing an example's key (BL-078) rebuilds it
// from those labels, so a lossy label would change the music.
func TestHarmonyLabelsReadBackAsTheSameChords(t *testing.T) {
	labels := []string{
		"I", "ii", "iii", "IV", "V", "vi", "vii°",
		"i", "ii°", "III", "iv", "v", "VI", "VII",
		"Imaj7", "ii7", "iii7", "IVmaj7", "V7", "vi7", "viiø7",
	}
	for _, mode := range []Mode{Major, Minor, Dorian, Mixolydian} {
		key := Key{Tonic: A, Mode: mode}
		for _, label := range labels {
			first, err := StepFromLabel(key, label, 4)
			if err != nil {
				continue // not every label exists in every mode
			}
			tl, err := FromProgression(key, Progression{first}, 100)
			if err != nil {
				t.Fatalf("%s %s: %v", mode, label, err)
			}
			shown := tl.Harmony[0]
			again, err := StepFromLabel(key, shown.RomanNumeral, 4)
			if err != nil {
				t.Fatalf("%s: %q shows as %q, which does not parse: %v", mode, label, shown.RomanNumeral, err)
			}
			tl2, err := FromProgression(key, Progression{again}, 100)
			if err != nil {
				t.Fatal(err)
			}
			if tl2.Harmony[0].Chord != shown.Chord {
				t.Errorf("%s: %q shows as %q, which reads back as %v, not %v",
					mode, label, shown.RomanNumeral, tl2.Harmony[0].Chord, shown.Chord)
			}
		}
	}
}
