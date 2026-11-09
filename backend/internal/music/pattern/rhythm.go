package pattern

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// placement classifies where in the bar a hit starts, and which beats a
// hit off the beat pushes over: a syncopation (BL-083, after Longuet-Higgins
// & Lee) is a hit between beats whose next beat does not sound. The
// pattern repeats every bar, so the beat after the last is the next bar's
// first. In a compound meter the eighths between the beats divide the beat
// in three rather than answer it; they count as offBeat like any other.
type placement struct {
	onBeat, offBeat, finer int
	compound               bool
	skipped                []int // 1-based beats left silent after a hit before them
}

func classify(starts []music.Tick, ts music.TimeSignature) placement {
	beat := music.BeatTicks(ts, music.DefaultPPQ)
	bar := music.BarTicks(ts, music.DefaultPPQ)
	p := placement{compound: ts.Compound()}
	sounds := map[music.Tick]bool{}
	for _, s := range starts {
		sounds[s%bar] = true
	}
	skipped := map[int]bool{}
	for _, s := range starts {
		switch {
		case s%beat == 0:
			p.onBeat++
			continue
		case s%eighth == 0:
			p.offBeat++
		default:
			p.finer++
		}
		next := (s/beat + 1) * beat % bar
		if !sounds[next] && !skipped[int(next/beat)+1] {
			skipped[int(next/beat)+1] = true
			p.skipped = append(p.skipped, int(next/beat)+1)
		}
	}
	return p
}

// syncopated is true when a hit off the beat leaves the next beat silent.
func (p placement) syncopated() bool { return len(p.skipped) > 0 }

func describeAccompaniment(a Accompaniment, ts music.TimeSignature) *music.RhythmPart {
	// Hits struck together (a bass and its chord) are one attack.
	starts := make([]music.Tick, 0, len(a.Hits))
	for _, h := range a.Hits {
		if !slices.Contains(starts, h.Start) {
			starts = append(starts, h.Start)
		}
	}
	return part(a.ID, a.Name, a.Description, starts, "伴奏", ts)
}

// describeDrums counts the strokes that carry the groove: the kick and
// snare of a kit, or the low tom and side stick of a hand drum such as the
// bodhrán. Hats and the light mid-tom eighths only mark time.
func describeDrums(d Drums, ts music.TimeSignature) *music.RhythmPart {
	kit := make([]music.Tick, 0, len(d.Hits))
	hand := make([]music.Tick, 0, len(d.Hits))
	var kick, snare bool
	for _, h := range d.Hits {
		switch h.Note {
		case Kick, Snare:
			kit = append(kit, h.Start)
			kick, snare = kick || h.Note == Kick, snare || h.Note == Snare
		case LowTom, SideStick:
			hand = append(hand, h.Start)
		}
	}
	if len(kit) == 0 && len(hand) == 0 { // a grid of cymbals or high toms alone
		all := make([]music.Tick, len(d.Hits))
		for i, h := range d.Hits {
			all[i] = h.Start
		}
		rp := part(d.ID, d.Name, d.Description, all, "打楽器", ts)
		rp.Explanation += densityText(d, ts) + nearestText(d, ts)
		return rp
	}
	if len(kit) == 0 && len(hand) > 0 {
		rp := part(d.ID, d.Name, d.Description, hand, "低いタムとサイドスティック", ts)
		rp.Explanation += densityText(d, ts) + nearestText(d, ts)
		return rp
	}
	// Name only what plays: a jazz ride or a bossa has a kick and no snare.
	subject := "キックとスネア"
	switch {
	case kick && !snare:
		subject = "キック"
	case snare && !kick:
		subject = "スネア"
	}
	rp := part(d.ID, d.Name, d.Description, kit, subject, ts)
	rp.Explanation += backbeatText(d, ts) + densityText(d, ts) + nearestText(d, ts)
	return rp
}

// backbeatText names a 4/4 backbeat (the snare or the cross-stick on 2 and
// 4) or half time (on 3 alone). Ghost and light strokes do not count.
func backbeatText(d Drums, ts music.TimeSignature) string {
	if ts != music.FourFour {
		return ""
	}
	on := map[music.Pitch]map[int]bool{Snare: {}, SideStick: {}}
	for _, h := range d.Hits {
		if beats, ok := on[h.Note]; ok && h.Velocity >= 60 && h.Start%quarter == 0 {
			beats[int(h.Start/quarter)+1] = true
		}
	}
	for _, n := range []struct {
		note music.Pitch
		name string
	}{{Snare, "スネア"}, {SideStick, "クロススティック"}} {
		b := on[n.note]
		switch {
		case b[2] && b[4]:
			return fmt.Sprintf("2・4 拍目の%sがバックビート。拍の裏側に重みが来て、体が前後に揺れる。", n.name)
		case b[3] && !b[1] && !b[2] && !b[4]:
			return fmt.Sprintf("%sが 3 拍目だけに来るハーフタイム。テンポの半分の速さで、ゆったり揺れて聞こえる。", n.name)
		}
	}
	return ""
}

// densityText says how full a grid is (BL-122): how many of its steps
// sound. Catalogue drums have no grid to count against.
func densityText(d Drums, ts music.TimeSignature) string {
	if d.ID != GridID || len(d.Hits) == 0 {
		return ""
	}
	step := d.Hits[0].Duration
	steps := int(music.BarTicks(ts, music.DefaultPPQ) / step)
	sounding := map[music.Tick]bool{}
	for _, h := range d.Hits {
		sounding[h.Start] = true
	}
	division := steps / ts.BeatsPerBar()
	return fmt.Sprintf("%sの格子 %d マスのうち %d マスで音が鳴る（密度 %d%%）。", divisionName(division, ts), steps, len(sounding), len(sounding)*100/steps)
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
		fmt.Fprintf(&b, "拍の頭が %d 回、裏拍が %d 回。裏拍で鳴ったあと %sの頭が鳴らないので、音が拍をまたいで前に押し出すシンコペーションになる。", p.onBeat, p.offBeat+p.finer, skippedNames(p.skipped))
	default:
		fmt.Fprintf(&b, "拍の頭が %d 回、裏拍が %d 回。土台は拍の頭にあり、裏拍が軽い推進力を足している。", p.onBeat, p.offBeat+p.finer)
	}
	return b.String()
}

// beatNames lists a bar's beats: "1・2・3" in 3/4.
func beatNames(beats int) string {
	names := make([]string, beats)
	for i := range names {
		names[i] = strconv.Itoa(i + 1)
	}
	return strings.Join(names, "・")
}

// skippedNames lists the skipped beats: "1 拍目", "2 拍目と 4 拍目".
func skippedNames(beats []int) string {
	names := make([]string, len(beats))
	for i, b := range beats {
		names[i] = fmt.Sprintf("%d 拍目", b)
	}
	return strings.Join(names, "と ")
}
