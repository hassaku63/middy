package music

import (
	"fmt"
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
}

var modeCharacter = map[Mode]string{
	Major:      "明るく安定して聞こえる。",
	Minor:      "自然短音階で、暗く内向きに聞こえる。",
	Dorian:     "短調に似ているが、6 度の音が半音高いので、暗さの中に明るさが混ざる。",
	Mixolydian: "長調に似ているが、7 度の音が半音低く導音がないので、開放的に聞こえる。",
}

// reference is the mode a modal key is heard against: dorian against
// natural minor, mixolydian against major. Each differs from it by the
// characteristic tone alone.
var reference = map[Mode]Mode{Dorian: Minor, Mixolydian: Major}

var modeName = map[Mode]string{Major: "長調", Minor: "短調", Dorian: "ドリアン", Mixolydian: "ミクソリディアン"}

// overview needs the per-chord analysis, so Analyze calls it last.
func overview(tl Timeline) *Overview {
	if len(tl.Harmony) == 0 {
		return nil
	}
	return &Overview{Whole: whole(tl), Flow: flow(tl)}
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
				joinChords(used), Key{Tonic: k.Tonic, Mode: Minor})
		}
	case k.Mode == Minor:
		if used := distinct(tl, func(h HarmonyEvent) bool { return isHarmonicMinorDominant(k, h.Chord) }); len(used) > 0 {
			s += fmt.Sprintf("%sは和声的短音階の形で、導音が主和音へ強く引き寄せる。", joinChords(used))
		}
		if used := distinct(tl, func(h HarmonyEvent) bool { return h.Analysis != nil && h.Analysis.Borrowed }); len(used) > 0 {
			s += fmt.Sprintf("この例では %sが短調の外から借りた和音（同主長調の I・IV、ナポリの ♭II）。", joinChords(used))
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
		if h.Analysis == nil {
			continue
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
	s := fmt.Sprintf("この例では %sが特徴音 %s を含む。", joinChords(used), k.PitchClassName(tone))
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
		s += fmt.Sprintf("%sから始まり、", chordName(first))
	}
	prev := h[len(h)-2]
	if cad := cadenceOfEvent(last); cad != NoCadence {
		s += fmt.Sprintf("最後は %s から %s への%s。", prev.RomanNumeral, last.RomanNumeral, endingText[cad])
	} else if degreeOfEvent(last) == 1 {
		s += "最後は主和音に戻って落ち着く。"
	} else {
		s += fmt.Sprintf("最後は %sで、主和音に戻らずに終わる。", chordName(last))
	}
	// A loop closes again when it repeats: say what the turn-around makes.
	if wrap := cadenceBetween(tl.Key.Mode, placedOf(last), placedOf(first)); wrap != NoCadence && degreeOfEvent(first) == 1 {
		s += fmt.Sprintf("繰り返すと、%s から頭の %s への%sになる。", last.RomanNumeral, first.RomanNumeral, cadenceName[wrap])
	}
	return s
}

var endingText = map[Cadence]string{
	Authentic: "正格終止で閉じる",
	Plagal:    "変格終止で、柔らかく閉じる",
	Modal:     "モーダル終止で、導音を使わずに主音へ戻る",
	Deceptive: "偽終止で、解決を先延ばしにしたまま終わる",
	Half:      "半終止で止まり、続きを感じさせる",
}

var cadenceName = map[Cadence]string{
	Authentic: "正格終止", Plagal: "変格終止", Modal: "モーダル終止", Deceptive: "偽終止", Half: "半終止",
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
	ThreeFour: "3 拍子（3/4）で、1 拍目が強く 2・3 拍目が弱い。ワルツのように 1 小節ごとに回る。",
}

func degreeOfEvent(h HarmonyEvent) int {
	if h.Analysis == nil {
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

func chordName(h HarmonyEvent) string { return fmt.Sprintf("%s（%s）", h.RomanNumeral, h.Symbol) }

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

func joinChords(hs []HarmonyEvent) string {
	names := make([]string, len(hs))
	for i, h := range hs {
		names[i] = fmt.Sprintf("%s（%s）", h.RomanNumeral, h.Symbol)
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
