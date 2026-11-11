package pattern

import (
	"fmt"
	"slices"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Style binds the patterns and feels that suit a kind of music (ADR-0012):
// the model picks a style first, and Misfits rejects patterns that do not
// belong to it, so "a Celtic reel on the eight-beat kit" is stopped by
// the engine, not left to the prompt. Block chords and no drums suit every
// style.
type Style struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Accompaniments []string `json:"accompaniments"`
	Drums          []string `json:"drums"`
	Feels          []Feel   `json:"feels"`
	// Program, when set, is the GM instrument the style plays a
	// keyboard-default pattern on (BL-110); Instrument names it.
	Program    *int   `json:"program,omitempty"`
	Instrument string `json:"instrument,omitempty"`
	// Grooves is the style's own groove in each of GrooveMeters (BL-087).
	Grooves []Groove `json:"grooves"`
	// Tensions, when set, is how the style colours its chords (BL-125).
	Tensions *Tensions `json:"tensions,omitempty"`
}

// GrooveMeters are the meters a style lists a groove for.
var GrooveMeters = []music.TimeSignature{music.FourFour, music.ThreeFour, music.SixEight, music.TwelveEight}

// Groove is what a style plays in a meter when the learner picks the style
// for an example that already has its harmony (BL-087): the first of the
// style's accompaniments, drums and feels that the meter takes, preferring
// a figure written for the meter to one that fits any (a drone). The
// catalogue lists the style's signature patterns first, so this is the
// style's own sound, chosen by the engine and not by the model.
type Groove struct {
	Meter         music.TimeSignature `json:"meter"`
	Accompaniment string              `json:"accompaniment"` // "block" when none is written for the meter
	Drums         string              `json:"drums,omitempty"`
	Feel          Feel                `json:"feel"`
}

// Groove picks the style's groove in the meter.
func (s Style) Groove(ts music.TimeSignature) Groove {
	g := Groove{Meter: ts, Accompaniment: "block", Feel: FeelStraight}
	for _, id := range s.Accompaniments {
		a, err := FindAccompaniment(id)
		if err != nil || !a.Fits(ts) {
			continue
		}
		if a.Meter != nil { // written for this meter: the style's figure
			g.Accompaniment = id
			break
		}
		if g.Accompaniment == "block" {
			g.Accompaniment = id // fits any meter; kept unless a figure follows
		}
	}
	for _, id := range s.Drums {
		if d, err := FindDrums(id); err == nil && d.Fits(ts) {
			g.Drums = id
			break
		}
	}
	for _, f := range s.Feels {
		if !f.Swings() || !ts.Compound() { // swing, shuffle and swing-16 are for simple meters
			g.Feel = f
			break
		}
	}
	return g
}

// Choice is the groove as a Render choice in the style.
func (g Groove) Choice(style string) Choice {
	return Choice{Accompaniment: g.Accompaniment, Drums: g.Drums, Feel: g.Feel, Style: style}
}

func program(p int) *int { return &p }

// tensions is the style's colouring; none without a style.
func (s *Style) tensions() *Tensions {
	if s == nil {
		return nil
	}
	return s.Tensions
}

// Styles lists the style catalogue, each with its grooves.
func Styles() []Style {
	out := slices.Clone(styles)
	for i := range out { // callers must not change the catalogue
		out[i].Accompaniments = slices.Clone(out[i].Accompaniments)
		out[i].Drums = slices.Clone(out[i].Drums)
		out[i].Feels = slices.Clone(out[i].Feels)
		if p := out[i].Program; p != nil {
			out[i].Program = program(*p)
		}
		if t := out[i].Tensions; t != nil {
			copied := *t // callers must not change the catalogue's rule
			out[i].Tensions = &copied
		}
		out[i].Grooves = nil
		for _, ts := range GrooveMeters {
			out[i].Grooves = append(out[i].Grooves, out[i].Groove(ts))
		}
	}
	return out
}

// FindStyle returns the style with the id, a copy the caller may change.
func FindStyle(id string) (Style, error) {
	for _, s := range Styles() {
		if s.ID == id {
			return s, nil
		}
	}
	return Style{}, fmt.Errorf("%w: style %q", ErrUnknownPattern, id)
}

// Misfits lists what does not belong to the style, each with what would:
// an accompaniment ("" and "block" always fit), drums ("" always fits) and
// a feel ("" is straight).
func (s Style) Misfits(acc, drums string, feel Feel) []string {
	var out []string
	if acc != "" && acc != "block" && !slices.Contains(s.Accompaniments, acc) {
		out = append(out, fmt.Sprintf("accompaniment %q does not fit the %s style; use one of %v or block", acc, s.ID, s.Accompaniments))
	}
	if drums != "" && !slices.Contains(s.Drums, drums) {
		out = append(out, fmt.Sprintf("drums %q does not fit the %s style; use one of %v or none", drums, s.ID, s.Drums))
	}
	if feel == "" {
		feel = FeelStraight
	}
	if !slices.Contains(s.Feels, feel) {
		out = append(out, fmt.Sprintf("feel %q does not fit the %s style; use one of %v", feel, s.ID, s.Feels))
	}
	return out
}
