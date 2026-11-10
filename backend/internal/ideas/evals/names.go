package evals

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Chord names in a reason that the candidate never plays (#331, B, stage
// one): the reason says "Eb7" where F7 sounds. Reported, not judged, and
// only in the evaluation: the app does not use it. Like the claim checks
// of ADR-0034, it reads only what is plainly a claim about this
// progression and skips the rest; the independent reviewer covers what it
// cannot read.

// chordName matches a chord symbol: root, accidental, quality suffix and an
// optional slash bass. The suffixes are tried longest first.
var chordName = regexp.MustCompile(`([A-G])(♭|♯|b|#)?(maj7|maj9|mM7|M7|m7b5|m7♭5|ø7|ø|m6|m7|m9|m|°7|dim7|°|dim|aug|\+|7sus4|sus4|sus2|add9|69|6|7|9|11|13|5)?(?:/([A-G])(♭|♯|b|#)?)?`)

// keyWord follows a key name ("A minor", "E♭ メジャー"), which is not a chord.
var keyWord = regexp.MustCompile(`^\s*(major|minor|メジャー|マイナー|調|dorian|ドリアン|mixolydian|ミクソリディアン|phrygian|フリジアン)`)

// runSep joins the names of a run, "Am–G–F–E" or "C → G".
var runSep = regexp.MustCompile(`^\s*(–|—|-|→|⇒|・|、)\s*$`)

// hedges mark a sentence that names a chord without claiming it sounds:
// a contrast, a substitution, another song or a supposition; or one that
// names notes, not chords (a bass line "A→G#→G").
var hedges = []string{"ではなく", "でなく", "ではない", "代わり", "置き換え", "本来", "原曲", "なら", "例えば", "ベース", "ライン", "音階"}

type nameMatch struct {
	text        string
	start, end  int
	root        music.PitchClass
	class       string
	bass        *music.PitchClass
	bare        bool // a letter alone, "E": a chord only inside a run
	followedKey bool
}

// unplayedNames returns the chord names of the reason that no chord of the
// harmony plays, once each, in the order they appear.
func unplayedNames(reason string, harmony []music.HarmonyEvent) []string {
	var out []string
	seen := map[string]bool{}
	for _, sentence := range strings.SplitAfter(reason, "。") {
		if hedged(sentence) {
			continue
		}
		for _, m := range claimedNames(sentence) {
			if !seen[m.text] && !plays(harmony, m) {
				seen[m.text] = true
				out = append(out, m.text)
			}
		}
	}
	return out
}

func hedged(sentence string) bool {
	for _, h := range hedges {
		if strings.Contains(sentence, h) {
			return true
		}
	}
	return false
}

// claimedNames finds the chord names of a sentence that read as chords:
// not inside an ASCII word ("BPM"), not a key name, and, for a bare
// letter, only next to another name in a run.
func claimedNames(sentence string) []nameMatch {
	var all []nameMatch
	for _, loc := range chordName.FindAllStringSubmatchIndex(sentence, -1) {
		start, end := loc[0], loc[1]
		if start > 0 && isWordByte(sentence[start-1]) || end < len(sentence) && isWordByte(sentence[end]) {
			continue
		}
		m := nameMatch{text: sentence[start:end], start: start, end: end}
		m.root = pitchOf(sentence[loc[2]:loc[3]], group(sentence, loc, 2))
		suffix := group(sentence, loc, 3)
		m.class = qualityClass(suffix)
		m.bare = suffix == "" && group(sentence, loc, 2) == "" && loc[8] < 0
		if loc[8] >= 0 {
			b := pitchOf(sentence[loc[8]:loc[9]], group(sentence, loc, 5))
			m.bass = &b
		}
		m.followedKey = keyWord.MatchString(sentence[end:])
		all = append(all, m)
	}
	var out []nameMatch
	for i, m := range all {
		if m.followedKey {
			continue
		}
		if m.bare && !inRun(sentence, all, i) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// inRun reports a name joined to the name before or after it by a run's
// separator.
func inRun(sentence string, all []nameMatch, i int) bool {
	if i > 0 && runSep.MatchString(sentence[all[i-1].end:all[i].start]) {
		return true
	}
	return i+1 < len(all) && runSep.MatchString(sentence[all[i].end:all[i+1].start])
}

func group(s string, loc []int, n int) string {
	if loc[2*n] < 0 {
		return ""
	}
	return s[loc[2*n]:loc[2*n+1]]
}

func isWordByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

func pitchOf(letter, accidental string) music.PitchClass {
	pc := map[string]int{"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}[letter]
	switch accidental {
	case "♭", "b":
		pc--
	case "♯", "#":
		pc++
	}
	return music.PitchClass((pc + 12) % 12)
}

// qualityClass groups a written suffix by the kind of chord it names, so
// "Fmaj9" for a sounding Fmaj7 is no mismatch, but "E" for Em is.
func qualityClass(suffix string) string {
	switch suffix {
	case "m", "m6", "m7", "m9", "mM7":
		return "minor"
	case "°", "dim":
		return "dim"
	case "°7", "dim7":
		return "dim7"
	case "ø", "ø7", "m7b5", "m7♭5":
		return "half-dim"
	case "7", "9", "11", "13":
		return "dominant"
	case "sus4", "sus2", "7sus4":
		return "sus"
	case "aug", "+":
		return "aug"
	case "5":
		return "power"
	}
	return "major" // "", maj7, M7, maj9, 6, 69, add9
}

// engineClass is qualityClass for the chords the engine plays.
func engineClass(q music.ChordQuality) string {
	switch q {
	case music.MinorTriad, music.Minor7, music.MinorSixth:
		return "minor"
	case music.DiminishedTriad:
		return "dim"
	case music.Diminished7:
		return "dim7"
	case music.HalfDiminished7:
		return "half-dim"
	case music.Dominant7:
		return "dominant"
	case music.Sus4, music.Dominant7Sus4:
		return "sus"
	case music.AugmentedTriad:
		return "aug"
	case music.Power:
		return "power"
	case music.MajorTriad, music.Major7:
		return "major"
	}
	return "major"
}

// plays reports a chord of the harmony on the name's root, of its kind
// (a plain major name also covers a dominant seventh, as "G" is often said
// for G7) and, for a slash name, over its bass.
func plays(harmony []music.HarmonyEvent, m nameMatch) bool {
	for _, h := range harmony {
		if h.Chord.Root != m.root {
			continue
		}
		class := engineClass(h.Chord.Quality)
		if class != m.class && (m.class != "major" || class != "dominant") {
			continue
		}
		if m.bass != nil {
			bass := h.Chord.Root
			if h.Bass != nil {
				bass = *h.Bass
			}
			if bass != *m.bass {
				continue
			}
		}
		return true
	}
	return false
}

// unplayedLine names the chords a candidate's reason claims and does not
// play, with the chords it does play: "2: Eb7（鳴るのは Dmaj7 F7 Em7 A7）".
func unplayedLine(i int, names []string, harmony []music.HarmonyEvent) string {
	played := make([]string, 0, len(harmony))
	for _, h := range harmony {
		played = append(played, h.Symbol)
	}
	return fmt.Sprintf("%d: %s（鳴るのは %s）", i, strings.Join(names, "、"), strings.Join(played, " "))
}
