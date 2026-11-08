package pattern

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// strokeNames names the drum notes for a learner.
var strokeNames = map[music.Pitch]string{
	Kick: "キック", Snare: "スネア", SideStick: "クロススティック", ClosedHat: "ハイハット",
	OpenHat: "オープンのハイハット", PedalHat: "ペダルのハイハット", Ride: "ライド", RideBell: "ライドのベル",
	Crash: "クラッシュ", LowTom: "低いタム", MidTom: "中くらいのタム", HighTom: "高いタム",
}

// maxListed is how many differing strokes the explanation lists; beyond
// it, the grid is its own groove and only the count is given.
const maxListed = 4

type stroke struct {
	note  music.Pitch
	start music.Tick
}

func strokesOf(d Drums) map[stroke]bool {
	out := map[stroke]bool{}
	for _, h := range d.Hits {
		out[stroke{h.Note, h.Start}] = true
	}
	return out
}

// nearestText compares a grid with the catalogue drums of its meter
// (BL-124, ADR-0012) and names the nearest, with the strokes the grid
// adds and drops; velocities are not compared.
func nearestText(d Drums, ts music.TimeSignature) string {
	if d.ID != GridID {
		return ""
	}
	mine := strokesOf(d)
	var best Drums
	var added, dropped []stroke
	bestDist := -1
	for _, c := range DrumPatternsFor(ts) {
		theirs := strokesOf(c)
		a, r := difference(mine, theirs), difference(theirs, mine)
		if dist := len(a) + len(r); bestDist < 0 || dist < bestDist {
			best, added, dropped, bestDist = c, a, r, dist
		}
	}
	switch {
	case bestDist < 0:
		return ""
	case bestDist == 0:
		return fmt.Sprintf("いちばん近いカタログの型は「%s」と同じ打ち方。", best.Name)
	case bestDist > maxListed:
		return fmt.Sprintf("いちばん近いカタログの型は「%s」だが、%d 打の打ち方が違う、独自のグルーヴ。", best.Name, bestDist)
	}
	clauses := append(changes(added, "に足", ts), changes(dropped, "から抜", ts)...)
	for k, c := range clauses { // 足し・抜き joining, 足した・抜いた at the end
		last := k == len(clauses)-1
		switch {
		case strings.HasSuffix(c, "足") && last:
			clauses[k] += "した"
		case strings.HasSuffix(c, "足"):
			clauses[k] += "し"
		case last:
			clauses[k] += "いた"
		default:
			clauses[k] += "き"
		}
	}
	return fmt.Sprintf("いちばん近いカタログの型は「%s」で、それに対して%sもの。", best.Name, strings.Join(clauses, "、"))
}

// difference lists the strokes in a and not in b, by voice then time.
func difference(a, b map[stroke]bool) []stroke {
	var out []stroke
	for s := range a {
		if !b[s] {
			out = append(out, s)
		}
	}
	slices.SortFunc(out, func(x, y stroke) int {
		if x.note != y.note {
			return int(x.note) - int(y.note)
		}
		return int(x.start - y.start)
	})
	return out
}

// changes groups strokes by voice: "キックを 1 拍目と 3 拍目に足", a verb
// stem the caller conjugates.
func changes(strokes []stroke, verb string, ts music.TimeSignature) []string {
	var out []string
	for i := 0; i < len(strokes); {
		j := i
		var at []string
		for ; j < len(strokes) && strokes[j].note == strokes[i].note; j++ {
			at = append(at, position(strokes[j].start, ts))
		}
		out = append(out, fmt.Sprintf("%sを %s%s", strokeNames[strokes[i].note], strings.Join(at, "と "), verb))
		i = j
	}
	return out
}

// position names a tick in the bar: "3 拍目", "3 拍目の裏", "2 拍目の 4
// つめの 16 分", "1 拍目の 3 連の 2 つめ".
func position(t music.Tick, ts music.TimeSignature) string {
	beat := music.BeatTicks(ts, music.DefaultPPQ)
	n, off := int(t/beat)+1, t%beat
	switch {
	case off == 0:
		return fmt.Sprintf("%d 拍目", n)
	case !ts.Compound() && off == beat/2:
		return fmt.Sprintf("%d 拍目の裏", n)
	case ts.Compound() && off%(beat/3) == 0:
		return fmt.Sprintf("%d 拍目の %d つめの 8 分", n, int(off/(beat/3))+1)
	case !ts.Compound() && off%(beat/4) == 0:
		return fmt.Sprintf("%d 拍目の %d つめの 16 分", n, int(off/(beat/4))+1)
	case !ts.Compound() && off%(beat/3) == 0:
		return fmt.Sprintf("%d 拍目の 3 連の %d つめ", n, int(off/(beat/3))+1)
	default:
		return fmt.Sprintf("%d 拍目の %d つめの 16 分", n, int(off/(beat/6))+1)
	}
}
