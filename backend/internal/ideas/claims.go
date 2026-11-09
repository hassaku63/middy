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

// bracketed is what a numeral is followed by in brackets: "（F7）",
// "(Bb7)", or the names of a whole run of numerals, "I–V–vi–IV（D–A–Bm–G）".
var bracketed = regexp.MustCompile(`^\s*[（(]([^）)]*)[）)]`)

// chordName reads a chord symbol's root: "F7", "E♭maj7", "Bm", "C/E".
var chordName = regexp.MustCompile(`^([A-G])([b♭#♯]?)`)

// between matches what may separate the numerals of a run, or the names
// in brackets: dashes, arrows, middle dots, commas and spaces.
var between = regexp.MustCompile(`^[\s–—\-→・、,]*$`)

var nameSeparators = regexp.MustCompile(`[\s–—\-→・、,]+`)

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

// claimProblem is a sentence of the answer that its chords contradict.
// Candidate is the candidate whose reason holds it, or -1 for the reply.
type claimProblem struct {
	candidate int
	sentence  string
	message   string
}

func (p claimProblem) String() string {
	where := "summary"
	if p.candidate >= 0 {
		where = fmt.Sprintf("candidates[%d].reason", p.candidate)
	}
	return where + ": " + p.message
}

// checkClaims returns the sentences of the answer that its candidates
// contradict, with what the retry should be told.
func checkClaims(res Result) []claimProblem {
	var problems []claimProblem
	keys := make([]music.Key, 0, len(res.Candidates))
	for i, c := range res.Candidates {
		keys = append(keys, c.Key)
		for _, s := range sentences(c.Reason) {
			for _, m := range namedChordProblems(s, []music.Key{c.Key}) {
				problems = append(problems, claimProblem{i, s, m})
			}
			for _, m := range motionProblems(s, c) {
				problems = append(problems, claimProblem{i, s, m})
			}
			for _, m := range bassProblems(s, c) {
				problems = append(problems, claimProblem{i, s, m})
			}
		}
	}
	// The reply speaks of the image, not of one candidate: a named chord
	// must fit at least one candidate's key.
	for _, s := range sentences(res.Reply) {
		for _, m := range namedChordProblems(s, keys) {
			problems = append(problems, claimProblem{-1, s, m})
		}
	}
	return problems
}

// withoutContradictions drops the contradicted sentences from the answer:
// what is left the chords bear out (ADR-0034, decision 5). Used when the
// retried answer still contradicts its chords, rather than failing a turn
// whose music is sound.
func withoutContradictions(res Result, problems []claimProblem) Result {
	out := res
	out.Candidates = append([]Candidate(nil), res.Candidates...)
	for _, p := range problems {
		if p.candidate < 0 {
			out.Reply = strings.TrimSpace(strings.Replace(out.Reply, p.sentence, "", 1))
			continue
		}
		c := &out.Candidates[p.candidate]
		c.Reason = strings.TrimSpace(strings.Replace(c.Reason, p.sentence, "", 1))
	}
	return out
}

// sentences splits a text after each "。", keeping the mark, so that a
// sentence can be removed from the text as it was written.
func sentences(text string) []string {
	var out []string
	for text != "" {
		i := strings.Index(text, "。")
		if i < 0 {
			out = append(out, text)
			break
		}
		out = append(out, text[:i+len("。")])
		text = text[i+len("。"):]
	}
	return out
}

// namedChordProblems checks numerals named in brackets: "numeral（chord）",
// or a run of numerals named name by name, "I–V–vi–IV（D–A–Bm–G）". Each
// name's root must be its numeral's root in one of the keys. Brackets
// that are not chord names, or whose names do not line up one to one
// with the run before them, are not checked.
func namedChordProblems(sentence string, keys []music.Key) []string {
	var problems []string
	ms := labelsIn(sentence)
	for j, m := range ms {
		g := bracketed.FindStringSubmatch(sentence[m.end:])
		if g == nil {
			continue
		}
		names := chordNames(g[1])
		if len(names) == 0 {
			continue
		}
		found, ok := checkNames(sentence, ms, j, names, keys)
		if !ok || len(found) == 0 {
			continue
		}
		// "C/G" is a slash chord or two names; it is wrong only if both
		// readings are.
		if len(names) == 1 && strings.Contains(names[0], "/") {
			if alt, ok := checkNames(sentence, ms, j, strings.Split(names[0], "/"), keys); ok && len(alt) == 0 {
				continue
			}
		}
		problems = append(problems, found...)
	}
	return problems
}

// checkNames pairs the names with the numerals they follow — one name with
// ms[j], several with the run ending at ms[j] — and returns the names
// their numerals contradict. It reports false when they cannot be paired.
func checkNames(sentence string, ms []mention, j int, names []string, keys []music.Key) ([]string, bool) {
	run := []mention{ms[j]}
	if len(names) > 1 {
		for _, n := range names {
			if !chordName.MatchString(n) {
				return nil, false
			}
		}
		run = runEndingAt(sentence, ms, j)
		if len(run) < len(names) {
			return nil, false
		}
		run = run[len(run)-len(names):]
	}
	var problems []string
	for i, name := range names {
		if p := namedChordProblem(run[i].label, name, keys); p != "" {
			problems = append(problems, p)
		}
	}
	return problems, true
}

// chordNames reads the chord names in a bracket: one name followed by
// other words ("F7、ブルースの属七") is that one name; a list of names
// ("D–A–Bm–G") is the list. Anything else names no chord.
func chordNames(inside string) []string {
	if strings.Contains(inside, "の") {
		return nil // "（D の V7）" names what the chord belongs to, not the chord
	}
	parts := nameSeparators.Split(strings.TrimSpace(inside), -1)
	var names []string
	for _, p := range parts {
		if !chordName.MatchString(p) {
			break
		}
		names = append(names, p)
	}
	if len(names) > 1 && len(names) < len(parts) {
		return nil // a list mixed with words: ambiguous
	}
	if len(names) > 1 {
		return names
	}
	if len(names) == 1 {
		return names[:1]
	}
	return nil
}

// runEndingAt is the numerals that end at ms[j] with only dashes, arrows,
// dots or commas between them.
func runEndingAt(sentence string, ms []mention, j int) []mention {
	first := j
	for first > 0 && between.MatchString(sentence[ms[first-1].end:ms[first].start]) {
		first--
	}
	return ms[first : j+1]
}

// namedChordProblem compares one name with its numeral's root in the keys.
func namedChordProblem(label, name string, keys []music.Key) string {
	g := chordName.FindStringSubmatch(name)
	claimed := namedRoot(g[1], g[2])
	readable := false
	var want []string
	for _, k := range keys {
		root, symbol, ok := chordOf(k, label)
		if !ok {
			continue
		}
		readable = true
		if root == claimed {
			return ""
		}
		want = append(want, symbol)
	}
	if !readable {
		return ""
	}
	return fmt.Sprintf("%s を %s と書いているが、この調の %s は %s", label, name, label, strings.Join(want, " / "))
}

// motionProblems checks "source … target へ（解決する、進む…）" in a
// sentence: the source is the nearest numeral before the target, and some
// chord written as the source must be followed by one on the target's
// root (the last chord is followed by the first, as the example loops).
func motionProblems(sentence string, c Candidate) []string {
	if hedged.MatchString(sentence) {
		return nil // "V7 は本来 I へ解決するが…": a contrast, not a claim about this example
	}
	var problems []string
	ms := labelsIn(sentence)
	for j := 1; j < len(ms); j++ {
		if !motionVerb.MatchString(sentence[ms[j].end:]) {
			continue
		}
		source, target := ms[j-1].label, ms[j].label
		if next, contradicted := nextChords(c, source, target); contradicted {
			problems = append(problems, fmt.Sprintf("%s が %s へ進むと書いているが、この候補で %s の次は %s", source, target, source, strings.Join(next, " / ")))
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
