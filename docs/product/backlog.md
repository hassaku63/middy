# バックログ

**AI エージェントはここから着手項目を選ぶ。** ルール:

- **Now** の中で上から順に、`Ready` の項目を 1 つ選ぶ。`Blocked` は依存が Done になるまで触らない。
- 1 項目 = 1 PR = 1 縦スライス。受け入れ条件をすべて満たして `Done` にする。
- Now にない機能は実装しない。思いついたら **Ideas に 1 行追記** して終わり。
- 項目を分割・追加・並べ替えるときは、この表を編集して PR に含める。並べ替えの理由も一言書く。
- サイズ: S = 半日以内 / M = 1 日 / L = 2〜3 日。L は着手前に分割を検討する。
- **Done 表は現在のマイルストーンと直前のマイルストーンの分だけ残す。** それより古い行はマイルストーン完了時に削除する（履歴は git にある）。削除した ID を参照している文書は `make docs-check` が検出するので、その文書を現状に合わせて書き直す。

ステータス: `Ready` / `In progress` / `Blocked(依存)` / `Done`

## Now（M1 — Hear & See）

| ID | スライス | 受け入れ条件 | サイズ | 依存 | 状態 |
| --- | --- | --- | --- | --- | --- |
| BL-009 | **ダウンロード UI** 再生中のタイムラインを .mid で保存するボタン | ファイル名に調と進行が入る | S | BL-007, BL-008 | Ready |

## Next（M2 — Imagine）

| ID | スライス | 受け入れ条件（要約） | サイズ | 依存 |
| --- | --- | --- | --- | --- |
| BL-010 | **MusicalIntent スキーマ** `internal/intent`: 調候補、モード、テンポ範囲、雰囲気タグ、進行テンプレート候補（ローマ数字）、理由。JSON Schema を Go で定義し、検証関数を持つ | 不正な値（存在しないローマ数字など）を弾く | S | — |
| BL-011 | **Anthropic アダプタ** `internal/llm/anthropic`: 公式 Go SDK、構造化出力、`fallbacks`、タイムアウト、エラー分類。`docs/architecture/llm-integration.md` 準拠 | 録画レスポンスでテスト。ネットワークなしで CI が通る | M | — |
| BL-012 | **意図解釈ユースケース** イメージ文 → `MusicalIntent`（LLM）→ 検証 → 候補 Timeline 2〜3 個（BL-001） | fake LLM の台本で end-to-end テスト | M | BL-010, BL-011 |
| BL-013 | **API: `POST /api/v1/ideas`** とチャット継続 `POST /api/v1/ideas/{id}/messages`（まずはステートレスに会話履歴を往復） | OpenAPI 更新 | S | BL-012 |
| BL-014 | **チャット UI** `features/chat`: 入力、候補カード（再生ボタン付き）、相談の継続 | 候補カードから M1 の再生・解説が開く | M | BL-013 |
| BL-015 | **プロンプト評価セット** 代表的なイメージ 20 件と期待される特徴（モード・テンポ帯）。手動実行スクリプトと結果の記録 | 回帰を検出できる | S | BL-012 |

## Later（M3 — Groove & Ship）

| ID | スライス | 要点 |
| --- | --- | --- |
| BL-020 | リズムパターン辞書 `music/pattern`: 伴奏（ブロック、アルペジオ、ストローク風）とドラム（GM ch.10） |
| BL-021 | パターン適用: 進行 + パターン → 複数トラックの Timeline |
| BL-022 | リズム解説: 拍の重心・シンコペーションの検出と可視化データ |
| BL-023 | 候補にリズムを含める（LLM の意図にパターン候補を追加） |
| BL-024 | ローカル永続化（IndexedDB）とプロジェクト一覧 |
| BL-025 | PWA 仕上げ: 更新通知、オフライン時の挙動、インストール導線 |
| BL-026 | 型生成: OpenAPI → TypeScript 型（手書き型の置き換え） |

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
| BL-030 | ドキュメント鮮度ハーネス（`make docs-check`、ルート↔OpenAPI 整合性テスト、`docs-sync` 自己点検、月次 gardening） | 2026-11-07 |
