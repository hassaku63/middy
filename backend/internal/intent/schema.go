package intent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
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
        "required": ["title", "style", "key", "tempoBpm", "progression", "meter", "beatsPerChord", "reason"],
        "properties": {
          "title": { "type": "string", "description": "候補の短い見出し（日本語）" },
          "style": { "type": "string", "enum": [%s], "description": "候補の音楽のスタイル。最初に選び、伴奏・ドラム・ノリはこのスタイルに合うものから選ぶ" },
          "key": {
            "type": "object",
            "additionalProperties": false,
            "required": ["tonic", "mode"],
            "properties": {
              "tonic": { "type": "integer", "enum": [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], "description": "0=C, 1=C#, 2=D, 3=D#, 4=E, 5=F, 6=F#, 7=G, 8=G#, 9=A, 10=A#, 11=B" },
              "mode": { "type": "string", "enum": ["major", "minor", "dorian", "mixolydian", "phrygian"], "description": "major / minor のほか、dorian（短調に似て 6 度が長い。IV が長三和音）、mixolydian（長調に似て 7 度が短い。VII が長三和音、導音なし）、phrygian（短調に似て 2 度が短い。II が長三和音）。ローマ数字はそのモードの音階に対して書く" }
            }
          },
          "tempoBpm": { "type": "number", "description": "テンポ（BPM）。%d〜%d" },
          "progression": {
            "type": "array",
            "minItems": 1,
            "items": { "type": "string" },
            "description": "%d 個までのコード。ダイアトニックなローマ数字と、長調では同主短調からの借用（iv, ivm6, ii°, bVI, bVII とその七の和音。♭ は b と書く）、短調では同主長調の I・IV とナポリの bII（三和音のみ）と、セカンダリードミナント V/x・V7/x（x は主和音以外の長・短三和音。調の中の和音と同じになるものは不可）と、裏コード subV7/x（x の半音上の属七。x は主和音も含む長・短三和音）と、経過の減七 vii°7/x（x の半音下の減七。x は主和音以外の長・短三和音）と、3 度のない和音 sus4・7sus4・5（大文字。例: Vsus4、V7sus4、I5。和音の音がすべて調の中にある度数だけ。長調では bIII5・bVI5・bVII5 も）と、長調のブルースの属七 I7・IV7。分数コードは「/」のあとにベースの調の度数 1〜7（♭・♯ は b・#。和音の何度かではない）: I/3、i/#7、V7/V/#4。ダイアトニック: I, ii, iii, IV, V, vi, vii° と 7 / maj7 / ø7 の七の和音（例: ii7, V7, IVmaj7）。コードの質は調から決まる。例外は短調の V / V7（大文字）で、和声的短音階のドミナント（長三和音・属七）になる。自然短音階の v が欲しければ小文字で書く"
          },
          "meter": { "type": "string", "enum": ["4/4", "3/4", "6/8", "12/8"], "description": "拍子。4/4 の候補でも書く（長さと伴奏・ドラムの型は、この拍子のものから選ぶ）。ワルツ、メヌエット、ジャズ・ワルツのように 3 拍子が典型のイメージでは 3/4。子守唄、舟歌、ジグのように付点 4 分の 2 拍で揺れるイメージでは 6/8。スロー・ブルースや 3 連のバラードのように、付点 4 分の 4 拍で 3 連に揺れるイメージでは 12/8" },
          "beatsPerChord": { "type": "integer", "enum": [1, 2, 3, 4, 6, 8], "description": "1 コードの長さ（拍子の拍の数）。4/4 では 1・2・4・8、3/4 では 3（1 小節）か 6（2 小節）、6/8 では拍が付点 4 分で 1（半小節）・2（1 小節）・4（2 小節）、12/8 では拍が付点 4 分で 1・2（半小節）・4（1 小節）・8（2 小節）" },
          "beats": { "type": "array", "items": { "type": "integer", "enum": [1, 2, 3, 4, 6, 8] }, "description": "和音ごとの長さ（拍子の拍の数）。省略するとすべて beatsPerChord。書くなら progression と同じ数にし、各長さは beatsPerChord と同じ値から選ぶ。1 小節より短い和音は小節を等分した位置（4/4 の 2 拍なら 1 拍目か 3 拍目）から始め、1 小節以上の和音は小節の頭から始め、合計は小節の切れ目で終える" },
          "bars": { "type": "integer", "description": "利用者が長さを小節数で頼んだときだけ書く（「12小節」なら 12、1〜32）。和音の長さの合計がちょうどこの小節数になるようにする（4/4 の 12 小節なら 48 拍、3/4 なら 36 拍）。頼まれていなければ省略" },
          "pushes": { "type": "array", "items": { "type": "string", "enum": ["none", "eighth", "sixteenth"] }, "description": "和音ごとの押し。省略するとどの和音も押さない。書くなら progression と同じ数にする。eighth はその和音を 8 分、sixteenth は 16 分だけ拍の頭より早く弾き、そのままつなぐ（前の小節の 4 拍目の裏で次の和音を先に鳴らす）。最初の和音は none" },
          "reason": { "type": "string", "description": "この候補の聴きどころの短い解説。調・テンポ・進行がイメージのどこに合うかを 1〜2 文の日本語で。この例で鳴るものだけを語り、鳴らない楽器やメロディには触れない" },
          "accompaniment": { "type": "string", "enum": [%s], "description": "伴奏の型。例の拍子の型から選ぶ。省略するとブロックコード" },
          "drums": { "type": "string", "enum": [%s], "description": "ドラムの型。例の拍子の型から選ぶ。省略するとドラムなし" },
          "drumGrid": {
            "type": "object",
            "additionalProperties": false,
            "required": ["division", "voices"],
            "description": "ドラムを型の代わりに 1 小節の格子で書く。drums とは同時に使わない。division は 1 拍のマスの数（4/4・3/4 では 2・3・4、6/8 では 3・6）。voices は声部ごとに 1 小節を 1 マス 1 文字で書き（X は強く、x は普通、o はゴースト、. は休み。拍の区切りに空白を入れてよい）、手で叩く声部は 1 マスに 2 つまで",
            "properties": {
              "division": { "type": "integer", "enum": [2, 3, 4, 6] },
              "voices": { "type": "array", "items": { "type": "object", "additionalProperties": false, "required": ["voice", "steps"], "properties": { "voice": { "type": "string", "enum": [%s] }, "steps": { "type": "string" } } } },
              "basedOn": { "type": "string", "enum": [%s], "description": "もとにしたドラムの型。スタイルの型から選ぶ" }
            }
          },
          "feel": { "type": "string", "enum": [%s], "description": "ノリ。straight は均等、swing は 8 分を長・短で跳ねさせる（跳ね方はテンポに合わせる）、shuffle は 3 連符で跳ねる（ブルースやロックのシャッフル）、swing-16 は 16 分をわずかに揺らす（ローファイやヒップホップ）。伴奏かドラムがあるときに効く。swing・shuffle・swing-16 は 6/8 と 12/8 では使わない。省略するとストレート" },
          "tensions": { "type": "string", "enum": [%s], "description": "和音に足すテンション。スタイルの決まりの代わりに使う。none は足さない、9 は 9 度（長三和音には add9）、9-13 は 9 度と 13 度。どの音を足すかはエンジンが調の音階から選び、ローマ数字には書かない。省略するとスタイルのまま" }
        }
      }
    }
  }
}`, MaxMoodTags, MinCandidates, MaxCandidates, quoted(StyleIDs()), MinTempoBPM, MaxTempoBPM, MaxChords,
	quoted(AccompanimentIDs()), quoted(DrumIDs()), quoted(pattern.GridVoices()), quoted(DrumIDs()), quoted(FeelIDs()), quoted(pattern.TensionIDs())))

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
	// Structured output does not guarantee the case of enum values, nor
	// that they carry no stray spaces.
	id := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	for i := range in.Candidates {
		c := &in.Candidates[i]
		c.Accompaniment, c.Drums, c.Feel, c.Style, c.Tensions = id(c.Accompaniment), id(c.Drums), id(c.Feel), id(c.Style), id(c.Tensions)
		for j, p := range c.Pushes {
			c.Pushes[j] = id(p)
		}
		if g := c.DrumGrid; g != nil {
			g.BasedOn = id(g.BasedOn)
			for j := range g.Voices {
				g.Voices[j].Voice = id(g.Voices[j].Voice)
			}
		}
		c.Meter = strings.TrimSpace(c.Meter)
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
	props, ok := root["properties"].(map[string]any)
	if !ok {
		panic("intent: Schema has no properties object")
	}
	props["unconstrained"] = map[string]any{
		"type":        "string",
		"description": "評価のための欄。このイメージのために使いたかったが、指示された語彙やカタログで表せないために使わなかったもの（和音、音階やモード、リズム、拍子など）を 1〜3 文で。表せる語彙で書けたものは含めない。なければ「なし」。音には使われない",
	}
	out, err := json.Marshal(root)
	if err != nil {
		panic("intent: " + err.Error())
	}
	return out
}
