# ドメインモデル

音楽理論の語彙を、コード上の型に対応づける。**ここに書かれた定義が `backend/internal/music` の仕様** であり、テストはこれを根拠にする。曖昧な点があれば、実装より先にこの文書を直す。

## 規約

| 項目 | 規約 |
| --- | --- |
| ピッチ | MIDI ノート番号 0..127。**C4 = 60**（middle C） |
| ピッチクラス | 0..11、C = 0。調に依存しない表示（`String()`）はシャープ。調の中での綴りは `Key.PitchClassName` / `Key.ChordSymbol`（B♭ major では `Bb`、`Bbmaj7`）。親となる長調の調号から決め、同数なら ♭ |
| 時間 | tick。`DefaultPPQ = 480`（四分音符 = 480 tick） |
| 度数 | 1 始まり（1 = 主音, 5 = 属音） |
| チャンネル | 0 始まり。9 = GM パーカッション |
| プログラム | GM プログラム番号 0 始まり（0 = Acoustic Grand Piano） |

## 型

### 音高

- `PitchClass` — `Normalize()`, `Add(semitones)`, `String()`, `FlatName()`
- `Pitch` — `NewPitch(pc, octave)`, `Valid()`, `Class()`, `Octave()`, `Transpose()`, `String()`（`"C#4"`）

### 調と音階

- `Mode` — `Major`, `Minor`（自然短音階）, `Dorian`, `Mixolydian`（ADR-0015）。JSON では `"major"` / `"minor"` / `"dorian"` / `"mixolydian"`。ローマ数字はそのモードの音階に対して書く（D ドリアンは `i ii III IV v vi° VII`、G ミクソリディアンは `I ii iii° IV v vi VII`）。質は音階から導くので、モードを足すのは `modeIntervals`・`parentOffset`（綴り）・`characteristicDegree`（解説）の 3 行。`Characteristic()` はモードの特徴音の度数（ドリアン 6、ミクソリディアン 7。長調・短調は 0）
- `Key{Tonic, Mode}` — `Scale()`（7 音）, `Degree(n)`, `Contains(pc)`, `Diatonic(degree, seventh)`

### コード

- `ChordQuality` — 三和音 4 種（maj, min, dim, aug）、七の和音 4 種（maj7, min7, dom7, m7b5）、六の和音 1 種（min6。長調で借用する ivm6 のため）。JSON は識別子文字列（`"dom7"`）
- `Chord{Root, Quality}` — `PitchClasses()`, `Voicing(octave)`（密集・基本形）, `String()`（`"G7"`）
- `RomanNumeral(degree, quality)` — `I`, `ii`, `vii°`, `V7`, `iiø7`, `IVmaj7` など。大文字 = 長/増、小文字 = 短/減。`RomanNumeralAlt(degree, accidental, quality)` は根音を半音下げた `♭VI`, `♭VII7` も書く
- `ParseRomanNumeral(s)` — 上記の表記（と `dim`, `aug`, `+`, `m7b5`）を度数と七の和音の有無に戻す。`RomanNumeral` の出力は必ず往復する。大文字・小文字の違いは `StepFromLabel` だけが（短調の V で）使う

**未対応（backlog）**: 転回形、分数コード、テンション、長調の承認済みセット以外の借用和音（短調側の借用、♭II、♭III など）、セカンダリドミナント。

### タイムライン

```
Timeline
├── PPQ, TempoBPM, TimeSignature, Key
├── Tracks[]            楽器レーン
│   └── Notes[]         NoteEvent{Pitch, Velocity, Start, Duration}
└── Harmony[]           HarmonyEvent{Start, Duration, Chord, RomanNumeral, Analysis?}
```

- `Timeline` は **再生・描画・解説・書き出しのすべてが消費する唯一の成果物**。
- `Harmony` は音を出さない注釈。解説レイヤー（和声機能、ケーデンス）はここに情報を足す。
- `Validate()` は API 境界で必ず呼ぶ。内部の関数は妥当な Timeline を前提にしてよい。
- `LengthTicks()` は最後のノート／注釈の終端。

### 進行

- `Step{Degree, Seventh, Beats, Quality?}` — 度数（1..7）を何拍（四分音符単位）鳴らすか。拍は拍子に依存しない。`Quality` は根音を保ったまま質を上書きする（今は短調の V のためだけに使う）
- `StepFromLabel(key, label, beats)` — ローマ数字の文字列から `Step` を作る唯一の入口（API と intent が使う）。短調の大文字 `V` / `V7` を和声的短音階のドミナントにする規則はここにある
- `Progression []Step` — 調に依存しない進行テンプレート。`Validate()` は全ステップの問題をまとめて返す
- `FromProgression(key, prog, tempoBPM)` — 進行を調で解決し、ブロックコード 1 トラック（GM ピアノ, ch.0, `ChordOctave`=4 の基本形, `ChordVelocity`=80）と `Harmony` 注釈を持つ `Timeline` を返す。リズム・ボイシングのパターンはこの上に重ねる（BL-020, 021）

### 分析（解説データ）

- `Function` — `T` / `S` / `D`。度数で決まる: I, III, VI = T、II, IV = S、V, VII = D。長調・自然短音階とも同じ割り当て（短調の v と VII は「弱いドミナント」扱い。借用和音は、その度数の機能をそのまま使う: iv は S、♭VI は T、♭VII は D）
- `Cadence` — 直前のコードからの進行の分類: `authentic`（V→I）、`plagal`（IV→I）、`half`（V 以外→V）、`deceptive`（V→vi）、`modal`（ドリアン・ミクソリディアンの VII→I。導音を使わない）。該当なしは空
- `Analysis{Degree, Function, Cadence, Explanation}` — `HarmonyEvent.Analysis` に付く。`Explanation` は UI にそのまま出せる日本語で、決定論的に組み立てる（LLM は使わない）。機能（T/S/D）は全モードで度数から付ける初学者向けの単純化。モードの特徴音を含む和音には、その音を名指しして「なぜそう聞こえるか」の一文を足す（ADR-0015）
- `Analyze(Timeline) (Timeline, error)` — 入力を変更せず、`Harmony` に `Analysis` を付けた複製を返す。ダイアトニックでも長調の借用和音でもないコードがあればエラー。借用和音には `Borrowed: true` を付け、解説に「同主短調（C minor）から借りた和音」の一文を足す

### 書き出し

- `smf.Write(Timeline) ([]byte, error)`（`internal/music/smf`）— Standard MIDI File format 0。テンポ・拍子・調号のメタイベント、トラックごとのプログラムチェンジ、ノートのオン／オフを 1 トラックに統合する。同一 tick ではオフ → オンの順（同じ音の連打を正しく切る）。`Analysis` やトラック名は含まれない

### 音楽的意図（`internal/intent`）

- `intent.Intent{Summary, Mood, Candidates}` — LLM が 1 ターンで返す構造化の答え。`Summary` は返答文、`Mood` はタグ、`Candidates` は 1〜3 個の提案
- `intent.Candidate{Title, Key, TempoBPM, Progression, BeatsPerChord, Reason, Accompaniment?, Drums?, Feel?}` — エンジンが描画できる具体案。`Progression` はローマ数字の文字列（`ParseRomanNumeral` の語彙）、`BeatsPerChord` は 1 / 2 / 4 / 8、`Accompaniment` / `Drums` はパターンカタログの id、`Feel` は 8 分音符のノリ（`straight` / `swing`。`Schema` の enum とシステムプロンプトの一覧はカタログから生成される）
- 綴り（`spelling.go`）: 調の親となる長調（短調は主音から長 6 度上。モードを足すときは `parentOffset` に 1 行）の主音の綴りを表で決め、その音階を 7 つの文字に乗せる。ダイアトニックな音は文字で、それ以外は調の側（♭ の調なら ♭）で綴る。`HarmonyEvent.Symbol` と解説文はこれを使い、画面は `symbol` をそのまま出す（`keyName` の表だけフィクスチャで一致を検査）
- `Validate()` が全問題をまとめて返す（テンポ 40〜220、進行 1〜16 個、ローマ数字の妥当性など）。ラベルの書き方（大文字・小文字、7 の種類）が調から決まる和音と食い違うもの（短調の `IV`、長調の `I7`）は、`music.CanonicalLabel` が「何として鳴るか」と正しい表記を添えて落とす。黙って直さない（ADR-0013）。`Parse` は通ったラベルをエンジンの表記にそろえる（`iim7b5` → `iiø7`）。`Schema` は同じ制約を LLM に伝えるが、構造化出力が受け付ける JSON Schema のサブセット（`maxItems` や `minimum` / `maximum` は不可）で書けない上限は説明文で伝え、`Validate` が最終判定する（サブセットの遵守は `TestSchemaUsesOnlySupportedKeywords` が検査）。`Parse(data)` は復号と検証
- `Candidate.ToProgression()` で `music.Progression` に変換し、`FromProgression` → `Analyze` の既存経路に乗せる

### パターン（`internal/music/pattern`）

- `pattern.Accompaniment{ID, Name, Description, Program, Hits}` — 1 小節（4/4、`DefaultPPQ`）の伴奏テンプレート。`Hit{Start, Duration, Voice, Velocity}` の `Voice` はコードの密集ボイシングに対する相対指定（`chord` / `bass` / `root` / `third` / `fifth` / `top` / `octave`）
- `pattern.Drums{ID, Name, Description, Hits}` — 1 小節のドラムテンプレート。`DrumHit.Note` は GM ドラムマップ（36 キック、38 スネア、42 クローズドハイハット…）
- カタログ: 伴奏 `block` / `pulse` / `ballad` / `arpeggio` / `strum`、ドラム `eight-beat` / `ballad` / `bossa`、ノリ `straight` / `swing`。`Accompaniments()` / `DrumPatterns()` / `Feels()` / `FindAccompaniment(id)` / `FindDrums(id)`
- `pattern.Apply(tl, acc, drums)` — `Harmony` を元にトラックを作り直す。コードごとに小節単位でテンプレートを繰り返し、小節より短いコードは先頭部分だけ鳴らす。ドラムは ch.9 にタイムライン長ぶん繰り返す。入力は変更しない。4/4 以外はエラー
- `pattern.Render(tl, Choice{Accompaniment, Drums, Feel})` — id で選んで `Apply` し、ノリを掛ける。HTTP と `ideas` の入口。伴奏もドラムもなければそのまま返し、ドラムだけならブロックコードを使う
- ノリ（`Feel`）: `swing` は、拍の頭の 8 分を拍の 3 分の 2 まで伸ばし、裏拍（8 分の 2 つ目）で始まる音を 3 連符の 3 つ目まで遅らせる（終わりの位置は保つ）。4 分音符や 16 分音符は動かさない。拍の位置の分析はテンプレートの位置で行うので、スウィングしても裏拍は裏拍として数える
- 説明文（`Description`）は解説レイヤーと LLM がそのまま使う前提で、日本語で「何が起きるか」を書く
- `music.Rhythm{Accompaniment, Drums *RhythmPart, Feel *RhythmFeel}` — `Apply` が `Timeline.Rhythm` に入れる解説データ。`RhythmPart` は 1 小節の打点を拍の頭 / 8 分の裏 / それより細かい位置に分類し（ドラムはキックとスネアのみ）、裏拍が 4 分の 1 以上なら `Syncopated`。`Explanation` は決定論的な日本語

### 未実装の語彙

この文書は実装済みの型だけを書き、将来の型は backlog を唯一の情報源とする。

## 設計上の判断

- **質はスケールから導出する。** `Key.Diatonic` は三度堆積の音程から質を判定するため、モードを追加してもコードの表が不要。
- **ローマ数字の文字列は境界でだけ扱う。** 内部の分析は `degree + quality` で行い、文字列は入力（API、将来は LLM の出力）を `ParseRomanNumeral` で度数と七の和音の有無に変換するときにだけ現れる。コードの質は常に調から導出され、文字列の大文字小文字や `°` は結果に影響しない。
- **自然短音階 + 和声的短音階のドミナント。** 短調で `V` / `V7`（大文字）と書くと `StepFromLabel` が `Step.Quality` を長三和音・属七に上書きし、導音（第 7 音の半音上げ）が鳴る。根音は自然短音階上にあるので `Analyze` はそのまま度数を求め、解説に和声的短音階の一文を足す。小文字 `v` は自然短音階の短三和音。ドリアンとミクソリディアンにはこの例外がなく、`V` と書けば `CanonicalLabel` が落とす。
- **長調の借用和音（BL-094）。** 長調では、同主短調から借りた `iv`（`iv7`、`ivm6` も）、`ii°`（`iiø7`）、`bVI`（`bVImaj7`）、`bVII`（`bVII7`）を書ける。`Step.Accidental`（0 か -1）が根音を半音下げ、`Step.Borrowed` が印になる。ラベルの ♭ は `♭` でも `b` でもよく、`CanonicalLabel` が `♭` にそろえる。和音の質は同主短調のその度数と一致しなければならず（長調の `IV7` や `bVI7` は落ちる）、綴りは調の文字から決める（C major の `bVI` は Ab、E major の `bVII` は D）。このセット以外の借用、短調側の借用、セカンダリドミナント、その他の非ダイアトニックなコードは未対応で、`Analyze` は拒否する。
