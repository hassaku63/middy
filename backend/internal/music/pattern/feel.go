package pattern

import (
	"fmt"
	"slices"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Feel is how evenly the eighth notes are played. It is a parameter on
// top of any template (BL-076): the LLM picks it, the engine applies it.
type Feel string

// Feels.
const (
	FeelStraight Feel = "straight" // even eighths; the default
	FeelSwing    Feel = "swing"    // long-short eighths, a 2:1 triplet pair
)

// FeelInfo names and explains a feel for the prompt and the panel.
type FeelInfo struct {
	ID          Feel   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var feels = []FeelInfo{
	{FeelStraight, "ストレート", "8 分音符を均等に刻む。"},
	{FeelSwing, "スウィング", "8 分音符を長・短で跳ねさせる。ジャズやシャッフルの揺れ。"},
}

// Feels lists the feels in a stable order.
func Feels() []FeelInfo {
	out := make([]FeelInfo, len(feels))
	copy(out, feels)
	return out
}

func findFeel(f Feel) (FeelInfo, error) {
	if f == "" {
		f = FeelStraight
	}
	for _, info := range feels {
		if info.ID == f {
			return info, nil
		}
	}
	return FeelInfo{}, fmt.Errorf("%w: feel %q", ErrUnknownPattern, f)
}

// swungAnd is where the "and" of a beat lands in swing: two thirds of the
// beat, the last note of a triplet.
const swungAnd = quarter * 2 / 3

// swing moves one note: an eighth on the beat stretches to the swung
// "and", and a note starting on the "and" starts there instead, keeping
// its end. Anything else (quarters, sixteenths) is untouched.
func swing(start, dur music.Tick) (music.Tick, music.Tick) {
	switch start % quarter {
	case 0:
		if dur == eighth {
			return start, swungAnd
		}
	case eighth:
		shift := swungAnd - eighth
		if dur > shift {
			return start + shift, dur - shift
		}
	}
	return start, dur
}

// applyFeel swings every track and adds the feel to the explanation. The
// beat-placement counts stay those of the template: swing changes how
// the "and" is played, not which beats are hit.
func applyFeel(tl music.Timeline, info FeelInfo) music.Timeline {
	if info.ID == FeelStraight {
		return tl
	}
	tracks := make([]music.Track, len(tl.Tracks))
	for i, tr := range tl.Tracks {
		notes := make([]music.NoteEvent, len(tr.Notes))
		for j, n := range tr.Notes {
			n.Start, n.Duration = swing(n.Start, n.Duration)
			notes[j] = n
		}
		tr.Notes = notes
		tracks[i] = tr
	}
	tl.Tracks = tracks
	rhythm := *tl.Rhythm
	rhythm.Feel = &music.RhythmFeel{
		ID:          string(info.ID),
		Name:        info.Name,
		Explanation: "裏拍（8 分の 2 つ目）を少し遅らせて、3 連符の 3 つ目に置いている。長・短の組み合わせで、跳ねるように揺れて聞こえる。",
	}
	tl.Rhythm = &rhythm
	return tl
}

// Choice is what a caller asks for by id: an accompaniment, drums, and a
// feel. Empty ids mean block chords and no drums.
type Choice struct {
	Accompaniment string
	Drums         string
	Feel          Feel
	// Style, when set, may choose the chords' instrument (BL-110) and is
	// recorded on the timeline. Render does not check that the patterns
	// fit it: the intent does, and a caller may mix on purpose.
	Style string
	// Grid, when set, writes the drums out instead of naming them
	// (BL-122); Drums must then be empty.
	Grid *music.DrumGrid
}

// Render layers a Choice on a block-chord timeline. With no accompaniment
// and no drums the timeline is returned unchanged (a feel alone has
// nothing to swing); drums alone keep block chords.
func Render(tl music.Timeline, c Choice) (music.Timeline, error) {
	info, err := findFeel(c.Feel)
	if err != nil {
		return music.Timeline{}, err
	}
	st, err := styleOf(c.Style)
	if err != nil {
		return music.Timeline{}, err
	}
	if c.Accompaniment == "" && c.Drums == "" && c.Grid == nil && (st == nil || st.Program == nil) {
		return tl, nil // block chords on the piano: nothing to render
	}
	accID := c.Accompaniment
	if accID == "" {
		accID = "block"
	}
	acc, err := FindAccompaniment(accID)
	if err != nil {
		return music.Timeline{}, err
	}
	drums, err := choiceDrums(c, tl)
	if err != nil {
		return music.Timeline{}, err
	}
	if info.ID == FeelSwing && tl.TimeSignature.Compound() {
		ts := tl.TimeSignature
		return music.Timeline{}, fmt.Errorf("pattern: swing is for simple meters; %d/%d already divides each beat in three", ts.Numerator, ts.Denominator)
	}
	out, err := Apply(tl, acc, drums)
	if err != nil {
		return music.Timeline{}, err
	}
	return applyStyle(applyFeel(out, info), st, acc), nil
}

// choiceDrums finds the named drums or renders the grid; nil for none.
func choiceDrums(c Choice, tl music.Timeline) (*Drums, error) {
	switch {
	case c.Grid != nil && c.Drums != "":
		return nil, fmt.Errorf("pattern: choose drums %q or a grid, not both", c.Drums)
	case c.Grid != nil:
		d, err := GridDrums(*c.Grid, tl.TimeSignature, tl.TempoBPM)
		if err != nil {
			return nil, err
		}
		return &d, nil
	case c.Drums != "":
		d, err := FindDrums(c.Drums)
		if err != nil {
			return nil, err
		}
		return &d, nil
	}
	return nil, nil
}

// styleOf finds the choice's style; nil when it names none.
func styleOf(id string) (*Style, error) {
	if id == "" {
		return nil, nil
	}
	st, err := FindStyle(id)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// applyStyle records the style on the timeline and, when the style has an
// instrument and the pattern plays on the keyboard default, puts the
// chords on that instrument (BL-110). A pattern with an instrument of its
// own (a guitar strum, the drone's organ) keeps it.
func applyStyle(tl music.Timeline, st *Style, acc Accompaniment) music.Timeline {
	if st == nil || tl.Rhythm == nil {
		return tl
	}
	rs := &music.RhythmStyle{ID: st.ID, Name: st.Name}
	if st.Program != nil && acc.Program == 0 && len(tl.Tracks) > 0 {
		tracks := slices.Clone(tl.Tracks)
		tracks[0].Program = *st.Program
		tl.Tracks = tracks
		rs.Instrument = st.Instrument
	}
	rhythm := *tl.Rhythm
	rhythm.Style = rs
	tl.Rhythm = &rhythm
	return tl
}
