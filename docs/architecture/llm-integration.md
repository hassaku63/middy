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

## Anthropic アダプタ実装メモ（BL-011 向け）

実装時は Claude Code の `claude-api` スキル（`/claude-api`）を参照し、以下を守る。

- SDK: `github.com/anthropics/anthropic-sdk-go/v2`。生 HTTP は書かない。
- モデル: 既定 `claude-opus-5-5`（`ANTHROPIC_MODEL` で上書き）。意図解釈は短い出力なので、コストが問題になったら評価セット（BL-015）で `claude-sonnet-5-5` を比較してから切り替える。
- 構造化出力: `output_config.format` に JSON Schema を渡す。`tool_choice` の強制は現行モデルで不可。
- 思考: パラメータは省略（既定の adaptive）。`output_config.effort` は `low`〜`medium` から始める。
- 拒否: `stop_reason == "refusal"` を必ず確認し、サーバー側 `fallbacks: "default"` を有効化する。
- タイムアウト: `option.WithRequestTimeout` で 30 秒程度。ユーザー向けには「考え中」の UI を出す。
- プロンプトキャッシュ: システムプロンプトは固定文字列にし、可変部分（ユーザー文）は後ろに置く。
- エラー: `errors.As` で `*anthropic.Error` を取り、429/5xx はリトライ対象、4xx は即時エラーとして API の `error.code` に対応づける。
- テスト: HTTP レベルで録画したレスポンスを `httptest.Server` から返す。実 API を叩くテストは `//go:build integration` タグで分離し、CI では実行しない。

## プロンプト設計の原則

- システムプロンプトは「音楽理論の家庭教師」として、**出力はスキーマに従う JSON のみ** と指示する。
- ローマ数字は `I, ii, iii, IV, V, vi, vii°` と七の和音記法のみを許可し、`intent.Validate()` で検証する。検証に落ちたら、その答えを assistant 発言として、検証エラーを user 発言として足し、1 回だけ再試行する（`ideas.Service`）。
- 理由文（`reason`）は 1〜2 文の日本語。解説レイヤーの文章はエンジンの分析結果を素材に、LLM に「言い換え」だけをさせる（M2 以降）。
- 評価セット（BL-015）で回帰を見る。プロンプト変更は必ず評価結果を PR に添える。
