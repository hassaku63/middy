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

// TestPushStaysTiedUnderSwing (review of BL-085/084): swing moves a pushed
// eighth to the swung "and", and the chord before it now ends there too,
// so no silence opens before the push.
func TestPushStaysTiedUnderSwing(t *testing.T) {
	prog := music.Progression{
		{Degree: 1, Beats: 4},
		{Degree: 5, Beats: 4, Push: music.PushEighth},
	}
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(tl, Choice{Accompaniment: "block", Feel: FeelSwing})
	if err != nil {
		t.Fatal(err)
	}
	var firstEnd, pushAt music.Tick
	for _, n := range out.Tracks[0].Notes {
		switch {
		case n.Start == 0:
			firstEnd = n.Start + n.Duration
		case pushAt == 0 || n.Start < pushAt:
			pushAt = n.Start
		}
	}
	if pushAt == 1680 || firstEnd != pushAt {
		t.Errorf("the push should swing and stay tied: first chord ends %d, push at %d", firstEnd, pushAt)
	}
}
