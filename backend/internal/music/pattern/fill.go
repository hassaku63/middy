package pattern

import (
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// phraseBars is the length of a phrase the drums mark with a fill (BL-126,
// ADR-0012 decision 5): four bars, the commonest phrase in songs.
const phraseBars = 4

type kit int

const (
	kitSticks kit = iota // kick and snare: toms down, then a crash
	kitHand              // a hand drum (bodhrán): a tom roll, no crash
	kitRide              // a jazz ride: light snare comping into a crash
)

func kitOf(d Drums) kit {
	var snare, kick, ride, hand bool
	for _, h := range d.Hits {
		switch h.Note {
		case Snare:
			snare = true
		case Kick:
			kick = true
		case Ride:
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
			out = append(out, DrumHit{at, eighth, descent[i*len(descent)/n], vel})
		}
	}
	return out
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
	for b := 0; music.Tick(b)*bar+fillFrom(ts) < length; b++ {
		if phraseEnd(b) {
			at = append(at, fmt.Sprint(b+1))
		}
	}
	if len(at) == 0 {
		return ""
	}
	k := kitOf(d)
	how := "タムを高い方から低い方へ回すフィルを入れ、次のフレーズの頭をクラッシュで示す"
	switch k {
	case kitHand:
		how = "太鼓を細かく転がすフィルを入れる"
	case kitRide:
		how = "スネアで軽くつなぐフィルを入れ、次のフレーズの頭をクラッシュで示す"
	}
	return fmt.Sprintf("%d 小節ごとのフレーズの終わり（%s 小節目）の後半に、%s。フレーズの区切りが耳でわかる。", phraseBars, strings.Join(at, "・"), how)
}
