# middy

作曲スキルと音楽理論の理解・実践を助ける Web アプリ（PWA）。

「雨の日曜の午後みたいな曲」といった **イメージを AI が解釈** し、それに合う **コード進行とリズムのパターンを提案** します。チャットで相談しながら候補を絞り、生成された **タイムラインをその場で再生** できます。さらに、そのトラックの **コード進行やリズムの成り立ちを視覚的な支援つきで解説** します。

> 何をどの順で作るかは [`docs/product/roadmap.md`](docs/product/roadmap.md)、進捗は [`docs/product/backlog.md`](docs/product/backlog.md) を参照してください。

## 何が特別か

- **聴いて、見て、理解する**: 理論の説明を読むだけでなく、自分のイメージから生まれた進行を再生し、ローマ数字分析やリズムの構造をハイライト付きで確認できる。
- **AI は「解釈役」、音楽理論エンジンは「生成役」**: LLM はイメージを音楽パラメータ（調・テンポ・進行テンプレートなど）に翻訳し、実際の音符は決定論的なエンジンが作る。再現性があり、テストでき、説明と音が必ず一致する（[ADR-0003](docs/architecture/adr/0003-llm-interprets-engine-generates.md)）。
- **PWA**: インストールしてスタンドアロンで使える。保存したプロジェクトはオフラインでも開いて再生できる。

## 技術スタック

| 層 | 採用 | 理由（詳細は [ADR-0002](docs/architecture/adr/0002-tech-stack.md)） |
| --- | --- | --- |
| フロントエンド | React 19 + TypeScript + Vite, Vitest, vite-plugin-pwa | 再生・ピアノロール・チャット UI はブラウザ側が主戦場 |
| バックエンド | Go 1.24, 標準 `net/http`, `log/slog` | 高速なビルド/テスト、少ない依存、AI が自律的に回しやすい |
| 音楽理論エンジン | Go `backend/internal/music`（純粋関数） | 決定論・テーブル駆動テスト |
| LLM | Anthropic Claude（`backend/internal/llm` 経由、既定は `fake`） | 構造化出力でパラメータを得る |
| API 契約 | OpenAPI 3.1（`docs/api/openapi.yaml`） | フロント・バックの唯一の合意点 |

## クイックスタート

前提: Go、Node.js、[golangci-lint](https://golangci-lint.run/)。版は `.tool-versions` が唯一の情報源で、CI もそれを読みます（asdf / mise なら `asdf install` / `mise install`）。

```bash
make doctor         # 手元の版が .tool-versions と合っているか
make setup          # 依存のインストール
make check          # フォーマット・lint・テスト・ビルド・ドキュメント検査を全部
make e2e            # Playwright の通しシナリオ（make run か make up で起動したサーバーに対して）
make dev-backend    # :8080 で API（LLM は fake）
make dev-frontend   # :5173 で UI（/api を :8080 へプロキシ）
```

本番相当で動かす（単一バイナリ、ADR-0008）:

```bash
make build          # vite build → バイナリに同梱 → go build
make run            # :8080 で UI と API を同一オリジンから配信
```

フロントを同梱せずに `go build` したバイナリは API だけが動き、`/` は 503 を返します。

コンテナで動かす（CI と同じイメージ）:

```bash
make up             # docker compose up --build → http://localhost:8080
make down
LLM_PROVIDER=anthropic ANTHROPIC_API_KEY=... make up   # 実 LLM を使う場合
```

イメージはマルチステージ（Node でフロント → Go でバイナリ → distroless、非 root）で、`HEALTHCHECK` はバイナリ自身の `-healthcheck` フラグが担います。

ブラウザで http://localhost:5173 を開き、「イメージから（AI）」で曲のイメージを言葉で送ると、調・テンポ・進行の候補が返ります（既定の `LLM_PROVIDER=fake` では気分語で変わる見本。実 LLM は `backend/.env` で設定）。「進行から」では調・モード・ローマ数字の進行・テンポを入力して「生成」します。候補を開く／生成すると、タイムラインが表示されます。「再生」で鳴らすと、鳴っているコードとノートがハイライトされ、解説パネルがコードの機能と終止を説明します。コードをクリックすると解説を固定できます。「MIDI を保存」で Standard MIDI File をダウンロードし、DAW に読み込めます。名前をつけて「保存」すると、プロジェクトとしてブラウザ（IndexedDB）に残り、次に開いたときも一覧から開けます。

実際の LLM を使うときは `backend/.env.example` を `backend/.env` にコピーして `LLM_PROVIDER=anthropic` と `ANTHROPIC_API_KEY` を設定します（`make dev-backend` と `make up` の両方が読みます）。

## リポジトリ構成

```
.
├── AGENTS.md              # AI エージェント向けの正典（Claude / Codex / Copilot 共通）
├── CLAUDE.md              # Claude Code 固有の補足（AGENTS.md を取り込む）
├── .tool-versions         # Go / Node / golangci-lint の版（CI・make doctor・Docker が読む）
├── Makefile               # make check = 唯一のゲート
├── Dockerfile / compose.yaml  # 本番相当のイメージとローカル検証（ADR-0008）
├── scripts/               # ドキュメント検査などの補助スクリプト
├── backend/               # Go API + 音楽理論エンジン
│   ├── cmd/server/        # エントリポイント
│   └── internal/
│       ├── config/        # 環境変数の読み込み
│       ├── httpapi/       # HTTP ハンドラ（薄い）と静的配信
│       ├── llm/           # LLM 境界（interface + fake）
│       ├── music/         # 音楽理論ドメイン（純粋・無依存）
│       └── web/           # ビルド済みフロントの embed
├── frontend/              # React PWA
│   └── src/
│       ├── app/           # アプリシェル
│       ├── features/      # 機能ごとの縦スライス（chat, progression-input, player, timeline, explanation, export, projects）
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
