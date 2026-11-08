package pattern

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// GridID is the id of drums written as a grid rather than taken from the
// catalogue (BL-122).
const GridID = "grid"

// More GM drum notes a grid can use.
const (
	Crash   music.Pitch = 49
	HighTom music.Pitch = 50
)

type gridVoice struct {
	id   string
	note music.Pitch
	foot bool // played with a foot, so it does not take a hand
}

var gridVoices = []gridVoice{
	{"kick", Kick, true}, {"snare", Snare, false}, {"cross-stick", SideStick, false},
	{"hat", ClosedHat, false}, {"open-hat", OpenHat, false}, {"pedal-hat", PedalHat, true},
	{"ride", Ride, false}, {"ride-bell", RideBell, false}, {"crash", Crash, false},
	{"low-tom", LowTom, false}, {"mid-tom", MidTom, false}, {"high-tom", HighTom, false},
}

// GridVoices lists the voices a grid may use.
func GridVoices() []string {
	out := make([]string, len(gridVoices))
	for i, v := range gridVoices {
		out[i] = v.id
	}
	return out
}

// GridDivisions are the steps per beat a grid may use in the meter: a
// simple beat splits in two, three (triplets) or four; a compound beat (a
// dotted quarter) in three or six.
func GridDivisions(ts music.TimeSignature) []int {
	if ts.Compound() {
		return []int{3, 6}
	}
	return []int{2, 3, 4}
}

// Stroke velocities: accent, normal, ghost.
var strokes = map[rune]int{'X': 104, 'x': 80, 'o': 44}

// minStepMS is the shortest step a grid may use at its tempo: faster than
// this, the strokes blur and a learner cannot follow them (ADR-0012).
const minStepMS = 100

// GridDrums validates a grid against the meter and tempo and renders it
// as one bar of drums. Every problem is reported at once.
func GridDrums(g music.DrumGrid, ts music.TimeSignature, tempoBPM float64) (Drums, error) {
	if !slices.Contains(GridDivisions(ts), g.Division) {
		return Drums{}, fmt.Errorf("pattern: grid division must be one of %v in %d/%d, got %d", GridDivisions(ts), ts.Numerator, ts.Denominator, g.Division)
	}
	steps := ts.BeatsPerBar() * g.Division
	stepTicks := music.BeatTicks(ts, music.DefaultPPQ) / music.Tick(g.Division)
	var errs []error
	if ms := 60000 / tempoBPM / float64(g.Division); ms < minStepMS {
		errs = append(errs, fmt.Errorf("pattern: grid steps of %d ms at %g BPM are shorter than %d ms; use a coarser division", int(math.Round(ms)), tempoBPM, minStepMS))
	}
	hands := make([][]string, steps)
	var hits []DrumHit
	for _, v := range gridVoices {
		line, ok := g.Voices[v.id]
		if !ok {
			continue
		}
		vh, err := v.read(line, steps, stepTicks)
		if err != nil {
			errs = append(errs, fmt.Errorf("pattern: grid voices[%q] %w (%d beats of %d)", v.id, err, ts.BeatsPerBar(), g.Division))
			continue
		}
		for _, h := range vh {
			if !v.foot {
				i := int(h.Start / stepTicks)
				hands[i] = append(hands[i], v.id)
			}
		}
		hits = append(hits, vh...)
	}
	errs = append(errs, unknownVoices(g)...)
	if err := basedOnErr(g, ts); err != nil {
		errs = append(errs, err)
	}
	for i, h := range hands {
		if len(h) > 2 {
			errs = append(errs, fmt.Errorf("pattern: grid step %d needs %d hands (%s); two hands play at most 2", i+1, len(h), strings.Join(h, ", ")))
		}
	}
	if len(hits) == 0 && len(errs) == 0 {
		errs = append(errs, errors.New("pattern: the grid has no strokes"))
	}
	if err := errors.Join(errs...); err != nil {
		return Drums{}, err
	}
	meter := ts
	return Drums{ID: GridID, Name: gridName(g), Description: gridDescription(g, ts), Meter: &meter, Hits: hits}, nil
}

// read turns one voice's line into hits, one per stroke.
func (v gridVoice) read(line string, steps int, stepTicks music.Tick) ([]DrumHit, error) {
	cells := []rune(strings.NewReplacer(" ", "", "|", "").Replace(line))
	if len(cells) != steps {
		return nil, fmt.Errorf("has %d steps, want %d", len(cells), steps)
	}
	var hits []DrumHit
	var errs []error
	for i, c := range cells {
		vel, ok := strokes[c]
		switch {
		case c == '.':
		case !ok:
			errs = append(errs, fmt.Errorf("step %d: %q is not one of X, x, o or a dot", i+1, string(c)))
		default:
			hits = append(hits, DrumHit{Start: music.Tick(i) * stepTicks, Duration: stepTicks, Note: v.note, Velocity: vel})
		}
	}
	return hits, errors.Join(errs...)
}

// basedOnErr checks that a grid's basedOn names drums of its meter.
func basedOnErr(g music.DrumGrid, ts music.TimeSignature) error {
	if g.BasedOn == "" {
		return nil
	}
	if d, err := FindDrums(g.BasedOn); err == nil && d.Fits(ts) {
		return nil
	}
	return fmt.Errorf("pattern: grid basedOn %q is not drums written for %d/%d; use one of %v", g.BasedOn, ts.Numerator, ts.Denominator, drumIDs(DrumPatternsFor(ts)))
}

func drumIDs(ds []Drums) []string {
	out := make([]string, len(ds))
	for i, d := range ds {
		out[i] = d.ID
	}
	return out
}

// unknownVoices reports voices outside the vocabulary, in a stable order.
func unknownVoices(g music.DrumGrid) []error {
	known := GridVoices()
	var bad []string
	for id := range g.Voices {
		if !slices.Contains(known, id) {
			bad = append(bad, id)
		}
	}
	slices.Sort(bad)
	errs := make([]error, len(bad))
	for i, id := range bad {
		errs[i] = fmt.Errorf("pattern: grid has unknown voice %q; use one of %v", id, known)
	}
	return errs
}

// gridName names the grid after the catalogue drums it starts from.
func gridName(g music.DrumGrid) string {
	if d, err := FindDrums(g.BasedOn); err == nil {
		return d.Name + "をもとにしたグルーヴ"
	}
	return "格子で書いたグルーヴ"
}

func gridDescription(g music.DrumGrid, ts music.TimeSignature) string {
	return fmt.Sprintf("%sの格子に、1 小節分の打ち方を書いたもの。", divisionName(g.Division, ts))
}

// divisionName names a grid's step: "16 分", "3 連符".
func divisionName(division int, ts music.TimeSignature) string {
	switch {
	case ts.Compound() && division == 3, !ts.Compound() && division == 2:
		return "8 分"
	case ts.Compound() && division == 6, !ts.Compound() && division == 4:
		return "16 分"
	default:
		return "3 連符"
	}
}
