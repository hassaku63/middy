package music

import (
	"slices"
	"strings"
	"testing"
)

func TestFromProgressionCMajorPopProgression(t *testing.T) {
	prog := Progression{
		{Degree: 1, Beats: 4},
		{Degree: 5, Beats: 4},
		{Degree: 6, Beats: 4},
		{Degree: 4, Beats: 4},
	}
	tl, err := FromProgression(Key{C, Major}, prog, 90)
	if err != nil {
		t.Fatalf("FromProgression: %v", err)
	}
	if err := tl.Validate(); err != nil {
		t.Fatalf("generated timeline is invalid: %v", err)
	}
	if tl.TempoBPM != 90 || tl.Key != (Key{C, Major}) || tl.PPQ != DefaultPPQ {
		t.Fatalf("unexpected header %+v", tl)
	}

	// Four whole-note chords in 4/4 = four bars.
	if got, want := tl.LengthTicks(), Tick(4*4*DefaultPPQ); got != want {
		t.Fatalf("LengthTicks = %d, want %d", got, want)
	}

	wantHarmony := []struct {
		numeral string
		chord   string
		start   Tick
	}{
		{"I", "C", 0},
		{"V", "G", 4 * DefaultPPQ},
		{"vi", "Am", 8 * DefaultPPQ},
		{"IV", "F", 12 * DefaultPPQ},
	}
	if len(tl.Harmony) != len(wantHarmony) {
		t.Fatalf("Harmony has %d events, want %d", len(tl.Harmony), len(wantHarmony))
	}
	for i, w := range wantHarmony {
		h := tl.Harmony[i]
		if h.RomanNumeral != w.numeral || h.Chord.String() != w.chord || h.Start != w.start || h.Duration != 4*DefaultPPQ {
			t.Errorf("harmony[%d] = %+v (%s), want %s %s at %d", i, h, h.Chord, w.numeral, w.chord, w.start)
		}
	}

	if len(tl.Tracks) != 1 {
		t.Fatalf("Tracks = %d, want 1", len(tl.Tracks))
	}
	tr := tl.Tracks[0]
	if tr.Channel != 0 || tr.Program != 0 {
		t.Fatalf("track should be a GM piano on channel 0, got %+v", tr)
	}
	// Block chords: every chord tone starts with the chord and lasts the whole step.
	if len(tr.Notes) != 4*3 {
		t.Fatalf("Notes = %d, want 12 (4 triads)", len(tr.Notes))
	}
	first := tr.Notes[:3]
	wantPitches := []Pitch{60, 64, 67} // C4 E4 G4
	for i, n := range first {
		if n.Pitch != wantPitches[i] || n.Start != 0 || n.Duration != 4*DefaultPPQ || n.Velocity != ChordVelocity {
			t.Errorf("note[%d] = %+v, want pitch %d start 0 duration %d", i, n, wantPitches[i], 4*DefaultPPQ)
		}
	}
	// The second chord starts at bar 2, voiced from the first (BL-086):
	// G under C E G is B3 D4 G4, the inversion that moves least.
	var second []Pitch
	for _, n := range tr.Notes[3:6] {
		if n.Start != 4*DefaultPPQ {
			t.Errorf("second chord should start at bar 2, got %+v", n)
		}
		second = append(second, n.Pitch)
	}
	if !slices.Equal(second, []Pitch{59, 62, 67}) {
		t.Errorf("second chord = %v, want B3 D4 G4", second)
	}
}

func TestFromProgressionMinorAndSevenths(t *testing.T) {
	prog := Progression{
		{Degree: 1, Beats: 2},
		{Degree: 4, Beats: 2},
		{Degree: 5, Beats: 2, Seventh: true},
		{Degree: 1, Beats: 2},
	}
	tl, err := FromProgression(Key{A, Minor}, prog, 120)
	if err != nil {
		t.Fatalf("FromProgression: %v", err)
	}
	if err := tl.Validate(); err != nil {
		t.Fatalf("generated timeline is invalid: %v", err)
	}

	var numerals, chords []string
	for _, h := range tl.Harmony {
		numerals = append(numerals, h.RomanNumeral)
		chords = append(chords, h.Chord.String())
	}
	if got := strings.Join(numerals, " "); got != "i iv v7 i" {
		t.Errorf("numerals = %q", got)
	}
	if got := strings.Join(chords, " "); got != "Am Dm Em7 Am" {
		t.Errorf("chords = %q", got)
	}
	// 3 + 3 + 4 + 3 chord tones.
	if n := len(tl.Tracks[0].Notes); n != 13 {
		t.Errorf("Notes = %d, want 13", n)
	}
	if got, want := tl.LengthTicks(), Tick(8*DefaultPPQ); got != want {
		t.Errorf("LengthTicks = %d, want %d", got, want)
	}
}

func TestFromProgressionRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		key   Key
		prog  Progression
		tempo float64
		want  string
	}{
		{"empty", Key{C, Major}, Progression{}, 120, "at least one step"},
		{"degree zero", Key{C, Major}, Progression{{Degree: 0, Beats: 4}}, 120, "steps[0]"},
		{"degree eight", Key{C, Major}, Progression{{Degree: 8, Beats: 4}}, 120, "steps[0]"},
		{"zero beats", Key{C, Major}, Progression{{Degree: 1, Beats: 0}}, 120, "beats must be >= 1"},
		{"second bad step", Key{C, Major}, Progression{{Degree: 1, Beats: 4}, {Degree: 2, Beats: -1}}, 120, "steps[1]"},
		{"zero tempo", Key{C, Major}, Progression{{Degree: 1, Beats: 4}}, 0, "tempoBpm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromProgression(tt.key, tt.prog, tt.tempo)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestProgressionValidateCollectsAllErrors(t *testing.T) {
	err := Progression{{Degree: 0, Beats: 0}, {Degree: 9, Beats: 1}}.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"steps[0]: degree", "steps[0]: beats", "steps[1]: degree"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}
