package music

import (
	"fmt"
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
	Degree      int      `json:"degree"`
	Function    Function `json:"function"`
	Cadence     Cadence  `json:"cadence,omitempty"`
	Explanation string   `json:"explanation"`
}

// Analyze returns a copy of the timeline whose Harmony events carry an
// Analysis: scale degree, harmonic function, the cadence formed with the
// previous chord, and a short Japanese explanation. Every chord must be
// diatonic to the timeline's key. The input is not modified.
func Analyze(tl Timeline) (Timeline, error) {
	out := tl
	out.Harmony = make([]HarmonyEvent, len(tl.Harmony))
	copy(out.Harmony, tl.Harmony)

	prevDegree := 0
	for i := range out.Harmony {
		h := &out.Harmony[i]
		degree, ok := degreeOf(tl.Key, h.Chord.Root)
		if !ok {
			return Timeline{}, fmt.Errorf("music: harmony[%d]: chord %s is not diatonic to %s", i, tl.Key.ChordSymbol(h.Chord), tl.Key)
		}
		fn := functionByDegree[degree]
		cad := cadenceBetween(tl.Key.Mode, prevDegree, degree)
		h.Analysis = &Analysis{
			Degree:      degree,
			Function:    fn,
			Cadence:     cad,
			Explanation: explain(tl.Key, *h, fn, cad),
		}
		prevDegree = degree
	}
	return out, nil
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

// cadenceBetween classifies the motion from the previous degree (0 when
// there is none) into the current one.
func cadenceBetween(mode Mode, prev, cur int) Cadence {
	switch {
	case prev == 0:
		return NoCadence
	case prev == 7 && cur == 1 && mode.Characteristic() != 0:
		return Modal
	case prev == 5 && cur == 1:
		return Authentic
	case prev == 4 && cur == 1:
		return Plagal
	case prev == 5 && cur == 6:
		return Deceptive
	case prev != 5 && cur == 5:
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

func explain(key Key, h HarmonyEvent, fn Function, cad Cadence) string {
	s := fmt.Sprintf("%s（%s）は%s", h.RomanNumeral, key.ChordSymbol(h.Chord), functionText[fn])
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
