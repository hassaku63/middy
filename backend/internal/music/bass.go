package music

import (
	"fmt"
	"strings"
)

// Bass is the note a slash chord puts under its chord (ADR-0020): a scale
// degree of the key, raised or lowered by Accidental (+1 sharp, -1 flat).
// The chord keeps its identity; only the lowest note changes, as in an
// inversion (I/3), a pedal (IV/1) or a line cliché (i, i/#7, i/7, i/#6).
type Bass struct {
	Degree     int `json:"degree"`               // 1..7
	Accidental int `json:"accidental,omitempty"` // -1, 0 or +1
}

// suffix writes the bass as a label writes it: "/3", "/b7", "/#7".
func (b Bass) suffix() string {
	switch b.Accidental {
	case -1:
		return fmt.Sprintf("/b%d", b.Degree)
	case 1:
		return fmt.Sprintf("/#%d", b.Degree)
	}
	return fmt.Sprintf("/%d", b.Degree)
}

// pitchClass is the bass note in the key.
func (b Bass) pitchClass(key Key) PitchClass {
	return key.Scale()[b.Degree-1].Add(b.Accidental)
}

// splitBass separates a slash-bass suffix ("/3", "/b7", "/#7") from the
// chord label. A "/" followed by a roman numeral is a secondary dominant
// (V/V) and is left in the label; a "/" followed by anything else that is
// not a degree 1..7 is an error.
func splitBass(label string) (string, *Bass, error) {
	trimmed := strings.TrimSpace(label)
	i := strings.LastIndex(trimmed, "/")
	if i < 0 {
		return trimmed, nil, nil
	}
	head, tail := trimmed[:i], trimmed[i+1:]
	acc := 0
	switch {
	case strings.HasPrefix(tail, "b"):
		acc, tail = -1, tail[1:]
	case strings.HasPrefix(tail, "#"):
		acc, tail = 1, tail[1:]
	}
	if len(tail) == 1 && tail[0] >= '0' && tail[0] <= '9' {
		d := int(tail[0] - '0')
		if d < 1 || d > 7 {
			return "", nil, fmt.Errorf("music: %q: the bass after \"/\" must be a scale degree 1..7, got %d", trimmed, d)
		}
		return head, &Bass{Degree: d, Accidental: acc}, nil
	}
	if acc != 0 || tail == "" || (tail[0] >= '0' && tail[0] <= '9') {
		return "", nil, fmt.Errorf("music: %q: the bass after \"/\" must be a scale degree 1..7 (optionally b or #)", trimmed)
	}
	return trimmed, nil, nil // a secondary dominant such as V/V
}

// bassOf is the note a chord stands on: its slash bass, or its root.
func bassOf(h HarmonyEvent) PitchClass {
	if h.Bass != nil {
		return *h.Bass
	}
	return h.Chord.Root
}

// bassName spells a chord's bass as its symbol does: the part after "/"
// for a slash chord ("E" in "C/E"), else the root ("Bb" in "Bbmaj7").
func bassName(key Key, h HarmonyEvent) string {
	s := h.Symbol
	if i := strings.LastIndex(s, "/"); h.Bass != nil && i >= 0 {
		return s[i+1:]
	}
	if s != "" {
		if len(s) > 1 && (s[1] == 'b' || s[1] == '#') {
			return s[:2]
		}
		return s[:1]
	}
	return key.PitchClassName(bassOf(h))
}

var inversionName = [4]string{"", "3 度", "5 度", "7 度"}
var inversionText = [4]string{
	"",
	"第 1 転回形）なので、基本形より軽く、次の和音へ流れやすい。",
	"第 2 転回形）。響きが浮いて不安定で、次の和音へ進む力が強い。",
	"第 3 転回形）。和音の 7 度の音が下にあり、次の和音で 1 つ下へ解決したがる。",
}

// bassText explains a slash chord's bass (BL-114): a chord tone is an
// inversion; a note outside the chord kept from a neighbour is a pedal;
// one that lies between the basses around it passes between them.
func bassText(key Key, harmony []HarmonyEvent, i int) string {
	h := harmony[i]
	if h.Bass == nil {
		return ""
	}
	name := bassName(key, h)
	prev, next := neighbourBasses(harmony, i)
	pedal := (prev != nil && *prev == *h.Bass) || (next != nil && *next == *h.Bass)
	if idx := inversionIndex(h); idx == 3 && h.Chord.Quality == MinorSixth {
		s := fmt.Sprintf(" 根音ではなく、足した 6 度の %s がベース。6 度が下に来ると、和音の重心が下がって陰りが深まる。", name)
		if pedal {
			s += "隣の和音と同じベースが続くので、ペダルとしても働く。"
		}
		return s
	} else if idx > 0 {
		s := fmt.Sprintf(" 根音ではなく %sの %s がベース（%s", inversionName[idx], name, inversionText[idx])
		if pedal { // e.g. I – IV/1 – I, the pedal six-four
			s += "隣の和音と同じベースが続くので、ペダルとしても働き、ベースは留まって和音だけが上で動く。"
		}
		return s
	}
	switch {
	case pedal:
		return fmt.Sprintf(" ベースの %s はこの和音にない音で、隣の和音から同じベースが続くペダル。ベースは留まり、和音だけが上で動く。", name)
	case prev != nil && next != nil && between(*prev, *h.Bass, *next):
		return fmt.Sprintf(" ベースの %s はこの和音にない音で、前後のベースのあいだを順につなぐ経過音。", name)
	}
	return fmt.Sprintf(" ベースの %s はこの和音にない音で、和音に緊張を加える。", name)
}

// neighbourBasses are the basses of the chords before and after i (nil at
// either end).
func neighbourBasses(harmony []HarmonyEvent, i int) (prev, next *PitchClass) {
	if i > 0 {
		b := bassOf(harmony[i-1])
		prev = &b
	}
	if i+1 < len(harmony) {
		b := bassOf(harmony[i+1])
		next = &b
	}
	return prev, next
}

// inversionIndex is 1, 2 or 3 when the slash bass is the chord's third,
// fifth or seventh; 0 when it is not a chord tone above the root.
func inversionIndex(h HarmonyEvent) int {
	for idx, p := range h.Chord.Voicing(ChordOctave) {
		if idx > 0 && idx < len(inversionName) && p.Class() == *h.Bass {
			return idx
		}
	}
	return 0
}

// step is the signed distance from a to b in semitones, -6..5.
func step(a, b PitchClass) int {
	d := (int(b) - int(a) + 12) % 12
	if d > 6 {
		d -= 12
	}
	return d
}

// between reports whether mid lies on the way from a to c by steps of a
// whole tone or less in one direction.
func between(a, mid, c PitchClass) bool {
	s1, s2 := step(a, mid), step(mid, c)
	return s1 != 0 && s2 != 0 && (s1 > 0) == (s2 > 0) && abs(s1) <= 2 && abs(s2) <= 2
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// lineText names the longest stepwise bass line of three or more chords
// that includes a slash chord (BL-114): "ベースが C → B → A → G と順に下がる".
// When the chord stays and the bass moves by semitones, it is a line cliché.
func lineText(key Key, harmony []HarmonyEvent) string {
	start, end := longestLine(harmony)
	if end == 0 {
		return ""
	}
	return describeLine(key, harmony[start:end])
}

// longestLine finds the longest run of three or more chords whose basses
// move by a whole tone or less in one direction and that includes a slash
// chord; end is 0 when there is none.
func longestLine(harmony []HarmonyEvent) (bestStart, bestEnd int) {
	for start := 0; start+2 < len(harmony); start++ {
		end := lineEnd(harmony, start)
		if end-start >= 3 && end-start > bestEnd-bestStart && hasSlash(harmony[start:end]) {
			bestStart, bestEnd = start, end
		}
	}
	return bestStart, bestEnd
}

// lineEnd is where the stepwise bass line starting at start stops.
func lineEnd(harmony []HarmonyEvent, start int) int {
	end, dir := start+1, 0
	for end < len(harmony) {
		d := step(bassOf(harmony[end-1]), bassOf(harmony[end]))
		if d == 0 || abs(d) > 2 || (dir != 0 && (d > 0) != (dir > 0)) {
			break
		}
		dir = d
		end++
	}
	return end
}

func describeLine(key Key, run []HarmonyEvent) string {
	names := make([]string, len(run))
	same, chromatic := true, true
	for i, h := range run {
		names[i] = bassName(key, h)
		if i > 0 {
			same = same && h.Chord == run[0].Chord
			chromatic = chromatic && abs(step(bassOf(run[i-1]), bassOf(h))) == 1
		}
	}
	way := "下がる"
	if step(bassOf(run[0]), bassOf(run[1])) > 0 {
		way = "上がる"
	}
	if same && chromatic {
		chord, _, _ := strings.Cut(run[0].Symbol, "/")
		return fmt.Sprintf("%s のまま、ベースが %s と半音ずつ%s（ライン・クリシェ）。", chord, strings.Join(names, " → "), way)
	}
	return fmt.Sprintf("ベースが %s と順に%s。和音どうしが滑らかにつながる。", strings.Join(names, " → "), way)
}

func hasSlash(harmony []HarmonyEvent) bool {
	for _, h := range harmony {
		if h.Bass != nil {
			return true
		}
	}
	return false
}

// LiteralSlash is LiteralChord with an optional slash bass, read against
// the tonic's major scale like the chord (BL-115): in A, "i/7" is Am over
// G# and "i/b7" Am over G. The bass is nil without a slash.
func LiteralSlash(tonic PitchClass, label string) (Chord, *PitchClass, error) {
	head, bass, err := splitBass(label)
	if err != nil {
		return Chord{}, nil, err
	}
	chord, err := LiteralChord(tonic, head)
	if err != nil || bass == nil {
		return chord, nil, err
	}
	pc := bass.pitchClass(Key{Tonic: tonic, Mode: Major})
	return chord, &pc, nil
}

// BassClash reports a slash bass whose sharp or flat lands it on another
// degree of the key's own scale (#331): A minor's 6 is already F, so
// "iv/b6" stands on Fb, the note of degree 5 (E), and was meant as "/6" or
// "/5". It returns nil for a label without such a bass, and for one the
// engine cannot read (the label's own error says why).
func BassClash(key Key, label string) error {
	_, b, _ := splitBass(label) // a label it cannot read has no bass: its own error says why
	if b == nil || b.Accidental == 0 {
		return nil
	}
	names := key.scaleNames()
	pc := b.pitchClass(key)
	for i, s := range key.Scale() {
		if s == pc {
			return fmt.Errorf("music: %q: the bass %s in %s is %s, the same note as degree %d (%s); write \"/%d\" for that note, or \"/%d\" for %s, the key's own %d",
				label, strings.TrimPrefix(b.suffix(), "/"), key, key.spellDegree(b.Degree, b.Accidental), i+1, names[i], i+1, b.Degree, names[b.Degree-1], b.Degree)
		}
	}
	return nil
}
