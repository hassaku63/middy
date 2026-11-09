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
	Beats []int `json:"beats,omitempty"`
	// Pushes, when set, pushes chords early (BL-085), one per progression
	// entry: "none", "eighth" or "sixteenth".
	Pushes []string `json:"pushes,omitempty"`
	Reason string   `json:"reason"` // why this fits the image, Japanese
	// Accompaniment and Drums are pattern ids from the catalogue
	// (internal/music/pattern); empty means block chords / no drums.
	Accompaniment string `json:"accompaniment,omitempty"`
	Drums         string `json:"drums,omitempty"`
	// DrumGrid writes the drums out instead of naming them (BL-077); the
	// engine checks it as it would any grid (pattern.GridDrums).
	DrumGrid *GridSpec `json:"drumGrid,omitempty"`
	// Feel is how the eighths are played (pattern.Feels); empty means straight.
	Feel string `json:"feel,omitempty"`
	// Style is the kind of music the candidate is (pattern.Styles, BL-082);
	// its patterns and feel must belong to it. Empty in older answers.
	Style string `json:"style,omitempty"`
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
		errs = append(errs, errors.New("key.mode must be major, minor, dorian or mixolydian"))
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
			errs = append(errs, fmt.Errorf("progression[%d]: %w%s%s%s", i, err, c.suggestLabel(label), c.dominantHint(label), c.substituteHint(label)))
		}
	}
	if _, ok := Meters[c.Meter]; c.Meter != "" && !ok {
		errs = append(errs, fmt.Errorf("meter must be one of %v, got %q", MeterIDs(), c.Meter))
	}
	ts := c.TimeSignature()
	if allowed := AllowedBeats[ts]; !slices.Contains(allowed, c.BeatsPerChord) {
		errs = append(errs, fmt.Errorf("beatsPerChord must be one of %v in %d/%d, got %d%s", allowed, ts.Numerator, ts.Denominator, c.BeatsPerChord, lengthHint(ts, c.BeatsPerChord)))
	}
	errs = append(errs, c.validateBeats(ts)...)
	return append(errs, c.validatePushes()...)
}

// PushIDs are the values a candidate's pushes may take.
var PushIDs = []string{"none", string(music.PushEighth), string(music.PushSixteenth)}

// validatePushes checks the pushes (BL-085): one per chord, known values,
// and none on the first chord.
func (c Candidate) validatePushes() []error {
	if c.Pushes == nil {
		return nil
	}
	if len(c.Pushes) != len(c.Progression) {
		return []error{fmt.Errorf("pushes has %d entries for %d chords", len(c.Pushes), len(c.Progression))}
	}
	var errs []error
	for i, p := range c.Pushes {
		switch {
		case !slices.Contains(PushIDs, p):
			errs = append(errs, fmt.Errorf("pushes[%d] must be one of %v, got %q", i, PushIDs, p))
		case i == 0 && p != "none":
			errs = append(errs, fmt.Errorf("pushes[0]: the first chord has nothing to push into; write none"))
		}
	}
	return errs
}

// push is chord i's push in the engine's terms.
func (c Candidate) push(i int) music.Push {
	if c.Pushes == nil || c.Pushes[i] == "none" {
		return music.PushNone
	}
	return music.Push(c.Pushes[i])
}

// suggestLabel offers the label one deletion away from a misspelt one
// ("Imaxj7" → "Imaj7"), so a retry can fix it (#133); "" when none reads.
// A label that already reads ("VIImaj7" in minor) gets no suggestion: the
// engine's error names the right spelling, and a deletion would point at
// another chord.
func (c Candidate) suggestLabel(label string) string {
	if _, _, err := music.ParseRomanNumeral(label); err == nil || strings.Contains(label, "/") {
		return "" // it reads, or it is a slash label whose error already says the fix
	}
	r := []rune(label)
	for i := range r {
		try := string(r[:i]) + string(r[i+1:])
		if _, err := music.CanonicalLabel(c.Key, try); err == nil {
			return fmt.Sprintf(`; did you mean %q?`, try)
		}
	}
	return ""
}

// substituteHint offers the tritone substitute a dominant seventh written by
// its root most likely meant (BL-129): "bII7" in C major is Db7, which the
// key writes "subV7/I". "" when the label is not such a seventh, or the
// chord a semitone below is not a target a substitute can aim at.
func (c Candidate) substituteHint(label string) string {
	head, ok := strings.CutSuffix(strings.TrimSpace(label), "7")
	numeral := strings.TrimPrefix(head, "b") // a flat root: bII, bVI
	if !ok || numeral == "" || strings.ContainsAny(numeral, "/°ø+m") || strings.ToUpper(numeral) != numeral {
		return ""
	}
	chord, err := music.LiteralChord(c.Key.Tonic, head+"7")
	if err != nil || chord.Quality != music.Dominant7 {
		return ""
	}
	for d := 1; d <= 7; d++ {
		pc, err := c.Key.Degree(d)
		if err != nil || pc != chord.Root.Add(-1) {
			continue
		}
		target, err := c.Key.Diatonic(d, false)
		if err != nil {
			return ""
		}
		rn, err := music.RomanNumeral(d, target.Quality)
		if err != nil {
			return ""
		}
		try := "subV7/" + rn
		if _, err := music.CanonicalLabel(c.Key, try); err == nil {
			return fmt.Sprintf(`; if you meant the tritone substitute resolving to %s, write %q`, rn, try)
		}
	}
	return ""
}

// dominantHint offers the secondary dominant an uppercase label most likely
// meant (review): "II7" in C major names D7, which the key writes "V7/V",
// not the borrowed iiø7 the key would otherwise make of it. "" when the
// label is not an uppercase major triad or seventh on a degree, or the
// key has no such secondary dominant.
func (c Candidate) dominantHint(label string) string {
	head, seventh := strings.CutSuffix(label, "7")
	if strings.ContainsAny(head, "/°ø+m") || head == "" || head != strings.ToUpper(head) {
		return ""
	}
	step, err := music.StepFromLabel(c.Key, head, 4)
	if err != nil || step.Accidental != 0 {
		return ""
	}
	root, err := c.Key.Degree(step.Degree)
	if err != nil {
		return ""
	}
	for d := 1; d <= 7; d++ {
		pc, err := c.Key.Degree(d)
		if err != nil || pc != root.Add(5) {
			continue
		}
		target, err := c.Key.Diatonic(d, false)
		if err != nil {
			return ""
		}
		rn, err := music.RomanNumeral(d, target.Quality)
		if err != nil {
			return ""
		}
		try := "V/" + rn
		if seventh {
			try = "V7/" + rn
		}
		if _, err := music.CanonicalLabel(c.Key, try); err == nil {
			return fmt.Sprintf(`; if you meant the dominant on that root, write %q`, try)
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
	// A pattern that is wrong for the meter is reported once, here; the
	// style check below then leaves it out, so the retry gets one fix.
	acc, drums, feel := c.Accompaniment, c.styleDrums(), c.Feel
	if c.Accompaniment != "" {
		if a, err := pattern.FindAccompaniment(c.Accompaniment); err != nil || !a.Fits(ts) {
			errs = append(errs, fmt.Errorf("accompaniment must be one of %v in %d/%d, got %q%s", AccompanimentIDsFor(ts), ts.Numerator, ts.Denominator, c.Accompaniment, patternHint(c.Accompaniment, a.Meter)))
			acc = ""
		}
	}
	if c.Drums != "" {
		if d, err := pattern.FindDrums(c.Drums); err != nil || !d.Fits(ts) {
			errs = append(errs, fmt.Errorf("drums must be one of %v in %d/%d, got %q%s", DrumIDsFor(ts), ts.Numerator, ts.Denominator, c.Drums, patternHint(c.Drums, d.Meter)))
			drums = ""
		}
	}
	errs = append(errs, c.validateGrid(ts)...)
	if c.badBasedOn(ts) {
		drums = "" // the grid check above already named the drums to base it on
	}
	if c.Feel != "" && !slices.Contains(FeelIDs(), c.Feel) {
		errs = append(errs, fmt.Errorf("feel must be one of %v, got %q", FeelIDs(), c.Feel))
		feel = ""
	}
	errs = append(errs, c.styleErrs(acc, drums, feel)...)
	if c.Feel == string(pattern.FeelSwing) && ts.Compound() {
		errs = append(errs, fmt.Errorf("feel swing is not used in %d/%d: its beat already divides in three", ts.Numerator, ts.Denominator))
	}
	return errs
}

// styleErrs holds the patterns that passed their own checks to the style;
// the ones that failed are passed in empty, so each mistake is one error.
func (c Candidate) styleErrs(acc, drums, feel string) []error {
	if c.Style == "" {
		return nil
	}
	st, err := pattern.FindStyle(c.Style)
	if err != nil {
		return []error{fmt.Errorf("style must be one of %v, got %q", StyleIDs(), c.Style)}
	}
	var errs []error
	for _, m := range st.Misfits(acc, drums, pattern.Feel(feel)) {
		errs = append(errs, errors.New(m))
	}
	return errs
}

// badBasedOn reports a grid basedOn that is not catalogue drums for ts.
func (c Candidate) badBasedOn(ts music.TimeSignature) bool {
	g := c.DrumGrid
	if g == nil || g.BasedOn == "" {
		return false
	}
	d, err := pattern.FindDrums(g.BasedOn)
	return err != nil || !d.Fits(ts)
}

// validateGrid checks a written-out drum grid (BL-077) with the engine's
// rules, at the candidate's meter and tempo.
func (c Candidate) validateGrid(ts music.TimeSignature) []error {
	g := c.DrumGrid
	if g == nil {
		return nil
	}
	var errs []error
	if c.Drums != "" {
		errs = append(errs, errors.New("write drums or drumGrid, not both"))
	}
	grid, err := g.Grid()
	if err != nil {
		return append(errs, err)
	}
	// An out-of-range tempo is reported on its own; timing the steps at it
	// would ask for a coarser grid the fixed tempo may not need, so the grid
	// is timed at the slowest tempo instead (the retry is checked again).
	tempo := c.TempoBPM
	if tempo < MinTempoBPM || tempo > MaxTempoBPM {
		tempo = MinTempoBPM
	}
	if _, err := pattern.GridDrums(grid, ts, tempo); err != nil {
		errs = append(errs, fmt.Errorf("drumGrid: %w", err))
	}
	return errs
}

// GridSpec is a drum grid as the model writes it (BL-077): the voices are
// a list of voice and steps, which keeps the output schema flat; Grid
// turns it into the engine's music.DrumGrid.
type GridSpec struct {
	Division int        `json:"division"`
	Voices   []GridLine `json:"voices"`
	BasedOn  string     `json:"basedOn,omitempty"`
}

// GridLine is one voice of a grid: one character a step.
type GridLine struct {
	Voice string `json:"voice"`
	Steps string `json:"steps"`
}

// Grid is the spec as the engine's grid; a voice written twice is an error.
func (g GridSpec) Grid() (music.DrumGrid, error) {
	out := music.DrumGrid{Division: g.Division, Voices: map[string]string{}, BasedOn: g.BasedOn}
	for _, l := range g.Voices {
		if _, dup := out.Voices[l.Voice]; dup {
			return music.DrumGrid{}, fmt.Errorf("drumGrid: voice %q is written twice; write each voice once", l.Voice)
		}
		out.Voices[l.Voice] = l.Steps
	}
	return out, nil
}

// Grid is the candidate's drum grid for the engine, nil when it names
// drums instead. It assumes Validate passed.
func (c Candidate) Grid() *music.DrumGrid {
	if c.DrumGrid == nil {
		return nil
	}
	g, err := c.DrumGrid.Grid()
	if err != nil {
		return nil
	}
	return &g
}

// styleDrums is what the style check holds to the style: the catalogue
// drums a grid is based on, or else the drums id (also when both are
// written, so the retry learns everything at once).
func (c Candidate) styleDrums() string {
	if c.DrumGrid != nil && c.DrumGrid.BasedOn != "" {
		return c.DrumGrid.BasedOn
	}
	return c.Drums
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

// StyleIDs lists the style catalogue's ids (the schema's enum).
func StyleIDs() []string {
	var ids []string
	for _, s := range pattern.Styles() {
		ids = append(ids, s.ID)
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
		step.Push = c.push(i)
		prog = append(prog, step)
	}
	return prog, nil
}
