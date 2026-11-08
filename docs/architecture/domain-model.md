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
- 長調の借用の範囲外でも、♭ 付きのラベルを主音からの距離で読んだ和音（`LiteralChord`）がその調の和音と同じなら、受け付けて調の表記にそろえる（ミクソリディアンの `bVII` は `VII`、短調の `bVI` は `VI`。#109）。音は書いたとおりなので、黙って直すことには当たらない。同じにならないもの（短調の `bII`）は落とす
- `LiteralChord(tonic, label)` — ラベルを、調に関係なく、主音の長音階に対して書いたとおりに読む（大文字・小文字と接尾辞が質、先頭の ♭ / b が根音を半音下げる）。`bVII` は主音の全音下の長三和音。評価の `wants` のように「どの調でも同じ響き」を指すときに使う
- `ParseRomanNumeral(s)` — 上記の表記（と `dim`, `aug`, `+`, `m7b5`）を度数と七の和音の有無に戻す。`RomanNumeral` の出力は必ず往復する。大文字・小文字の違いは `StepFromLabel` だけが（短調の V で）使う

**未対応（backlog）**: 転回形、分数コード、テンション、承認済みのセット以外の借用和音（長調の ♭II・♭III、短調の七の和音の借用など）、V 以外を当てる和音（`vii°/x` など）。

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
- `FromProgressionIn(ts, key, prog, tempoBPM)` — 同じことを別の拍子で行う（4/4、3/4、6/8。ADR-0017）。`Step.Beats` は拍子の拍で数える（4/4 と 3/4 では 4 分音符、6/8 では付点 4 分。`BeatTicks(ts, ppq)`）。6/8 のような複合拍子（`TimeSignature.Compound()`）は、拍が 8 分音符 3 つに分かれ、1 小節を 2 拍と数える（`BeatsPerBar()`）。`FromProgression` は 4/4 のこれ

### 分析（解説データ）

- `Function` — `T` / `S` / `D`。度数で決まる: I, III, VI = T、II, IV = S、V, VII = D。長調・自然短音階とも同じ割り当て（短調の v と VII は「弱いドミナント」扱い。VII は下の段落のとおり隣の和音で働きを言う。借用和音は、その度数の機能をそのまま使う: iv は S、♭VI は T、♭VII は D）
- `Cadence` — 直前のコードからの進行の分類: `authentic`（V→I）、`plagal`（IV→I）、`half`（V 以外→V）、`deceptive`（V→vi）、`modal`（ドリアン・ミクソリディアンの VII→I。導音を使わない）。該当なしは空。`authentic` / `deceptive` / `half` の V は、導音を持つ V（長三和音か属七）に限る。自然短音階・ドリアン・ミクソリディアンの短三和音の `v` からの進行には、終止の名前を付けない（ADR-0012 の判断 2）
- 短調の `VII` は、隣の和音で働きを言う（ADR-0012 の判断 2）: 次が `III` なら III の V にあたる働き（V/III）、次が `i` なら導音のない弱いドミナント、`i` と `VI` のあいだなら V へ下りる線の途中（経過）。機能の印は度数どおり D のまま、解説の一文で補う
- `Overview{Whole, Flow}` — `Timeline.Overview` に付く例全体の解説で、`Analyze` が最後に作る（BL-098）。全体は、モードとその性格に加えて、ドリアンとミクソリディアンでは特徴音を含む和音と、比べる相手のモード（ドリアンは短調、ミクソリディアンは長調）ならその位置が何になるかを名指しする。長調では借用和音、短調では和声的短音階の V を名指しする。流れは、和音の数と小節数、始まりの和音、最後の終止、繰り返したときに頭へ戻る終止を言う
- `Analysis{Degree, Function, Cadence, Explanation}` — `HarmonyEvent.Analysis` に付く。`Explanation` は UI にそのまま出せる日本語で、決定論的に組み立てる（LLM は使わない）。機能（T/S/D）は全モードで度数から付ける初学者向けの単純化。モードの特徴音を含む和音には、その音を名指しして「なぜそう聞こえるか」の一文を足す（ADR-0015）
- `Analyze(Timeline) (Timeline, error)` — 入力を変更せず、`Harmony` に `Analysis` を付けた複製を返す。ダイアトニックでも長調の借用和音でもないコードがあればエラー。借用和音には `Borrowed: true` を付け、解説に「同主短調（C minor）から借りた和音」の一文を足す

### 書き出し

- `smf.Write(Timeline) ([]byte, error)`（`internal/music/smf`）— Standard MIDI File format 0。テンポ・拍子・調号のメタイベント、トラックごとのプログラムチェンジ、ノートのオン／オフを 1 トラックに統合する。同一 tick ではオフ → オンの順（同じ音の連打を正しく切る）。`Analysis` やトラック名は含まれない

### 音楽的意図（`internal/intent`）

- `intent.Intent{Summary, Mood, Candidates}` — LLM が 1 ターンで返す構造化の答え。`Summary` は返答文、`Mood` はタグ、`Candidates` は 1〜3 個の提案
- `intent.Candidate{Title, Key, TempoBPM, Progression, BeatsPerChord, Reason, Accompaniment?, Drums?, Feel?, Meter?}` — エンジンが描画できる具体案。`Progression` はローマ数字の文字列（`ParseRomanNumeral` の語彙）、`Meter` は `4/4`（省略時）か `3/4`（ADR-0017）、`BeatsPerChord` は拍子の拍の数で 4/4 なら 1 / 2 / 4 / 8、3/4 なら 3 / 6（1 小節か 2 小節）、伴奏とドラムはその拍子の型だけ、`Accompaniment` / `Drums` はパターンカタログの id、`Feel` は 8 分音符のノリ（`straight` / `swing`。`Schema` の enum とシステムプロンプトの一覧はカタログから生成される）
- 綴り（`spelling.go`）: 調の親となる長調（短調は主音から長 6 度上。モードを足すときは `parentOffset` に 1 行）の主音の綴りを表で決め、その音階を 7 つの文字に乗せる。ダイアトニックな音は文字で、それ以外は調の側（♭ の調なら ♭）で綴る。`HarmonyEvent.Symbol` と解説文はこれを使い、画面は `symbol` をそのまま出す（`keyName` の表だけフィクスチャで一致を検査）
- `Validate()` が全問題をまとめて返す（テンポ 40〜220、進行 1〜16 個、ローマ数字の妥当性など）。ラベルの書き方（大文字・小文字、7 の種類）が調から決まる和音と食い違うもの（短調の `IV`、長調の `I7`）は、`music.CanonicalLabel` が「何として鳴るか」と正しい表記を添えて落とす。黙って直さない（ADR-0013）。`Parse` は通ったラベルをエンジンの表記にそろえる（`iim7b5` → `iiø7`）。`Schema` は同じ制約を LLM に伝えるが、構造化出力が受け付ける JSON Schema のサブセット（`maxItems` や `minimum` / `maximum` は不可）で書けない上限は説明文で伝え、`Validate` が最終判定する（サブセットの遵守は `TestSchemaUsesOnlySupportedKeywords` が検査）。`Parse(data)` は復号と検証
- `Candidate.ToProgression()` で `music.Progression` に変換し、`FromProgression` → `Analyze` の既存経路に乗せる

### パターン（`internal/music/pattern`）

- `pattern.Accompaniment{ID, Name, Description, Program, Meter, Hits}` — 1 小節（`DefaultPPQ`）の伴奏テンプレート。`Meter` はその型が書かれた拍子で、`nil` はどの拍子でも使える（`block`。`Duration` 0 は小節の終わりまで）。`Hit{Start, Duration, Voice, Velocity}` の `Voice` はコードの密集ボイシングに対する相対指定（`chord` / `bass` / `root` / `third` / `fifth` / `top` / `octave`）
- `pattern.Drums{ID, Name, Description, Meter, Hits}` — 1 小節のドラムテンプレート。拍子を必ず宣言する。`DrumHit.Note` は GM ドラムマップ（36 キック、38 スネア、42 クローズドハイハット…）
- カタログ: 伴奏は 4/4 の `pulse` / `ballad` / `arpeggio` / `strum`、3/4 の `waltz`（1 拍目にベース、2・3 拍目に和音）/ `arpeggio-3`、6/8 の `rocking-6`（拍の頭にベース、残りの 8 分 2 つに和音）/ `arpeggio-6`、どの拍子でも使える `block`。ドラムは 4/4 の `eight-beat` / `ballad` / `bossa`、3/4 の `waltz`、6/8 の `six-eight` / `jig`。ノリ `straight` / `swing`。`Accompaniments()` / `DrumPatterns()` / `AccompanimentsFor(ts)` / `DrumPatternsFor(ts)` / `Feels()` / `FindAccompaniment(id)` / `FindDrums(id)`。AI に見せるのは、インテントが許す拍子の型だけ
- `pattern.Apply(tl, acc, drums)` — `Harmony` を元にトラックを作り直す。コードごとに小節単位でテンプレートを繰り返し、小節より短いコードは先頭部分だけ鳴らす。ドラムは ch.9 にタイムライン長ぶん繰り返す。入力は変更しない。小節の長さは拍子から決める（`music.BarTicks`）。4/4・3/4・6/8 以外、または型の拍子がタイムラインと違うときはエラー（ADR-0017）。リズムの解説は拍子の拍で拍の頭を数え、6/8 では拍を 3 つに分けた 8 分をシンコペーションに数えない。`pattern.Render` は 6/8 でのスウィングを拒む（拍がすでに 3 つに分かれている）
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
- **セカンダリードミナント（BL-074）。** `V/x` と `V7/x` は、度数 x の和音の 5 度上の長三和音か属七で、x を一時的な主音のように聞かせて引き寄せる（C major の `V7/V` は D7 で G へ向かう）。`Step.Applied` が立つと、`Step.Degree` は向かう先の度数を表す。向かう先は、主和音以外の長三和音か短三和音に限る（減三和音は自分のドミナントで導かれず、`V/I` は `V` そのもの）。和音が調の中にあるものは、別名にすぎないので落とす（C major の `V/IV` は主和音の C で、`V7/IV` の C7 だけがセカンダリードミナント。C minor の `V/III` は `VII`、ミクソリディアンの `V7/IV` は `I7`）。根音は、向かう先の文字の 5 度上の文字で綴る（C# minor の `V/v` は D#）。`Analyze` は、調・和声的短音階の V・借用で説明できない和音のうち、この形のものを機能 D、`AppliedTo` に向かう先の度数を入れて分析し、次の和音がその先なら「解決する」、違えば「かわす」と書く。セカンダリードミナント自体には終止を付けない。
- **長調の借用和音（BL-094）。** 長調では、同主短調から借りた `iv`（`iv7`、`ivm6` も）、`ii°`（`iiø7`）、`bVI`（`bVImaj7`）、`bVII`（`bVII7`）を書ける。`Step.Accidental`（0 か -1）が根音を半音下げ、`Step.Borrowed` が印になる。ラベルの ♭ は `♭` でも `b` でもよく、`CanonicalLabel` が `♭` にそろえる。和音の質は同主短調のその度数と一致しなければならず（長調の `IV7` や `bVI7` は落ちる）、綴りは調の文字から決める（C major の `bVI` は Ab、E major の `bVII` は D）。このセット以外の借用、その他の非ダイアトニックなコード（下のセカンダリードミナントを除く）は未対応で、`Analyze` は拒否する。
- **短調の借用和音（BL-075）。** 短調では、同主長調の `I`（曲の終わりならピカルディの 3 度）と `IV`（旋律的短音階やドリアンの 6 度）、ナポリの `♭II`（主音の半音上の長三和音）を書ける。どれも三和音で、その調の音階に対して書き、音階にない根音（ナポリ）にだけ ♭ を付ける（ADR-0012 の判断 3）。機能は度数どおり（I は T、IV と ♭II は S）。C minor の C major のように、同じ和音がセカンダリードミナント（`V/iv`）にもなる場合は、ラベルで決める（`/` を含めば向かう先のある和音、なければ借用）
