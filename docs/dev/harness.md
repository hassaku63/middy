# 開発ハーネスの一覧

`make check` と CI、定期ジョブが行う検査の全リスト。**検査を足すときは必ずここに 1 行足す。** 四半期に一度、各行を見直し「6 か月以上なにも捕まえておらず、コストだけがある検査」は削除候補にする（ADR-0007）。

実行はすべて `make check`（= CI）に含まれ、個別には右の列のコマンドで走らせる。

## コードの正しさ

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| Go テスト（race, cover） | 振る舞いの退行 | `make backend-test` | `make check`, CI `backend` |
| Vitest + Testing Library | 振る舞いの退行 | `make frontend-test` | `make check`, CI `frontend` |
| `tsc -b` | 型の矛盾 | `make frontend-typecheck` | `make check`, CI `frontend` |
| ルート ↔ OpenAPI 整合性テスト | 契約と実装のズレ（ADR-0004） | `cd backend && go test ./internal/httpapi/ -run TestRoutesMatchOpenAPI` | `make check`, CI `backend` |
| 言語またぎフィクスチャ（Go が生成、TS が照合） | `notation.ts` 等のバックエンドの写しがずれる。JSON の陳腐化は Go テストが検出 | `make fixtures` で再生成、`cd backend && go test ./internal/fixtures/` | `make check`, CI `backend` + `frontend` |

## テストの健全性（検知）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| テスト所要時間レポート（上位 10 件、合計、パッケージ別。予算: backend 60 秒 / frontend 30 秒超で警告） | 気づかないうちに遅くなるスイート、価値に見合わない重いテスト | `make test-report` | CI `backend` / `frontend` の各ステップ |

## 構造と規約（矯正）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| gofmt / goimports / Prettier | 書式のばらつき | `make fmt-check` | `make check`, CI |
| golangci-lint 標準セット + revive, gocritic, misspell | 典型的なバグと不統一 | `make backend-lint` | `make check`, CI `backend` |
| `depguard`（Go） | 境界違反: `music` の I/O、`httpapi` からアダプタ直 import、`net/http` の漏れ | `make backend-lint` | `make check`, CI `backend` |
| `funlen` 80 行 / `gocognit` 20 / `gocyclo` 15 | 肥大した関数（テストは除外） | `make backend-lint` | `make check`, CI `backend` |
| ESLint 型つき推奨 + react-hooks | 典型的なバグ | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint feature 境界 | feature 間 import、`../` での越境 | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint コンポーネント純粋性 | `*.tsx` での API 呼び出し・`fetch` / `AudioContext` / ストレージ / タイマー | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint `max-lines` 400 / `max-lines-per-function` 120 / `complexity` 15 | 肥大したファイル・関数（テストの関数長は除外） | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint スナップショット禁止 | 実装に結合したテスト | `make frontend-lint` | `make check`, CI `frontend` |
| PR サイズ予算（400 警告 / 600 失敗） | レビュー不能な PR | `bash scripts/check-pr-size.sh main` | CI `pr-size`（PR のみ） |

## ドキュメント（ADR-0006）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| リンク・パス・索引・BL-ID | 存在しないものを指す文書 | `make docs-check` | `make check`, CI `docs` |
| AGENTS.md / CLAUDE.md の行数予算（150） | 読まれない指示書 | `make docs-check` | `make check`, CI `docs` |
| 鮮度レポート（In progress 放置、120 日未更新、腐りやすい言い回し） | 古い記述の放置 | `bash scripts/docs-staleness.sh` | 月次 `docs gardening` |

## ビルドと配備（ADR-0008）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `go build` / `vite build` | ビルド不能 | `make build` | `make check`, CI |
| コンテナイメージのビルドとスモーク（health の SHA、`/` の HTML、`sw.js` の `no-cache`、`HEALTHCHECK` healthy） | Dockerfile の腐敗、配信の退行 | `make image` | CI `image` |
| E2E（Playwright, S1 シナリオ・API エラー・SPA フォールバック・Service Worker） | モックでは見えない通しの退行 | `make e2e`（サーバー起動が前提） | `e2e` ワークフロー（手動・週次） |

## 依存（削減）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `go mod tidy -diff` | `go.mod` / `go.sum` の不整合、不要な依存 | `make deps-check` | CI `deps` |
| `npm audit --omit=dev --audit-level=high` | 本番依存の既知脆弱性 | `make deps-check` | CI `deps` |
| `govulncheck` + `npm audit`（全依存） | 新たに公開された脆弱性 | （CI のみ） | 週次 `security audit` |
| Dependabot（週次、minor/patch と major を別グループ） | 依存の陳腐化 | — | GitHub |

**Dependabot の方針**: minor/patch のグループは CI が緑ならマージしてよい（人間でもエージェントでも）。major は人間が判断し、規約や書き方が変わるなら ADR を書く。`typescript` と `@types/node` の major は除外（理由は `.github/dependabot.yml`）。CI が落ちた更新は閉じて原因を backlog に書く。

## ツールチェーン（ADR-0008）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `.tool-versions` を CI の setup-go / setup-node / golangci-lint-action と Docker のビルド引数が読む | CI と手元・コンテナの版ズレ | — | CI、`make image` / `make up` |
| `make doctor` | 手元の Go / Node / golangci-lint の major.minor ズレ | `make doctor` | セッション開始時（`docs/dev/workflow.md`） |

## 手順（機械化されていない）

| 手順 | 目的 | いつ |
| --- | --- | --- |
| `docs-sync` 自己点検 | 差分で言及が変わる文書の修正 | PR 前（`docs/dev/docs-hygiene.md`） |
| 自己レビュー（`/code-review` 等） | 人間のレビュー負荷の低減 | PR 前 |
| `make e2e` を 1 回通す | モックでは見えない失敗 | UI スライスの PR 前（`docs/dev/testing.md`） |
| WIP 上限 3 | レビュー処理能力の保護 | 常時（`AGENTS.md`） |
| ハーネスの見直し | 効かない検査の削除、CI 時間（予算 10 分）の確認 | 四半期 |
