package pattern

import (
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// placement classifies where in the bar a hit starts. In a compound meter
// the eighths between the beats divide the beat in three rather than
// answer it, so they are counted as offBeat but are not syncopation.
type placement struct {
	onBeat, offBeat, finer int
	compound               bool
}

func classify(starts []music.Tick, ts music.TimeSignature) placement {
	beat := music.BeatTicks(ts, music.DefaultPPQ)
	p := placement{compound: ts.Compound()}
	for _, s := range starts {
		switch {
		case s%beat == 0:
			p.onBeat++
		case s%eighth == 0:
			p.offBeat++
		default:
			p.finer++
		}
	}
	return p
}

// syncopated is true when a noticeable share of the hits avoids the beats
// and their division.
func (p placement) syncopated() bool {
	total := p.onBeat + p.offBeat + p.finer
	off := p.offBeat + p.finer
	if p.compound {
		off = p.finer
	}
	return total > 0 && off*4 >= total
}

func describeAccompaniment(a Accompaniment, ts music.TimeSignature) *music.RhythmPart {
	starts := make([]music.Tick, 0, len(a.Hits))
	for _, h := range a.Hits {
		starts = append(starts, h.Start)
	}
	return part(a.ID, a.Name, a.Description, starts, "伴奏", ts)
}

func describeDrums(d Drums, ts music.TimeSignature) *music.RhythmPart {
	// Hats mark time; the groove is in the kick and snare, so count only those.
	starts := make([]music.Tick, 0, len(d.Hits))
	for _, h := range d.Hits {
		if h.Note == Kick || h.Note == Snare {
			starts = append(starts, h.Start)
		}
	}
	return part(d.ID, d.Name, d.Description, starts, "キックとスネア", ts)
}

func part(id, name, desc string, starts []music.Tick, subject string, ts music.TimeSignature) *music.RhythmPart {
	p := classify(starts, ts)
	rp := &music.RhythmPart{
		ID: id, Name: name, Description: desc,
		Hits: len(starts), OnBeat: p.onBeat, OffBeat: p.offBeat, Finer: p.finer,
		Syncopated: p.syncopated(),
	}
	rp.Explanation = explainPlacement(subject, p, ts.BeatsPerBar())
	return rp
}

// explainPlacement turns the counts into the sentence a tutor would say;
// beats is the number of beats in the bar, for naming them.
func explainPlacement(subject string, p placement, beats int) string {
	var b strings.Builder
	total := p.onBeat + p.offBeat + p.finer
	fmt.Fprintf(&b, "1 小節に%sが %d 回。", subject, total)
	switch {
	case total == 0:
		return b.String()
	case p.offBeat == 0 && p.finer == 0:
		fmt.Fprintf(&b, "すべて拍の頭（%s）に乗るので、安定して重心がはっきり聞こえる。", beatNames(beats))
	case p.compound && p.finer == 0:
		fmt.Fprintf(&b, "拍の頭（%s）が %d 回、拍を 3 つに分けた残りの 8 分が %d 回。付点 4 分の拍が 3 つに分かれ、「タタタ」と回る揺れが出る。", beatNames(beats), p.onBeat, p.offBeat)
	case p.syncopated():
		fmt.Fprintf(&b, "拍の頭が %d 回、裏拍が %d 回。裏拍の割合が多く、拍をまたいで前に押し出すシンコペーションの感じが出る。", p.onBeat, p.offBeat+p.finer)
	default:
		fmt.Fprintf(&b, "拍の頭が %d 回、裏拍が %d 回。土台は拍の頭にあり、裏拍が軽い推進力を足している。", p.onBeat, p.offBeat+p.finer)
	}
	return b.String()
}

// beatNames lists a bar's beats: "1・2・3" in 3/4.
func beatNames(beats int) string {
	names := make([]string, beats)
	for i := range names {
		names[i] = fmt.Sprint(i + 1)
	}
	return strings.Join(names, "・")
}
