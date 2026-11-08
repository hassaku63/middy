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
	var in map[string]any
	switch {
	case modal:
		in = map[string]any{
			"summary": "ケルトの舞曲の情景。ドリアンの明るい IV と、ミクソリディアンの導音のない開放感が合います。",
			"mood":    []string{"素朴", "軽やか"},
			"candidates": []map[string]any{
				{"title": "緑の丘のリール", "key": map[string]any{"tonic": 2, "mode": "dorian"}, "tempoBpm": 112,
					"progression": []string{"i", "IV", "i", "VII"}, "beatsPerChord": 4,
					"reason":        "D ドリアンの i と IV（G）の往復。長 6 度の B が短調の暗さを和らげ、ケルトらしい明るさを作ります。",
					"accompaniment": "strum", "drums": "eight-beat"},
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
					"progression": []string{"I", "V", "vi", "IV"}, "beatsPerChord": 4,
					"reason":        "もっとも親しみやすいポップ進行で、明るさと懐かしさを両立します。",
					"accompaniment": "strum", "drums": "eight-beat"},
				{"title": "出発の朝", "key": map[string]any{"tonic": 7, "mode": "major"}, "tempoBpm": 120,
					"progression": []string{"Imaj7", "ii7", "V7", "Imaj7"}, "beatsPerChord": 4,
					"reason": "ii–V–I で前へ進む推進力を作り、七の和音で少し洗練させます。"},
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
