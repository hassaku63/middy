package pattern

import (
	"fmt"
	"math"
	"slices"
	"strconv"

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

// SwingUnit is the note value a swing lengthens and shortens in pairs.
type SwingUnit string

// Swing units.
const (
	UnitEighth    SwingUnit = "eighth"
	UnitSixteenth SwingUnit = "sixteenth"
)

// Swing is how much the feel swings (BL-084): Ratio is the long note of
// each pair to the short one, from 1.0 (straight) to 3.0, and Unit the note
// value paired. The zero Swing means the tempo's default ratio on eighths.
type Swing struct {
	Ratio float64
	Unit  SwingUnit
}

// The swing ratios a feel may use: 1:1 is straight, 3:1 a dotted eighth
// and a sixteenth.
const (
	minSwingRatio = 1.0
	maxSwingRatio = 3.0
)

// DefaultSwingRatio is the ratio a tempo (in beats a minute) swings at by
// default: about 3:1 at 60, 2:1 at 160 and no less than 1.5:1 when fast,
// the tendency measured in jazz drummers (Friberg & Sundström 2002), in
// tenths.
func DefaultSwingRatio(beatBPM float64) float64 {
	r := 3.0 - (beatBPM-60)*0.01
	return math.Round(min(maxSwingRatio, max(1.5, r))*10) / 10
}

func (s Swing) validate() error {
	if s.Ratio != 0 && (s.Ratio < minSwingRatio || s.Ratio > maxSwingRatio) {
		return fmt.Errorf("pattern: swing ratio must be %g to %g, got %g", minSwingRatio, maxSwingRatio, s.Ratio)
	}
	if s.Unit != "" && s.Unit != UnitEighth && s.Unit != UnitSixteenth {
		return fmt.Errorf("pattern: swing unit must be %q or %q, got %q", UnitEighth, UnitSixteenth, s.Unit)
	}
	return nil
}

// resolved fills the defaults: the tempo's ratio, eighths.
func (s Swing) resolved(beatBPM float64) Swing {
	if s.Ratio == 0 {
		s.Ratio = DefaultSwingRatio(beatBPM)
	}
	if s.Unit == "" {
		s.Unit = UnitEighth
	}
	return s
}

// move swings one note: a unit on the first of a pair stretches to the
// long note, and a note starting on the second of a pair starts there
// instead, keeping its end. Anything else is untouched.
func (s Swing) move(start, dur music.Tick) (music.Tick, music.Tick) {
	unit := eighth
	if s.Unit == UnitSixteenth {
		unit = eighth / 2
	}
	pair := 2 * unit
	long := music.Tick(math.Round(float64(pair) * s.Ratio / (s.Ratio + 1)))
	switch start % pair {
	case 0:
		if dur == unit {
			return start, long
		}
	case unit:
		if shift := long - unit; dur > shift {
			return start + shift, dur - shift
		}
	}
	return start, dur
}

// applyFeel swings every track and adds the feel to the explanation. The
// beat-placement counts stay those of the template: swing changes how
// the "and" is played, not which beats are hit.
func applyFeel(tl music.Timeline, info FeelInfo, sw Swing) music.Timeline {
	if info.ID == FeelStraight {
		return tl
	}
	sw = sw.resolved(tl.TempoBPM)
	tracks := make([]music.Track, len(tl.Tracks))
	for i, tr := range tl.Tracks {
		notes := make([]music.NoteEvent, len(tr.Notes))
		for j, n := range tr.Notes {
			n.Start, n.Duration = sw.move(n.Start, n.Duration)
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
		Ratio:       sw.Ratio,
		Unit:        string(sw.Unit),
		Explanation: swingText(sw, tl.TempoBPM),
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
	// Swing sets how much a swing feel swings (BL-084); the zero value is
	// the tempo's default ratio on eighths.
	Swing Swing
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
	if err := c.Swing.validate(); err != nil {
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
	out, err := apply(tl, acc, drums, st.tensions())
	if err != nil {
		return music.Timeline{}, err
	}
	return withGrid(applyStyle(applyFeel(out, info, c.Swing), st, acc), c.Grid), nil
}

// withGrid records the grid the drums were written as (BL-123), so a
// rebuild can send it back.
func withGrid(tl music.Timeline, g *music.DrumGrid) music.Timeline {
	if g == nil || tl.Rhythm == nil {
		return tl
	}
	rhythm := *tl.Rhythm
	rhythm.DrumGrid = g
	tl.Rhythm = &rhythm
	return tl
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
	if st.Tensions != nil && coloured(tl, st.Tensions, acc) {
		rs.Tensions = st.Name + "のスタイルなので、" + st.Tensions.Text + "（調の音階にある音だけ）。ローマ数字と分析は 7 度までのまま。"
	}
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

// swingText explains the swing: which notes, how far, and the tendency
// the default ratio follows.
func swingText(sw Swing, beatBPM float64) string {
	unit := "8 分"
	if sw.Unit == UnitSixteenth {
		unit = "16 分"
	}
	ratio := strconv.FormatFloat(sw.Ratio, 'f', -1, 64)
	text := fmt.Sprintf("%sを 2 つずつ組にし、裏の%sを少し遅らせて、長・短をおよそ %s : 1 にしている（2 : 1 なら 3 連符の 3 つ目）。跳ねるように揺れて聞こえる。", unit, unit, ratio)
	if sw.Ratio == DefaultSwingRatio(beatBPM) {
		text += "遅いテンポほど深く、速いほど浅く跳ねるのがジャズの傾向で、この比はテンポに合わせている。"
	}
	return text
}

// coloured is true when the style's tensions sound: the pattern plays the
// chord voice and some chord takes a tension.
func coloured(tl music.Timeline, t *Tensions, acc Accompaniment) bool {
	if !slices.ContainsFunc(acc.Hits, func(h Hit) bool { return h.Voice == VoiceChord }) {
		return false
	}
	return slices.ContainsFunc(tl.Harmony, func(h music.HarmonyEvent) bool { return len(tensionsOf(t, tl.Key, h.Chord)) > 0 })
}
