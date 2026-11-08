package pattern

import (
	"fmt"
	"slices"
)

// Style binds the patterns and feels that suit a kind of music (ADR-0012):
// the model picks a style first, and Misfits rejects patterns that do not
// belong to it, so "a Celtic reel on the eight-beat kit" is stopped by
// the engine, not left to the prompt. Block chords and no drums suit every
// style.
type Style struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Accompaniments []string `json:"accompaniments"`
	Drums          []string `json:"drums"`
	Feels          []Feel   `json:"feels"`
	// Program, when set, is the GM instrument the style plays a
	// keyboard-default pattern on (BL-110); Instrument names it.
	Program    *int   `json:"program,omitempty"`
	Instrument string `json:"instrument,omitempty"`
}

func program(p int) *int { return &p }

var styles = []Style{
	{
		ID: "pop", Name: "ポップ", Description: "歌ものの土台。8 ビートや 4 分刻みで拍をはっきりさせる。",
		Accompaniments: []string{"pulse", "ballad", "arpeggio", "strum", "waltz", "arpeggio-3", "rocking-6", "arpeggio-6"},
		Drums:          []string{"eight-beat", "ballad", "waltz", "six-eight"},
		Feels:          []Feel{FeelStraight, FeelSwing},
	},
	{
		ID: "ballad", Name: "バラード", Description: "ゆっくり歌う曲。間を広く取り、和音を伸ばしたり分散したりする。",
		Accompaniments: []string{"ballad", "arpeggio", "waltz", "arpeggio-3", "rocking-6", "arpeggio-6", "drone"},
		Drums:          []string{"ballad", "waltz", "six-eight"},
		Feels:          []Feel{FeelStraight},
	},
	{
		ID: "rock", Name: "ロック", Description: "ギターのストロークと 8 ビートで、強く前に進む。",
		Accompaniments: []string{"strum", "pulse"},
		Drums:          []string{"eight-beat"},
		Feels:          []Feel{FeelStraight, FeelSwing},
		Program:        program(29), Instrument: "歪んだギター",
	},
	{
		ID: "jazz", Name: "ジャズ", Description: "ライドシンバルの 4 ビートと短いコンピング。スウィングと組み合わせる。",
		Accompaniments: []string{"comp", "arpeggio", "ballad", "waltz", "arpeggio-3"},
		Drums:          []string{"jazz-ride", "jazz-waltz"},
		Feels:          []Feel{FeelSwing, FeelStraight},
	},
	{
		ID: "bossa", Name: "ボサノバ", Description: "ブラジルの穏やかなリズム。表と裏が交互に来るキックの上で、和音を軽く流す。",
		Accompaniments: []string{"arpeggio", "comp", "ballad"},
		Drums:          []string{"bossa"},
		Feels:          []Feel{FeelStraight},
		Program:        program(24), Instrument: "ナイロン弦ギター",
	},
	{
		ID: "celtic", Name: "ケルト", Description: "アイルランドやスコットランドの舞曲と民謡。リールの裏打ちとボーラン、6/8 のジグ、ドローン。",
		Accompaniments: []string{"reel", "drone", "arpeggio", "arpeggio-3", "waltz", "rocking-6", "arpeggio-6"},
		Drums:          []string{"bodhran", "jig", "six-eight"},
		Feels:          []Feel{FeelStraight},
		Program:        program(46), Instrument: "ハープ",
	},
}

// Styles lists the style catalogue.
func Styles() []Style { return slices.Clone(styles) }

// FindStyle returns the style with the id.
func FindStyle(id string) (Style, error) {
	for _, s := range styles {
		if s.ID == id {
			return s, nil
		}
	}
	return Style{}, fmt.Errorf("%w: style %q", ErrUnknownPattern, id)
}

// Misfits lists what does not belong to the style, each with what would:
// an accompaniment ("" and "block" always fit), drums ("" always fits) and
// a feel ("" is straight).
func (s Style) Misfits(acc, drums string, feel Feel) []string {
	var out []string
	if acc != "" && acc != "block" && !slices.Contains(s.Accompaniments, acc) {
		out = append(out, fmt.Sprintf("accompaniment %q does not fit the %s style; use one of %v or block", acc, s.ID, s.Accompaniments))
	}
	if drums != "" && !slices.Contains(s.Drums, drums) {
		out = append(out, fmt.Sprintf("drums %q does not fit the %s style; use one of %v or none", drums, s.ID, s.Drums))
	}
	if feel == "" {
		feel = FeelStraight
	}
	if !slices.Contains(s.Feels, feel) {
		out = append(out, fmt.Sprintf("feel %q does not fit the %s style; use one of %v", feel, s.ID, s.Feels))
	}
	return out
}
