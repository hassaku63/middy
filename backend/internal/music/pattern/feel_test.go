package pattern

import (
	"errors"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func TestSwingDelaysOffBeatEighths(t *testing.T) {
	// A swung pair of eighths becomes a 2:1 triplet pair: the "and" moves
	// from half way (240) to two thirds of the beat (320).
	tests := []struct {
		name            string
		start, dur      music.Tick
		wantStart, want music.Tick
	}{
		{"on-beat eighth stretches to the swung position", 0, eighth, 0, 320},
		{"off-beat eighth starts late and keeps to its beat", eighth, eighth, 320, 160},
		{"off-beat in a later beat", 3*quarter + eighth, eighth, 3*quarter + 320, 160},
		{"quarter notes are untouched", quarter, quarter, quarter, quarter},
		{"long off-beat note keeps its end", eighth, 3 * eighth, 320, 3*eighth - 80},
		{"sixteenths are untouched", quarter / 4, quarter / 4, quarter / 4, quarter / 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, dur := swing(tt.start, tt.dur)
			if start != tt.wantStart || dur != tt.want {
				t.Fatalf("swing(%d, %d) = (%d, %d), want (%d, %d)", tt.start, tt.dur, start, dur, tt.wantStart, tt.want)
			}
		})
	}
}

func TestRenderAppliesFeelToEveryTrack(t *testing.T) {
	tl := fourBars(t)
	straight, err := Render(tl, Choice{Accompaniment: "arpeggio", Drums: "eight-beat"})
	if err != nil {
		t.Fatal(err)
	}
	swung, err := Render(tl, Choice{Accompaniment: "arpeggio", Drums: "eight-beat", Feel: FeelSwing})
	if err != nil {
		t.Fatal(err)
	}
	for i, track := range swung.Tracks {
		moved := 0
		for j, n := range track.Notes {
			s := straight.Tracks[i].Notes[j]
			if n.Pitch != s.Pitch {
				t.Fatalf("track %s note %d: swing must not change pitches", track.Name, j)
			}
			if n.Start != s.Start {
				moved++
				if s.Start%quarter != eighth {
					t.Fatalf("track %s: only off-beat eighths move, %d moved", track.Name, s.Start)
				}
			}
		}
		if moved == 0 {
			t.Fatalf("track %s: nothing swung", track.Name)
		}
	}
	if swung.Rhythm == nil || swung.Rhythm.Feel == nil || swung.Rhythm.Feel.ID != "swing" || swung.Rhythm.Feel.Explanation == "" {
		t.Fatalf("the explanation must name the feel: %+v", swung.Rhythm)
	}
	if straight.Rhythm.Feel != nil {
		t.Fatal("straight is the default and needs no explanation")
	}
	// The placement counts describe the template, so swing does not turn
	// "and" into "finer".
	if swung.Rhythm.Accompaniment.OffBeat != straight.Rhythm.Accompaniment.OffBeat {
		t.Fatal("swing must not change the beat-placement analysis")
	}
}

func TestRenderChoices(t *testing.T) {
	tl := fourBars(t)
	same, err := Render(tl, Choice{})
	if err != nil || len(same.Tracks) != len(tl.Tracks) || same.Rhythm != nil {
		t.Fatalf("no choice leaves the block chords untouched: %v", err)
	}
	drumsOnly, err := Render(tl, Choice{Drums: "ballad"})
	if err != nil || drumsOnly.Rhythm.Accompaniment.ID != "block" {
		t.Fatalf("drums alone keep block chords: %v", err)
	}
	if _, err := Render(tl, Choice{Accompaniment: "polka"}); !errors.Is(err, ErrUnknownPattern) {
		t.Fatalf("unknown ids are errors, got %v", err)
	}
	if _, err := Render(tl, Choice{Accompaniment: "block", Feel: "shuffle"}); !errors.Is(err, ErrUnknownPattern) {
		t.Fatalf("unknown feels are errors, got %v", err)
	}
}

func fourBars(t *testing.T) music.Timeline {
	t.Helper()
	return progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 4, 5, 1)
}
