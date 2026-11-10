package music

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Overview explains an analysed example as a whole (BL-098): Whole says
// what its mode sounds like and which chords make it so; Flow says how the
// progression runs and where it closes. The explanation panel shows both
// above the per-chord notes. Like the rest of the analysis, the prose is
// deterministic (ADR-0003).
type Overview struct {
	Whole string `json:"whole"`
	Flow  string `json:"flow"`
	// Degrees are the seven diatonic triads of the key's mode, in degree
	// order, for the degree map (BL-099).
	Degrees []DegreeChord `json:"degrees"`
}

// DegreeChord is one diatonic triad of the example's mode on the degree
// map: where it sits, what it is called in this key, and its function.
type DegreeChord struct {
	Degree       int      `json:"degree"`
	RomanNumeral string   `json:"romanNumeral"`
	Symbol       string   `json:"symbol"`
	Function     Function `json:"function"`
	// Used is true when the example plays a chord on this degree as part
	// of the key: a seventh or minor's harmonic V counts, a borrowed or
	// applied chord does not.
	Used bool `json:"used,omitempty"`
	// Altered lists the chords the example plays on this degree whose tones
	// leave the mode's scale, as spelled ("B7" for E minor's harmonic V),
	// so the map shows what sounds rather than only the diatonic triad
	// (#296). A seventh or suspension inside the scale is not altered.
	Altered []string `json:"altered,omitempty"`
	// Role is what the chord does in a mode (#366, ADR-0047); empty in
	// major and minor.
	Role ModalRole `json:"role,omitempty"`
}

var modeCharacter = map[Mode]string{
	Major:      "明るく安定して聞こえる。",
	Minor:      "自然短音階で、暗く内向きに聞こえる。",
	Dorian:     "短調に似ているが、音階の第 6 音が半音高いので、暗さの中に明るさが混ざる。",
	Mixolydian: "長調に似ているが、音階の第 7 音が半音低く導音（主音の半音下の音）がないので、開放的に聞こえる。",
	Phrygian:   "短調に似ているが、音階の第 2 音が半音低いので、さらに暗く、スペインのフラメンコのような張りつめた響きになる。",
}

// reference is the mode a modal key is heard against: dorian and phrygian
// against natural minor, mixolydian against major. Each differs from it by
// the characteristic tone alone.
var reference = map[Mode]Mode{Dorian: Minor, Mixolydian: Major, Phrygian: Minor}

var modeName = map[Mode]string{Major: "長調", Minor: "短調", Dorian: "ドリアン", Mixolydian: "ミクソリディアン", Phrygian: "フリジアン"}

// overview needs the per-chord analysis, so Analyze calls it last.
func overview(tl Timeline) *Overview {
	if len(tl.Harmony) == 0 {
		return nil
	}
	return &Overview{Whole: whole(tl), Flow: flow(tl), Degrees: degrees(tl)}
}

// degrees lays out the key's seven triads and marks the ones the example
// plays as part of the key.
func degrees(tl Timeline) []DegreeChord {
	var used [8]bool
	var altered [8][]string
	for _, h := range tl.Harmony {
		if a := h.Analysis; a != nil && !a.Borrowed && a.AppliedTo == 0 && a.Degree >= 1 && a.Degree <= 7 {
			used[a.Degree] = true
			if sym := symbolOf(tl.Key, h); leavesScale(tl.Key, h.Chord) && !slices.Contains(altered[a.Degree], sym) {
				altered[a.Degree] = append(altered[a.Degree], sym)
			}
		}
	}
	out := make([]DegreeChord, 0, 7)
	for d := 1; d <= 7; d++ {
		c, err := tl.Key.Diatonic(d, false)
		if err != nil {
			continue // every mode stacks a triad on each degree; unreachable
		}
		numeral, err := RomanNumeral(d, c.Quality)
		if err != nil {
			continue
		}
		out = append(out, DegreeChord{
			Degree: d, RomanNumeral: numeral, Symbol: tl.Key.ChordSymbol(c),
			Function: functionByDegree[d], Used: used[d], Altered: altered[d], Role: modalRole(tl.Key, d),
		})
	}
	return out
}

// leavesScale reports whether a chord has a tone outside the key's scale.
func leavesScale(k Key, c Chord) bool {
	for _, pc := range c.PitchClasses() {
		if !k.Contains(pc) {
			return true
		}
	}
	return false
}

func whole(tl Timeline) string {
	k := tl.Key
	s := fmt.Sprintf("%s（%s）。%s", k, modeName[k.Mode], modeCharacter[k.Mode])
	switch {
	case k.Mode.Characteristic() != 0:
		s += characteristicSentence(tl)
	case k.Mode == Major:
		if used := distinct(tl, func(h HarmonyEvent) bool { return h.Analysis != nil && h.Analysis.Borrowed }); len(used) > 0 {
			s += fmt.Sprintf("この例では %sが同主短調（%s）から借りた和音で、明るさの中に陰りを差す。",
				joinChords(k, used), Key{Tonic: k.Tonic, Mode: Minor})
		}
	case k.Mode == Minor:
		if used := distinct(tl, func(h HarmonyEvent) bool { return isHarmonicMinorDominant(k, h.Chord) }); len(used) > 0 {
			s += fmt.Sprintf("%sは和声的短音階の形で、導音（主音の半音下の音）が主和音へ強く引き寄せる。", joinChords(k, used))
		}
		if used := distinct(tl, func(h HarmonyEvent) bool { return h.Analysis != nil && h.Analysis.Borrowed }); len(used) > 0 {
			s += fmt.Sprintf("この例では %sが短調の外から借りた和音（同主長調の I・IV、ナポリの ♭II）。", joinChords(k, used))
		}
	}
	return s
}

// characteristicSentence names the chords that carry the mode's
// characteristic tone and what each would be in the reference mode, or
// says that none does, so the example sounds like the reference.
func characteristicSentence(tl Timeline) string {
	k := tl.Key
	deg := k.Mode.Characteristic()
	tone, _ := k.Degree(deg)
	ref := Key{Tonic: k.Tonic, Mode: reference[k.Mode]}
	used := distinct(tl, func(h HarmonyEvent) bool { _, ok := characteristicTone(k, h.Chord); return ok })
	if len(used) == 0 {
		return fmt.Sprintf("この例の和音は特徴音（%s）を含まないので、響きは%sとほとんど変わらない。",
			k.PitchClassName(tone), modeName[ref.Mode])
	}
	var instead []string
	for _, h := range used {
		if h.Analysis == nil || h.Analysis.AppliedTo != 0 {
			continue // a secondary dominant has no counterpart in the other mode
		}
		c, err := ref.Diatonic(h.Analysis.Degree, len(h.Chord.Quality.Intervals()) == 4)
		if err != nil {
			continue
		}
		rn, err := RomanNumeral(h.Analysis.Degree, c.Quality)
		if err != nil {
			continue
		}
		instead = append(instead, fmt.Sprintf("%s（%s）", rn, ref.ChordSymbol(c)))
	}
	s := fmt.Sprintf("この例では %sが特徴音 %s を含む。", joinChords(k, used), k.PitchClassName(tone))
	if len(instead) == 1 {
		s += fmt.Sprintf("%sなら、ここは %sになる。", modeName[ref.Mode], instead[0])
	} else if len(instead) > 1 {
		s += fmt.Sprintf("%sなら、それぞれ %sになる。", modeName[ref.Mode], strings.Join(instead, "と "))
	}
	return s
}

func flow(tl Timeline) string {
	h := tl.Harmony
	first, last := h[0], h[len(h)-1]
	s := meterText[tl.TimeSignature] + fmt.Sprintf("%d 個の和音で %s 小節。", len(h), bars(tl))
	if len(h) == 1 {
		return s
	}
	if degreeOfEvent(first) == 1 {
		s += "主和音から始まり、"
	} else {
		s += fmt.Sprintf("%sから始まり、", chordName(tl.Key, first))
	}
	prev := h[len(h)-2]
	if cad := cadenceOfEvent(last); cad != NoCadence {
		s += fmt.Sprintf("最後は %s から %s への%s。", prev.RomanNumeral, last.RomanNumeral, endingText[cad])
	} else if degreeOfEvent(last) == 1 {
		s += "最後は主和音に戻って落ち着く。"
	} else {
		s += fmt.Sprintf("最後は %sで、主和音に戻らずに終わる。", chordName(tl.Key, last))
	}
	s += lineText(tl.Key, h)
	// A loop closes again when it repeats: say what the turn-around makes.
	if wrap := cadenceBetween(tl.Key.Mode, placedOf(last), placedOf(first)); wrap != NoCadence && degreeOfEvent(first) == 1 {
		s += fmt.Sprintf(wrapText[wrap], last.RomanNumeral, first.RomanNumeral)
	}
	return s
}

// wrapText says how a loop closes when it repeats, without naming a
// cadence (#378, ADR-0052): the head's tonic begins the next phrase, it
// is no arrival, and the last chord has its own cadence already.
var wrapText = map[Cadence]string{
	Authentic: "繰り返すと、最後の %s は頭の %s へ解決し、ひと回りが閉じる。",
	Plagal:    "繰り返すと、最後の %s から頭の %s へ柔らかく戻り、ひと回りが閉じる。",
	Modal:     "繰り返すと、最後の %s から頭の %s へ、導音を使わずに戻る。",
}

var endingText = map[Cadence]string{
	Authentic: "正格終止で閉じる",
	Plagal:    "変格終止で、柔らかく閉じる",
	Modal:     "モーダル終止で、導音を使わずに主和音へ戻る",
	Deceptive: "偽終止で、解決を先延ばしにしたまま終わる",
	Half:      "半終止で止まり、続きを感じさせる",
}

// bars is the length in bars of the time signature, without a trailing .0.
func bars(tl Timeline) string {
	var end Tick
	for _, h := range tl.Harmony {
		if e := h.Start + h.Duration; e > end {
			end = e
		}
	}
	perBar := BarTicks(tl.TimeSignature, tl.PPQ)
	if perBar <= 0 {
		return "?"
	}
	return strconv.FormatFloat(float64(end)/float64(perBar), 'f', -1, 64)
}

// meterText describes a meter other than the default 4/4 (ADR-0017).
var meterText = map[TimeSignature]string{
	ThreeFour:   "3 拍子（3/4）で、1 拍目が強く 2・3 拍目が弱い。ワルツのように 1 小節ごとに回る。",
	SixEight:    "6/8 拍子で、付点 4 分の 2 拍に数え、それぞれの拍が 8 分音符 3 つに分かれる。1 拍目が強く、ゆりかごのように 1 小節に 2 回揺れる。",
	TwelveEight: "12/8 拍子で、付点 4 分の 4 拍に数え、それぞれの拍が 8 分音符 3 つに分かれる。1 拍目と 3 拍目が強く、スロー・ブルースや 3 連のバラードのように、3 連符のように揺れながらゆったり進む。",
}

// degreeOfEvent is the chord's scale degree; 0 for a secondary dominant,
// whose root is not the function it has (V7/IV on the tonic's root pulls
// away from home).
func degreeOfEvent(h HarmonyEvent) int {
	if h.Analysis == nil || h.Analysis.AppliedTo != 0 {
		return 0
	}
	return h.Analysis.Degree
}

func cadenceOfEvent(h HarmonyEvent) Cadence {
	if h.Analysis == nil {
		return NoCadence
	}
	return h.Analysis.Cadence
}

func chordName(k Key, h HarmonyEvent) string {
	return fmt.Sprintf("%s（%s）", h.RomanNumeral, symbolOf(k, h))
}

// distinct returns the first event of each roman numeral that matches.
func distinct(tl Timeline, match func(HarmonyEvent) bool) []HarmonyEvent {
	seen := map[string]bool{}
	var out []HarmonyEvent
	for _, h := range tl.Harmony {
		if match(h) && !seen[h.RomanNumeral] {
			seen[h.RomanNumeral] = true
			out = append(out, h)
		}
	}
	return out
}

func joinChords(k Key, hs []HarmonyEvent) string {
	names := make([]string, len(hs))
	for i, h := range hs {
		names[i] = chordName(k, h)
	}
	return strings.Join(names, "と ")
}

// placedOf rebuilds a chord's place in the key from its analysis.
func placedOf(h HarmonyEvent) placed {
	if h.Analysis == nil {
		return placed{}
	}
	a := h.Analysis
	return placed{degree: a.Degree, borrowed: a.Borrowed, applied: a.AppliedTo,
		leading: a.AppliedTo == 0 && hasLeadingTone(a.Degree, h.Chord)}
}
