package music

import (
	"errors"
	"fmt"
)

// Tick is a position or duration in the timeline. DefaultPPQ ticks equal one
// quarter note.
type Tick int

// DefaultPPQ is the pulses-per-quarter-note resolution used by middy.
const DefaultPPQ = 480

// The meters the engine renders (ADR-0017): 4/4, 3/4 for the waltz, the
// compound 6/8 for lullabies and jigs, and the compound 12/8 for the slow
// blues and the triplet ballad (#308).
var (
	FourFour    = TimeSignature{Numerator: 4, Denominator: 4}
	ThreeFour   = TimeSignature{Numerator: 3, Denominator: 4}
	SixEight    = TimeSignature{Numerator: 6, Denominator: 8}
	TwelveEight = TimeSignature{Numerator: 12, Denominator: 8}
)

// Compound reports whether ts is a compound meter (6/8, 9/8, 12/8): its
// beat is a dotted note divided in three, not the denominator's note.
func (ts TimeSignature) Compound() bool {
	return ts.Denominator == 8 && ts.Numerator > 3 && ts.Numerator%3 == 0
}

// BeatsPerBar is how many beats a musician counts in one bar: the
// numerator in a simple meter, a third of it in a compound one (6/8 is
// counted in two).
func (ts TimeSignature) BeatsPerBar() int {
	if ts.Compound() {
		return ts.Numerator / 3
	}
	return ts.Numerator
}

// BeatTicks is the length of one beat of ts at ppq ticks per quarter note:
// a quarter in 4/4 and 3/4, a dotted quarter in 6/8. Chord lengths
// (Step.Beats) are counted in these beats.
func BeatTicks(ts TimeSignature, ppq int) Tick {
	if beats := ts.BeatsPerBar(); beats > 0 {
		return BarTicks(ts, ppq) / Tick(beats)
	}
	return 0
}

// QuarterBPM is the tempo in quarter notes per minute, the unit of a MIDI
// tempo and of seconds per tick: TempoBPM times the beat's length in
// quarters (1.5 in 6/8).
func (t Timeline) QuarterBPM() float64 {
	if t.PPQ <= 0 {
		return t.TempoBPM
	}
	return t.TempoBPM * float64(BeatTicks(t.TimeSignature, t.PPQ)) / float64(t.PPQ)
}

// PhraseBars is the length of a phrase in bars: four, the commonest in
// songs (ADR-0012 decision 5). The drums mark its end with a fill
// (BL-126), and a cadence is read only where one ends (#371, ADR-0051).
const PhraseBars = 4

// BarTicks is the length of one bar of ts at ppq ticks per quarter note.
func BarTicks(ts TimeSignature, ppq int) Tick {
	if ts.Denominator <= 0 {
		return 0
	}
	return Tick(ppq * 4 * ts.Numerator / ts.Denominator)
}

// TimeSignature is numerator/denominator, e.g. 4/4.
type TimeSignature struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

// NoteEvent is one sounding note on a track.
type NoteEvent struct {
	Pitch    Pitch `json:"pitch"`
	Velocity int   `json:"velocity"` // 1..127
	Start    Tick  `json:"start"`
	Duration Tick  `json:"duration"` // > 0
}

// End returns the tick at which the note stops sounding.
func (n NoteEvent) End() Tick {
	return n.Start + n.Duration
}

// Track is an instrument lane. Channel is 0-based (9 is percussion in GM),
// Program is the General MIDI program number (0-based).
type Track struct {
	Name    string      `json:"name"`
	Channel int         `json:"channel"`
	Program int         `json:"program"`
	Notes   []NoteEvent `json:"notes"`
}

// HarmonyEvent annotates a time span with the chord that governs it. It is
// the raw material for the visual explanation layer; it does not sound.
// Analysis is nil until Analyze has run.
type HarmonyEvent struct {
	Start        Tick   `json:"start"`
	Duration     Tick   `json:"duration"`
	Chord        Chord  `json:"chord"`
	RomanNumeral string `json:"romanNumeral"`
	// Symbol is the chord spelled as the key wants it ("Bbmaj7"); the UI
	// shows it instead of deriving a sharp-only name from Chord.Root.
	Symbol string `json:"symbol,omitempty"`
	// Bass is the note under the chord when it is a slash chord (ADR-0020);
	// nil means the chord's root.
	Bass *PitchClass `json:"bass,omitempty"`
	// Tensions names the tensions that actually sound on the chord, in
	// order ("9", "13", or "add9" on a major triad; BL-307, #450). The
	// symbol and the analysis stay at the seventh (ADR-0037); the UI writes
	// them as Fmaj7(9).
	Tensions []string `json:"tensions,omitempty"`
	// Push, when set, strikes the chord before Start (BL-085); Start stays
	// on the beat the chord belongs to.
	Push     Push      `json:"push,omitempty"`
	Analysis *Analysis `json:"analysis,omitempty"`
}

// Timeline is the central artifact of middy: everything the player, the
// piano roll, the explanation panel and the MIDI exporter consume.
type Timeline struct {
	PPQ int `json:"ppq"`
	// TempoBPM counts the beat of the meter (ADR-0019): a quarter in 4/4
	// and 3/4, a dotted quarter in 6/8. QuarterBPM converts it for MIDI
	// and the clock.
	TempoBPM      float64        `json:"tempoBpm"`
	TimeSignature TimeSignature  `json:"timeSignature"`
	Key           Key            `json:"key"`
	Tracks        []Track        `json:"tracks"`
	Harmony       []HarmonyEvent `json:"harmony"`
	// Rhythm is present once patterns have been applied (see pattern.Apply).
	Rhythm *Rhythm `json:"rhythm,omitempty"`
	// Overview explains the example as a whole; Analyze fills it.
	Overview *Overview `json:"overview,omitempty"`
}

// NewTimeline returns an empty timeline with sensible defaults.
func NewTimeline(key Key, tempoBPM float64) Timeline {
	return Timeline{
		PPQ:           DefaultPPQ,
		TempoBPM:      tempoBPM,
		TimeSignature: TimeSignature{Numerator: 4, Denominator: 4},
		Key:           key,
	}
}

// LengthTicks is the end of the last note or harmony event.
func (t Timeline) LengthTicks() Tick {
	var end Tick
	for _, tr := range t.Tracks {
		for _, n := range tr.Notes {
			if e := n.End(); e > end {
				end = e
			}
		}
	}
	for _, h := range t.Harmony {
		if e := h.Start + h.Duration; e > end {
			end = e
		}
	}
	return end
}

// Validate checks structural invariants. It is called at the API boundary
// so that every other function may assume a valid timeline. All problems
// are reported together.
func (t Timeline) Validate() error {
	errs := t.validateHeader()
	for ti, tr := range t.Tracks {
		errs = append(errs, tr.validate(fmt.Sprintf("tracks[%d]", ti))...)
	}
	for hi, h := range t.Harmony {
		errs = append(errs, h.validate(fmt.Sprintf("harmony[%d]", hi))...)
		if h.Attack(t.PPQ) < 0 {
			errs = append(errs, fmt.Errorf("harmony[%d]: a push before the start of the timeline", hi))
		}
	}
	return errors.Join(errs...)
}

func (t Timeline) validateHeader() []error {
	var errs []error
	// The bounds a Standard MIDI File can hold: a 15-bit division, a
	// 24-bit tempo (at least about 4 quarters a minute) and one-byte meter
	// fields; a timeline from outside (a MIDI export) is checked against them.
	if t.PPQ <= 0 || t.PPQ > 0x7FFF {
		errs = append(errs, fmt.Errorf("ppq must be 1..32767, got %d", t.PPQ))
	}
	if t.TempoBPM < minTempoBPM || t.TempoBPM > maxTempoBPM {
		errs = append(errs, fmt.Errorf("tempoBpm must be %g..%g, got %v", minTempoBPM, maxTempoBPM, t.TempoBPM))
	}
	ts := t.TimeSignature
	if ts.Numerator <= 0 || ts.Numerator > 32 || !isPowerOfTwo(ts.Denominator) || ts.Denominator > 32 {
		errs = append(errs, fmt.Errorf("invalid time signature %d/%d", ts.Numerator, ts.Denominator))
	}
	return errs
}

func (tr Track) validate(where string) []error {
	var errs []error
	if tr.Channel < 0 || tr.Channel > 15 {
		errs = append(errs, fmt.Errorf("%s: channel must be 0..15, got %d", where, tr.Channel))
	}
	if tr.Program < 0 || tr.Program > 127 {
		errs = append(errs, fmt.Errorf("%s: program must be 0..127, got %d", where, tr.Program))
	}
	for ni, n := range tr.Notes {
		errs = append(errs, n.validate(fmt.Sprintf("%s.notes[%d]", where, ni))...)
	}
	return errs
}

func (n NoteEvent) validate(where string) []error {
	var errs []error
	if !n.Pitch.Valid() {
		errs = append(errs, fmt.Errorf("%s: pitch out of range: %d", where, n.Pitch))
	}
	if n.Velocity < 1 || n.Velocity > 127 {
		errs = append(errs, fmt.Errorf("%s: velocity must be 1..127, got %d", where, n.Velocity))
	}
	if n.Start < 0 {
		errs = append(errs, fmt.Errorf("%s: start must be >= 0, got %d", where, n.Start))
	}
	if n.Duration <= 0 {
		errs = append(errs, fmt.Errorf("%s: duration must be > 0, got %d", where, n.Duration))
	}
	return errs
}

func (h HarmonyEvent) validate(where string) []error {
	var errs []error
	if h.Start < 0 || h.Duration <= 0 {
		errs = append(errs, fmt.Errorf("%s: invalid span start=%d duration=%d", where, h.Start, h.Duration))
	}
	if _, err := h.Chord.Quality.info(); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", where, err))
	}
	if h.Push != PushNone && h.Push != PushEighth && h.Push != PushSixteenth {
		errs = append(errs, fmt.Errorf("%s: unknown push %q", where, h.Push))
	}
	return errs
}

// The tempo range a timeline may have, in beats of its meter.
const (
	minTempoBPM = 10.0
	maxTempoBPM = 1000.0
)

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
