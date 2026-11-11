package pattern

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// phraseBars is the length of a phrase the drums mark with a fill
// (BL-126): the engine's phrase (music.PhraseBars).
const phraseBars = music.PhraseBars

type kit int

const (
	kitSticks kit = iota // kick and snare: toms down, then a crash
	kitHand              // a hand drum (bodhrán): a tom roll, no crash
	kitRide              // a ride keeps time (jazz, bossa): light snare comping over it, into a crash
)

func kitOf(d Drums) kit {
	var snare, kick, ride, hand bool
	for _, h := range d.Hits {
		switch h.Note {
		case Snare:
			snare = true
		case Kick:
			kick = true
		case Ride, RideBell: // the bossa's bell keeps time like the jazz ride
			ride = true
		case LowTom, SideStick:
			hand = true
		}
	}
	switch {
	case ride && !snare: // the jazz ride keeps a soft kick, but no snare backbeat
		return kitRide
	case hand && !snare && !kick:
		return kitHand
	default:
		return kitSticks
	}
}

// fillFrom is where a fill starts in the bar: the back half of a 4/4 bar,
// the last beat of a 3/4 bar or a 6/8 bar.
func fillFrom(ts music.TimeSignature) music.Tick {
	beat := music.BeatTicks(ts, music.DefaultPPQ)
	bar := music.BarTicks(ts, music.DefaultPPQ)
	if ts.BeatsPerBar() >= 4 {
		return bar - 2*beat
	}
	return bar - beat
}

// fillHits are the fill's eighths from fillFrom to the bar's end, louder
// as they go.
func fillHits(k kit, ts music.TimeSignature) []DrumHit {
	from := fillFrom(ts)
	n := int((music.BarTicks(ts, music.DefaultPPQ) - from) / eighth)
	descent := []music.Pitch{Snare, Snare, HighTom, HighTom, MidTom, MidTom, LowTom, LowTom}
	out := make([]DrumHit, 0, n+1)
	for i := range n {
		at := from + music.Tick(i)*eighth
		vel := 70 + 30*i/max(1, n-1)
		switch k {
		case kitHand:
			note := MidTom
			if i%2 == 1 || i == n-1 {
				note = LowTom
			}
			out = append(out, DrumHit{at, eighth, note, vel - 10})
		case kitRide:
			out = append(out, DrumHit{at, eighth, Snare, vel - 25})
		default:
			// Spread over the whole descent, so a short fill still ends low.
			out = append(out, DrumHit{at, eighth, descent[i*(len(descent)-1)/max(1, n-1)], vel})
		}
	}
	return out
}

// keepsTime reports a hit the fill bar keeps after the fill starts: a ride
// kit's cymbal and hi-hat go on under the comping, as a drummer keeps time.
func keepsTime(k kit, h DrumHit) bool {
	if k != kitRide {
		return false
	}
	switch h.Note {
	case Ride, RideBell, PedalHat, ClosedHat:
		return true
	}
	return false
}

// fillBarHits is a phrase-end bar: the figure up to the fill (and, on a
// ride kit, its timekeeping through it), then the fill.
func fillBarHits(dr Drums, k kit, ts music.TimeSignature) []DrumHit {
	from := fillFrom(ts)
	var hits []DrumHit
	for _, h := range dr.Hits {
		if h.Start < from || keepsTime(k, h) {
			hits = append(hits, h)
		}
	}
	return append(hits, fillHits(k, ts)...)
}

// phraseEnd reports whether bar b (0-based) ends a phrase.
func phraseEnd(b int) bool { return (b+1)%phraseBars == 0 }

// fills reports whether the drums take phrase fills: a grid plays as it
// was written.
func fills(d Drums) bool { return d.ID != GridID }

// fillText explains the fills, naming the phrase-end bars whose fill
// sounds before the timeline ends.
func fillText(d Drums, ts music.TimeSignature, length music.Tick) string {
	if !fills(d) {
		return ""
	}
	bar := music.BarTicks(ts, music.DefaultPPQ)
	var at []string
	crashes := 0 // fills followed by another phrase, whose first beat gets a crash
	for b := 0; music.Tick(b)*bar+fillFrom(ts) < length; b++ {
		if phraseEnd(b) {
			at = append(at, strconv.Itoa(b+1))
			if music.Tick(b+1)*bar < length {
				crashes++
			}
		}
	}
	if len(at) == 0 {
		return ""
	}
	k := kitOf(d)
	how := "タムを高い方から低い方へ回すフィルを入れ"
	// A fill of two eighths (3/4's last beat) has room for its first and
	// last strokes only, the snare and the low tom: it does not go round
	// the toms (#404 S9).
	if (bar-fillFrom(ts))/eighth < 3 {
		how = "スネアから低いタムへ下りる短いフィルを入れ"
	}
	switch k {
	case kitHand:
		how = "太鼓を細かく転がすフィルを入れ"
	case kitRide:
		how = "シンバルで拍を刻んだまま、スネアで軽くつなぐフィルを入れ"
	default: // kitSticks: the tom fill above
	}
	switch {
	case k == kitHand || crashes == 0:
		how += "る"
	case crashes == len(at):
		how += "、次のフレーズの頭をクラッシュで示す"
	default:
		how += "、次のフレーズが続くところでは、その頭をクラッシュで示す"
	}
	where := "後半"
	if ts.BeatsPerBar() < 4 {
		where = "最後の拍"
	}
	return fmt.Sprintf("%d 小節ごとのフレーズの終わり（%s 小節目）の%sに、%s。フレーズの区切りが耳でわかる。", phraseBars, strings.Join(at, "・"), where, how)
}
