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
// data/drums.json, one hit per line; styles are still a Go table (BL-179).
// An accompaniment without a meter fits any, and a hit of it with no
// duration lasts to the end of whatever bar it plays in.

//go:embed data/accompaniments.json
var accompanimentsJSON []byte

//go:embed data/drums.json
var drumsJSON []byte

// Catalogue is the patterns the engine can play.
type Catalogue struct {
	Accompaniments []Accompaniment `json:"accompaniments,omitempty"`
	Drums          []Drums         `json:"drums,omitempty"`
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
	return errors.Join(errs...)
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
var accompaniments, drums = mustLoadCatalogue()

func mustLoadCatalogue() ([]Accompaniment, []Drums) {
	var c Catalogue
	var errs []error
	for _, data := range [][]byte{accompanimentsJSON, drumsJSON} {
		part, err := decodeCatalogue(data)
		errs = append(errs, err)
		c.Accompaniments = append(c.Accompaniments, part.Accompaniments...)
		c.Drums = append(c.Drums, part.Drums...)
	}
	err := errors.Join(errs...)
	if err == nil {
		err = c.Validate()
	}
	if err != nil {
		panic(fmt.Sprintf("pattern: the built-in catalogue is broken: %v", err))
	}
	return c.Accompaniments, c.Drums
}
