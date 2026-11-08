package music

import (
	"fmt"
	"strings"
)

// Function is the harmonic function of a chord within its key, in the
// classic three-function model: tonic (rest), subdominant (departure),
// dominant (tension that pulls back to the tonic).
type Function string

// Harmonic functions. The JSON values are the single letters used in
// textbooks so the UI can show them directly.
const (
	Tonic       Function = "T"
	Subdominant Function = "S"
	Dominant    Function = "D"
)

// functionByDegree is the textbook assignment, used for both modes:
// I, III, VI are tonic; II, IV subdominant; V, VII dominant. In natural
// minor this makes v (minor) a weak dominant and VII a subtonic dominant,
// which is the usual beginner-level simplification; refinements (harmonic
// minor V, borrowed chords) are backlog material.
var functionByDegree = [8]Function{"", Tonic, Subdominant, Tonic, Subdominant, Dominant, Tonic, Dominant}

// Cadence names how the previous chord resolves into this one.
type Cadence string

// Cadence kinds. NoCadence is the empty string and is omitted from JSON.
const (
	NoCadence Cadence = ""
	Authentic Cadence = "authentic" // V -> I
	Plagal    Cadence = "plagal"    // IV -> I
	Half      Cadence = "half"      // anything (except V) -> V
	Deceptive Cadence = "deceptive" // V -> vi
	Modal     Cadence = "modal"     // VII -> I in dorian and mixolydian (no leading tone)
)

// Analysis is the explanation data attached to a HarmonyEvent. It is the
// raw material for the explanation panel (features/explanation); the prose is
// deterministic so that the UI can rely on it without an LLM.
type Analysis struct {
	Degree   int      `json:"degree"`
	Function Function `json:"function"`
	Cadence  Cadence  `json:"cadence,omitempty"`
	// Borrowed is true for a chord taken from the parallel minor (BL-094).
	Borrowed bool `json:"borrowed,omitempty"`
	// AppliedTo is the degree a secondary dominant pulls towards (BL-074);
	// zero for every other chord.
	AppliedTo   int    `json:"appliedTo,omitempty"`
	Explanation string `json:"explanation"`
}

// Analyze returns a copy of the timeline whose Harmony events carry an
// Analysis: scale degree, harmonic function, the cadence formed with the
// previous chord, and a short Japanese explanation. Every chord must be
// diatonic to the timeline's key. The input is not modified.
func Analyze(tl Timeline) (Timeline, error) {
	out := tl
	out.Harmony = make([]HarmonyEvent, len(tl.Harmony))
	copy(out.Harmony, tl.Harmony)

	ps, err := place(tl.Key, tl.Harmony)
	if err != nil {
		return Timeline{}, err
	}
	for i := range out.Harmony {
		h := &out.Harmony[i]
		p := ps[i]
		if p.applied != 0 {
			h.Analysis = &Analysis{
				Degree:      p.degree,
				Function:    Dominant,
				AppliedTo:   p.applied,
				Explanation: explainApplied(tl.Key, *h, p.applied, next(tl.Harmony, i)),
			}
			continue
		}
		var prev, following placed // the zero value means "none"
		if i > 0 {
			prev = ps[i-1]
		}
		if i+1 < len(ps) {
			following = ps[i+1]
		}
		fn := functionByDegree[p.degree]
		cad := cadenceBetween(tl.Key.Mode, prev, p)
		h.Analysis = &Analysis{
			Degree:      p.degree,
			Function:    fn,
			Cadence:     cad,
			Borrowed:    p.borrowed,
			Explanation: explain(tl.Key, *h, fn, cad, p.borrowed) + subtonicText(tl.Key, p, prev, following) + majorAlias(tl.Key, p.degree, h.Chord),
		}
	}
	for i := range out.Harmony {
		out.Harmony[i].Analysis.Explanation += bassText(tl.Key, out.Harmony, i) + pushText(out, i)
	}
	out.Overview = overview(out)
	return out, nil
}

// placed is a chord's position in the key, found for every chord before
// the analysis so that each can look at its neighbours.
type placed struct {
	degree   int
	borrowed bool
	applied  int  // the target degree of a secondary dominant, else 0
	leading  bool // a V that carries the leading tone (major triad or dominant seventh)
}

func place(k Key, hs []HarmonyEvent) ([]placed, error) {
	out := make([]placed, len(hs))
	for i, h := range hs {
		// The same sound can be two chords: C major in C minor is the
		// borrowed I (a Picardy third) or V/iv, the dominant of iv. The label
		// says which the author meant.
		if chordLabel, _, _ := splitBass(h.RomanNumeral); strings.Contains(chordLabel, "/") {
			if target, ok := appliedTargetOf(k, h.Chord); ok {
				d, _ := degreeOf(k, h.Chord.Root)
				out[i] = placed{degree: d, applied: target}
				continue
			}
		}
		if target, ok := locateApplied(k, h.Chord); ok {
			d, _ := degreeOf(k, h.Chord.Root)
			out[i] = placed{degree: d, applied: target}
			continue
		}
		d, borrowed, ok := locate(k, h.Chord)
		if !ok {
			return nil, fmt.Errorf("music: harmony[%d]: chord %s is not diatonic to %s", i, k.ChordSymbol(h.Chord), k)
		}
		out[i] = placed{degree: d, borrowed: borrowed, leading: hasLeadingTone(d, h.Chord)}
	}
	return out, nil
}

// hasLeadingTone reports a V that carries the leading tone, the major
// triad or dominant seventh on degree 5. The minor v of natural minor,
// dorian and mixolydian has none, so it makes no authentic, deceptive or
// half cadence (ADR-0012, decision 2).
func hasLeadingTone(degree int, c Chord) bool {
	return degree == 5 && (c.Quality == MajorTriad || c.Quality == Dominant7)
}

// subtonicText says what a minor key's VII does, from its neighbours
// (ADR-0012, decision 2): heading to III it acts as III's dominant, going
// back to i it is a weak dominant without a leading tone, and between i
// and VI it passes on the way down to V.
func subtonicText(k Key, p, prev, following placed) string {
	if k.Mode != Minor || p.degree != 7 || p.borrowed {
		return ""
	}
	switch {
	case prev.degree == 1 && following.degree == 6:
		return " ここでは機能というより、i から VI、V へ下りていく線の途中（経過）。"
	case following.degree == 3:
		return " ここでは次の III へ向かう。III を一時的な主音とみれば、その V にあたる（V/III の働き）。"
	case following.degree == 1:
		return " ここでは導音を持たないまま i へ戻る。V ほど強くない、弱いドミナント。"
	}
	return ""
}

// locate finds the chord's scale degree in the key, or, in a major key, in
// the parallel minor it was borrowed from (same degree, possibly a lowered
// root). A diatonic root with the parallel minor's quality (Fm on IV in C
// major) is borrowed too.
func locate(k Key, c Chord) (degree int, borrowed bool, ok bool) {
	if d, ok := degreeOf(k, c.Root); ok {
		if dia, err := k.Diatonic(d, len(c.Quality.Intervals()) == 4); err == nil && dia.Quality == c.Quality {
			return d, false, true
		}
		if isHarmonicMinorDominant(k, c) {
			return d, false, true
		}
	}
	switch k.Mode {
	case Major:
		parallel := Key{Tonic: k.Tonic, Mode: Minor}
		d, ok := degreeOf(parallel, c.Root)
		if !ok || !borrowable(d, c.Quality) {
			return 0, false, false
		}
		return d, true, true
	case Minor:
		if d, ok := minorBorrowedDegree(k, c); ok {
			return d, true, true
		}
	}
	return 0, false, false
}

// minorBorrowedDegree recognises a minor key's borrowed chords (BL-075):
// the major I and IV, and the Neapolitan bII, a major triad a semitone above
// the tonic.
func minorBorrowedDegree(k Key, c Chord) (int, bool) {
	if c.Quality != MajorTriad {
		return 0, false
	}
	switch c.Root.Normalize() {
	case k.Tonic.Normalize():
		return 1, true
	case k.Scale()[3]:
		return 4, true
	case k.Tonic.Add(1).Normalize():
		return 2, true
	}
	return 0, false
}

// isNeapolitan reports the minor key's bII.
func isNeapolitan(k Key, c Chord) bool {
	return k.Mode == Minor && c.Quality == MajorTriad && c.Root.Normalize() == k.Tonic.Add(1).Normalize()
}

// borrowable lists what a major key may take from its parallel minor today
// (BL-094): iv (also iv7, ivm6), ii° (also iiø7), ♭VI (also ♭VImaj7), ♭VII
// (also ♭VII7). The rest waits for BL-075.
func borrowable(degree int, q ChordQuality) bool {
	switch degree {
	case 4:
		return q == MinorTriad || q == Minor7 || q == MinorSixth
	case 2:
		return q == DiminishedTriad || q == HalfDiminished7
	case 6:
		return q == MajorTriad || q == Major7
	case 7:
		return q == MajorTriad || q == Dominant7
	}
	return false
}

// degreeOf returns the 1-based scale degree whose pitch class is pc.
func degreeOf(k Key, pc PitchClass) (int, bool) {
	pc = pc.Normalize()
	for i, s := range k.Scale() {
		if s == pc {
			return i + 1, true
		}
	}
	return 0, false
}

// cadenceBetween classifies the motion from the previous chord (degree 0
// when there is none) into the current one. Authentic, deceptive and half
// cadences need a V with the leading tone (hasLeadingTone).
func cadenceBetween(mode Mode, prev, cur placed) Cadence {
	switch {
	case prev.degree == 0, prev.applied != 0:
		// A secondary dominant is named for what it pulls towards, not
		// for its root's degree (V/VII is not a IV).
		return NoCadence
	case prev.degree == 7 && cur.degree == 1 && mode.Characteristic() != 0:
		return Modal
	case prev.leading && cur.degree == 1:
		return Authentic
	case prev.degree == 4 && cur.degree == 1:
		return Plagal
	case prev.leading && cur.degree == 6:
		return Deceptive
	case prev.degree != 5 && cur.leading:
		return Half
	default:
		return NoCadence
	}
}

var functionText = map[Function]string{
	Tonic:       "トニック（T）。調の中心で、落ち着いて聞こえる。",
	Subdominant: "サブドミナント（S）。中心から離れて、動きを作る。",
	Dominant:    "ドミナント（D）。トニックへ戻ろうとする緊張を作る。",
}

var cadenceText = map[Cadence]string{
	Authentic: "直前の V から I へ解決する正格終止。もっとも強い区切り。",
	Plagal:    "直前の IV から I へ落ち着く変格終止。「アーメン終止」とも呼ばれる、柔らかい区切り。",
	Half:      "V で止まる半終止。「続きがある」と感じさせる。",
	Deceptive: "V から I ではなく vi へ進む偽終止。解決を先延ばしにして意外性を出す。",
	Modal:     "直前の VII から主和音へ進むモーダル終止。導音を使わずに主音へ戻る、ロックやケルトに多い動き。",
}

// characteristicText explains why a chord carrying the mode's
// characteristic tone sounds the way it does. %s is the tone's name.
var characteristicText = map[Mode]string{
	Dorian:     "この和音はドリアンの特徴音（主音から長 6 度の %s）を含む。短調に似た響きの中で明るく聞こえる理由。",
	Mixolydian: "この和音はミクソリディアンの特徴音（主音から短 7 度の %s）を含む。導音がないので、長調より開放的に聞こえる理由。",
}

const harmonicMinorText = "短調の V を長三和音にするのは和声的短音階（第 7 音を半音上げる）の形で、導音がトニックへ強く引き寄せる。"

const (
	borrowedText      = "同主短調（%s）から借りた和音。長調の中に短調の陰りを一瞬差す。"
	minorBorrowedText = "同主長調（%s）から借りた和音。短調の中に一瞬明るさを差す。"
	picardyText       = "曲の終わりに置けば、短調を長三和音で閉じる「ピカルディの 3 度」になる。"
	neapolitanText    = "ナポリの和音（♭II）。主音の半音上の長三和音で、強い陰りを作り、V へ向かう前によく置かれる。"
)

// borrowedSentence says where a borrowed chord comes from.
func borrowedSentence(key Key, h HarmonyEvent) string {
	switch {
	case key.Mode == Major:
		return fmt.Sprintf(borrowedText, Key{Tonic: key.Tonic, Mode: Minor})
	case isNeapolitan(key, h.Chord):
		return neapolitanText
	case h.Chord.Root.Normalize() == key.Tonic.Normalize():
		return fmt.Sprintf(minorBorrowedText, Key{Tonic: key.Tonic, Mode: Major}) + picardyText
	}
	return fmt.Sprintf(minorBorrowedText, Key{Tonic: key.Tonic, Mode: Major})
}

// symbolOf is the chord's spelled symbol, from the key when the event
// carries none (a timeline built elsewhere or stored before symbols).
func symbolOf(key Key, h HarmonyEvent) string {
	if h.Symbol != "" {
		return h.Symbol
	}
	return key.ChordSymbol(h.Chord)
}

func explain(key Key, h HarmonyEvent, fn Function, cad Cadence, borrowed bool) string {
	symbol := symbolOf(key, h)
	s := fmt.Sprintf("%s（%s）は%s", h.RomanNumeral, symbol, functionText[fn])
	if borrowed {
		s += " " + borrowedSentence(key, h)
	}
	if isHarmonicMinorDominant(key, h.Chord) {
		s += " " + harmonicMinorText
	}
	if tone, ok := characteristicTone(key, h.Chord); ok {
		s += " " + fmt.Sprintf(characteristicText[key.Mode], key.PitchClassName(tone))
	}
	if cad != NoCadence {
		s += " " + cadenceText[cad]
	}
	return s
}

// isHarmonicMinorDominant reports a major V / V7 in a minor key, which is
// not diatonic to natural minor and deserves a word in the explanation.
func isHarmonicMinorDominant(key Key, c Chord) bool {
	if key.Mode != Minor {
		return false
	}
	dominant, _ := key.Degree(5)
	return c.Root == dominant && (c.Quality == MajorTriad || c.Quality == Dominant7)
}

// characteristicTone returns the mode's characteristic pitch when the chord
// contains it.
func characteristicTone(key Key, c Chord) (PitchClass, bool) {
	deg := key.Mode.Characteristic()
	if deg == 0 {
		return 0, false
	}
	tone, _ := key.Degree(deg)
	for _, pc := range c.PitchClasses() {
		if pc == tone {
			return tone, true
		}
	}
	return 0, false
}

// locateApplied reports a secondary dominant: a chord that the key, its
// harmonic-minor V and its borrowing do not explain, a fifth above a major
// or minor triad (appliedTargetOf). Applied chords mark no cadence of their
// own: they depart, towards their target.
func locateApplied(k Key, c Chord) (int, bool) {
	if _, _, ok := locate(k, c); ok {
		return 0, false
	}
	return appliedTargetOf(k, c)
}

// next returns the chord after i, if any.
func next(hs []HarmonyEvent, i int) *Chord {
	if i+1 >= len(hs) {
		return nil
	}
	return &hs[i+1].Chord
}

const (
	appliedText   = "%s（%s）は、%s（%s）へ向かうセカンダリードミナント。%s を一時的な主音のように聞かせて、そこへ強く引き寄せる。"
	resolvesText  = "次の %s へ解決する。"
	deflectedText = "ここでは %s へは進まず、引き寄せをかわす。"
)

func explainApplied(k Key, h HarmonyEvent, target int, following *Chord) string {
	t, _ := k.Diatonic(target, false)
	rn, _ := RomanNumeral(target, t.Quality)
	symbol := k.ChordSymbol(t)
	s := fmt.Sprintf(appliedText, h.RomanNumeral, symbolOf(k, h), rn, symbol, symbol)
	switch {
	case following == nil:
	case following.Root == t.Root:
		s += " " + fmt.Sprintf(resolvesText, rn)
	default:
		s += " " + fmt.Sprintf(deflectedText, rn)
	}
	return s
}
