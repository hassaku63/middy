# ドメインモデル

音楽理論の語彙を、コード上の型に対応づける。**ここに書かれた定義が `backend/internal/music` の仕様** であり、テストはこれを根拠にする。曖昧な点があれば、実装より先にこの文書を直す。

## 規約

| 項目 | 規約 |
| --- | --- |
| ピッチ | MIDI ノート番号 0..127。**C4 = 60**（middle C） |
| ピッチクラス | 0..11、C = 0。表示は既定でシャープ（`C#`）、必要なら `FlatName()` |
| 時間 | tick。`DefaultPPQ = 480`（四分音符 = 480 tick） |
| 度数 | 1 始まり（1 = 主音, 5 = 属音） |
| チャンネル | 0 始まり。9 = GM パーカッション |
| プログラム | GM プログラム番号 0 始まり（0 = Acoustic Grand Piano） |

## 型

### 音高

- `PitchClass` — `Normalize()`, `Add(semitones)`, `String()`, `FlatName()`
- `Pitch` — `NewPitch(pc, octave)`, `Valid()`, `Class()`, `Octave()`, `Transpose()`, `String()`（`"C#4"`）

### 調と音階

- `Mode` — `Major`, `Minor`（自然短音階）。JSON では `"major"` / `"minor"`。モード追加は `modeIntervals` に 1 行足すだけで `Diatonic` が追従する
- `Key{Tonic, Mode}` — `Scale()`（7 音）, `Degree(n)`, `Contains(pc)`, `Diatonic(degree, seventh)`

### コード

- `ChordQuality` — 三和音 4 種（maj, min, dim, aug）と七の和音 4 種（maj7, min7, dom7, m7b5）。JSON は識別子文字列（`"dom7"`）
- `Chord{Root, Quality}` — `PitchClasses()`, `Voicing(octave)`（密集・基本形）, `String()`（`"G7"`）
- `RomanNumeral(degree, quality)` — `I`, `ii`, `vii°`, `V7`, `iiø7`, `IVmaj7` など。大文字 = 長/増、小文字 = 短/減

**未対応（backlog）**: 転回形、分数コード、テンション、借用和音、セカンダリドミナント。

### タイムライン

```
Timeline
├── PPQ, TempoBPM, TimeSignature, Key
├── Tracks[]            楽器レーン
│   └── Notes[]         NoteEvent{Pitch, Velocity, Start, Duration}
└── Harmony[]           HarmonyEvent{Start, Duration, Chord, RomanNumeral}
```

- `Timeline` は **再生・描画・解説・書き出しのすべてが消費する唯一の成果物**。
- `Harmony` は音を出さない注釈。解説レイヤー（和声機能、ケーデンス）はここに情報を足す。
- `Validate()` は API 境界で必ず呼ぶ。内部の関数は妥当な Timeline を前提にしてよい。
- `LengthTicks()` は最後のノート／注釈の終端。

### 進行

- `Step{Degree, Seventh, Beats}` — 度数（1..7）を何拍（四分音符単位）鳴らすか。拍は拍子に依存しない
- `Progression []Step` — 調に依存しない進行テンプレート。`Validate()` は全ステップの問題をまとめて返す
- `FromProgression(key, prog, tempoBPM)` — 進行を調で解決し、ブロックコード 1 トラック（GM ピアノ, ch.0, `ChordOctave`=4 の基本形, `ChordVelocity`=80）と `Harmony` 注釈を持つ `Timeline` を返す。リズム・ボイシングのパターンはこの上に重ねる（BL-020, 021）

### 今後追加される型（M1〜M3）

| 型 | 役割 | 予定 |
| --- | --- | --- |
| `Analysis` / `Function` (T/S/D) / `Cadence` | 解説の素材 | BL-002 |
| `MusicalIntent` | LLM が返す構造化パラメータ | BL-010（`internal/intent`） |
| `Pattern`（伴奏・リズム） | 進行に適用するテンプレート | BL-020 |

## 設計上の判断

- **質はスケールから導出する。** `Key.Diatonic` は三度堆積の音程から質を判定するため、モードを追加してもコードの表が不要。
- **ローマ数字は表示専用の文字列。** 解析の入力としては `degree + quality` を使い、文字列をパースしない（M2 で LLM からローマ数字文字列を受けるときは `intent` パッケージ側でパーサを持つ）。
- **自然短音階のみ。** 和声的短音階の V（長三和音）は M1 の分析（BL-002）で「借用」として扱うか、進行テンプレート側で `V` を明示指定して解決する。ADR 候補。
