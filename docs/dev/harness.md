# 開発ハーネスの一覧

`make check` と CI、定期ジョブが行う検査の全リスト。**検査を足すときは必ずここに 1 行足す。** 四半期に一度、各行を見直し「6 か月以上なにも捕まえておらず、コストだけがある検査」は削除候補にする（ADR-0007）。

実行はすべて `make check`（= CI）に含まれ、個別には右の列のコマンドで走らせる。

## コードの正しさ

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| Go テスト（race, cover） | 振る舞いの退行 | `make backend-test`（自動ダウンロードされた toolchain に `covdata` がない環境では `GOTEST_FLAGS=-race` で上書き） | `make check`, CI `backend` |
| Vitest + Testing Library | 振る舞いの退行 | `make frontend-test` | `make check`, CI `frontend` |
| `tsc -b` | 型の矛盾 | `make frontend-typecheck` | `make check`, CI `frontend` |
| ルート ↔ OpenAPI 整合性テスト | 契約と実装のズレ（ADR-0004） | `cd backend && go test ./internal/httpapi/ -run TestRoutesMatchOpenAPI` | `make check`, CI `backend` |
| 生成 API 型の鮮度（OpenAPI → `schema.gen.ts`、ADR-0009） | TS の型と契約のズレ、YAML として壊れた契約 | `bash scripts/check-api-types.sh`（直すときは `cd frontend && npm run gen:api`） | `make check`, CI `frontend` |
| 言語またぎフィクスチャ（Go が生成、TS が照合） | `notation.ts` 等のバックエンドの写しがずれる。JSON の陳腐化は Go テストが検出 | `make fixtures` で再生成、`cd backend && go test ./internal/fixtures/` | `make check`, CI `backend` + `frontend` |

## テストの健全性（検知）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| テスト所要時間レポート（上位 10 件、合計、パッケージ別。予算: backend 60 秒 / frontend 30 秒超で警告） | 気づかないうちに遅くなるスイート、価値に見合わない重いテスト | `make test-report` | CI `backend` / `frontend` の各ステップ（同じテストの実行が書いた JSON から作る。再実行しない。ADR-0016） |
| バンドルサイズ予算（gzip: 最大 JS チャンク 120 kB、JS 合計 160 kB、CSS 20 kB） | PWA の初回ダウンロードの肥大 | `bash scripts/check-bundle-size.sh`（`make frontend-build` が実行） | `make check`, CI `frontend` |
| 1,000 ノート描画の予算テスト（緩い上限 3 秒。実測は `make test-report` に出る） | タイムライン表示の計算量の退行 | `cd frontend && npx vitest run src/features/timeline/TimelineView.perf.test.tsx` | `make check`, CI `frontend` |

## 構造と規約（矯正）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| gofmt / goimports / Prettier | 書式のばらつき | `make fmt-check` | `make check`, CI |
| golangci-lint 標準セット + revive（`exported` とエラーの規則 5 つ）, gocritic, misspell | 典型的なバグと不統一、エラーの書き方（小文字で句点なしの文、`fmt.Errorf`、`err`/`Err` で始まる名前、最後に返す `error`、`return` のあとの `else`） | `make backend-lint` | `make check`, CI `backend` |
| `bodyclose`（テストを除く）、`contextcheck`、`copyloopvar`、`dupl`、`errname`、`exhaustive`（`default` は網羅とみなす。`Pitch`・`PitchClass` は列挙として扱わない）、`forcetypeassert`、`gosec`（除外は行ごとに理由つき）、`intrange`、`nakedret`、`nilerr`、`noctx`、`nolintlint`、`perfsprint`（文字列の連結の提案は止める）、`sloglint`（グローバルの logger と `err` のキーを禁じる）、`unparam`、`usestdlibvars`（Go。ADR-0029。1 つずつ入れた） | HTTP の応答の本文の閉じ忘れ、context の取りこぼし、古いループ変数のコピー、重複コード、エラー型の命名、列挙の switch の取りこぼし、確かめない型アサーション、典型的な安全上の誤り（権限、整数の縮小変換など）、整数の range で書けるループ、裸の return、`nil` を返して握りつぶすエラー、context のない HTTP の要求、注入されない logger、いつも同じ値の引数や戻り値、理由のない `//nolint`、書式の要らない `fmt` の呼び出し、標準ライブラリの定数の手書き | `make backend-lint` | `make check`, CI `backend` |
| shellcheck（`scripts/*.sh`。意図した一重引用符などは、行ごとに理由つきで止める） | 失敗を見ない `cd`、引用符のない展開、使われない変数など、シェルの典型的な誤り | `make shell-lint` | `make check`, CI `shell` |
| actionlint（`.github/workflows/*.yml`。`run:` のスクリプトは ShellCheck に渡す。版は `.tool-versions`） | ワークフローの書き間違い（式、イベント、ジョブの依存、権限）と、`run:` の中のシェルの誤り | `make actions-lint` | `make check`, CI `shell` |
| hadolint（`Dockerfile`。版は `.tool-versions`。手元に hadolint か Docker がないことがあるので `make check` には入れない） | Dockerfile の典型的な誤り（固定しない版、`apt` の後始末、`RUN` の書き方）と、`RUN` の中のシェルの誤り | `make dockerfile-lint`（hadolint か Docker が要る） | CI `shell`（固定した `ghcr.io/hadolint/hadolint` のイメージ） |
| `depguard`（Go） | 境界違反: `music` の I/O、`httpapi` からアダプタ直 import、`net/http` の漏れ | `make backend-lint` | `make check`, CI `backend` |
| `funlen` 80 行 / `gocognit` 20 / `gocyclo` 15 | 肥大した関数（テストは除外） | `make backend-lint` | `make check`, CI `backend` |
| ESLint 型つき strict と stylistic（typescript-eslint `strictTypeChecked`・`stylisticTypeChecked`。矢印の短縮形が返す `void` と、テンプレート文字列の数値は許す。型の形は `type` で書く）+ react-hooks | 典型的なバグ、要らない条件、非推奨の API、`void` の値の取り違え、書き方の不統一 | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint Testing Library（`eslint-plugin-testing-library` の `flat/react`。単体テストに。タイムラインの SVG のテストは、role のない図形をクラスで引くので `no-container`・`no-node-access` を外す） | 実装の詳細に結びついたテスト（`container` や DOM のたどり）、待たない非同期、慣例から外れた名前 | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint Vitest（`@vitest/eslint-plugin` の `recommended`。単体テストに） | 同じ名前のテスト、`expect` のないテスト、`.only` や消し忘れのテスト | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint feature 境界 | feature 間 import、`../` での越境 | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint コンポーネント純粋性 | `*.tsx` での API 呼び出し・`fetch` / `AudioContext` / ストレージ / タイマー | `make frontend-lint` | `make check`, CI `frontend` |
| knip（未使用の export・ファイル・依存、ADR-0007） | 削除や移動のあとに残った死んだコードと依存 | `make frontend-lint`（`npm run knip`）。設定は `frontend/knip.json` | `make check`, CI `frontend` |
| ESLint `max-lines` 400 / `max-lines-per-function` 120 / `complexity` 15 | 肥大したファイル・関数（テストの関数長は除外） | `make frontend-lint` | `make check`, CI `frontend` |
| ESLint スナップショット禁止 | 実装に結合したテスト | `make frontend-lint` | `make check`, CI `frontend` |
| PR サイズ予算（400 警告 / 600 失敗） | レビュー不能な PR | `bash scripts/check-pr-size.sh main` | CI `pr-size`（PR のみ） |

## ドキュメント（ADR-0006）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| リンク・パス・索引・BL-ID・競合マーカー | 存在しないものを指す文書、マージ時に残った `<<<<<<<` | `make docs-check` | `make check`, CI `docs` |
| AGENTS.md / CLAUDE.md の行数予算（150） | 読まれない指示書 | `make docs-check` | `make check`, CI `docs` |
| 鮮度レポート（In progress 放置、Next / Later / Ideas に 90 日以上ある backlog 項目、120 日未更新の文書、腐りやすい言い回し） | 古い記述と、決めないまま溜まる backlog | `bash scripts/docs-staleness.sh` | 月次 `docs gardening` |
| 再実行で通った CI の一覧（直近 30 日） | flaky なテスト・ジョブの放置 | `bash scripts/flaky-report.sh <owner/repo>`（`gh` が必要） | 月次 `docs gardening` |
| 道具と依存の陳腐化の報告（直接の npm・Go の依存のうち、非推奨のもの、365 日以上公開のないもの、大きな版が遅れているもの。golangci-lint の版。見直しの確認項目） | 古びた道具と依存の放置（ADR-0029 の決定 5） | `bash scripts/tool-staleness.sh` | 四半期 `tool review`（`tool-review` の Issue）。Claude の定期のタスクがその Issue を読み、一次情報で調べて企画の Issue にする（実装はしない） |

## ビルドと配備（ADR-0008）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `go build` / `vite build` | ビルド不能 | `make build` | `make check`, CI |
| コンテナイメージのビルドとスモーク（`scripts/smoke.sh`: health の SHA、`/` の HTML とキャッシュ、`assets/` の `immutable`、存在しないファイルの 404、画面の URL、API の 404、本文の上限。加えて `HEALTHCHECK` healthy） | Dockerfile の腐敗、配信の退行 | `make image` と `make smoke`（`make up` に対して） | CI `image` と `e2e`（どちらも compose で 2 つのコンテナ） |
| E2E（Playwright, axe-core による主な画面のアクセシビリティの検査（WCAG 2.2 A・AA の機械で見える部分。明るい・暗いテーマ）, イメージから例を聴く通しのシナリオ・言い直し・画面幅ごとの配置・保存・オフライン・SPA フォールバック・Service Worker） | モックでは見えない通しの退行 | `make e2e`（サーバー起動が前提） | `e2e` ワークフロー（手動・週次） |

## 依存（削減）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `go mod tidy -diff` | `go.mod` / `go.sum` の不整合、不要な依存 | `make deps-check` | CI `backend` |
| `npm audit --omit=dev --audit-level=high` | 本番依存の既知脆弱性 | `make deps-check` | CI `frontend` |
| `govulncheck` + `npm audit`（全依存） | 新たに公開された脆弱性 | （CI のみ） | 週次 `security audit` |
| Dependabot（週次、minor/patch と major を別グループ） | 依存の陳腐化 | — | GitHub |

**Dependabot の方針**: CI が `main` に追従した状態で緑なら、major を含めてマージしてよい（人間でもエージェントでも）。古いブランチは先に `main` へ追従させて CI を回し直す。major で規約や書き方が変わる（設定ファイルの形式変更など）なら、マージと同じ PR かその直後に ADR か文書を更新する。`typescript` と `@types/node` の major は除外（理由は `.github/dependabot.yml`）。CI が落ちた更新は閉じて原因を backlog に書く。

## ツールチェーン（ADR-0008）

| 検査 | 防ぐもの | 個別実行 | 場所 |
| --- | --- | --- | --- |
| `.tool-versions` を CI の setup-go / setup-node / golangci-lint-action と Docker のビルド引数が読む | CI と手元・コンテナの版ズレ | — | CI、`make image` / `make up` |
| `make doctor` | 手元の Go / Node / golangci-lint / actionlint の major.minor ズレ | `make doctor` | セッション開始時（`docs/dev/workflow.md`） |

## CI が外部から取得するもの（ADR-0016）

ジョブはこれらをキャッシュから取り、ないときだけ取得元へ行く。すべてのジョブに `timeout-minutes` がある。

PR と `main` への push では、変更に関係するジョブだけが走る（`backend`、`frontend`、`image`。`shell`・`docs` は毎回、`pr size` は PR だけ）。PR は base ブランチと、push は push の前の `main` の先頭と比べる（ADR-0032）。どのパスがどのジョブを走らせるかは `scripts/ci-changes.sh` だけが持つ。ジョブが新しいパスを読むようになったら、そこに足す。PR では新しい push が古い実行を取り消し、`main` の実行は取り消さない。必須チェックは `required (make check)` の 1 つだけ（ADR-0028）。`changes`・`backend`・`frontend`・`shell`・`docs` の結果をまとめ、スキップは通し、失敗と取り消しは落とす。`image`・`pr size`・`eval`・CodeQL は必須ではない。

CodeQL は `.github/workflows/codeql.yml` で動く（ADR-0031）。`docs/**` と `**/*.md` だけの変更では走らず、週に 1 回は全体を解析する。

| 取得するもの | 取得元 | キャッシュ | 取得元が不調のとき |
| --- | --- | --- | --- |
| Go 本体（`.tool-versions` の版） | GitHub の `actions/go-versions` のリリース | ランナーのツールキャッシュを `actions/cache` で版ごとに | `setup-go` が go.dev から取る |
| Go のモジュールとビルド | proxy.golang.org | `setup-go`（キーは `backend/go.sum`） | — |
| golangci-lint | GitHub のリリース（チェックサムを検証、再試行あり） | `actions/cache` で版ごとに | `scripts/install-golangci-lint.sh` がモジュールプロキシからビルドする |
| npm パッケージ | npm レジストリ | `setup-node`（キーは `frontend/package-lock.json`） | — |
| Docker のベースイメージと層 | Docker Hub、gcr.io | buildx の `type=gha` | — |

## 手順（機械化されていない）

| 手順 | 目的 | いつ |
| --- | --- | --- |
| `docs-sync` 自己点検 | 差分で言及が変わる文書の修正 | PR 前（`docs/dev/docs-hygiene.md`） |
| 自己レビュー（`/code-review` 等） | 人間のレビュー負荷の低減 | PR 前 |
| `ui-review`（画面を撮り、文脈を共有しないサブエージェントがレビューする。`.claude/skills/ui-review`） | 作った文脈を知らない目での画面の見直し。指摘は仮説で、CI には組み込まない（#286） | 画面の構成・配置・状態の見せ方を変える PR の前、配置の案を見比べるとき |
| `make e2e` を 1 回通す | モックでは見えない失敗 | UI スライスの PR 前（`docs/dev/testing.md`） |
| PR に `needs-eval` ラベルを付けて実 API の評価を走らせる（結果は PR のコメント。ADR-0018）。`needs-eval:<tag>`（例: `needs-eval:meter`）は、そのタグと smoke のケースだけ（ADR-0038）。全件は Message Batches で送るので、結果のコメントまで数分から 1 時間ほどかかる（ADR-0039） | プロンプトの退行、語彙を足したのに使われないこと | プロンプト・スキーマ・インテント・語彙・評価のケースを変えた PR で、マージ前（`docs/dev/workflow.md`） |
| WIP 上限 3 | レビュー処理能力の保護 | 常時（`AGENTS.md`） |
| ハーネスの見直し | 効かない検査の削除、CI 時間（予算 10 分）の確認 | 四半期 |
