// Package intent defines MusicalIntent: the structured interpretation of a
// user's image ("a rainy Sunday afternoon") that the LLM returns and the
// theory engine renders (ADR-0003). The LLM never emits notes; it emits
// this, and Validate decides whether the engine can honour it.
package intent

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// Limits the validator enforces. They are also baked into Schema so the
// model is told the rules up front.
const (
	MinCandidates   = 1
	MaxCandidates   = 3
	MinTempoBPM     = 40
	MaxTempoBPM     = 220
	MaxChords       = 16
	MaxMoodTags     = 6
	MaxSummaryRunes = 300
	MaxReasonRunes  = 300
)

// AllowedBeats are the per-chord lengths a candidate may use, in beats of
// its meter (ADR-0017): in 4/4 a beat, half a bar, a bar or two; in 3/4 a
// bar or two, since a chord on two of three beats would cross the beat
// pattern (a hemiola, a special effect left out for now); in 6/8, whose
// beat is a dotted quarter, half a bar, a bar or two.
var AllowedBeats = map[music.TimeSignature][]int{
	music.FourFour:  {1, 2, 4, 8},
	music.ThreeFour: {3, 6},
	music.SixEight:  {1, 2, 4},
}

// Meters are the time signatures a candidate may ask for; omitted is 4/4.
var Meters = map[string]music.TimeSignature{"4/4": music.FourFour, "3/4": music.ThreeFour, "6/8": music.SixEight}

// MeterIDs lists Meters in a stable order.
func MeterIDs() []string { return []string{"4/4", "3/4", "6/8"} }

// Intent is the LLM's whole answer for one turn of the conversation.
type Intent struct {
	// Summary restates the image in one or two Japanese sentences; shown as
	// the assistant's reply and sent back as history on later turns.
	Summary string `json:"summary"`
	// Mood are short tags such as "静か" or "切ない".
	Mood []string `json:"mood"`
	// Candidates are concrete musical proposals, best first.
	Candidates []Candidate `json:"candidates"`
	// Unconstrained is what the model would have written without the
	// engine's limits. Asked for only by DiagnosticSchema (evaluation);
	// never rendered, never validated (ADR-0013).
	Unconstrained string `json:"unconstrained,omitempty"`
}

// Candidate is one proposal the engine can render as a Timeline.
type Candidate struct {
	Title         string    `json:"title"`
	Key           music.Key `json:"key"`
	TempoBPM      float64   `json:"tempoBpm"`
	Progression   []string  `json:"progression"`   // roman numerals, see music.ParseRomanNumeral
	BeatsPerChord int       `json:"beatsPerChord"` // one of AllowedBeats
	// Beats, when set, gives each chord its own length in beats of the
	// meter (BL-109), one per progression entry; BeatsPerChord is then
	// only the default the model also states.
	Beats  []int  `json:"beats,omitempty"`
	Reason string `json:"reason"` // why this fits the image, Japanese
	// Accompaniment and Drums are pattern ids from the catalogue
	// (internal/music/pattern); empty means block chords / no drums.
	Accompaniment string `json:"accompaniment,omitempty"`
	Drums         string `json:"drums,omitempty"`
	// Feel is how the eighths are played (pattern.Feels); empty means straight.
	Feel string `json:"feel,omitempty"`
	// Meter is one of MeterIDs (ADR-0017); empty means 4/4.
	Meter string `json:"meter,omitempty"`
}

// TimeSignature is the candidate's meter, 4/4 when omitted or unknown
// (Validate rejects unknown ones).
func (c Candidate) TimeSignature() music.TimeSignature {
	if ts, ok := Meters[c.Meter]; ok {
		return ts
	}
	return music.FourFour
}

// Validate reports every problem at once so the caller can feed them back
// to the model in one retry.
func (in Intent) Validate() error {
	var errs []error
	if strings.TrimSpace(in.Summary) == "" {
		errs = append(errs, errors.New("summary is required"))
	}
	if n := len([]rune(in.Summary)); n > MaxSummaryRunes {
		errs = append(errs, fmt.Errorf("summary is too long (%d > %d characters)", n, MaxSummaryRunes))
	}
	if len(in.Mood) > MaxMoodTags {
		errs = append(errs, fmt.Errorf("mood has %d tags, at most %d allowed", len(in.Mood), MaxMoodTags))
	}
	for i, m := range in.Mood {
		if strings.TrimSpace(m) == "" {
			errs = append(errs, fmt.Errorf("mood[%d] is empty", i))
		}
	}
	if len(in.Candidates) < MinCandidates || len(in.Candidates) > MaxCandidates {
		errs = append(errs, fmt.Errorf("candidates must have %d..%d entries, got %d", MinCandidates, MaxCandidates, len(in.Candidates)))
	}
	for i, c := range in.Candidates {
		for _, err := range c.validate() {
			errs = append(errs, fmt.Errorf("candidates[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

func (c Candidate) validate() []error {
	var errs []error
	if strings.TrimSpace(c.Title) == "" {
		errs = append(errs, errors.New("title is required"))
	}
	errs = append(errs, c.validateKeyAndTempo()...)
	errs = append(errs, c.validateProgression()...)
	if strings.TrimSpace(c.Reason) == "" {
		errs = append(errs, errors.New("reason is required"))
	}
	if n := len([]rune(c.Reason)); n > MaxReasonRunes {
		errs = append(errs, fmt.Errorf("reason is too long (%d > %d characters)", n, MaxReasonRunes))
	}
	errs = append(errs, c.validatePatterns()...)
	return errs
}

func (c Candidate) validateKeyAndTempo() []error {
	var errs []error
	if c.Key.Tonic < 0 || c.Key.Tonic > 11 {
		errs = append(errs, fmt.Errorf("key.tonic must be 0..11, got %d", c.Key.Tonic))
	}
	if _, err := c.Key.Mode.MarshalText(); err != nil {
		errs = append(errs, errors.New("key.mode must be major or minor"))
	}
	if c.TempoBPM < MinTempoBPM || c.TempoBPM > MaxTempoBPM {
		errs = append(errs, fmt.Errorf("tempoBpm must be %d..%d, got %v", MinTempoBPM, MaxTempoBPM, c.TempoBPM))
	}
	return errs
}

func (c Candidate) validateProgression() []error {
	var errs []error
	if len(c.Progression) == 0 || len(c.Progression) > MaxChords {
		errs = append(errs, fmt.Errorf("progression must have 1..%d chords, got %d", MaxChords, len(c.Progression)))
	}
	for i, label := range c.Progression {
		// A label that names a chord the key would silently change (IV in
		// minor, I7 in major) is an error the model can act on, not
		// something to fix quietly (ADR-0013).
		if _, err := music.CanonicalLabel(c.Key, label); err != nil {
			errs = append(errs, fmt.Errorf("progression[%d]: %w%s", i, err, c.suggestLabel(label)))
		}
	}
	if _, ok := Meters[c.Meter]; c.Meter != "" && !ok {
		errs = append(errs, fmt.Errorf("meter must be one of %v, got %q", MeterIDs(), c.Meter))
	}
	ts := c.TimeSignature()
	if allowed := AllowedBeats[ts]; !slices.Contains(allowed, c.BeatsPerChord) {
		errs = append(errs, fmt.Errorf("beatsPerChord must be one of %v in %d/%d, got %d%s", allowed, ts.Numerator, ts.Denominator, c.BeatsPerChord, lengthHint(ts, c.BeatsPerChord)))
	}
	return append(errs, c.validateBeats(ts)...)
}

// suggestLabel offers the label one deletion away from a misspelt one
// ("Imaxj7" → "Imaj7"), so a retry can fix it (#133); "" when none reads.
func (c Candidate) suggestLabel(label string) string {
	r := []rune(label)
	for i := range r {
		try := string(r[:i]) + string(r[i+1:])
		if _, err := music.CanonicalLabel(c.Key, try); err == nil {
			return fmt.Sprintf(`; did you mean %q?`, try)
		}
	}
	return ""
}

// lengthHint names the meter a length belongs to when the candidate's
// meter does not allow it (a waltz's 3 without meter "3/4"; #133).
func lengthHint(ts music.TimeSignature, beats int) string {
	for _, id := range MeterIDs() {
		if m := Meters[id]; m != ts && slices.Contains(AllowedBeats[m], beats) && !slices.Contains(AllowedBeats[ts], beats) {
			return fmt.Sprintf(`; %d is a length in %s; set meter to %q to use it`, beats, id, id)
		}
	}
	return ""
}

// patternHint names the meter a pattern is written for when it does not
// fit the candidate's (#133); "" for an unknown id or a pattern of any meter.
func patternHint(id string, meter *music.TimeSignature) string {
	if meter == nil {
		return ""
	}
	name := fmt.Sprintf("%d/%d", meter.Numerator, meter.Denominator)
	if _, ok := Meters[name]; !ok {
		return ""
	}
	return fmt.Sprintf(`; %q is a %s pattern; set meter to %q to use it`, id, name, name)
}

// validateBeats checks per-chord lengths (BL-109) against the rule of
// ADR-0017: chords change on a bar line or at an equal division of the
// bar. A chord shorter than a bar starts at a multiple of its own length
// within the bar, so it never crosses a bar line; a bar or longer starts
// on a bar line; the progression ends on a bar line.
func (c Candidate) validateBeats(ts music.TimeSignature) []error {
	if c.Beats == nil {
		return nil
	}
	if len(c.Beats) != len(c.Progression) {
		return []error{fmt.Errorf("beats has %d lengths for %d chords", len(c.Beats), len(c.Progression))}
	}
	var errs []error
	allowed := AllowedBeats[ts]
	bar := ts.BeatsPerBar()
	at := 0 // beats from the start
	for i, b := range c.Beats {
		pos := at % bar
		switch {
		case !slices.Contains(allowed, b):
			errs = append(errs, fmt.Errorf("beats[%d] must be one of %v in %d/%d, got %d", i, allowed, ts.Numerator, ts.Denominator, b))
		case b < bar && pos%b != 0:
			errs = append(errs, fmt.Errorf("beats[%d]: a chord of %d beats must start on beat %s of the bar, not beat %d", i, b, startBeats(bar, b), pos+1))
		case b >= bar && pos != 0:
			errs = append(errs, fmt.Errorf("beats[%d]: a chord of %d beats must start on a bar line, not beat %d", i, b, pos+1))
		}
		at += b
	}
	if bar > 0 && at%bar != 0 {
		errs = append(errs, fmt.Errorf("the progression ends mid-bar (%d beats, bars of %d)", at, bar))
	}
	return errs
}

// startBeats lists where a chord of length b may start in a bar: "1 or 3".
func startBeats(bar, b int) string {
	var s []string
	for p := 0; p < bar; p += b {
		s = append(s, fmt.Sprint(p+1))
	}
	return strings.Join(s, " or ")
}

func (c Candidate) validatePatterns() []error {
	var errs []error
	// Patterns must be written for the candidate's meter (ADR-0017).
	ts := c.TimeSignature()
	if c.Accompaniment != "" {
		if a, err := pattern.FindAccompaniment(c.Accompaniment); err != nil || !a.Fits(ts) {
			errs = append(errs, fmt.Errorf("accompaniment must be one of %v in %d/%d, got %q%s", AccompanimentIDsFor(ts), ts.Numerator, ts.Denominator, c.Accompaniment, patternHint(c.Accompaniment, a.Meter)))
		}
	}
	if c.Drums != "" {
		if d, err := pattern.FindDrums(c.Drums); err != nil || !d.Fits(ts) {
			errs = append(errs, fmt.Errorf("drums must be one of %v in %d/%d, got %q%s", DrumIDsFor(ts), ts.Numerator, ts.Denominator, c.Drums, patternHint(c.Drums, d.Meter)))
		}
	}
	if c.Feel != "" && !slices.Contains(FeelIDs(), c.Feel) {
		errs = append(errs, fmt.Errorf("feel must be one of %v, got %q", FeelIDs(), c.Feel))
	}
	if c.Feel == string(pattern.FeelSwing) && ts.Compound() {
		errs = append(errs, fmt.Errorf("feel swing is not used in %d/%d: its beat already divides in three", ts.Numerator, ts.Denominator))
	}
	return errs
}

// AccompanimentIDs lists every catalogue id, in any meter (the schema's enum).
func AccompanimentIDs() []string {
	var ids []string
	for _, a := range pattern.Accompaniments() {
		ids = append(ids, a.ID)
	}
	return ids
}

// AccompanimentIDsFor lists the ids that can play in ts.
func AccompanimentIDsFor(ts music.TimeSignature) []string {
	var ids []string
	for _, a := range pattern.AccompanimentsFor(ts) {
		ids = append(ids, a.ID)
	}
	return ids
}

// FeelIDs lists the feels the model may use.
func FeelIDs() []string {
	var ids []string
	for _, f := range pattern.Feels() {
		ids = append(ids, string(f.ID))
	}
	return ids
}

// DrumIDs lists every drum id, in any meter (the schema's enum).
func DrumIDs() []string {
	var ids []string
	for _, d := range pattern.DrumPatterns() {
		ids = append(ids, d.ID)
	}
	return ids
}

// DrumIDsFor lists the drum ids that can play in ts.
func DrumIDsFor(ts music.TimeSignature) []string {
	var ids []string
	for _, d := range pattern.DrumPatternsFor(ts) {
		ids = append(ids, d.ID)
	}
	return ids
}

// ToProgression converts the candidate's labels into the engine's type.
// It assumes Validate passed.
func (c Candidate) ToProgression() (music.Progression, error) {
	prog := make(music.Progression, 0, len(c.Progression))
	for i, label := range c.Progression {
		beats := c.BeatsPerChord
		if c.Beats != nil {
			beats = c.Beats[i]
		}
		step, err := music.StepFromLabel(c.Key, label, beats)
		if err != nil {
			return nil, fmt.Errorf("intent: progression[%d]: %w", i, err)
		}
		prog = append(prog, step)
	}
	return prog, nil
}
