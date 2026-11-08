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
	Symbol   string    `json:"symbol,omitempty"`
	Analysis *Analysis `json:"analysis,omitempty"`
}

// Timeline is the central artifact of middy: everything the player, the
// piano roll, the explanation panel and the MIDI exporter consume.
type Timeline struct {
	PPQ           int            `json:"ppq"`
	TempoBPM      float64        `json:"tempoBpm"`
	TimeSignature TimeSignature  `json:"timeSignature"`
	Key           Key            `json:"key"`
	Tracks        []Track        `json:"tracks"`
	Harmony       []HarmonyEvent `json:"harmony"`
	// Rhythm is present once patterns have been applied (see pattern.Apply).
	Rhythm *Rhythm `json:"rhythm,omitempty"`
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
	}
	return errors.Join(errs...)
}

func (t Timeline) validateHeader() []error {
	var errs []error
	if t.PPQ <= 0 {
		errs = append(errs, fmt.Errorf("ppq must be positive, got %d", t.PPQ))
	}
	if t.TempoBPM <= 0 {
		errs = append(errs, fmt.Errorf("tempoBpm must be positive, got %v", t.TempoBPM))
	}
	if t.TimeSignature.Numerator <= 0 || !isPowerOfTwo(t.TimeSignature.Denominator) {
		errs = append(errs, fmt.Errorf("invalid time signature %d/%d", t.TimeSignature.Numerator, t.TimeSignature.Denominator))
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
	return errs
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
