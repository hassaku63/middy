package music

import (
	"errors"
	"fmt"
)

// Step is one chord of a progression: a scale degree played for a number of
// beats. Beats are quarter notes (DefaultPPQ ticks each), independent of the
// time signature, so a 4-beat step is one bar in 4/4 and two bars in 2/4.
type Step struct {
	Degree  int  `json:"degree"`  // 1..7
	Seventh bool `json:"seventh"` // stack a seventh on the triad
	Beats   int  `json:"beats"`   // >= 1
	// Quality, when set, replaces the diatonic quality while keeping the
	// diatonic root. Today it exists for one idiom: the harmonic-minor
	// dominant (a major V / V7 in a minor key). See StepFromLabel.
	Quality *ChordQuality `json:"quality,omitempty"`
}

// Progression is an ordered list of steps, e.g. I-V-vi-IV. It is key-agnostic:
// the same progression renders in any key via FromProgression.
type Progression []Step

// Validate checks every step and reports all problems at once.
func (p Progression) Validate() error {
	if len(p) == 0 {
		return errors.New("music: progression needs at least one step")
	}
	var errs []error
	for i, s := range p {
		if s.Degree < 1 || s.Degree > 7 {
			errs = append(errs, fmt.Errorf("music: steps[%d]: degree must be 1..7, got %d", i, s.Degree))
		}
		if s.Beats < 1 {
			errs = append(errs, fmt.Errorf("music: steps[%d]: beats must be >= 1, got %d", i, s.Beats))
		}
		if s.Quality != nil {
			if _, err := s.Quality.info(); err != nil {
				errs = append(errs, fmt.Errorf("music: steps[%d]: %w", i, err))
			}
		}
	}
	return errors.Join(errs...)
}

// Rendering defaults for FromProgression. They are exported so that later
// pattern work (backlog BL-020) and tests can refer to them by name.
const (
	// ChordVelocity is the MIDI velocity of generated block chords.
	ChordVelocity = 80
	// ChordOctave is the octave of the chord root in generated voicings (C4 = middle C).
	ChordOctave = 4
	// ChordTrackName names the single accompaniment track.
	ChordTrackName = "chords"
)

// FromProgression renders a progression in a key as block chords: every
// chord tone sounds for the whole step on one GM piano track, and each step
// is annotated in Timeline.Harmony with its chord and roman numeral.
//
// This is deliberately the plainest possible realisation; rhythmic and
// voicing patterns are layered on top of it later (BL-020, BL-021).
func FromProgression(key Key, prog Progression, tempoBPM float64) (Timeline, error) {
	if err := prog.Validate(); err != nil {
		return Timeline{}, err
	}
	if tempoBPM <= 0 {
		return Timeline{}, fmt.Errorf("music: tempoBpm must be positive, got %v", tempoBPM)
	}

	tl := NewTimeline(key, tempoBPM)
	track := Track{Name: ChordTrackName, Channel: 0, Program: 0}

	var cursor Tick
	for i, s := range prog {
		chord, err := key.Diatonic(s.Degree, s.Seventh)
		if err != nil {
			return Timeline{}, fmt.Errorf("music: steps[%d]: %w", i, err)
		}
		if s.Quality != nil {
			chord.Quality = *s.Quality
		}
		numeral, err := RomanNumeral(s.Degree, chord.Quality)
		if err != nil {
			return Timeline{}, fmt.Errorf("music: steps[%d]: %w", i, err)
		}
		duration := Tick(s.Beats) * Tick(tl.PPQ)

		for _, p := range chord.Voicing(ChordOctave) {
			track.Notes = append(track.Notes, NoteEvent{
				Pitch:    p,
				Velocity: ChordVelocity,
				Start:    cursor,
				Duration: duration,
			})
		}
		tl.Harmony = append(tl.Harmony, HarmonyEvent{
			Start:        cursor,
			Duration:     duration,
			Chord:        chord,
			RomanNumeral: numeral,
			Symbol:       key.ChordSymbol(chord),
		})
		cursor += duration
	}

	tl.Tracks = []Track{track}
	return tl, nil
}
