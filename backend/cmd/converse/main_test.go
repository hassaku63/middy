package main

import (
	"testing"

	"github.com/hassaku63/middy/backend/internal/ideas"
	"github.com/hassaku63/middy/backend/internal/intent"
	"github.com/hassaku63/middy/backend/internal/music"
)

// TestMeasure (#453): bars, sus4, a sounding ninth and a claimed one.
func TestMeasure(t *testing.T) {
	f := music.Chord{Root: music.F, Quality: music.MajorTriad}
	tl := music.Timeline{
		PPQ:           480,
		TimeSignature: music.FourFour,
		Harmony: []music.HarmonyEvent{
			{Start: 0, Duration: 1920, Chord: f},
			{Start: 1920, Duration: 1920 + 960, Chord: f},
		},
		Tracks: []music.Track{{Name: "chords", Notes: []music.NoteEvent{
			{Pitch: 65, Start: 0, Duration: 1920},
			{Pitch: 67, Start: 0, Duration: 960},    // G, the add9, inside the first chord
			{Pitch: 67, Start: 1680, Duration: 480}, // pushed across the line: not the first chord's
		}}},
	}
	c := ideas.Candidate{Candidate: intent.Candidate{Progression: []string{"I", "Isus4"}, Reason: "9 度で浮遊感"}, Timeline: tl}
	m := measure(c, "")
	if m.Bars != "2.50" || !m.Sus4 || m.Tensioned != 1 || m.Chords != 2 || !m.Claims {
		t.Fatalf("got %+v", m)
	}
}
