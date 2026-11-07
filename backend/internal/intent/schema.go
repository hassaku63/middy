package intent

import (
	"encoding/json"
	"fmt"
)

// Schema is the JSON Schema handed to the LLM as the required output
// format. It mirrors Intent and the limits above; Validate remains the
// authority because schema validation cannot check roman numerals or that
// the engine can render the result.
var Schema = json.RawMessage(fmt.Sprintf(`{
  "type": "object",
  "additionalProperties": false,
  "required": ["summary", "mood", "candidates"],
  "properties": {
    "summary": {
      "type": "string",
      "description": "ユーザーのイメージを 1〜2 文の日本語で言い換えたもの。返答としてそのまま表示される"
    },
    "mood": {
      "type": "array",
      "maxItems": %d,
      "items": { "type": "string" },
      "description": "雰囲気を表す短い日本語のタグ"
    },
    "candidates": {
      "type": "array",
      "minItems": %d,
      "maxItems": %d,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["title", "key", "tempoBpm", "progression", "beatsPerChord", "reason"],
        "properties": {
          "title": { "type": "string", "description": "候補の短い見出し（日本語）" },
          "key": {
            "type": "object",
            "additionalProperties": false,
            "required": ["tonic", "mode"],
            "properties": {
              "tonic": { "type": "integer", "minimum": 0, "maximum": 11, "description": "0=C, 1=C#, ... 11=B" },
              "mode": { "type": "string", "enum": ["major", "minor"] }
            }
          },
          "tempoBpm": { "type": "number", "minimum": %d, "maximum": %d },
          "progression": {
            "type": "array",
            "minItems": 1,
            "maxItems": %d,
            "items": { "type": "string" },
            "description": "ダイアトニックなローマ数字のみ: I, ii, iii, IV, V, vi, vii° と 7 / maj7 / ø7 の七の和音（例: ii7, V7, IVmaj7）。コードの質は調から決まるため、短調でも V と書けば v になる"
          },
          "beatsPerChord": { "type": "integer", "enum": [1, 2, 4, 8], "description": "1 コードの長さ（四分音符単位）" },
          "reason": { "type": "string", "description": "なぜこの調・テンポ・進行がイメージに合うかの 1〜2 文（日本語）" }
        }
      }
    }
  }
}`, MaxMoodTags, MinCandidates, MaxCandidates, MinTempoBPM, MaxTempoBPM, MaxChords))

// Parse decodes the model's JSON answer and validates it.
func Parse(data []byte) (Intent, error) {
	var in Intent
	if err := json.Unmarshal(data, &in); err != nil {
		return Intent{}, fmt.Errorf("intent: not valid JSON: %w", err)
	}
	if err := in.Validate(); err != nil {
		return Intent{}, fmt.Errorf("intent: invalid: %w", err)
	}
	return in, nil
}
