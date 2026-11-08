package music

import (
	"errors"
	"fmt"
	"slices"
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
	// Accidental lowers the root by one semitone from the scale degree
	// (-1 for the borrowed ♭VI, ♭VII). Zero for diatonic roots.
	Accidental int `json:"accidental,omitempty"`
	// Bass, when set, puts that note under the chord (a slash chord,
	// ADR-0020); the chord itself is unchanged.
	Bass *Bass `json:"bass,omitempty"`
	// Borrowed marks a chord taken from the parallel minor (BL-094); the
	// analysis says so.
	Borrowed bool `json:"borrowed,omitempty"`
	// Push strikes the chord an eighth or a sixteenth early (BL-085).
	Push Push `json:"push,omitempty"`
	// Applied makes the step the secondary dominant of Degree (V/x, V7/x
	// with Seventh; BL-074) instead of the chord on it.
	Applied bool `json:"applied,omitempty"`
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
		errs = append(errs, s.problems(i)...)
	}
	return errors.Join(errs...)
}

// problems lists what is wrong with step i.
func (s Step) problems(i int) []error {
	var errs []error
	if s.Degree < 1 || s.Degree > 7 {
		errs = append(errs, fmt.Errorf("music: steps[%d]: degree must be 1..7, got %d", i, s.Degree))
	}
	if s.Beats < 1 {
		errs = append(errs, fmt.Errorf("music: steps[%d]: beats must be >= 1, got %d", i, s.Beats))
	}
	if s.Accidental != 0 && s.Accidental != -1 {
		errs = append(errs, fmt.Errorf("music: steps[%d]: accidental must be 0 or -1, got %d", i, s.Accidental))
	}
	if s.Applied && (s.Accidental != 0 || s.Quality != nil || s.Borrowed) {
		errs = append(errs, fmt.Errorf("music: steps[%d]: an applied chord has no accidental, quality or borrowing of its own", i))
	}
	if err := s.Push.check(i); err != nil {
		errs = append(errs, err)
	}
	if s.Quality != nil {
		if _, err := s.Quality.info(); err != nil {
			errs = append(errs, fmt.Errorf("music: steps[%d]: %w", i, err))
		}
	}
	return errs
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
	return FromProgressionIn(FourFour, key, prog, tempoBPM)
}

// SupportedMeter reports whether the engine renders ts (ADR-0017).
func SupportedMeter(ts TimeSignature) error {
	if ts != FourFour && ts != ThreeFour && ts != SixEight {
		return fmt.Errorf("music: meter %d/%d is not supported yet (4/4, 3/4, 6/8)", ts.Numerator, ts.Denominator)
	}
	return nil
}

// FromProgressionIn is FromProgression in another meter (ADR-0017). Step
// lengths are in beats of the meter: quarter notes in 4/4 and 3/4, dotted
// quarters in 6/8.
func FromProgressionIn(ts TimeSignature, key Key, prog Progression, tempoBPM float64) (Timeline, error) {
	if err := SupportedMeter(ts); err != nil {
		return Timeline{}, err
	}
	if err := prog.Validate(); err != nil {
		return Timeline{}, err
	}
	if tempoBPM <= 0 {
		return Timeline{}, fmt.Errorf("music: tempoBpm must be positive, got %v", tempoBPM)
	}

	tl := NewTimeline(key, tempoBPM)
	tl.TimeSignature = ts
	track := Track{Name: ChordTrackName, Channel: 0, Program: 0}

	beat := BeatTicks(ts, tl.PPQ)
	chords := make([]Chord, len(prog))
	numerals, symbols := make([]string, len(prog)), make([]string, len(prog))
	for i, s := range prog {
		chord, numeral, symbol, err := stepChord(key, s)
		if err != nil {
			return Timeline{}, fmt.Errorf("music: steps[%d]: %w", i, err)
		}
		chords[i], numerals[i], symbols[i] = chord, numeral, symbol
	}
	led := LeadVoicings(chords) // each chord voiced from the one before (BL-086)
	var cursor Tick
	for i, s := range prog {
		chord, numeral, symbol := chords[i], numerals[i], symbols[i]
		duration := Tick(s.Beats) * beat

		pitches := slices.Clone(led[i])
		var bass *PitchClass
		if s.Bass != nil {
			pc := s.Bass.pitchClass(key)
			bass = &pc
			pitches = append([]Pitch{NewPitch(pc, ChordOctave-1)}, pitches...) // under the chord
		}
		for _, p := range pitches {
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
			Symbol:       symbol,
			Bass:         bass,
			Push:         s.Push,
		})
		cursor += duration
	}
	track.Notes = PushNotes(tl.Harmony, track.Notes, tl.PPQ)

	tl.Tracks = []Track{track}
	return tl, nil
}

// stepChord resolves a step to its chord, roman numeral and spelled symbol.
func stepChord(key Key, s Step) (Chord, string, string, error) {
	chord, numeral, symbol, err := chordOfStep(key, s)
	if err != nil || s.Bass == nil {
		return chord, numeral, symbol, err
	}
	return chord, numeral + s.Bass.suffix(), symbol + "/" + key.spellDegree(s.Bass.Degree, s.Bass.Accidental), nil
}

// chordOfStep resolves the chord, numeral and symbol without the bass.
func chordOfStep(key Key, s Step) (Chord, string, string, error) {
	if s.Applied {
		chord, err := key.appliedChord(s.Degree, s.Seventh)
		if err != nil {
			return Chord{}, "", "", err
		}
		numeral, err := key.appliedNumeral(s.Degree, s.Seventh)
		if err != nil {
			return Chord{}, "", "", err
		}
		return chord, numeral, key.spellApplied(s.Degree) + chord.Quality.Symbol(), nil
	}
	chord, err := key.Diatonic(s.Degree, s.Seventh)
	if err != nil {
		return Chord{}, "", "", err
	}
	chord.Root = chord.Root.Add(s.Accidental)
	if s.Quality != nil {
		chord.Quality = *s.Quality
	}
	numeral, err := RomanNumeralAlt(s.Degree, s.Accidental, chord.Quality)
	if err != nil {
		return Chord{}, "", "", err
	}
	return chord, numeral, key.spellDegree(s.Degree, s.Accidental) + chord.Quality.Symbol(), nil
}
