package pattern

import (
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestPushesComeBeforeTheFeel (BL-085): a turnaround of two beats a chord
// with V7 pushed an eighth lands on the "and" of beat 2, and swing then
// delays that off-beat like any other.
func TestPushesComeBeforeTheFeel(t *testing.T) {
	prog := music.Progression{
		{Degree: 2, Seventh: true, Beats: 2},
		{Degree: 5, Seventh: true, Beats: 2, Push: music.PushEighth},
		{Degree: 1, Seventh: true, Beats: 4},
	}
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, prog, 120)
	if err != nil {
		t.Fatal(err)
	}
	at := func(tl music.Timeline, tick music.Tick) bool {
		for _, n := range tl.Tracks[0].Notes {
			if n.Start == tick {
				return true
			}
		}
		return false
	}
	straight, err := Render(tl, Choice{Accompaniment: "ballad"})
	if err != nil {
		t.Fatal(err)
	}
	if !at(straight, 720) || at(straight, 960) {
		t.Errorf("V7 should be struck at 720, not 960")
	}
	swung, err := Render(tl, Choice{Accompaniment: "ballad", Feel: FeelSwing, Swing: Swing{Ratio: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !at(swung, 800) { // 480 + 2/3 of a beat
		t.Errorf("the pushed eighth should swing to 800")
	}
}
