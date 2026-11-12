# 型の足し方（伴奏・ドラム・スタイル）

伴奏・ドラム・スタイルの型は、エンジンの一部のデータである（ADR-0035）。ここでは、型を 1 つ足す手順を書く。型の意味（声部、拍子、グルーヴ）は `docs/architecture/domain-model.md` の「パターン」にある。

## 置き場所と形

- `backend/internal/music/pattern/data/accompaniments.json`: 伴奏
- `backend/internal/music/pattern/data/drums.json`: ドラム
- `backend/internal/music/pattern/data/styles.json`: スタイル

形は `docs/architecture/pattern-catalogue.schema.json`（JSON Schema）にある。エディタに読ませると、項目名と値の範囲を補完し、検査する。

ここで使う単位と記法は次のとおり。

- 長さ: tick で書く。4 分音符が 480、8 分音符が 240。1 小節は、4/4 なら 1920、3/4 なら 1440、6/8 なら 1440。
- 打: 1 行に 1 つ書く。
- 音色: General MIDI の番号で書く。伴奏の `program` は 0 始まり。ドラムの `note` は GM のパーカッションの鍵（36 がキック、38 がスネア、42 がクローズのハイハット）。

## 手順

1. **型を書く。** 既存の、近い型を写して直す。
   - `id` は小文字とハイフン。
   - `description` は、学ぶ人に向けた 1〜2 文。何が鳴り、どう聞こえるかを書く。この文は、解説とプロンプトのカタログにそのまま出る。
   - 伴奏で `meter` を省くと、どの拍子にも合う型になる。そのときだけ、`duration` を 0 にした打は小節の終わりまで伸びる。
2. **スタイルに入れる。** 型はどれかのスタイルの `accompaniments` か `drums` に入れる。入れないと `Validate` が落ちる（block は例外）。スタイルが挙げる型の順番は、グルーヴ（スタイルの選び直しで鳴る型）の選び方に効く。先頭に近いほど、そのスタイルらしい型として扱う。
3. **検査する。** `make test`。カタログは起動時に `pattern.Catalogue.Validate` が検査する。壊れていれば、どの項目のどの打かを添えて、全部を一度に返す。
4. **ゴールデンを書き直す。** 次を実行し、差分を確かめる。足した型の描画が増え、ほかの型の描画が変わっていないこと。

   ```bash
   cd backend && go test ./internal/golden ./internal/httpapi ./internal/ideas -run Golden -update
   ```
5. **評価を走らせる。** プロンプトのカタログの一部が変わるので、PR に `needs-eval-cases:style` か `needs-eval-cases:groove` を付ける（ADR-0038、`docs/dev/workflow.md` の 4.1）。AI が新しい型を選ぶ場面が評価のケースにないなら、ケースを足す。

## してはいけないこと

- **プロンプトの固定文で、型を名指しで勧めない。** 新しい型を使わせたいときは、まず `description` とスタイルで足りるかを見る。固定文に型の id を書いた場合は、その id がカタログにあることをテストが確かめる（`TestPromptNamesOnlyCatalogueIDs`）。
- **型の中で和音や調を決めない。** 型が決めるのはリズムと声部だけで、和音はエンジンが決める（ADR-0003）。
