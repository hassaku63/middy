package ideas

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// The model's own words (the reply and each candidate's reason) may make
// claims about harmony that the engine can check (ADR-0034, decision 4):
// that a numeral is a given chord ("subV7/vi（F7）"), and that a chord
// moves to another ("V7/vi … vi7 へ解決する"). A claim the candidate's
// chords contradict makes the answer unusable, so it is retried like any
// other invalid answer. Only claims that can be read without doubt are
// checked: a numeral the engine cannot read, a chord named without a
// numeral, or a sentence with no numeral before the target, is left alone.
// Wording the checks cannot read (progression names, how one candidate
// varies another) is for the evaluation's reviewer (BL-172).

// numeralPattern matches a roman-numeral label as the prompt teaches them:
// optional "sub", accidental, numeral, decoration, seventh or open-chord
// suffix, and an optional "/x" target. It is matched only where ASCII
// letters do not continue it on either side (see labelsIn).
var numeralPattern = regexp.MustCompile(`(?:sub)?(?:b|♭|#|♯)?(?:VII|VI|IV|V|III|II|I|vii|vi|iv|v|iii|ii|i)(?:°|ø|o)?(?:maj7|m7b5|ø7|°7|7sus4|sus4|m6|7|5)?(?:/(?:b|♭|#|♯)?(?:VII|VI|IV|V|III|II|I|vii|vi|iv|v|iii|ii|i))?`)

// namedChord matches a chord symbol in brackets right after a numeral:
// "（F7）", "(Bb7)", "（E♭maj7）". Only its root is compared.
var namedChord = regexp.MustCompile(`^\s*[（(]\s*(([A-G])([b♭#♯]?)[^）)、,\s]*)`)

// motion matches a target numeral followed by a particle and a verb of
// moving to it: "vi7 へ滑り込む", "IVmaj7 へ戻る", "I に解決する".
var motionVerb = regexp.MustCompile(`^\s*(?:（[^）]*）|\([^)]*\))?\s*(?:へ|に)\s*(?:強く|まっすぐ|半音で|自然に|静かに)?\s*(?:解決|進|向か|戻|滑り込|引き込|着地|落ち|帰)`)

// hedged marks a sentence that contrasts the example with the usual
// motion; its "へ" may name what does not happen.
var hedged = regexp.MustCompile(`本来|通常|ふつう|普通は|一般に|代わりに|ではなく|せず|しない|避け`)

type mention struct {
	label      string
	start, end int
}

// labelsIn finds the numerals in a sentence, skipping matches that are
// part of a longer ASCII word ("Vm7", "Ivy").
func labelsIn(s string) []mention {
	var out []mention
	for _, m := range numeralPattern.FindAllStringIndex(s, -1) {
		if m[0] > 0 && isASCIILetter(s[m[0]-1]) || m[1] < len(s) && isASCIILetter(s[m[1]]) {
			continue
		}
		out = append(out, mention{label: s[m[0]:m[1]], start: m[0], end: m[1]})
	}
	return out
}

func isASCIILetter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }

// chordOf is a numeral's chord root and spelled symbol in a key, or false
// if the engine cannot read it there.
func chordOf(key music.Key, label string) (music.PitchClass, string, bool) {
	step, err := music.StepFromLabel(key, normalizeAccidentals(label), 4)
	if err != nil {
		return 0, "", false
	}
	tl, err := music.FromProgression(key, music.Progression{step}, 120)
	if err != nil || len(tl.Harmony) != 1 {
		return 0, "", false
	}
	return tl.Harmony[0].Chord.Root, tl.Harmony[0].Symbol, true
}

// rootOf is chordOf's root.
func rootOf(key music.Key, label string) (music.PitchClass, bool) {
	root, _, ok := chordOf(key, label)
	return root, ok
}

func normalizeAccidentals(s string) string {
	return strings.NewReplacer("♭", "b", "♯", "#").Replace(s)
}

var letterPC = map[string]music.PitchClass{"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}

func namedRoot(letter, accidental string) music.PitchClass {
	pc := letterPC[letter]
	switch accidental {
	case "b", "♭":
		return pc.Add(-1)
	case "#", "♯":
		return pc.Add(1)
	}
	return pc
}

// checkClaims returns the claims in the answer that its candidates
// contradict, as sentences the retry can quote.
func checkClaims(res Result) []string {
	var problems []string
	for i, c := range res.Candidates {
		where := fmt.Sprintf("candidates[%d].reason", i)
		problems = append(problems, namedChordProblems(where, c.Reason, []music.Key{c.Key})...)
		problems = append(problems, motionProblems(where, c.Reason, c)...)
	}
	keys := make([]music.Key, 0, len(res.Candidates))
	for _, c := range res.Candidates {
		keys = append(keys, c.Key)
	}
	// The reply speaks of the image, not of one candidate: a named chord
	// must fit at least one candidate's key.
	return append(problems, namedChordProblems("summary", res.Reply, keys)...)
}

// namedChordProblems checks "numeral（chord）": the chord's root must be
// the numeral's root in one of the keys.
func namedChordProblems(where, text string, keys []music.Key) []string {
	var problems []string
	for _, m := range labelsIn(text) {
		g := namedChord.FindStringSubmatch(text[m.end:])
		if g == nil {
			continue
		}
		claimed := namedRoot(g[2], g[3])
		readable, fits := false, false
		var want []string
		for _, k := range keys {
			root, symbol, ok := chordOf(k, m.label)
			if !ok {
				continue
			}
			readable = true
			if root == claimed {
				fits = true
				break
			}
			want = append(want, symbol)
		}
		if readable && !fits {
			problems = append(problems, fmt.Sprintf("%s: %s を %s と書いているが、この調の %s は %s", where, m.label, g[1], m.label, strings.Join(want, " / ")))
		}
	}
	return problems
}

// motionProblems checks "source … target へ（解決する、進む…）" in each
// sentence: the source is the nearest numeral before the target, and some
// chord written as the source must be followed by one on the target's
// root (the last chord is followed by the first, as the example loops).
func motionProblems(where, text string, c Candidate) []string {
	var problems []string
	for _, sentence := range strings.Split(text, "。") {
		if hedged.MatchString(sentence) {
			continue // "V7 は本来 I へ解決するが…": a contrast, not a claim about this example
		}
		ms := labelsIn(sentence)
		for j := 1; j < len(ms); j++ {
			if !motionVerb.MatchString(sentence[ms[j].end:]) {
				continue
			}
			source, target := ms[j-1].label, ms[j].label
			if next, contradicted := nextChords(c, source, target); contradicted {
				problems = append(problems, fmt.Sprintf("%s: %s が %s へ進むと書いているが、この候補で %s の次は %s", where, source, target, source, strings.Join(next, " / ")))
			}
		}
	}
	return problems
}

// nextChords reports whether the candidate contradicts "source moves to
// target": no chord written as the source is followed by one on the
// target's root. It returns what does follow. A target the engine cannot
// read, or a source the candidate does not have, is not a contradiction.
func nextChords(c Candidate, source, target string) ([]string, bool) {
	targetRoot, ok := rootOf(c.Key, target)
	if !ok {
		return nil, false
	}
	labels := c.Progression
	var next []string
	for _, at := range indexesOf(labels, source) {
		following := labels[(at+1)%len(labels)]
		if root, ok := rootOf(c.Key, following); ok && root == targetRoot {
			return nil, false
		}
		next = append(next, following)
	}
	return next, len(next) > 0
}

func indexesOf(labels []string, label string) []int {
	var out []int
	for i, l := range labels {
		if normalizeAccidentals(l) == normalizeAccidentals(label) {
			out = append(out, i)
		}
	}
	return out
}
