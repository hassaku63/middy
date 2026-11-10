package music

import "fmt"

// Push is how early a chord arrives before its beat (BL-085): the chord
// is struck an eighth or a sixteenth ahead and tied over, as rock, pop and
// jazz players anticipate a change. The harmony keeps the chord on its
// beat, so bars count the same; only the notes move. Pushes are applied
// before the feel, so a pushed eighth swings with the other off-beats.
type Push string

// The pushes: none, an eighth early, a sixteenth early.
const (
	PushNone      Push = ""
	PushEighth    Push = "eighth"
	PushSixteenth Push = "sixteenth"
)

// Pushes lists the pushes a step may ask for.
var Pushes = []Push{PushEighth, PushSixteenth}

// Ticks is how many ticks early the push sounds.
func (p Push) Ticks(ppq int) Tick {
	switch p {
	case PushEighth:
		return Tick(ppq / 2)
	case PushSixteenth:
		return Tick(ppq / 4)
	default:
		return 0
	}
}

// check validates the push of step i: a known value, and none on the
// first chord, which has nothing before it to push into.
func (p Push) check(i int) error {
	switch {
	case p != PushNone && p != PushEighth && p != PushSixteenth:
		return fmt.Errorf("music: steps[%d]: push must be %q or %q, got %q", i, PushEighth, PushSixteenth, p)
	case i == 0 && p != PushNone:
		return fmt.Errorf("music: steps[%d]: the first chord has nothing to push into", i)
	}
	return nil
}

// Attack is when the chord is actually struck: its start, less its push.
func (h HarmonyEvent) Attack(ppq int) Tick { return h.Start - h.Push.Ticks(ppq) }

// pushText explains a pushed chord: how early, and on which off-beat when
// an eighth push lands on one in a simple meter.
func pushText(tl Timeline, i int) string {
	h := tl.Harmony[i]
	if h.Push == PushNone {
		return ""
	}
	value := "8 分音符"
	if h.Push == PushSixteenth {
		value = "16 分音符"
	}
	where := ""
	if ts := tl.TimeSignature; h.Push == PushEighth && !ts.Compound() {
		beat := BeatTicks(ts, tl.PPQ)
		bar := BarTicks(ts, tl.PPQ)
		n := int((h.Attack(tl.PPQ)%bar)/beat) + 1
		where = fmt.Sprintf("（%d 拍目の裏で）", n)
	}
	return fmt.Sprintf(" %s は拍の頭を待たず、%sひとつ早く%s入る（押し）。次の和音を先に鳴らすので、前へ進む勢いが出る。", h.RomanNumeral, value, where)
}

// PushNotes moves the attacks of pushed chords earlier (BL-085). A note
// that starts on a pushed chord's Start starts at its attack instead and
// is held the longer; a note of the chord before that would sound past
// the attack is cut there, and one that would start after it is dropped.
func PushNotes(harmony []HarmonyEvent, notes []NoteEvent, ppq int) []NoteEvent {
	out := make([]NoteEvent, 0, len(notes))
	for _, n := range notes {
		i := chordAt(harmony, n.Start)
		if i < 0 {
			out = append(out, n)
			continue
		}
		h := harmony[i]
		if n.Start == h.Start && h.Push != PushNone {
			early := h.Push.Ticks(ppq)
			n.Start -= early
			n.Duration += early
		}
		if i+1 < len(harmony) {
			cut := harmony[i+1].Attack(ppq)
			if n.Start >= cut {
				continue // the next chord has already been struck
			}
			n.Duration = min(n.Duration, cut-n.Start)
		}
		out = append(out, n)
	}
	return out
}

// chordAt is the index of the harmony event sounding at t, or -1.
func chordAt(harmony []HarmonyEvent, t Tick) int {
	for i, h := range harmony {
		if t >= h.Start && t < h.Start+h.Duration {
			return i
		}
	}
	return -1
}
