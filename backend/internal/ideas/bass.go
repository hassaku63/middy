package ideas

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// A reason may say how the bass moves: "ベースが F→E→D→C と下がる"
// (BL-187, the error the evaluation's reviewer found most often). Such a
// run of three or more note names is checked against what the candidate
// plays underneath: the harmony's basses (the slash bass, else the root)
// or the lowest notes the accompaniment sounds, each with repeated notes
// merged and the example looping. Notes are compared as pitch classes, so
// Bb and A# are the same note.

// bassWords marks a sentence about the bass.
var bassWords = regexp.MustCompile(`ベース|低音|バス音|バスライン|バス・ライン`)

// noteRun matches three or more bare note names joined by arrows, dashes
// or dots: "F→E→D→C", "C - B - A", "E♭・D・C".
var noteRun = regexp.MustCompile(`[A-G][b♭#♯]?(?:\s*(?:→|->|–|—|-|・|、)\s*[A-G][b♭#♯]?){2,}`)

var noteName = regexp.MustCompile(`([A-G])([b♭#♯]?)`)

// bassProblems returns the bass runs of a sentence that the candidate does
// not play.
func bassProblems(sentence string, c Candidate) []string {
	if !bassWords.MatchString(sentence) || hedged.MatchString(sentence) {
		return nil
	}
	var problems []string
	for _, m := range noteRun.FindAllStringIndex(sentence, -1) {
		// A letter or digit right after a name makes it a chord ("Am", "F7"),
		// and one right before makes it part of a word.
		if m[0] > 0 && isASCIILetter(sentence[m[0]-1]) || m[1] < len(sentence) && isNameTail(sentence[m[1]]) {
			continue
		}
		run := sentence[m[0]:m[1]]
		claimed := notesOf(run)
		harmony, lowest := harmonyBasses(c.Timeline), lowestNotes(c.Timeline)
		if len(harmony) == 0 || containsLooped(harmony, claimed) || containsLooped(lowest, claimed) {
			continue
		}
		problems = append(problems, fmt.Sprintf("ベースを %s と書いているが、この候補のベースは %s", run, spell(c.Timeline, harmony)))
	}
	return problems
}

func isNameTail(b byte) bool { return isASCIILetter(b) || b >= '0' && b <= '9' || b == '/' }

func notesOf(run string) []music.PitchClass {
	var out []music.PitchClass
	for _, g := range noteName.FindAllStringSubmatch(run, -1) {
		out = append(out, namedRoot(g[1], g[2]))
	}
	return merged(out)
}

// harmonyBasses is the bass under each chord: the slash bass, else the root.
func harmonyBasses(tl music.Timeline) []music.PitchClass {
	out := make([]music.PitchClass, 0, len(tl.Harmony))
	for _, h := range tl.Harmony {
		if h.Bass != nil {
			out = append(out, *h.Bass)
		} else {
			out = append(out, h.Chord.Root)
		}
	}
	return merged(out)
}

// lowestNotes is the lowest note of the accompaniment at each onset, which
// follows a walking bass's passing notes too.
func lowestNotes(tl music.Timeline) []music.PitchClass {
	low := map[music.Tick]music.Pitch{}
	for _, t := range tl.Tracks {
		if t.Name != music.ChordTrackName {
			continue
		}
		for _, n := range t.Notes {
			if p, ok := low[n.Start]; !ok || n.Pitch < p {
				low[n.Start] = n.Pitch
			}
		}
	}
	starts := make([]music.Tick, 0, len(low))
	for s := range low {
		starts = append(starts, s)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })
	out := make([]music.PitchClass, 0, len(starts))
	for _, s := range starts {
		out = append(out, low[s].Class())
	}
	return merged(out)
}

// merged drops a note that repeats the one before it: a bass held over
// two chords moves once.
func merged(notes []music.PitchClass) []music.PitchClass {
	return slices.Compact(slices.Clone(notes))
}

// containsLooped reports whether run occurs in seq read as a loop.
func containsLooped(seq, run []music.PitchClass) bool {
	if len(run) == 0 || len(seq) == 0 {
		return false
	}
	looped := append(slices.Clone(seq), seq...)
	looped = append(looped, seq...)
	for i := 0; i+len(run) <= len(looped); i++ {
		if slices.Equal(looped[i:i+len(run)], run) {
			return true
		}
	}
	return false
}

// spell names the basses in the key's spelling, for the retry message.
func spell(tl music.Timeline, basses []music.PitchClass) string {
	names := make([]string, len(basses))
	for i, b := range basses {
		names[i] = tl.Key.PitchClassName(b)
	}
	return strings.Join(names, "→")
}
