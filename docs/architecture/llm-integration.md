# LLM 連携の方針

## 役割分担（ADR-0003）

| 担当 | やること | やらないこと |
| --- | --- | --- |
| LLM | イメージ文の解釈、音楽パラメータ（`MusicalIntent`）の提案、ユーザーとの対話、解説文の言い換え | 音符の生成、コードの構成音の決定、分析結果の捏造 |
| `music` エンジン | 進行→タイムライン生成、分析、SMF | 自然言語の理解 |

理由: 音符レベルの生成を LLM に任せると、再現性・検証可能性・説明との一致が失われる。パラメータだけなら JSON Schema で縛れ、理論エンジンで検証できる。

## 境界

`backend/internal/llm`:

```go
type Client interface {
    Complete(ctx context.Context, req Request) (Response, error)
}
```

- `Request.OutputSchema` を指定すると構造化出力（JSON）を要求する。
- `fake.Client` は台本どおりに返す。台本がなく構造化出力を求められたときは、入力の気分語から決定論的な見本 Intent を返す（開発と E2E 用）。**ユースケースのテストはすべて fake で書き、CI はネットワークなしで通す。**
- 実アダプタは `internal/llm/anthropic`（BL-011）。プロバイダ固有の型はそこから外に出さない。

## Anthropic アダプタ（`internal/llm/anthropic`）

- SDK: `github.com/anthropics/anthropic-sdk-go/v2`（Go 1.25 以上を要求するため、`.tool-versions` も 1.25）。生 HTTP は書かない。
- モデル: 既定 `claude-opus-5-5`（`ANTHROPIC_MODEL` で上書き）。意図解釈は短い出力なので、コストが問題になったら評価セット（BL-015）で `claude-sonnet-5-5` を比較してから切り替える。
- 構造化出力: `llm.Request.OutputSchema` を `output_config.format`（`json_schema`）として送る。`tool_choice` の強制は使わない。API は JSON Schema のサブセットだけを受け付け、それ以外のキーワードはリクエスト時に 400 になる（fake では検出できない）。`intent.Schema` は `TestSchemaUsesOnlySupportedKeywords` で守る。
- 思考: パラメータは省略（モデル既定の adaptive）。`output_config.effort` は `low`（`Config.Effort` で変更可）。
- 拒否・打ち切り: `stop_reason` が `refusal` なら `llm.ErrRefused`（アダプタが包んで返す）、`max_tokens` なら `ErrTruncated`（構造化出力の途中切れは使えないためエラー扱い）。
- 拒否への対応: 安全分類器は確率的に判定し、無害な音楽の相談でも誤検出がある（評価セットで `reasoning_extraction` を 1 件観測）。`ideas.Service` は拒否されたら同じ要求で 1 回だけ再試行し、2 回とも拒否なら `ErrDeclined` を返す。HTTP では 422 `llm_declined` になり、UI は言い回しを変えるよう促す。公式ドキュメントは、出力の項目でモデルの推論や思考過程を求めるとこの拒否が起きやすいとしているので、`reason` は「短い解説」として求める。サーバー側 fallback（別モデルへの切り替え）は SDK の beta API が必要なので導入していない。再試行しても拒否が残るようなら検討する。
- enum の大文字小文字: 構造化出力は enum の値の大文字小文字を保証しない。`music.Mode` は大文字小文字を区別せずに読み、`intent.Parse` はパターン id を小文字にそろえる。ローマ数字は大文字小文字に意味があるので触らない。
- タイムアウト: 30 秒（`Config.Timeout`）。SDK の既定リトライ（2 回）は 429 / 5xx / 接続エラーに効く。
- プロンプトキャッシュ: システムプロンプトを 1 ブロックで送り、`cache_control: ephemeral` を付ける。可変部分（会話）は後ろ。
- エラー: `*anthropic.Error` はステータスとリクエスト ID を含めて包み直す。`ideas.Service` はすべて `ErrLLMUnavailable` として扱い、HTTP は 502 にする。
- テスト: `httptest.Server` が録画相当の Messages API 応答を返し、送信した本文（system のキャッシュ指定、messages の役割、`output_config`）を検証する。実 API を叩くテストは置かない（評価セット BL-015 が手動実行で担う）。
- 配線: `LLM_PROVIDER=anthropic` と `ANTHROPIC_API_KEY` を、起動時の環境変数で渡す（ADR-0011）。鍵がなければ起動時に設定エラー。実 LLM での起動は人が行い、エージェントは起動済みのサーバーに触る。

## プロンプト設計の原則

- システムプロンプトは `ideas.SystemPrompt()` がプロセス内で 1 度だけ組み立てる固定文字列。伴奏・ドラム・ノリの一覧は `pattern` のカタログから生成するので、カタログを変えればプロンプトと `intent.Schema` の enum が同時に追従する。

- システムプロンプトは「音楽理論の家庭教師」として役割と守ることを伝える。出力の形式は構造化出力が保証するので、プロンプトでは指示しない。
- ローマ数字は `I, ii, iii, IV, V, vi, vii°` と七の和音記法のみを許可し、`intent.Validate()` で検証する。書き方が調の和音と食い違うラベル（旋法やブルースを狙ったもの）は、「`IV` は D minor では Gm（`iv`）になる」と伝えて再試行させる。検証に落ちたら、その答えを assistant 発言として、検証エラーを user 発言として足し、1 回だけ再試行する（`ideas.Service`）。
- 理由文（`reason`）は、その候補の聴きどころの短い解説（1〜2 文の日本語）として求める。解説パネルの文章（和声機能・終止・リズム）はエンジンが決定論的に組み立て、LLM は関与しない。
- 評価セット（`backend/internal/ideas/evals/cases.json`）で回帰を見る。`make eval`（鍵が必要、CI では走らない）が各ケースを実 API に送り、最初の候補のモード・テンポ帯・ノリ（ケースが指定したとき）を判定して、進行つきの Markdown の表を出力する。表のあとに提案の幅の指標（進行の種類、最も多い進行、調の種類、使った語彙の割合、期待した語彙 `wants` の命中）と、モデルが「制約がなければ書きたかったこと」を出す。この欄は評価のときだけ `intent.DiagnosticSchema` で求め（`ideas.NewDiagnostic`）、本番の要求には入らない（ADR-0013）。`wants` の外れは失敗ではなく、表せない語彙を測るための指標。4 分の 1 を超えて失敗したら回帰とみなす。**プロンプトを変えた PR には表を貼る。** 結果はリポジトリにコミットしない（ADR-0006）。判定器と実行経路は fake で常にテストされる。
