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
// which is the usual beginner-level simplification. The harmonic-minor V,
// borrowed and applied chords are placed before this table is read.
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
	AppliedTo int `json:"appliedTo,omitempty"`
	// Substitute is true when that dominant is a tritone substitute
	// (subV7/x, BL-128).
	Substitute bool `json:"substitute,omitempty"`
	// LeadingTone is true when the applied chord is the diminished seventh
	// on the target's leading tone (vii°7/x, BL-130).
	LeadingTone bool   `json:"leadingTone,omitempty"`
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
				Substitute:  p.sub,
				LeadingTone: p.lead,
				Explanation: explainApplied(tl.Key, *h, p, next(tl.Harmony, i)) + passingText(tl.Harmony, i, p),
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
			Explanation: explain(tl.Key, *h, fn, cad, p.borrowed) + subtonicText(tl.Key, p, prev, following) + bluesText(tl.Key, p, following) + majorAlias(tl.Key, p.degree, h.Chord),
		}
	}
	for i := range out.Harmony {
		out.Harmony[i].Analysis.Explanation += openText(out.Harmony, i) + bassText(tl.Key, out.Harmony, i) + pushText(out, i)
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
	sub      bool // that dominant is a tritone substitute (subV7/x)
	lead     bool // the applied chord is a leading-tone diminished seventh (vii°7/x)
	leading  bool // a V that carries the leading tone (major triad or dominant seventh)
	blues    bool // I7 or IV7 in a major key, written so (BL-135)
}

func place(k Key, hs []HarmonyEvent) ([]placed, error) {
	out := make([]placed, len(hs))
	for i, h := range hs {
		if d, ok := bluesDegree(k, h); ok { // before placeApplied: C7 in C is V7/IV only when written so
			out[i] = placed{degree: d, blues: true}
			continue
		}
		if p, ok := placeApplied(k, h); ok {
			out[i] = p
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

// placeApplied places an applied chord: a secondary dominant, its tritone
// substitute or a leading-tone diminished seventh. The same sound can be
// two chords: C major in C minor is the borrowed I (a Picardy third) or
// V/iv, the dominant of iv; Bb7 in C major is the borrowed bVII7 or
// subV7/vi. The label says which the author meant; an unlabelled chord is
// applied only when nothing in the key explains it.
func placeApplied(k Key, h HarmonyEvent) (placed, bool) {
	chordLabel, _, _ := splitBass(h.RomanNumeral)
	at := func(target int, sub, lead bool) placed {
		d, _ := degreeOf(k, h.Chord.Root)
		return placed{degree: d, applied: target, sub: sub, lead: lead}
	}
	if strings.HasPrefix(chordLabel, "subV") {
		if target, ok := substituteTargetOf(k, h.Chord); ok {
			return at(target, true, false), true
		}
	}
	if target, ok := leadingTargetOf(k, h.Chord); ok {
		return at(target, false, true), true // no key's own chord is a diminished seventh
	}
	if strings.Contains(chordLabel, "/") {
		if target, ok := appliedTargetOf(k, h.Chord); ok {
			return at(target, false, false), true
		}
	}
	if target, ok := locateApplied(k, h.Chord); ok {
		return at(target, false, false), true
	}
	if target, ok := locateSubstitute(k, h.Chord); ok {
		return at(target, true, false), true
	}
	return placed{}, false
}

// bluesDegree places a dominant seventh written I7 or IV7 in a major key:
// the blues dominant, which keeps its function (BL-135, ADR-0026). The
// same chord written V7/IV, or unlabelled, is left to placeApplied.
func bluesDegree(k Key, h HarmonyEvent) (int, bool) {
	if h.Chord.Quality != Dominant7 {
		return 0, false
	}
	chordLabel, _, _ := splitBass(h.RomanNumeral)
	degree, _, upper, suffix, accidental, err := parseLabel(chordLabel)
	if err != nil || !bluesDominant(k, degree, upper, suffix, accidental) {
		return 0, false
	}
	if d, ok := degreeOf(k, h.Chord.Root); ok && d == degree {
		return d, true
	}
	return 0, false
}

// bluesText explains a blues dominant: the seventh it adds and that it
// does not resolve. On IV that seventh is the key's flat third, the blue
// note. An I7 that moves on to the key's own IV is not said to stay put:
// its seventh falls a semitone to IV's third, as V7/IV's would (ADR-0034,
// decision 2: the explanation says what the chord actually does).
func bluesText(k Key, p placed, following placed) string {
	if !p.blues {
		return ""
	}
	if p.degree == 4 {
		seventh := k.spellDegree(3, -1) // the seventh of IV is the key's lowered third
		return fmt.Sprintf(" ブルースの属七。IV に %s を重ねた属七で、%s は調の ♭3（ブルー・ノート）にあたる。どこかへ解決する和音ではなく、サブドミナントのまま鳴る。", seventh, seventh)
	}
	seventh := k.spellDegree(7, -1)
	if following.degree == 4 && following.applied == 0 && !following.borrowed {
		return fmt.Sprintf(" ブルースの属七。主和音に %s（♭7）を重ねた属七で、機能はトニックのまま。次が IV なので、%s が IV の 3 度の %s へ半音で下り、同じ音の V7/IV のように IV を引き寄せて聞こえる。", seventh, seventh, k.spellDegree(6, 0))
	}
	return fmt.Sprintf(" ブルースの属七。主和音に %s（♭7）を重ねた属七で、解決を求めずトニックのまま鳴る。V7/IV と同じ音だが、IV へ向かう働きではない。", seventh)
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
	// A secondary dominant's degree is only its root's: V7/VI on III is no III.
	prevDeg, nextDeg := prev.degree, following.degree
	if prev.applied != 0 {
		prevDeg = 0
	}
	if following.applied != 0 {
		nextDeg = 0
	}
	switch {
	case prevDeg == 1 && nextDeg == 6:
		return " ここでは機能というより、i から VI、V へ下りていく線の途中（経過）。"
	case nextDeg == 3:
		return " ここでは次の III へ向かう。III を一時的な主音とみれば、その V にあたる（V/III の働き）。"
	case nextDeg == 1:
		return " ここでは導音を持たないまま i へ戻る。V ほど強くない、弱いドミナント。"
	}
	return ""
}

// locate finds the chord's scale degree in the key, or, in a major key, in
// the parallel minor it was borrowed from (same degree, possibly a lowered
// root). A diatonic root with the parallel minor's quality (Fm on IV in C
// major) is borrowed too.
func locate(k Key, c Chord) (degree int, borrowed bool, ok bool) {
	if c.Quality.isOpen() { // checked note by note when the label was read (openStep)
		if d, ok := degreeOf(k, c.Root); ok {
			return d, false, true
		}
		if d, ok := degreeOf(Key{Tonic: k.Tonic, Mode: Minor}, c.Root); ok && k.Mode == Major {
			return d, true, true
		}
		return 0, false, false
	}
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
	default: // the modes borrow nothing (ADR-0012)
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
// (also ♭VII7). Other borrowings (♭II and ♭III in major) are not supported.
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
	case prev.degree == 0:
		return NoCadence
	case prev.applied != 0:
		// A secondary dominant is named for what it pulls towards, not
		// for its root's degree (V/VII is not a IV); arriving on V from
		// one (V/V – V) is still a half cadence.
		if cur.leading {
			return Half
		}
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
	Phrygian:   "この和音はフリジアンの特徴音（主音から短 2 度の %s）を含む。主音のすぐ上に半音があるので、短調よりさらに暗く、張りつめて聞こえる理由。",
}

// phrygianSecondText is said of phrygian's major chord on the second
// degree (BL-134), the chord that marks the mode most strongly.
const phrygianSecondText = "主音の半音上の長三和音（♭II）は、フリジアンをいちばん強く印象づける和音。ここから主和音へ根音が半音で下りる動きは、フリジアン終止と呼ばれる。"

// isPhrygianSecond reports the major triad (or its seventh) on phrygian's
// second degree.
func isPhrygianSecond(key Key, c Chord) bool {
	return key.Mode == Phrygian && c.Root.Normalize() == key.Tonic.Add(1).Normalize() &&
		(c.Quality == MajorTriad || c.Quality == Major7)
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
	if isPhrygianSecond(key, h.Chord) {
		s += " " + phrygianSecondText
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

// locateSubstitute reports a tritone substitute among the chords nothing
// else explains (an unlabelled one; a label decides first, in place).
func locateSubstitute(k Key, c Chord) (int, bool) {
	if _, _, ok := locate(k, c); ok {
		return 0, false
	}
	return substituteTargetOf(k, c)
}

// next returns the chord after i, if any.
func next(hs []HarmonyEvent, i int) *Chord {
	if i+1 >= len(hs) {
		return nil
	}
	return &hs[i+1].Chord
}

const (
	appliedText = "%s（%s）は、%s（%s）へ向かうセカンダリードミナント。%s を一時的な主音のように聞かせて、そこへ強く引き寄せる。"
	// %s（%s）: the substitute; %s（%s）: the target; %s（%s）: the dominant it replaces; %s: the target.
	substituteText  = "%s（%s）は、%s（%s）へ向かう裏コード（トライトーン代理）。%s（%s）の代わりに、根音が三全音離れた属七を置いている。同じ三全音（3 度と 7 度）を持つので同じように引き寄せ、根音は半音上から %s の根音へ降りる。"
	leadingText     = "%s（%s）は、%s（%s）の導音の上に積んだ減七。根音が %s の根音へ半音で上がり、そこへ引き寄せる。"
	passingSentence = "前の和音から根音が半音ずつ上がる、経過の減七（パッシング・ディミニッシュ）。" //nolint:gosec // G101 reads "passing" as a password; this is explanation text
	resolvesText    = "次の %s へ解決する。"
	deflectedText   = "ここでは %s へは進まず、引き寄せをかわす。"
)

func explainApplied(k Key, h HarmonyEvent, p placed, following *Chord) string {
	target := p.applied
	t, _ := k.Diatonic(target, false)
	rn, _ := RomanNumeral(target, t.Quality)
	symbol := k.ChordSymbol(t)
	s := fmt.Sprintf(appliedText, h.RomanNumeral, symbolOf(k, h), rn, symbol, symbol)
	if p.sub {
		dominant := "V7"
		if target != 1 {
			dominant, _ = k.appliedNumeral(target, true, dominantOf)
		}
		s = fmt.Sprintf(substituteText, h.RomanNumeral, symbolOf(k, h), rn, symbol, dominant, k.spellApplied(target, dominantOf)+"7", symbol)
	}
	if p.lead {
		s = fmt.Sprintf(leadingText, h.RomanNumeral, symbolOf(k, h), rn, symbol, symbol)
	}
	switch {
	case following == nil:
	case following.Root == t.Root:
		s += " " + fmt.Sprintf(resolvesText, rn)
	default:
		s += " " + fmt.Sprintf(deflectedText, rn)
	}
	return s
}

// passingText names a leading-tone diminished seventh that passes: the
// root climbs a semitone from the chord before and another to the next
// (I – vii°7/ii – ii). "" for any other chord.
func passingText(hs []HarmonyEvent, i int, p placed) string {
	if !p.lead || i == 0 || i+1 >= len(hs) {
		return ""
	}
	root := hs[i].Chord.Root
	if hs[i-1].Chord.Root.Add(1) != root || root.Add(1) != hs[i+1].Chord.Root {
		return ""
	}
	return " " + passingSentence
}

const (
	susText        = " 3 度の代わりに 4 度を鳴らす掛留（sus4）。長調とも短調とも決まらない、宙に浮いた響きになる。"
	susResolveText = "次の和音で 4 度が 3 度へ下がって解決する。"
	powerText      = " 3 度を抜き、根音と 5 度だけを鳴らすパワーコード。長調か短調かを決める 3 度がないので、開いた力強い響きになる（ロックのギターに多い）。"
)

// openText explains an open chord (BL-132): a suspension, which resolves
// when the next chord on the same root brings its third back, or a power
// chord. "" for any other chord.
func openText(hs []HarmonyEvent, i int) string {
	c := hs[i].Chord
	switch c.Quality {
	case Power:
		return powerText
	case Sus4, Dominant7Sus4:
		if i+1 < len(hs) && hs[i+1].Chord.Root == c.Root && !hs[i+1].Chord.Quality.isOpen() {
			return susText + susResolveText
		}
		return susText
	default: // a chord with a third: nothing to add here
	}
	return ""
}
