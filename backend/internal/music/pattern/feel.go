package pattern

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Feel is how evenly the eighth notes are played. It is a parameter on
// top of any template (BL-076): the LLM picks it, the engine applies it.
type Feel string

// Feels.
const (
	FeelStraight Feel = "straight" // even eighths; the default
	FeelSwing    Feel = "swing"    // long-short eighths, as deep as the tempo calls for
	// FeelShuffle is the blues and rock shuffle (#308, ADR-0036): eighths
	// swung on the triplet, 2:1, whatever the tempo.
	FeelShuffle Feel = "shuffle"
	// FeelSwing16 is the sixteenth swing of hip-hop and lo-fi beats
	// (BL-243, ADR-0060): sixteenths swung lightly, about 57 %, whatever
	// the tempo.
	FeelSwing16 Feel = "swing-16"
)

// shuffleRatio is the shuffle's long-short: the first and third notes of
// a triplet.
const shuffleRatio = 2.0

// swing16Ratio is the sixteenth swing's long-short: 1.3 : 1, the long
// sixteenth about 57 % of the pair, within the 54–58 % a drum machine's
// swing is set to for hip-hop (the owner's choice, #308).
const swing16Ratio = 1.3

// Swings reports whether the feel delays the offbeats, which a compound
// meter does not take: its beat is already divided in three.
func (f Feel) Swings() bool { return f == FeelSwing || f == FeelShuffle || f == FeelSwing16 }

// FeelInfo names and explains a feel for the prompt and the panel.
type FeelInfo struct {
	ID          Feel   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var feels = []FeelInfo{
	{FeelStraight, "ストレート", "8 分音符を均等に刻む。"},
	{FeelSwing, "スウィング", "8 分音符を長・短で跳ねさせる。跳ね方はテンポに合わせる（遅いほど深い）。ジャズの揺れ。"},
	{FeelShuffle, "シャッフル", "8 分音符を 3 連符の 1 つ目と 3 つ目で弾く（長・短が 2 : 1）。ブルースやロックの「タッカ・タッカ」と跳ねる揺れ。"},
	{FeelSwing16, "16 分のスウィング", "16 分音符を長・短でわずかに跳ねさせる（長・短がおよそ 1.3 : 1、組の 57 %）。跳ねるとは聞こえない程度にノリがほどける。ヒップホップやローファイのビートの揺れ。"},
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

// move swings one note by warping time within each pair of units: the
// first half of the pair stretches to the long note and the second half
// shrinks to the short one. A unit on the first of a pair becomes the long
// note, a note starting on the second starts late, a held note ending on
// the second (the chord before a push) ends there too, and finer notes
// (sixteenths under eighth swing) keep their order and spacing.
func (s Swing) move(start, dur music.Tick) (music.Tick, music.Tick) {
	unit := eighth
	if s.Unit == UnitSixteenth {
		unit = eighth / 2
	}
	pair := 2 * unit
	long := music.Tick(math.Round(float64(pair) * s.Ratio / (s.Ratio + 1)))
	warp := func(t music.Tick) music.Tick {
		pos := t % pair
		if pos <= unit {
			return t - pos + pos*long/unit
		}
		return t - pos + long + (pos-unit)*(pair-long)/unit
	}
	from, to := warp(start), warp(start+dur)
	return from, to - from
}

// applyFeel swings every track and adds the feel to the explanation. The
// beat-placement counts stay those of the template: swing changes how
// the "and" is played, not which beats are hit.
func applyFeel(tl music.Timeline, info FeelInfo, sw Swing) music.Timeline {
	if info.ID == FeelStraight {
		return tl
	}
	if info.ID == FeelShuffle && sw.Ratio == 0 {
		sw.Ratio = shuffleRatio
	}
	if info.ID == FeelSwing16 {
		sw.Unit = UnitSixteenth
		if sw.Ratio == 0 {
			sw.Ratio = swing16Ratio
		}
	}
	sw = sw.resolved(tl.TempoBPM)
	finer := sw.Unit == UnitEighth && finerThanEighths(tl.Tracks)
	tracks := make([]music.Track, len(tl.Tracks))
	beat := music.BeatTicks(tl.TimeSignature, tl.PPQ)
	for i, tr := range tl.Tracks {
		if onTriplets(tr.Notes, beat) {
			tracks[i] = tr
			continue
		}
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
		Explanation: feelText(info.ID, sw, tl.TempoBPM) + finerText(finer, tl.Harmony),
	}
	tl.Rhythm = &rhythm
	return tl
}

// finerThanEighths reports notes between the eighths (sixteenths) on a
// track that is not in triplets.
func finerThanEighths(tracks []music.Track) bool {
	return slices.ContainsFunc(tracks, func(tr music.Track) bool {
		return !onTriplets(tr.Notes, 2*eighth) && slices.ContainsFunc(tr.Notes, func(n music.NoteEvent) bool { return n.Start%eighth != 0 })
	})
}

// finerText says what a swing of eighths does to finer notes (#404 S8):
// it stretches them with their pair, so sixteenths go long, long, short,
// short, and a chord pushed a sixteenth comes in just before the beat,
// sooner than a straight sixteenth.
func finerText(finer bool, harmony []music.HarmonyEvent) string {
	s := ""
	if finer {
		s += "16 分の音も 8 分の組と一緒に伸び縮みするので、前の 2 つが長く、後の 2 つが短くなる。"
	}
	if slices.ContainsFunc(harmony, func(h music.HarmonyEvent) bool { return h.Push == music.PushSixteenth }) {
		s += "16 分の押しも縮んだ側に入るので、拍のすぐ手前で鳴る。"
	}
	return s
}

// onTriplets reports a track already written in triplets (a grid of three
// steps a beat): it swings by itself, and warping it by eighths would pull
// the triplets apart, off the shuffled eighths of the other parts (#404
// S3). A note between the eighths that falls on a triplet marks it.
func onTriplets(notes []music.NoteEvent, beat music.Tick) bool {
	return slices.ContainsFunc(notes, func(n music.NoteEvent) bool {
		pos := n.Start % beat
		return pos%(beat/2) != 0 && pos%(beat/3) == 0
	})
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
	// Tensions, when set, is the colour the candidate asked for in place of
	// its style's: TensionsNone, TensionsNinths or TensionsNinths13
	// (ADR-0037).
	Tensions string
}

// Render layers a Choice on a block-chord timeline. With no accompaniment,
// drums or grid, and a style with neither an instrument nor tensions, the
// notes are returned unchanged with only the style recorded (a feel alone
// has nothing to swing); drums alone, or such a style alone, keep block
// chords.
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
	tensions, err := chosenTensions(c, st)
	if err != nil {
		return music.Timeline{}, err
	}
	if plain(c, st, tensions) {
		return styleOnly(tl, st), nil // block chords on the piano: nothing to render
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
	if info.ID.Swings() && tl.TimeSignature.Compound() {
		ts := tl.TimeSignature
		return music.Timeline{}, fmt.Errorf("pattern: %s is for simple meters; %d/%d already divides each beat in three", info.ID, ts.Numerator, ts.Denominator)
	}
	if acc.ID == "block" && drums == nil && !slices.ContainsFunc(tl.Harmony, func(h music.HarmonyEvent) bool { return h.Push != music.PushNone }) {
		info, _ = findFeel(FeelStraight) // held block chords, unpushed, have no eighths to swing
	}
	out, err := apply(tl, acc, drums, tensions)
	if err != nil {
		return music.Timeline{}, err
	}
	out = applyStyle(applyFeel(out, info, c.Swing), st, acc, c.Tensions == "")
	return withGrid(withTensions(namedTensions(out, tensions), c.Tensions, tensions), c.Grid), nil
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

// plain is true when the choice changes nothing on block chords: no
// patterns, and no style instrument or colouring.
func plain(c Choice, st *Style, tensions *Tensions) bool {
	if c.Accompaniment != "" || c.Drums != "" || c.Grid != nil || tensions != nil {
		return false
	}
	return st == nil || st.Program == nil
}

// chosenTensions is the colour the chords take: the one the candidate
// asked for (ADR-0037), else its style's rule.
func chosenTensions(c Choice, st *Style) (*Tensions, error) {
	if c.Tensions != "" {
		return requestedTensions(c.Tensions)
	}
	return st.tensions(), nil
}

// withTensions records the colour the candidate asked for (ADR-0037) and
// what it added, so the notes say so and a rebuild sends it back.
func withTensions(tl music.Timeline, id string, t *Tensions) music.Timeline {
	if id == "" || tl.Rhythm == nil {
		return tl
	}
	rt := &music.RhythmTensions{ID: id}
	if added := soundedTensions(tl, t); added != "" {
		rt.Explanation = "この例の色付けとして、" + added + "（調の音階にある音だけ）。ローマ数字と分析は 7 度までのまま。"
	}
	rhythm := *tl.Rhythm
	rhythm.Tensions = rt
	tl.Rhythm = &rhythm
	return tl
}

// styleOnly records the style on a timeline it does not otherwise change,
// so a rebuild in another key keeps it.
func styleOnly(tl music.Timeline, st *Style) music.Timeline {
	if st == nil {
		return tl
	}
	rhythm := music.Rhythm{}
	if tl.Rhythm != nil {
		rhythm = *tl.Rhythm
	}
	rhythm.Style = &music.RhythmStyle{ID: st.ID, Name: st.Name}
	tl.Rhythm = &rhythm
	return tl
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
func applyStyle(tl music.Timeline, st *Style, acc Accompaniment, styleColour bool) music.Timeline {
	if st == nil || tl.Rhythm == nil {
		return tl
	}
	rs := &music.RhythmStyle{ID: st.ID, Name: st.Name}
	if added := soundedTensions(tl, st.Tensions); added != "" && styleColour {
		rs.Tensions = st.Name + "のスタイルなので、" + added + "（調の音階にある音だけ）。ローマ数字と分析は 7 度までのまま。"
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

// feelText explains the feel: the shuffle by its triplet, a swing by
// swingText.
func feelText(f Feel, sw Swing, beatBPM float64) string {
	if f == FeelShuffle && sw.Ratio == shuffleRatio && sw.Unit == UnitEighth {
		return "シャッフル。8 分音符を 2 つずつ組にし、3 連符の 1 つ目と 3 つ目で弾く（長・短が 2 : 1）。「タッカ・タッカ」と跳ね、ブルースやロックの重いノリになる。テンポが変わっても跳ね方は変わらない。"
	}
	if f == FeelSwing16 && sw.Ratio == swing16Ratio {
		return "16 分のスウィング。16 分音符を 2 つずつ組にし、裏の 16 分を少しだけ遅らせて、長・短をおよそ 1.3 : 1（組の 57 %）にしている。8 分と拍の頭は動かない。跳ねるとは聞こえない程度にノリがほどけ、ヒップホップやローファイのビートの揺れになる。テンポが変わっても揺れ方は変わらない。"
	}
	return swingText(sw, beatBPM)
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

// soundedTensions says which of the style's tensions sounded, read from
// the notes the timeline plays (BL-306, #449): a tension the key and the
// chord allowed (a ii7–Imaj7 adds no 13th, whatever the rule says) that a
// pitched note sounds inside the chord. A note running past the chord's
// end is the next chord pushed early, not this one's colour. Empty when
// none sounded.
func soundedTensions(tl music.Timeline, t *Tensions) string {
	var ninth7, thirteenth, add9 bool
	for _, h := range tl.Harmony {
		for _, pc := range sounding(tl, h, tensionsOf(t, tl.Key, h)) {
			switch {
			case h.Chord.Quality == music.MajorTriad:
				add9 = true
			case pc == h.Chord.Root.Add(9):
				thirteenth = true
			default:
				ninth7 = true
			}
		}
	}
	var parts []string
	if add9 {
		parts = append(parts, "長三和音に 9 度を足して（add9）柔らかくし")
	}
	switch {
	case ninth7 && thirteenth:
		parts = append(parts, "七の和音に 9 度、属七には 13 度も足し")
	case ninth7:
		parts = append(parts, "七の和音に 9 度を足し")
	case thirteenth:
		parts = append(parts, "属七に 13 度を足し")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "、") + "ている"
}

// namedTensions writes on each chord the tensions its notes sound (BL-307):
// "add9" on a major triad, else "9" and "13", in that order.
func namedTensions(tl music.Timeline, t *Tensions) music.Timeline {
	if t == nil {
		return tl
	}
	harmony := slices.Clone(tl.Harmony)
	for i, h := range harmony {
		var names []string
		for _, pc := range sounding(tl, h, tensionsOf(t, tl.Key, h)) {
			switch {
			case h.Chord.Quality == music.MajorTriad:
				names = append(names, "add9")
			case pc == h.Chord.Root.Add(9):
				names = append(names, "13")
			default:
				names = append(names, "9")
			}
		}
		slices.Sort(names) // "13" before "9" by bytes; put the ninth first
		slices.Reverse(names)
		harmony[i].Tensions = names
	}
	tl.Harmony = harmony
	return tl
}

// sounding is those of the pitch classes that a pitched note plays inside
// the chord's span.
func sounding(tl music.Timeline, h music.HarmonyEvent, pcs []music.PitchClass) []music.PitchClass {
	var out []music.PitchClass
	for _, pc := range pcs {
		if slices.ContainsFunc(tl.Tracks, func(tr music.Track) bool {
			return tr.Channel != DrumChannel && slices.ContainsFunc(tr.Notes, func(n music.NoteEvent) bool {
				return n.Start >= h.Start && n.Start+n.Duration <= h.Start+h.Duration && n.Pitch.Class() == pc.Normalize()
			})
		}) {
			out = append(out, pc)
		}
	}
	return out
}
