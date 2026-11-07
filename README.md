# middy

作曲スキルと音楽理論の理解・実践を助ける Web アプリ（PWA）。

「雨の日曜の午後みたいな曲」といった **イメージを AI が解釈** し、それに合う **コード進行とリズムのパターンを提案** します。チャットで相談しながら候補を絞り、生成された **タイムラインをその場で再生** できます。さらに、そのトラックの **コード進行やリズムの成り立ちを視覚的な支援つきで解説** します。

> 何をどの順で作るかは [`docs/product/roadmap.md`](docs/product/roadmap.md)、進捗は [`docs/product/backlog.md`](docs/product/backlog.md) を参照してください。

## 何が特別か

- **聴いて、見て、理解する**: 理論の説明を読むだけでなく、自分のイメージから生まれた進行を再生し、ローマ数字分析やリズムの構造をハイライト付きで確認できる。
- **AI は「解釈役」、音楽理論エンジンは「生成役」**: LLM はイメージを音楽パラメータ（調・テンポ・進行テンプレートなど）に翻訳し、実際の音符は決定論的なエンジンが作る。再現性があり、テストでき、説明と音が必ず一致する（[ADR-0003](docs/architecture/adr/0003-llm-interprets-engine-generates.md)）。
- **PWA**: インストールしてスタンドアロンで使える。

## 技術スタック

| 層 | 採用 | 理由（詳細は [ADR-0002](docs/architecture/adr/0002-tech-stack.md)） |
| --- | --- | --- |
| フロントエンド | React 19 + TypeScript + Vite, Vitest, vite-plugin-pwa | 再生・ピアノロール・チャット UI はブラウザ側が主戦場 |
| バックエンド | Go 1.24, 標準 `net/http`, `log/slog` | 高速なビルド/テスト、少ない依存、AI が自律的に回しやすい |
| 音楽理論エンジン | Go `backend/internal/music`（純粋関数） | 決定論・テーブル駆動テスト |
| LLM | Anthropic Claude（`backend/internal/llm` 経由、既定は `fake`） | 構造化出力でパラメータを得る |
| API 契約 | OpenAPI 3.1（`docs/api/openapi.yaml`） | フロント・バックの唯一の合意点 |

## クイックスタート

前提: Go 1.24+, Node.js 22+, [golangci-lint](https://golangci-lint.run/) v2。

```bash
make setup          # 依存のインストール
make check          # フォーマット・lint・テスト・ビルド・ドキュメント検査を全部
make dev-backend    # :8080 で API（LLM は fake）
make dev-frontend   # :5173 で UI（/api を :8080 へプロキシ）
```

ブラウザで http://localhost:5173 を開くと、バックエンドのヘルスチェック結果が表示されます。

実際の LLM を使うときは `backend/.env.example` を `backend/.env` にコピーして `LLM_PROVIDER=anthropic` と `ANTHROPIC_API_KEY` を設定します（Anthropic アダプタ自体は未実装。[backlog](docs/product/backlog.md) 参照）。

## リポジトリ構成

```
.
├── AGENTS.md              # AI エージェント向けの正典（Claude / Codex / Copilot 共通）
├── CLAUDE.md              # Claude Code 固有の補足（AGENTS.md を取り込む）
├── Makefile               # make check = 唯一のゲート
├── scripts/               # ドキュメント検査などの補助スクリプト
├── backend/               # Go API + 音楽理論エンジン
│   ├── cmd/server/        # エントリポイント
│   └── internal/
│       ├── config/        # 環境変数の読み込み
│       ├── httpapi/       # HTTP ハンドラ（薄い）
│       ├── llm/           # LLM 境界（interface + fake）
│       └── music/         # 音楽理論ドメイン（純粋・無依存）
├── frontend/              # React PWA
│   └── src/
│       ├── app/           # アプリシェル
│       ├── features/      # 機能ごとの縦スライス（今後）
│       └── shared/        # API クライアント等の共有コード
└── docs/
    ├── product/           # ビジョン・ロードマップ・バックログ
    ├── architecture/      # 全体像・ドメインモデル・ADR
    ├── dev/               # 開発ワークフロー・規約・テスト方針
    └── api/openapi.yaml   # API 契約
```

## ドキュメント

- [docs/README.md](docs/README.md) — ドキュメントの索引
- [プロダクトビジョンとシナリオ](docs/product/vision.md)
- [ロードマップ](docs/product/roadmap.md) / [バックログ](docs/product/backlog.md)
- [アーキテクチャ概要](docs/architecture/overview.md) / [ドメインモデル](docs/architecture/domain-model.md)
- [開発ワークフロー（AI 自律開発のハーネス）](docs/dev/workflow.md)
- [ドキュメント鮮度の維持](docs/dev/docs-hygiene.md)

## ライセンス

未定（リポジトリ所有者が決めるまで All rights reserved 扱い）。
