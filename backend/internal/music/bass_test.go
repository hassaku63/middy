package music

import (
	"fmt"
	"strings"
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

// TestSlashBassExplanations (BL-114): a chord tone in the bass is an
// inversion, a note outside the chord kept from a neighbour is a pedal, and
// any other outside note passes between the basses around it. The flow
// names a stepwise bass line, and a line cliché when the chord stays.
func TestSlashBassExplanations(t *testing.T) {
	tests := []struct {
		key    Key
		labels []string
		at     int
		chord  string // in the chord's explanation
		flow   string // in the overview's flow ("" to skip)
	}{
		{Key{C, Major}, []string{"I/3", "IV", "V", "I"}, 0, "3 度の E がベース（第 1 転回形）", ""},
		{Key{C, Major}, []string{"I/5", "V", "I"}, 0, "5 度の G がベース（第 2 転回形）", ""},
		{Key{C, Major}, []string{"I", "IV/1", "I"}, 1, "5 度の C がベース（第 2 転回形）", ""}, // the pedal six-four
		{Key{C, Major}, []string{"I", "IV/1", "I"}, 1, "ペダルとしても働き", ""},
		{Key{C, Major}, []string{"I", "ii/1", "I"}, 1, "この和音にない音で、隣の和音から同じベースが続くペダル", ""},
		{Key{C, Major}, []string{"I", "I/7", "vi", "V"}, 1, "経過", "ベースが C → B → A → G と順に下がる"},
		{Key{A, Minor}, []string{"i", "i/#7", "i/7", "i/#6"}, 2, "経過", "ライン・クリシェ"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.labels, " "), func(t *testing.T) {
			tl := analysed(t, tt.key, 4, tt.labels...)
			if e := tl.Harmony[tt.at].Analysis.Explanation; !strings.Contains(e, tt.chord) {
				t.Errorf("chord %d explanation lacks %q: %s", tt.at, tt.chord, e)
			}
			if tt.flow != "" && !strings.Contains(tl.Overview.Flow, tt.flow) {
				t.Errorf("flow lacks %q: %s", tt.flow, tl.Overview.Flow)
			}
		})
	}
	// A chord without a slash says nothing about its bass, and a plain
	// stepwise progression of roots is not called a line.
	tl := analysed(t, Key{C, Major}, 4, "I", "ii", "iii", "IV")
	if strings.Contains(tl.Harmony[0].Analysis.Explanation, "ベース") || strings.Contains(tl.Overview.Flow, "順に") {
		t.Errorf("no bass talk without a slash chord: %s / %s", tl.Harmony[0].Analysis.Explanation, tl.Overview.Flow)
	}
}

// TestSlashBassIsNotASecondaryDominant: the "/" of a slash bass is not
// the "/" of V/x; C/B in C major is still the tonic I.
func TestSlashBassIsNotASecondaryDominant(t *testing.T) {
	tl := analysed(t, Key{C, Major}, 4, "I", "I/7", "IV", "I/3", "V7/V/#4", "V")
	for _, i := range []int{1, 3} {
		a := tl.Harmony[i].Analysis
		if a.AppliedTo != 0 || a.Function != Tonic || strings.Contains(a.Explanation, "セカンダリードミナント") {
			t.Errorf("%s should be the tonic, not applied: %+v", tl.Harmony[i].RomanNumeral, a)
		}
	}
	if a := tl.Harmony[4].Analysis; a.AppliedTo != 5 {
		t.Errorf("V7/V/#4 is still the dominant of V: %+v", a)
	}
}
