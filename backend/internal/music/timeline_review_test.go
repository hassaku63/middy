package music

import (
	"strings"
	"testing"
)

// TestValidateKeepsWhatAMIDIFileCanHold (review): a timeline from outside
// (a MIDI export) is checked against the bounds of the file it becomes,
// and its pushes are checked as the engine's are.
func TestValidateKeepsWhatAMIDIFileCanHold(t *testing.T) {
	ok, err := FromProgression(Key{Tonic: C, Mode: Major}, Progression{{Degree: 1, Beats: 4}, {Degree: 5, Beats: 4}}, 100)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		edit func(*Timeline)
		want string
	}{
		{"a tempo too slow for the file", func(tl *Timeline) { tl.TempoBPM = 1 }, "tempoBpm must be"},
		{"a meter a byte cannot hold", func(tl *Timeline) { tl.TimeSignature.Numerator = 300 }, "invalid time signature"},
		{"a division the file cannot hold", func(tl *Timeline) { tl.PPQ = 40000 }, "ppq must be"},
		{"an unknown push", func(tl *Timeline) { tl.Harmony[1].Push = "quarter" }, "unknown push"},
		{"a push before the start", func(tl *Timeline) { tl.Harmony[0].Push = PushEighth }, "before the start"},
	}
	for _, tt := range tests {
		tl := ok
		tl.Harmony = append([]HarmonyEvent(nil), ok.Harmony...)
		tt.edit(&tl)
		if err := tl.Validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: want %q, got %v", tt.name, tt.want, err)
		}
	}
	if err := ok.Validate(); err != nil {
		t.Errorf("an engine timeline is valid: %v", err)
	}
}
