package music

import (
	"encoding/json"
	"strings"
	"testing"
)

func validTimeline() Timeline {
	tl := NewTimeline(Key{C, Major}, 120)
	tl.Tracks = []Track{{
		Name:    "piano",
		Channel: 0,
		Program: 0,
		Notes: []NoteEvent{
			{Pitch: 60, Velocity: 100, Start: 0, Duration: DefaultPPQ},
			{Pitch: 64, Velocity: 100, Start: DefaultPPQ, Duration: DefaultPPQ * 2},
		},
	}}
	tl.Harmony = []HarmonyEvent{
		{Start: 0, Duration: DefaultPPQ * 4, Chord: Chord{C, MajorTriad}, RomanNumeral: "I"},
	}
	return tl
}

func TestTimelineDefaultsAndLength(t *testing.T) {
	tl := validTimeline()
	if tl.PPQ != DefaultPPQ || tl.TimeSignature != (TimeSignature{4, 4}) {
		t.Fatalf("unexpected defaults %+v", tl)
	}
	if err := tl.Validate(); err != nil {
		t.Fatalf("valid timeline rejected: %v", err)
	}
	if got := tl.LengthTicks(); got != DefaultPPQ*4 {
		t.Fatalf("LengthTicks = %d, want %d", got, DefaultPPQ*4)
	}
	if got := (Timeline{}).LengthTicks(); got != 0 {
		t.Fatalf("empty LengthTicks = %d", got)
	}
}

func TestTimelineValidate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Timeline)
		want   string
	}{
		{"zero ppq", func(tl *Timeline) { tl.PPQ = 0 }, "ppq must be 1..32767"},
		{"zero tempo", func(tl *Timeline) { tl.TempoBPM = 0 }, "tempoBpm must be 10..1000"},
		{"bad denominator", func(tl *Timeline) { tl.TimeSignature.Denominator = 3 }, "invalid time signature"},
		{"bad channel", func(tl *Timeline) { tl.Tracks[0].Channel = 16 }, "channel must be 0..15"},
		{"bad program", func(tl *Timeline) { tl.Tracks[0].Program = 128 }, "program must be 0..127"},
		{"bad pitch", func(tl *Timeline) { tl.Tracks[0].Notes[0].Pitch = 200 }, "pitch out of range"},
		{"bad velocity", func(tl *Timeline) { tl.Tracks[0].Notes[0].Velocity = 0 }, "velocity must be 1..127"},
		{"negative start", func(tl *Timeline) { tl.Tracks[0].Notes[0].Start = -1 }, "start must be >= 0"},
		{"zero duration", func(tl *Timeline) { tl.Tracks[0].Notes[1].Duration = 0 }, "duration must be > 0"},
		{"bad harmony span", func(tl *Timeline) { tl.Harmony[0].Duration = 0 }, "harmony[0]: invalid span"},
		{"bad harmony quality", func(tl *Timeline) { tl.Harmony[0].Chord.Quality = 99 }, "unknown chord quality"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := validTimeline()
			tt.mutate(&tl)
			err := tl.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestTimelineValidateCollectsAllErrors(t *testing.T) {
	tl := validTimeline()
	tl.PPQ = 0
	tl.TempoBPM = 0
	err := tl.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"ppq", "tempoBpm"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

func TestTimelineJSONRoundTrip(t *testing.T) {
	tl := validTimeline()
	b, err := json.Marshal(tl)
	if err != nil {
		t.Fatal(err)
	}
	var back Timeline
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Validate() != nil || back.LengthTicks() != tl.LengthTicks() || back.Key != tl.Key {
		t.Fatalf("round trip mismatch: %+v", back)
	}
}

// TestSoundingSymbol (BL-307): the tensions that sound go in brackets
// before any slash bass, as the UI writes them.
func TestSoundingSymbol(t *testing.T) {
	tests := []struct {
		symbol   string
		tensions []string
		want     string
	}{
		{"Fmaj7", nil, "Fmaj7"},
		{"Fmaj7", []string{"9"}, "Fmaj7(9)"},
		{"C7", []string{"9", "13"}, "C7(9,13)"},
		{"F", []string{"add9"}, "F(add9)"},
		{"Cmaj7/E", []string{"9"}, "Cmaj7(9)/E"},
	}
	for _, tt := range tests {
		if got := (HarmonyEvent{Symbol: tt.symbol, Tensions: tt.tensions}).SoundingSymbol(); got != tt.want {
			t.Errorf("%s %v: got %q, want %q", tt.symbol, tt.tensions, got, tt.want)
		}
	}
}
