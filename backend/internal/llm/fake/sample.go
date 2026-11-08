package fake

import (
	"encoding/json"
	"strings"

	"github.com/hassaku63/middy/backend/internal/llm"
)

// Sample intents returned by an unscripted Client when the caller asks
// for structured output. Deterministic and keyword-driven so that local
// development and end-to-end tests work without an API key. The shapes
// follow internal/intent; fake must not import it (it would invert the
// dependency direction), so they are plain maps.
func sampleIntent(req llm.Request) string {
	text := strings.ToLower(lastUserContent(req))
	minor := false
	for _, kw := range []string{"雨", "悲", "切な", "夜", "寂", "暗", "rain", "sad", "night", "dark", "melanch"} {
		if strings.Contains(text, kw) {
			minor = true
			break
		}
	}
	modal := false
	for _, kw := range []string{"ケルト", "アイルランド", "ドリアン", "ミクソリディアン", "celtic", "irish", "dorian", "mixolydian"} {
		if strings.Contains(text, kw) {
			modal = true
			break
		}
	}
	in := meterSample(text)
	switch {
	case in != nil:
	case modal:
		in = map[string]any{
			"summary": "ケルトの舞曲の情景。ドリアンの明るい IV と、ミクソリディアンの導音のない開放感が合います。",
			"mood":    []string{"素朴", "軽やか"},
			"candidates": []map[string]any{
				{"title": "緑の丘のリール", "key": map[string]any{"tonic": 2, "mode": "dorian"}, "tempoBpm": 112,
					"progression": []string{"i", "IV", "i", "VII"}, "beatsPerChord": 4, "beats": []int{8, 4, 2, 2},
					"reason":        "主和音の Dm を 2 小節保ってから、D ドリアンの IV（G）へ。長 6 度の B が短調の暗さを和らげ、裏打ちのギターとボーラン風の太鼓でリールの推進力を作ります。",
					"accompaniment": "reel", "drums": "bodhran"},
				{"title": "港町のジグ", "key": map[string]any{"tonic": 7, "mode": "mixolydian"}, "tempoBpm": 120,
					"progression": []string{"I", "VII", "IV", "I"}, "beatsPerChord": 4,
					"reason": "G ミクソリディアンの I–VII–IV–I。導音のない VII（F）がそのまま I へ戻る、モーダルな終止です。"},
			},
		}
	case minor:
		in = map[string]any{
			"summary": "静かで少し切ない情景。ゆっくりした短調で、落ち着いた動きが合います。",
			"mood":    []string{"静か", "切ない"},
			"candidates": []map[string]any{
				{"title": "窓辺の雨", "key": map[string]any{"tonic": 9, "mode": "minor"}, "tempoBpm": 72,
					"progression": []string{"i", "VI", "III", "VII"}, "beatsPerChord": 4,
					"reason":        "A minor の定番の下行進行で、落ち着いた切なさを出します。",
					"accompaniment": "arpeggio", "drums": "ballad"},
				{"title": "遠い記憶", "key": map[string]any{"tonic": 4, "mode": "minor"}, "tempoBpm": 66,
					"progression": []string{"i", "iv", "VII", "III", "VI", "iiø7", "V7", "i"}, "beatsPerChord": 2,
					"reason": "iiø7 → V7 → i の進行で、戻ってくる安心感と緊張を両方入れます。"},
			},
		}
	default:
		in = map[string]any{
			"summary": "明るく開けた情景。軽快な長調で、前に進む感じが合います。",
			"mood":    []string{"明るい", "軽やか"},
			"candidates": []map[string]any{
				{"title": "朝の光", "key": map[string]any{"tonic": 0, "mode": "major"}, "tempoBpm": 104,
					"progression": []string{"I", "V7/vi", "vi", "IV"}, "beatsPerChord": 4,
					"reason":        "親しみやすいポップ進行に、vi へ向かうセカンダリードミナント（E7）を挟んで、胸が詰まる一瞬を作ります。",
					"accompaniment": "strum", "drums": "eight-beat"},
				{"title": "出発の朝", "key": map[string]any{"tonic": 7, "mode": "major"}, "tempoBpm": 120,
					"progression": []string{"I", "IV", "iv", "I"}, "beatsPerChord": 4,
					"reason": "IV から同主短調の iv（Cm）へ落とす借用で、明るさの中に一瞬の切なさを差します。"},
			},
		}
	}
	// A swing word swings the first candidate (BL-076), so the whole chain
	// can be exercised without a key.
	for _, kw := range []string{"スウィング", "ジャズ", "跳ね", "swing", "jazz"} {
		if strings.Contains(text, kw) {
			in["candidates"].([]map[string]any)[0]["feel"] = "swing"
			break
		}
	}
	// A "long" word repeats the first candidate's progression four times
	// (16 chords), so long-timeline layouts can be tried without a key.
	for _, kw := range []string{"長い", "long"} {
		if strings.Contains(text, kw) {
			first := in["candidates"].([]map[string]any)[0]
			prog := first["progression"].([]string)
			first["progression"] = append(append(append(append([]string{}, prog...), prog...), prog...), prog...)
			first["title"] = first["title"].(string) + "（長め）"
			break
		}
	}
	b, _ := json.Marshal(in)
	return string(b)
}

// meterSample answers images with a typical shape of their own: three-beat
// images in 3/4 and rocking ones in 6/8 (ADR-0017), a graduation song with
// a falling slash-chord bass (ADR-0020); nil for any other.
func meterSample(text string) map[string]any {
	switch {
	case hasAny(text, "ワルツ", "3拍子", "三拍子", "舞踏会", "waltz"):
		return waltzSample
	case hasAny(text, "ゆりかご", "子守唄", "舟歌", "ジグ", "6/8", "lullaby", "jig"):
		return sixEightSample
	case hasAny(text, "卒業", "graduation"):
		return graduationSample
	}
	return nil
}

// hasAny reports whether text contains any of the words.
func hasAny(text string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

// waltzSample answers three-beat images in 3/4 (ADR-0017), so the meter can
// be exercised end to end without a key.
var waltzSample = map[string]any{
	"summary": "舞踏会で回るワルツの情景。3 拍子で、1 拍目の重みがくるくると回る感じを作ります。",
	"mood":    []string{"優雅", "華やか"},
	"candidates": []map[string]any{
		{"title": "ウィーンの円舞曲", "key": map[string]any{"tonic": 5, "mode": "major"}, "tempoBpm": 160,
			"meter": "3/4", "progression": []string{"I", "IV", "V7", "I"}, "beatsPerChord": 3,
			"reason":        "1 拍目のベースと 2・3 拍目の和音の「ズン・チャッ・チャッ」で、1 小節ごとに回るワルツの揺れを作ります。",
			"accompaniment": "waltz", "drums": "waltz"},
		{"title": "揺れる 3 拍子のバラード", "key": map[string]any{"tonic": 2, "mode": "major"}, "tempoBpm": 84,
			"meter": "3/4", "progression": []string{"I", "vi", "IV", "V7"}, "beatsPerChord": 6,
			"reason":        "2 小節ずつの和音を 3 拍のアルペジオで流し、ゆったりとした 3 拍子の揺れにします。",
			"accompaniment": "arpeggio-3"},
	},
}

// sixEightSample answers rocking images in 6/8 (ADR-0017, ADR-0019): two
// dotted-quarter beats a bar, each divided in three, and the tempo counted
// in those beats.
var sixEightSample = map[string]any{
	"summary": "ゆりかごがゆっくり揺れる子守唄の情景。6/8 拍子で、1 小節に 2 回、大きく揺れる感じを作ります。",
	"mood":    []string{"穏やか", "やさしい"},
	"candidates": []map[string]any{
		{"title": "ゆりかごの 6/8", "key": map[string]any{"tonic": 7, "mode": "major"}, "tempoBpm": 56,
			"meter": "6/8", "progression": []string{"I", "IV", "V7", "I"}, "beatsPerChord": 2,
			"reason":        "付点 4 分の拍の頭にベース、残りの 8 分 2 つに和音を置き、1 小節に 2 回揺れる 6/8 の子守唄にします。",
			"accompaniment": "rocking-6"},
		{"title": "跳ねるジグ", "key": map[string]any{"tonic": 2, "mode": "major"}, "tempoBpm": 116,
			"meter": "6/8", "progression": []string{"I", "V", "I", "IV", "I", "V", "I", "I"}, "beatsPerChord": 1,
			"reason":        "半小節ごとに和音を替え、「タタタ・タタタ」と回るジグの推進力を、拍の 3 つめから次の拍へ跳ねるドラムで出します。",
			"accompaniment": "arpeggio-6", "drums": "jig"},
	},
}

// graduationSample answers a graduation song with a bass that walks down a
// step at a time under the chords (slash chords, ADR-0020), and a minor
// line cliché.
var graduationSample = map[string]any{
	"summary": "卒業式の合唱の情景。ベースが一歩ずつ下がっていき、別れの寂しさと前向きさが重なる感じを作ります。",
	"mood":    []string{"感動", "前向き"},
	"candidates": []map[string]any{
		{"title": "一歩ずつ下がるベース", "key": map[string]any{"tonic": 0, "mode": "major"}, "tempoBpm": 76,
			"progression": []string{"I", "I/7", "vi", "vi/5", "IV", "I/3", "ii7", "V7"}, "beatsPerChord": 4,
			"reason": "ベースが C から D まで一歩ずつ下がる分数コードの進行。和音を替えながらベースだけが滑らかに下り、卒業ソングらしい胸の高まりを作ります。"},
		{"title": "短調のライン・クリシェ", "key": map[string]any{"tonic": 9, "mode": "minor"}, "tempoBpm": 72,
			"progression": []string{"i", "i/#7", "i/7", "i/#6"}, "beatsPerChord": 4,
			"reason": "Am を保ったまま、ベースだけが A から F# へ半音ずつ下がるライン・クリシェ。動かない和音の下で時間が流れていく切なさを出します。"},
	},
}
