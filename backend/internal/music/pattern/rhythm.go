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
	if rp := ridePart(d, ts); rp != nil {
		return rp
	}
	return kitPart(d, ts)
}

// kitPart describes the kick and snare (or the hand drum) of a kit.
func kitPart(d Drums, ts music.TimeSignature) *music.RhythmPart {
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

// feathered is the loudest a kick may be and still only "feather" the
// beat, felt more than heard, as a swing drummer plays it (#378).
const feathered = 50

// ridePart describes a ride kit whose kick only feathers the beat (jazz
// ride, jazz waltz): the time is in the ride and the pedal hi-hat, so it
// counts the ride and says where the hi-hat closes, rather than crediting
// the quiet kick with the groove's weight (#378).
func ridePart(d Drums, ts music.TimeSignature) *music.RhythmPart {
	var ride []music.Tick
	hat := map[int]bool{}
	kick := false
	beat, bar := music.BeatTicks(ts, music.DefaultPPQ), music.BarTicks(ts, music.DefaultPPQ)
	for _, h := range d.Hits {
		switch h.Note {
		case Ride, RideBell:
			ride = append(ride, h.Start)
		case PedalHat: // on which beat of the bar, in the meter's beats
			if at := h.Start % bar; at%beat == 0 {
				hat[int(at/beat)+1] = true
			}
		case Kick:
			if h.Velocity >= feathered {
				return nil
			}
			kick = true
		case Snare:
			if h.Velocity >= 60 {
				return nil
			}
		}
	}
	if len(ride) == 0 || !kick { // a ride grid alone is counted as it is
		return nil
	}
	rp := part(d.ID, d.Name, d.Description, ride, "ライド", ts)
	var beats []string
	for b := 1; b <= ts.BeatsPerBar(); b++ {
		if hat[b] {
			beats = append(beats, strconv.Itoa(b))
		}
	}
	if len(beats) > 0 {
		rp.Explanation += fmt.Sprintf("足のハイハットが %s 拍目で閉じて、拍を刻む。", strings.Join(beats, "・"))
	}
	rp.Explanation += "キックはごく弱く支えるだけで、重さはライドとハイハットにある。"
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
	}{{Snare, "スネア"}, {SideStick, "サイドスティック"}} {
		b := on[n.note]
		switch {
		case b[2] && b[4]:
			return fmt.Sprintf("2・4 拍目の%sがバックビート。弱拍の 2・4 拍目に重みが来て、体が前後に揺れる。", n.name)
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
		// 裏拍 is the eighth off-beat (#390): a figure struck only on
		// finer positions is said to sound between the beats.
		between := "裏拍"
		if p.offBeat == 0 {
			between = "拍の頭のあいだ"
		}
		fmt.Fprintf(&b, "拍の頭が %d 回、%s。%sで鳴ったあと %sの頭が鳴らないので、音が拍をまたいで前に押し出すシンコペーションになる。", p.onBeat, offBeatCounts(p), between, skippedNames(p.skipped))
	case p.offBeat == 0:
		// No hit on the eighth off-beat: the push comes from finer ones,
		// which 裏拍 (defined as the eighth off-beat) does not name (#390).
		fmt.Fprintf(&b, "拍の頭が %d 回、%s。土台は拍の頭にあり、拍の頭以外の細かい音が軽い推進力を足している。", p.onBeat, offBeatCounts(p))
	default:
		fmt.Fprintf(&b, "拍の頭が %d 回、%s。土台は拍の頭にあり、裏拍が軽い推進力を足している。", p.onBeat, offBeatCounts(p))
	}
	return b.String()
}

// offBeatCounts says how many hits fall between the beats: on the
// eighth-note offbeat, and finer (a sixteenth, or a triplet from a grid),
// which the sixteen-beat brings (#308).
func offBeatCounts(p placement) string {
	if p.finer == 0 {
		return fmt.Sprintf("裏拍（8 分の裏）が %d 回", p.offBeat)
	}
	if p.offBeat == 0 {
		return fmt.Sprintf("8 分の裏より細かい位置（16 分など）が %d 回", p.finer)
	}
	return fmt.Sprintf("裏拍（8 分の裏）が %d 回、それより細かい位置（16 分など）が %d 回", p.offBeat, p.finer)
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
