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
		// Secondary dominants (BL-074); each exists only in some modes.
		"V/ii", "V7/ii", "V/iii", "V7/iii", "V7/IV", "V/V", "V7/V", "V/vi", "V7/vi",
		"V/III", "V7/iv", "V/v", "V7/VI", "V/VII", "V7/VII", "V/IV",
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

// TestEngineLabelsPassTheirOwnCheck (review): every numeral the engine
// writes on a timeline passes CanonicalLabel, so an API that rejects
// labels the key would change (ADR-0013) still takes back its own output
// on a key or style change.
func TestEngineLabelsPassTheirOwnCheck(t *testing.T) {
	labels := []string{
		"I", "ii", "iii", "IV", "V", "vi", "vii°", "i", "ii°", "III", "iv", "v", "VI", "VII",
		"Imaj7", "ii7", "iii7", "IVmaj7", "V7", "vi7", "viiø7", "iiø7", "i7", "III+",
		"V/ii", "V7/ii", "V/iii", "V7/IV", "V/V", "V7/V", "V/vi", "V7/vi", "V/III", "V7/iv", "V/VII", "V7/VI",
		"iv", "ivm6", "bVI", "bVII", "bIII", "bII", "♭VII", "II7", "I/3", "I/5", "V/7", "IV/1", "i/#7", "i/7", "V7/V/#4", "vi/5",
	}
	for tonic := range PitchClass(12) {
		for _, mode := range []Mode{Major, Minor, Dorian, Mixolydian} {
			key := Key{Tonic: tonic, Mode: mode}
			for _, label := range labels {
				step, err := StepFromLabel(key, label, 4)
				if err != nil {
					continue
				}
				if _, err := CanonicalLabel(key, label); err != nil {
					continue // the check refuses what the model wrote; the engine never writes it
				}
				tl, err := FromProgression(key, Progression{step}, 100)
				if err != nil {
					t.Fatal(err)
				}
				written := tl.Harmony[0].RomanNumeral
				if _, err := CanonicalLabel(key, written); err != nil {
					t.Errorf("%v: %q is written as %q, which its own check refuses: %v", key, label, written, err)
				}
			}
		}
	}
}
