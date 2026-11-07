# バックログ

**AI エージェントはここから着手項目を選ぶ。** ルール:

- **Now** の中で上から順に、`Ready` の項目を 1 つ選ぶ。`Blocked` は依存が Done になるまで触らない。
- 1 項目 = 1 PR = 1 縦スライス。受け入れ条件をすべて満たして `Done` にする。
- Now にない機能は実装しない。思いついたら **Ideas に 1 行追記** して終わり。
- 項目を分割・追加・並べ替えるときは、この表を編集して PR に含める。並べ替えの理由も一言書く。
- サイズ: S = 半日以内 / M = 1 日 / L = 2〜3 日。L は着手前に分割を検討する。
- **Done 表は現在のマイルストーンと直前のマイルストーンの分だけ残す。** それより古い行はマイルストーン完了時に削除する（履歴は git にある）。削除した ID を参照している文書は `make docs-check` が検出するので、その文書を現状に合わせて書き直す。

ステータス: `Ready` / `In progress` / `Blocked(依存)` / `Done`

## Now（M2 — Imagine）

M3（Groove & Ship）を Later から順に進める。実 API での疎通確認（`make eval`）は鍵の登録後に行う。

| ID | スライス | 受け入れ条件 | サイズ | 依存 | 状態 |
| --- | --- | --- | --- | --- | --- |

## Next

M2 が終わったら M3（Later）から Now へ。


## Later（M3 — Groove & Ship）

| ID | スライス | 要点 |
| --- | --- | --- |
| BL-050 | `knip` による未使用 export / ファイル / 依存の検出（ADR-0007）。発動条件: 最初の feature 削除 |
| BL-051 | `music` のミューテーションテスト（`gremlins`）、四半期実行（ADR-0007）。発動条件: `music` が 3k 行 |
| BL-053 | プロンプト監査の四半期実行（`claude-api` の `prompt-audit`）。発動条件: BL-011 以降 |

## Ideas（未評価。実装しない）

- モード拡張（ドリアン、ミクソリディアン等）、借用和音、セカンダリドミナント
- メロディ生成（LLM ではなくルールベースから）
- Web MIDI 出力、外部音源
- 共有リンク、サーバー側保存とアカウント
- 耳トレーニングモード（進行を聴いてローマ数字を当てる）
- 楽譜表示（VexFlow 等）
- 多言語 UI

## Done

| ID | スライス | 完了日 |
| --- | --- | --- |
| BL-000 | M0 足場（モノレポ、ドメイン基礎型、LLM 境界、ハーネス、ドキュメント） | 2026-11-07 |
| BL-001 | 進行テンプレート → Timeline 生成（`music.Progression`, `music.FromProgression`） | 2026-11-07 |
| BL-002 | 和声機能とケーデンス分析（`music.Analyze`, `music.Function`, `music.Cadence`） | 2026-11-07 |
| BL-003 | API `POST /api/v1/timelines/from-progression`（`music.ParseRomanNumeral` 含む） | 2026-11-07 |
| BL-004 | ブラウザ再生（`features/player`: TimelinePlayer, WebAudioSynth, usePlayer, PlayerControls） | 2026-11-07 |
| BL-005 | タイムライン表示（`features/timeline`: コードレーン + ピアノロール + 再生位置） | 2026-11-07 |
| BL-006 | 解説パネル（`features/explanation`: 再生追従とクリック固定） | 2026-11-07 |
| BL-007 | 入力 UI と結線（`features/progression-input`、App で S1 シナリオが通しで動く） | 2026-11-07 |
| BL-008 | SMF エクスポート（`music/smf`、`POST /api/v1/export/midi`） | 2026-11-07 |
| BL-009 | ダウンロード UI（`features/export`、ファイル名に調と進行） | 2026-11-07 |
| BL-054 | 月次 gardening の拡張（backlog の老化、再実行で通った CI の集計） | 2026-11-07 |
| BL-026 | OpenAPI → TypeScript 型生成（`openapi-typescript`、`npm run gen:api`、鮮度検査、ADR-0009） | 2026-11-07 |
| BL-052 | バンドルサイズ予算（`scripts/check-bundle-size.sh`）と 1,000 ノート描画の予算テスト | 2026-11-07 |
| BL-025 | PWA 仕上げ（更新は prompt モードでバナーから、オフライン表示、インストール導線、オフライン E2E） | 2026-11-07 |
| BL-024 | ローカル保存（IndexedDB の `ProjectStore`、保存・一覧・開く・削除、リロード後も残る） | 2026-11-07 |
| BL-023 | 候補にパターンを含める（intent の `accompaniment` / `drums`、Schema の enum とプロンプトをカタログから生成、候補カードに表示） | 2026-11-07 |
| BL-022 | リズム解説（`Timeline.rhythm`: 打点の分類とシンコペーション判定、解説パネルの「リズム」） | 2026-11-07 |
| BL-021 | パターン適用（API の `accompaniment` / `drums`、`GET /patterns`、フォームの選択、ドラムレーン、打楽器の音） | 2026-11-07 |
| BL-020 | リズムパターン辞書（`music/pattern`: 伴奏 5 種・ドラム 3 種、`Apply`） | 2026-11-07 |
| BL-016 | 短調の大文字 V / V7 を和声的短音階のドミナントとして扱う（`music.StepFromLabel`、解説に一文、プロンプトと Schema の記述） | 2026-11-07 |
| BL-015 | プロンプト評価セット（`internal/ideas/evals`、`make eval`、判定器は fake でテスト） | 2026-11-07 |
| BL-011 | Anthropic アダプタ（`internal/llm/anthropic`、公式 SDK、構造化出力、録画応答テスト） | 2026-11-07 |
| BL-014 | チャット UI（`features/chat`、入力方法の切り替え、E2E に S2 を追加） | 2026-11-07 |
| BL-013 | API `POST /api/v1/ideas`（502 `llm_unavailable` / `llm_invalid_output`、TS クライアント `imagine`） | 2026-11-07 |
| BL-012 | 意図解釈ユースケース（`internal/ideas`、1 回再試行、fake の見本 Intent） | 2026-11-07 |
| BL-010 | MusicalIntent スキーマ（`internal/intent`: 型・検証・JSON Schema・Parse） | 2026-11-07 |
| BL-043 | 通しの E2E（Playwright、`make e2e`、`e2e` ワークフロー） | 2026-11-07 |
| BL-042 | ツールチェーンの固定（`.tool-versions` を CI・Docker・`make doctor` が読む） | 2026-11-07 |
| BL-035 | テスト所要時間の可視化と予算（`make test-report`、CI の各ジョブで出力） | 2026-11-07 |
| BL-034 | 言語またぎの正準フィクスチャ（`internal/fixtures` → `frontend/src/shared/api/fixtures`、`make fixtures`） | 2026-11-07 |
| BL-033 | 依存の衛生（`go mod tidy -diff`、本番 npm 依存の audit、週次 `govulncheck`、Dependabot の方針） | 2026-11-07 |
| BL-032 | サイズ・複雑度・PR の予算、ハーネス一覧、レビュー導線（自己レビュー、WIP 上限、リスク区分） | 2026-11-07 |
| BL-031 | アーキテクチャ境界の lint（Go `depguard`、TS の feature 間 import / コンポーネント副作用 / スナップショット禁止） | 2026-11-07 |
| BL-041 | Dockerfile と compose（マルチステージ、distroless 非 root、`-healthcheck`、CI の image ジョブ） | 2026-11-07 |
| BL-040 | 単一バイナリ配信（`internal/web` の embed、SPA フォールバック、キャッシュ方針、`make build` / `make run`） | 2026-11-07 |
| BL-030 | ドキュメント鮮度ハーネス（`make docs-check`、ルート↔OpenAPI 整合性テスト、`docs-sync` 自己点検、月次 gardening） | 2026-11-07 |
