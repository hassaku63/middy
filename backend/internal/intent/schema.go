package intent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Schema is the JSON Schema handed to the LLM as the required output
// format (structured outputs). It uses only the subset the Messages API
// accepts: no maxItems, no minimum/maximum, minItems only 0 or 1, and
// additionalProperties false on every object (TestSchemaUsesOnlySupportedKeywords
// enforces this). Limits the subset cannot express are stated in the
// descriptions; Validate remains the authority and the service retries
// once with its errors.
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
      "items": { "type": "string" },
      "description": "雰囲気を表す短い日本語のタグ。%d 個まで"
    },
    "candidates": {
      "type": "array",
      "minItems": 1,
      "description": "候補。%d〜%d 個",
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
              "tonic": { "type": "integer", "enum": [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], "description": "0=C, 1=C#, 2=D, 3=D#, 4=E, 5=F, 6=F#, 7=G, 8=G#, 9=A, 10=A#, 11=B" },
              "mode": { "type": "string", "enum": ["major", "minor", "dorian", "mixolydian"], "description": "major / minor のほか、dorian（短調に似て 6 度が長い。IV が長三和音）と mixolydian（長調に似て 7 度が短い。VII が長三和音、導音なし）。ローマ数字はそのモードの音階に対して書く" }
            }
          },
          "tempoBpm": { "type": "number", "description": "テンポ（BPM）。%d〜%d" },
          "progression": {
            "type": "array",
            "minItems": 1,
            "items": { "type": "string" },
            "description": "%d 個までのコード。ダイアトニックなローマ数字と、長調では同主短調からの借用（iv, ivm6, ii°, bVI, bVII とその七の和音。♭ は b と書く）。ダイアトニック: I, ii, iii, IV, V, vi, vii° と 7 / maj7 / ø7 の七の和音（例: ii7, V7, IVmaj7）。コードの質は調から決まる。例外は短調の V / V7（大文字）で、和声的短音階のドミナント（長三和音・属七）になる。自然短音階の v が欲しければ小文字で書く"
          },
          "beatsPerChord": { "type": "integer", "enum": [1, 2, 4, 8], "description": "1 コードの長さ（四分音符単位）" },
          "reason": { "type": "string", "description": "この候補の聴きどころの短い解説。調・テンポ・進行がイメージのどこに合うかを 1〜2 文の日本語で" },
          "accompaniment": { "type": "string", "enum": [%s], "description": "伴奏の型。省略するとブロックコード" },
          "drums": { "type": "string", "enum": [%s], "description": "ドラムの型。省略するとドラムなし" },
          "feel": { "type": "string", "enum": [%s], "description": "8 分音符のノリ。straight は均等、swing は長・短で跳ねる。伴奏かドラムがあるときに効く。省略するとストレート" }
        }
      }
    }
  }
}`, MaxMoodTags, MinCandidates, MaxCandidates, MinTempoBPM, MaxTempoBPM, MaxChords,
	quoted(AccompanimentIDs()), quoted(DrumIDs()), quoted(FeelIDs())))

func quoted(ids []string) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = fmt.Sprintf("%q", id)
	}
	return strings.Join(out, ", ")
}

// Parse decodes the model's JSON answer and validates it.
func Parse(data []byte) (Intent, error) {
	var in Intent
	if err := json.Unmarshal(data, &in); err != nil {
		return Intent{}, fmt.Errorf("intent: not valid JSON: %w", err)
	}
	// Structured output does not guarantee the case of enum values, and
	// catalogue ids are lowercase.
	for i := range in.Candidates {
		in.Candidates[i].Accompaniment = strings.ToLower(in.Candidates[i].Accompaniment)
		in.Candidates[i].Drums = strings.ToLower(in.Candidates[i].Drums)
		in.Candidates[i].Feel = strings.ToLower(in.Candidates[i].Feel)
	}
	if err := in.Validate(); err != nil {
		return Intent{}, fmt.Errorf("intent: invalid: %w", err)
	}
	// Labels are valid here; spell them as the engine will, so the card
	// and the timeline agree ("iim7b5" becomes "iiø7").
	for i := range in.Candidates {
		c := &in.Candidates[i]
		for j, label := range c.Progression {
			if canonical, err := music.CanonicalLabel(c.Key, label); err == nil {
				c.Progression[j] = canonical
			}
		}
	}
	return in, nil
}

// DiagnosticSchema is Schema plus an optional free-text field in which the
// model says what it would have proposed without the engine's limits
// (diatonic chords, major and minor only, the pattern catalogue). The
// evaluation set uses it to tell "cannot be expressed" from "was not
// chosen" (ADR-0013); production requests use Schema.
var DiagnosticSchema = diagnosticSchema()

func diagnosticSchema() json.RawMessage {
	var root map[string]any
	if err := json.Unmarshal(Schema, &root); err != nil {
		panic("intent: Schema is not JSON: " + err.Error())
	}
	props := root["properties"].(map[string]any)
	props["unconstrained"] = map[string]any{
		"type":        "string",
		"description": "評価のための欄。エンジンの制約（ダイアトニックの和音だけ、長調・短調だけ、伴奏とドラムの型のカタログ）がなければ、このイメージに本当は何を書きたかったか。和音名、音階やモード、リズムを自由に 1〜3 文で。音には使われない",
	}
	out, err := json.Marshal(root)
	if err != nil {
		panic("intent: " + err.Error())
	}
	return out
}
