package pattern

import (
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// placement classifies where in the bar a hit starts.
type placement struct{ onBeat, offBeat, finer int }

func classify(starts []music.Tick) placement {
	var p placement
	for _, s := range starts {
		switch {
		case s%quarter == 0:
			p.onBeat++
		case s%eighth == 0:
			p.offBeat++
		default:
			p.finer++
		}
	}
	return p
}

// syncopated is true when a noticeable share of the hits avoids the beats.
func (p placement) syncopated() bool {
	total := p.onBeat + p.offBeat + p.finer
	return total > 0 && (p.offBeat+p.finer)*4 >= total
}

func describeAccompaniment(a Accompaniment, beats int) *music.RhythmPart {
	starts := make([]music.Tick, 0, len(a.Hits))
	for _, h := range a.Hits {
		starts = append(starts, h.Start)
	}
	return part(a.ID, a.Name, a.Description, starts, "伴奏", beats)
}

func describeDrums(d Drums, beats int) *music.RhythmPart {
	// Hats mark time; the groove is in the kick and snare, so count only those.
	starts := make([]music.Tick, 0, len(d.Hits))
	for _, h := range d.Hits {
		if h.Note == Kick || h.Note == Snare {
			starts = append(starts, h.Start)
		}
	}
	return part(d.ID, d.Name, d.Description, starts, "キックとスネア", beats)
}

func part(id, name, desc string, starts []music.Tick, subject string, beats int) *music.RhythmPart {
	p := classify(starts)
	rp := &music.RhythmPart{
		ID: id, Name: name, Description: desc,
		Hits: len(starts), OnBeat: p.onBeat, OffBeat: p.offBeat, Finer: p.finer,
		Syncopated: p.syncopated(),
	}
	rp.Explanation = explainPlacement(subject, p, beats)
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
