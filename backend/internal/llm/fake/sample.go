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
	var in map[string]any
	if minor {
		in = map[string]any{
			"summary": "静かで少し切ない情景。ゆっくりした短調で、落ち着いた動きが合います。",
			"mood":    []string{"静か", "切ない"},
			"candidates": []map[string]any{
				{"title": "窓辺の雨", "key": map[string]any{"tonic": 9, "mode": "minor"}, "tempoBpm": 72,
					"progression": []string{"i", "VI", "III", "VII"}, "beatsPerChord": 4,
					"reason": "A minor の定番の下行進行で、落ち着いた切なさを出します。"},
				{"title": "遠い記憶", "key": map[string]any{"tonic": 4, "mode": "minor"}, "tempoBpm": 66,
					"progression": []string{"i", "iv", "VII", "III", "VI", "iiø7", "V7", "i"}, "beatsPerChord": 2,
					"reason": "iiø7 → V7 → i の進行で、戻ってくる安心感と緊張を両方入れます。"},
			},
		}
	} else {
		in = map[string]any{
			"summary": "明るく開けた情景。軽快な長調で、前に進む感じが合います。",
			"mood":    []string{"明るい", "軽やか"},
			"candidates": []map[string]any{
				{"title": "朝の光", "key": map[string]any{"tonic": 0, "mode": "major"}, "tempoBpm": 104,
					"progression": []string{"I", "V", "vi", "IV"}, "beatsPerChord": 4,
					"reason": "もっとも親しみやすいポップ進行で、明るさと懐かしさを両立します。"},
				{"title": "出発の朝", "key": map[string]any{"tonic": 7, "mode": "major"}, "tempoBpm": 120,
					"progression": []string{"Imaj7", "ii7", "V7", "Imaj7"}, "beatsPerChord": 4,
					"reason": "ii–V–I で前へ進む推進力を作り、七の和音で少し洗練させます。"},
			},
		}
	}
	b, _ := json.Marshal(in)
	return string(b)
}
