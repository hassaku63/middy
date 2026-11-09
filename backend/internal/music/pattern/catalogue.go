package pattern

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hassaku63/middy/backend/internal/music"
)

// The catalogue is data the engine renders (#207): versioned with the
// code, read at start-up and checked as a whole, so a broken entry fails
// every test and never reaches a user. Moving it out of Go tables does not
// move the line of ADR-0003: the patterns are still the engine's, not the
// model's. Accompaniments live in data/accompaniments.json and drums in
// data/drums.json, one hit per line; styles, which name the patterns that
// suit them, in data/styles.json.
// An accompaniment without a meter fits any, and a hit of it with no
// duration lasts to the end of whatever bar it plays in.

//go:embed data/accompaniments.json
var accompanimentsJSON []byte

//go:embed data/drums.json
var drumsJSON []byte

//go:embed data/styles.json
var stylesJSON []byte

// Catalogue is the patterns the engine can play.
type Catalogue struct {
	Accompaniments []Accompaniment `json:"accompaniments,omitempty"`
	Drums          []Drums         `json:"drums,omitempty"`
	Styles         []Style         `json:"styles,omitempty"`
}

// voices are the roles an accompaniment's hit may play.
var voices = map[Voice]bool{
	VoiceChord: true, VoiceBass: true, VoiceRoot: true, VoiceThird: true, VoiceFifth: true,
	VoiceTop: true, VoiceOctave: true, VoiceDrone: true, VoiceWalk: true,
}

// Validate reports every problem of the catalogue at once.
func (c Catalogue) Validate() error {
	var errs []error
	seen := map[string]bool{}
	for i, a := range c.Accompaniments {
		where := fmt.Sprintf("accompaniments[%d] %q", i, a.ID)
		errs = append(errs, entryProblems(where, a.ID, a.Name, a.Description, len(a.Hits), seen["acc:"+a.ID])...)
		seen["acc:"+a.ID] = true
		errs = append(errs, accompanimentProblems(where, a)...)
	}
	for i, d := range c.Drums {
		where := fmt.Sprintf("drums[%d] %q", i, d.ID)
		errs = append(errs, entryProblems(where, d.ID, d.Name, d.Description, len(d.Hits), seen["drums:"+d.ID])...)
		seen["drums:"+d.ID] = true
		errs = append(errs, drumProblems(where, d)...)
	}
	errs = append(errs, c.styleProblems()...)
	return errors.Join(errs...)
}

// styleProblems checks the styles: what every entry needs, that every
// pattern a style names is in the catalogue and every pattern but block
// belongs to a style, known feels, and an instrument named with its
// program. Grooves are worked out by the engine, not written.
func (c Catalogue) styleProblems() []error {
	var errs []error
	known, used, seen := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, a := range c.Accompaniments {
		known["acc:"+a.ID] = true
	}
	for _, d := range c.Drums {
		known["drums:"+d.ID] = true
	}
	for i, st := range c.Styles {
		where := fmt.Sprintf("styles[%d] %q", i, st.ID)
		errs = append(errs, entryProblems(where, st.ID, st.Name, st.Description, len(st.Feels), seen[st.ID])...)
		seen[st.ID] = true
		errs = append(errs, oneStyleProblems(where, st, known, used)...)
	}
	if len(c.Styles) == 0 {
		return errs // a catalogue of patterns alone has no styles to belong to
	}
	for ref := range known {
		if ref != "acc:block" && !used[ref] {
			errs = append(errs, fmt.Errorf("%s belongs to no style", ref))
		}
	}
	return errs
}

// oneStyleProblems checks one style, marking the patterns it names as used.
func oneStyleProblems(where string, st Style, known, used map[string]bool) []error {
	var errs []error
	for _, ref := range append(prefixed("acc:", st.Accompaniments), prefixed("drums:", st.Drums)...) {
		if !known[ref] {
			errs = append(errs, fmt.Errorf("%s: names %s, which is not in the catalogue", where, ref))
		}
		used[ref] = true
	}
	for _, f := range st.Feels {
		if _, err := findFeel(f); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", where, err))
		}
	}
	if (st.Program == nil) != (st.Instrument == "") || st.Program != nil && (*st.Program < 0 || *st.Program > 127) {
		errs = append(errs, fmt.Errorf("%s: an instrument needs both a program in 0..127 and its name", where))
	}
	if st.Tensions != nil && st.Tensions.Text == "" {
		errs = append(errs, fmt.Errorf("%s: tensions need a text", where))
	}
	if len(st.Grooves) > 0 {
		errs = append(errs, fmt.Errorf("%s: grooves are worked out, not written", where))
	}
	return errs
}

func prefixed(prefix string, ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = prefix + id
	}
	return out
}

// accompanimentProblems checks the hits against the template's bar; one
// written for any meter is measured against a 4/4 bar.
func accompanimentProblems(where string, a Accompaniment) []error {
	length := bar
	if a.Meter != nil {
		if err := music.SupportedMeter(*a.Meter); err != nil {
			return []error{fmt.Errorf("%s: %w", where, err)}
		}
		length = music.BarTicks(*a.Meter, music.DefaultPPQ)
	}
	var errs []error
	for j, h := range a.Hits {
		// Without a meter, a zero duration lasts to the end of whatever bar.
		toBarEnd := a.Meter == nil && h.Duration == 0
		if h.Start < 0 || h.Start+h.Duration > length || h.Duration <= 0 && !toBarEnd {
			errs = append(errs, fmt.Errorf("%s: hits[%d] at %d for %d does not fit a bar of %d", where, j, h.Start, h.Duration, length))
		}
		errs = append(errs, velocityProblem(where, j, h.Velocity)...)
		if !voices[h.Voice] {
			errs = append(errs, fmt.Errorf("%s: hits[%d] voice %q is not a voice", where, j, h.Voice))
		}
	}
	return errs
}

// drumProblems checks that the figure declares a meter the engine plays
// and that every hit is a note inside its bar.
func drumProblems(where string, d Drums) []error {
	if d.Meter == nil {
		return []error{fmt.Errorf("%s: drums must declare their meter", where)}
	}
	if err := music.SupportedMeter(*d.Meter); err != nil {
		return []error{fmt.Errorf("%s: %w", where, err)}
	}
	length := music.BarTicks(*d.Meter, music.DefaultPPQ)
	var errs []error
	for j, h := range d.Hits {
		if h.Start < 0 || h.Duration <= 0 || h.Start+h.Duration > length {
			errs = append(errs, fmt.Errorf("%s: hits[%d] at %d for %d does not fit a bar of %d", where, j, h.Start, h.Duration, length))
		}
		if !h.Note.Valid() {
			errs = append(errs, fmt.Errorf("%s: hits[%d] note %d is not a MIDI note", where, j, h.Note))
		}
		errs = append(errs, velocityProblem(where, j, h.Velocity)...)
	}
	return errs
}

func velocityProblem(where string, j, velocity int) []error {
	if velocity < 1 || velocity > 127 {
		return []error{fmt.Errorf("%s: hits[%d] velocity %d is outside 1..127", where, j, velocity)}
	}
	return nil
}

// entryProblems checks what every entry needs: an id of its own, a name,
// a description and at least one hit.
func entryProblems(where, id, name, description string, hits int, seen bool) []error {
	var errs []error
	if id == "" {
		errs = append(errs, fmt.Errorf("%s: id is required", where))
	}
	if seen {
		errs = append(errs, fmt.Errorf("%s: id is used twice", where))
	}
	if name == "" {
		errs = append(errs, fmt.Errorf("%s: name is required", where))
	}
	if description == "" {
		errs = append(errs, fmt.Errorf("%s: description is required", where))
	}
	if hits == 0 {
		errs = append(errs, fmt.Errorf("%s: hits are required", where))
	}
	return errs
}

// decodeCatalogue reads a catalogue, refusing fields it does not know.
func decodeCatalogue(data []byte) (Catalogue, error) {
	var c Catalogue
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Catalogue{}, fmt.Errorf("pattern: catalogue: %w", err)
	}
	return c, nil
}

// The catalogue, read once at start-up and checked as a whole; a broken
// one stops the program before it serves.
var accompaniments, drums, styles = mustLoadCatalogue()

func mustLoadCatalogue() ([]Accompaniment, []Drums, []Style) {
	var c Catalogue
	var errs []error
	for _, data := range [][]byte{accompanimentsJSON, drumsJSON, stylesJSON} {
		part, err := decodeCatalogue(data)
		errs = append(errs, err)
		c.Accompaniments = append(c.Accompaniments, part.Accompaniments...)
		c.Drums = append(c.Drums, part.Drums...)
		c.Styles = append(c.Styles, part.Styles...)
	}
	err := errors.Join(errs...)
	if err == nil {
		err = c.Validate()
	}
	if err != nil {
		panic(fmt.Sprintf("pattern: the built-in catalogue is broken: %v", err))
	}
	return c.Accompaniments, c.Drums, c.Styles
}
