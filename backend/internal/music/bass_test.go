package music

import (
	"fmt"
	"testing"
)

// TestSlashBassLabels (BL-113, ADR-0020): a label ends with "/" and a
// scale degree (with b or #) to put that note in the bass; "/" followed by
// a roman numeral stays a secondary dominant.
func TestSlashBassLabels(t *testing.T) {
	tests := []struct {
		key     Key
		label   string
		numeral string
		symbol  string
		bass    PitchClass
	}{
		{Key{C, Major}, "I/3", "I/3", "C/E", E},
		{Key{C, Major}, "I/7", "I/7", "C/B", B},
		{Key{C, Major}, "IV/1", "IV/1", "F/C", C},
		{Key{C, Major}, "V7/V/#4", "V7/V/#4", "D7/F#", Fs}, // V7/V in first inversion: the bass is the key's raised 4th
		{Key{A, Minor}, "i/#7", "i/#7", "Am/G#", Gs},
		{Key{A, Minor}, "i/7", "i/7", "Am/G", G},
		{Key{A, Minor}, "i/#6", "i/#6", "Am/F#", Fs},
		{Key{F, Major}, "I/b7", "I/b7", "F/Eb", Ds},
	}
	for _, tt := range tests {
		t.Run(tt.key.String()+" "+tt.label, func(t *testing.T) {
			step, err := StepFromLabel(tt.key, tt.label, 4)
			if err != nil {
				t.Fatal(err)
			}
			tl, err := FromProgression(tt.key, Progression{step}, 100)
			if err != nil {
				t.Fatal(err)
			}
			h := tl.Harmony[0]
			if h.RomanNumeral != tt.numeral || h.Symbol != tt.symbol || h.Bass == nil || *h.Bass != tt.bass {
				t.Errorf("got %s %s bass %v, want %s %s bass %v", h.RomanNumeral, h.Symbol, h.Bass, tt.numeral, tt.symbol, tt.bass)
			}
			// The bass sounds below the chord, an octave under middle C's octave.
			low := tl.Tracks[0].Notes[0]
			for _, n := range tl.Tracks[0].Notes {
				if n.Pitch < low.Pitch {
					low = n
				}
			}
			if low.Pitch != NewPitch(tt.bass, ChordOctave-1) {
				t.Errorf("lowest note %v, want the bass %v", low.Pitch, NewPitch(tt.bass, ChordOctave-1))
			}
			canon, err := CanonicalLabel(tt.key, tt.label)
			if err != nil || canon != tt.numeral {
				t.Errorf("CanonicalLabel = %q, %v; want %q", canon, err, tt.numeral)
			}
		})
	}
}

func TestSlashBassRejectsBadDegrees(t *testing.T) {
	for _, label := range []string{"I/8", "I/0", "I/x3", "I/"} {
		if _, err := StepFromLabel(Key{C, Major}, label, 4); err == nil {
			t.Errorf("%q should be rejected", label)
		}
	}
	// A chord without a slash has no bass of its own.
	step, err := StepFromLabel(Key{C, Major}, "IV", 4)
	if err != nil || step.Bass != nil {
		t.Errorf("IV: bass %v, err %v", step.Bass, err)
	}
	tl, err := FromProgression(Key{C, Major}, Progression{step}, 100)
	if err != nil || tl.Harmony[0].Bass != nil || fmt.Sprint(len(tl.Tracks[0].Notes)) != "3" {
		t.Errorf("a plain IV stays three notes with no bass: %+v", tl.Harmony[0])
	}
}
