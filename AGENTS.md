# AGENTS.md — AI エージェント向け開発ガイド

このファイルは middy で作業するすべての AI コーディングエージェント（Claude Code, OpenAI Codex, GitHub Copilot など）と人間の共通の正典です。ツール固有の補足は `CLAUDE.md` と `.github/copilot-instructions.md` にありますが、ルールの本体はここだけに書きます。

## 1. このプロダクトは何か

作曲スキルと音楽理論の学習を助ける PWA。ユーザーのイメージ（言葉）を AI が音楽パラメータに解釈し、決定論的な音楽理論エンジンがコード進行・リズムのタイムラインを生成する。ユーザーはそれを再生し、成り立ちを視覚的な解説で理解する。詳細は `docs/product/vision.md`。

**最重要の設計原則（ADR-0003）**: LLM は「解釈と対話」を担当し、音符は書かない。音符・コード・分析はすべて `backend/internal/music` の純粋関数が作る。この境界を崩す変更は ADR なしに行わない。

## 2. 作業の進め方（開発ループ）

1. **`docs/product/backlog.md` を開き、「Now」の中で最も上にある `Ready` な項目を 1 つ選ぶ。** 自分で新しい機能を思いついても、backlog にない限り着手しない。思いついたものは backlog の「Later」か「Ideas」に 1 行で追記して終わり。使っていて見つけた問題は `observation` の Issue に、検討が要るものは `/propose` で `proposal` の Issue にする（`docs/dev/workflow.md` の「Issue の使い方」）。
2. 着手前に受け入れ条件を読み、曖昧なら **実装せずに質問する**（または最も保守的な解釈を選び、PR に明記する）。
3. **テストを先に書く**（または同じコミットで書く）。ドメインロジックはテーブル駆動テスト、HTTP は `httptest`、UI は Testing Library。
4. 実装する。ディレクトリ配置と規約は `docs/architecture/overview.md` と `docs/dev/conventions.md` に、画面や操作の判断は `docs/design/ux-principles.md` の手順に従う。
5. **`make check` を通す。** 通らないものは完成ではない。CI では `required (make check)` がマージを止める（ADR-0028）。
6. 影響するドキュメントを更新する: API 変更 → `docs/api/openapi.yaml`、設計判断 → `docs/architecture/adr/`、backlog のステータス。PR を出す前に `docs-sync` の自己点検（`docs/dev/docs-hygiene.md`）を行う。
7. PR を開く前に自己レビューを行い（Claude Code では `/code-review`）、要約を PR の「レビューの見どころ」に書く。Conventional Commits でコミットし、PR テンプレートを埋める。1 PR = 1 backlog 項目 = 1 縦スライス。
8. **未マージの自分の PR が 3 本になったら止まり、人間にマージか方針の判断を仰ぐ**（レビュー処理能力を超えない）。
9. 完了したら backlog の項目を `Done` にし、次の項目へ。

ワークフローの詳細と「完了の定義」は `docs/dev/workflow.md`。

## 3. ガードレール（やってはいけないこと）

- **スコープを広げない。** 「あったほうがいい」機能・リファクタ・依存追加は、今の backlog 項目に必要でなければやらない。backlog に書いて次に回す。
- **テストを消す・スキップする・lint を無効化して通す** ことで緑にしない。根本原因を直す。
- **依存を追加しない**（Go module / npm package）。本当に必要なら ADR を 1 枚書き、PR で理由を述べる。候補は `docs/architecture/adr/0002-tech-stack.md` の「将来の候補」を優先。
- **API の形を黙って変えない。** `docs/api/openapi.yaml` が正典。Go ハンドラは整合性テストが、TS の型は生成物（`npm run gen:api` → `schema.gen.ts`）の鮮度検査が一致を強制する。型を手で書かない。
- **秘密情報を扱わない。** `.env` を読まない・コミットしない。API キーはコードに書かない。
- **`main` に直接 push しない。** `git push --force` をしない。
- **巨大な PR を作らない。** 生成物を除き 400 行差分で警告、600 行で CI が落ちる（`scripts/check-pr-size.sh`）。超えそうなら項目を分割して backlog を更新する。
- **ファイル全体の書き換えより、差分編集を優先する。** 既存のスタイルに合わせる。
- ファイル形式・記法・プロトコル・番号の割り当てを決めるときは、業界標準の索引 `.claude/skills/standards/index.md` を引き、該当する標準があれば一次情報に当たる（#213）。
- 画面・操作・フィードバック・配置・色を決める・変えるときは、HCI の知見の索引 `.claude/skills/hci/index.md` を引き、判断の問いへの答えを PR の「原則の点検」に書く（#274）。
- 振る舞いが音楽理論的に正しいか自信がないときは、`docs/architecture/domain-model.md` の定義と既存テストを根拠にし、推測で「それらしい」出力を作らない。

## 4. 技術スタックと配置

| 場所 | 役割 | 守ること |
| --- | --- | --- |
| `backend/cmd/server` | エントリポイント。配線だけ | ロジックを書かない |
| `backend/internal/config` | 環境変数 | 新しい設定は `Load` とテストに追加 |
| `backend/internal/httpapi` | HTTP ハンドラ・ミドルウェア | 薄く保つ。decode → ユースケース → encode。LLM アダプタを直接 import しない（`depguard` で強制） |
| `backend/internal/music` | 音楽理論ドメイン | 標準ライブラリのみ、I/O なし、純粋関数（`depguard` で強制） |
| `backend/internal/music/smf` | Timeline → SMF エンコーダ | `music` にのみ依存。バイト列の期待値テストを壊す変更は意図を PR に書く |
| `backend/internal/intent` | LLM が返す MusicalIntent の型・検証・Schema | `music` にのみ依存。制約を変えたら `Validate` と `Schema` を同時に |
| `backend/internal/ideas` | S2 のユースケース（会話 → LLM → intent → music） | システムプロンプトは `prompt.go` の `SystemPrompt()`（固定文 + パターンカタログ）。変更したら PR に `needs-eval` ラベルを付けて実 API の評価を走らせる（ADR-0018） |
| `backend/internal/llm` | LLM 境界。`Client` interface、`fake`、`anthropic` アダプタ | プロバイダ依存はサブパッケージに閉じる |
| `backend/internal/fixtures` | フロント照合用の正準 JSON の生成 | 表記を変えたら `make fixtures` を実行してコミット |
| `frontend/src/app` | アプリシェル・ルーティング | |
| `frontend/src/features/<name>` | 機能ごとの縦スライス（UI + hooks + 状態） | 機能間で直接 import しない。共有は `shared` へ（ESLint で強制）。`*.tsx` は API 呼び出しや `fetch` / `AudioContext` 等の副作用を持たず、hook に置く（ESLint で強制） |
| `frontend/src/shared` | API クライアント、共通 UI、ユーティリティ | |
| `deploy/nginx` | `web` コンテナの nginx の設定（ADR-0027） | 配信の規則を変えたら `scripts/smoke.sh` の期待値も同じ PR で直す |
| `docs/api/openapi.yaml` | API 契約 | ハンドラ変更と同じ PR で更新 |

- Go: 標準 `net/http`（Go 1.22+ のメソッド付きパターン）、`log/slog`、`errors.Join`。フレームワークは使わない。
- TypeScript: `strict` + `noUncheckedIndexedAccess`。`any` 禁止。型 import は `import type`。
- スタイル: gofmt / goimports、Prettier（singleQuote, printWidth 100）。

## 5. コマンド

```bash
make check            # すべてのゲート（これが通れば CI も通る）
make test             # テストのみ
make lint             # lint のみ
make fmt              # 自動整形
make dev-backend      # API サーバー（:8080, LLM は fake）
make dev-frontend     # Vite dev server（:5173）
cd backend && go test ./internal/music/ -run TestDiatonic -v   # 単体で回す例
cd frontend && npx vitest run src/app                             # 単体で回す例
```

## 6. コミットと PR

- Conventional Commits: `feat(music): add harmonic function analysis`, `fix(httpapi): ...`, `docs:`, `test:`, `chore:`, `refactor:`。スコープは `music` / `httpapi` / `llm` / `config` / `frontend` / `docs` / `ci`。
- PR 本文は `.github/pull_request_template.md` に従い、**スコープ外にしたこと** を必ず書く。
- コミットメッセージ・コード・コメントにモデル名や生成ツールの宣伝を書かない。

## 7. ドキュメントの鮮度（ADR-0006）

- 文書は **現在の状態だけ** を書く。経緯は git log と ADR にだけ残す。Issue と PR の本文も同じで、見解の変化や訂正はコメントで足す（`docs/dev/workflow.md` の「Issue の使い方」）。
- 事実は 1 箇所にだけ書き、他は参照する。数値・日付・「予定」リストをコピーしない。将来のことは backlog の BL-ID で指す。
- `make docs-check` が、リンク切れ・存在しないパス・索引漏れ・存在しない BL-ID を落とす。Go のテストがルートと `docs/api/openapi.yaml` の一致を落とす。
- backlog の Done 表は現在と直前のマイルストーン分だけ残す。古い ID への参照が落ちたら、その文書を書き直す合図。
- 手順の詳細は `docs/dev/docs-hygiene.md`。ハーネスの全検査は `docs/dev/harness.md`。

## 8. 迷ったら

1. `docs/product/vision.md` の「コアシナリオ」に近い方を選ぶ。
2. 決定論的に解ける方を選ぶ（LLM に頼らない）。
3. 小さい方を選ぶ。
4. それでも決まらなければ、実装せずに質問する。
