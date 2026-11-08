package pattern

import (
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestReelPlaysBassOnTheBeatsAndChordsOffThem (BL-107): bass on 1 and 3,
// a short chord on every offbeat, on a steel-string guitar.
func TestReelPlaysBassOnTheBeatsAndChordsOffThem(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.D, Mode: music.Dorian}, 4, 1)
	got, err := Apply(tl, mustAcc(t, "reel"), nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := map[music.Tick]int{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		starts[n.Start]++
		if n.Start%480 != 0 && n.Duration > 240 {
			t.Errorf("offbeat chords should be short: %+v", n)
		}
	}
	want := map[music.Tick]int{0: 1, 240: 3, 720: 3, 960: 1, 1200: 3, 1680: 3}
	for s, n := range want {
		if starts[s] != n {
			t.Errorf("notes at %d = %d, want %d (all: %v)", s, starts[s], n, starts)
		}
	}
	if got.Tracks[0].Program != 25 {
		t.Errorf("reel should use a steel-string guitar (25), got %d", got.Tracks[0].Program)
	}
}

// TestBodhranKeepsSnareAndHatsOut (BL-107): low tom on 1 and 3, a mid tom
// on every eighth, a side stick on 2 and 4; no kick, snare or hi-hat, and
// the description counts the low tom and side stick as the groove.
func TestBodhranKeepsSnareAndHatsOut(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.D, Mode: music.Dorian}, 4, 1)
	d, err := FindDrums("bodhran")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "reel"), &d)
	if err != nil {
		t.Fatal(err)
	}
	count := map[music.Pitch]int{}
	for _, n := range notesIn(got, "drums") {
		count[n.Pitch]++
	}
	if count[Kick]+count[Snare]+count[ClosedHat]+count[OpenHat] != 0 {
		t.Errorf("a bodhrán has no kick, snare or hi-hat: %v", count)
	}
	if count[LowTom] != 2 || count[MidTom] != 6 || count[SideStick] != 2 {
		t.Errorf("low tom on 1 and 3, mid tom on the other eighths, side stick on 2 and 4: %v", count)
	}
	part := got.Rhythm.Drums
	if part.Hits != 4 || part.OnBeat != 4 {
		t.Errorf("the groove is the low tom and the side stick, all on the beats: %+v", part)
	}
	if !strings.Contains(part.Explanation, "低いタムとサイドスティック") {
		t.Errorf("explanation should name the bodhrán's strokes: %s", part.Explanation)
	}
	kit, _ := FindDrums("eight-beat")
	eight, err := Apply(tl, mustAcc(t, "block"), &kit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(eight.Rhythm.Drums.Explanation, "キックとスネア") {
		t.Errorf("a kit keeps its wording: %s", eight.Rhythm.Drums.Explanation)
	}
}

// TestDroneHoldsTheTonicUnderChangingChords (BL-108): the drone voice plays
// the key's tonic and fifth below the chords, the same notes under I and
// under VII, in any meter.
func TestDroneHoldsTheTonicUnderChangingChords(t *testing.T) {
	key := music.Key{Tonic: music.D, Mode: music.Dorian}
	tl := progression(t, key, 4, 1, 7) // i (Dm) then VII (C)
	got, err := Apply(tl, mustAcc(t, "drone"), nil)
	if err != nil {
		t.Fatal(err)
	}
	drone := map[music.Tick][]music.Pitch{}
	for _, n := range notesIn(got, music.ChordTrackName) {
		if n.Pitch < music.NewPitch(music.C, music.ChordOctave) {
			drone[n.Start] = append(drone[n.Start], n.Pitch)
		}
	}
	want := []music.Pitch{music.NewPitch(music.D, 3), music.NewPitch(music.A, 3)} // D3 and A3
	for _, start := range []music.Tick{0, 1920} {
		ps := drone[start]
		slices.Sort(ps)
		if !slices.Equal(ps, want) {
			t.Errorf("drone at %d = %v, want %v (D and A under both chords)", start, ps, want)
		}
	}
	if !strings.Contains(got.Rhythm.Accompaniment.Description, "和音が動いても主音が鳴り続ける") {
		t.Errorf("the description should say the tonic keeps sounding: %s", got.Rhythm.Accompaniment.Description)
	}
	if _, err := Apply(sixEightTimeline(t, 2, 1, 5), mustAcc(t, "drone"), nil); err != nil {
		t.Errorf("the drone fits any meter: %v", err)
	}
}

// TestBassVoiceFollowsTheSlashBass (BL-113): a pattern's bass voice plays
// the slash chord's bass (E under C/E), not the root.
func TestBassVoiceFollowsTheSlashBass(t *testing.T) {
	key := music.Key{Tonic: music.C, Mode: music.Major}
	step, err := music.StepFromLabel(key, "I/3", 4)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := music.FromProgression(key, music.Progression{step}, 100)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "ballad"), nil)
	if err != nil {
		t.Fatal(err)
	}
	var bass []music.Pitch
	for _, n := range notesIn(got, music.ChordTrackName) {
		if n.Pitch < music.NewPitch(music.C, music.ChordOctave) {
			bass = append(bass, n.Pitch)
		}
	}
	want := music.NewPitch(music.E, music.ChordOctave-1)
	if len(bass) != 2 || bass[0] != want || bass[1] != want {
		t.Errorf("ballad's bass on 1 and 3 should be E3 under C/E: %v", bass)
	}
}
