package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestJazzRideKeepsTimeOnTheRide (BL-081): the ride plays the jazz
// pattern (1, 2, and-of-2, 3, 4, and-of-4), the hi-hat pedal closes on 2
// and 4, and the kick only feathers the quarters; no snare backbeat.
func TestJazzRideKeepsTimeOnTheRide(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.F, Mode: music.Major}, 4, 2, 5)
	d, err := FindDrums("jazz-ride")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "comp"), &d)
	if err != nil {
		t.Fatal(err)
	}
	ride, pedal := map[music.Tick]bool{}, map[music.Tick]bool{}
	for _, n := range notesIn(got, "drums") {
		switch n.Pitch {
		case Ride:
			ride[n.Start%1920] = true
		case PedalHat:
			pedal[n.Start%1920] = true
		case Snare:
			t.Errorf("no snare backbeat in the jazz ride pattern: %+v", n)
		case Kick:
			if n.Velocity > 50 {
				t.Errorf("the kick only feathers: %+v", n)
			}
		}
	}
	for _, s := range []music.Tick{0, 480, 720, 960, 1440, 1680} {
		if !ride[s] {
			t.Errorf("ride should play at %d: %v", s, ride)
		}
	}
	if len(pedal) != 2 || !pedal[480] || !pedal[1440] {
		t.Errorf("hi-hat pedal on 2 and 4: %v", pedal)
	}
	if e := got.Rhythm.Drums.Explanation; !strings.Contains(e, "1 小節にキックが 4 回") || strings.Contains(e, "スネア") {
		t.Errorf("the explanation names only what plays (a kick, no snare): %s", e)
	}
	if !strings.Contains(got.Rhythm.Drums.Description, "ライド") {
		t.Errorf("the description names the ride: %s", got.Rhythm.Drums.Description)
	}
}

// TestCompPlaysTheCharleston (BL-081): bass on 1 and 3, short chords on 1
// and the and-of-2 (the Charleston rhythm) in every bar.
func TestCompPlaysTheCharleston(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.F, Mode: music.Major}, 4, 2)
	got, err := Apply(tl, mustAcc(t, "comp"), nil)
	if err != nil {
		t.Fatal(err)
	}
	chords := map[music.Tick]int{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		chords[n.Start]++
		if n.Start != 0 && n.Start != 960 && n.Duration > 240 {
			t.Errorf("comping chords are short: %+v", n)
		}
	}
	// bass alone at 960; bass + a 3-note chord at 0; a chord at 720.
	want := map[music.Tick]int{0: 4, 720: 3, 960: 1}
	for s, n := range want {
		if chords[s] != n {
			t.Errorf("notes at %d = %d, want %d (all: %v)", s, chords[s], n, chords)
		}
	}
}

// TestJazzWaltzRide (jazz waltz): in 3/4 the ride plays 1, 2, and-of-2
// and 3, the hi-hat pedal closes on 2, the kick feathers beat 1.
func TestJazzWaltzRide(t *testing.T) {
	tl := waltzTimeline(t, 3, 1, 5)
	d, err := FindDrums("jazz-waltz")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "waltz"), &d)
	if err != nil {
		t.Fatal(err)
	}
	ride, pedal := map[music.Tick]bool{}, map[music.Tick]bool{}
	for _, n := range notesIn(got, "drums") {
		switch n.Pitch {
		case Ride:
			ride[n.Start%1440] = true
		case PedalHat:
			pedal[n.Start%1440] = true
		case Snare:
			t.Errorf("no snare in the jazz waltz ride: %+v", n)
		}
	}
	for _, s := range []music.Tick{0, 480, 720, 960} {
		if !ride[s] {
			t.Errorf("ride should play at %d: %v", s, ride)
		}
	}
	if len(pedal) != 1 || !pedal[480] {
		t.Errorf("hi-hat pedal on 2: %v", pedal)
	}
	if _, err := Apply(progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1), mustAcc(t, "block"), &d); err == nil {
		t.Error("the jazz waltz ride is a 3/4 pattern")
	}
}
