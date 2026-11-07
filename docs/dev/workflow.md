# 開発ワークフロー — AI が自律的に回すためのハーネス

このドキュメントは、人間でも AI エージェントでも同じ手順で開発を進められるようにする「ハーネス」の説明。ルールの要約は `AGENTS.md`、ここは詳細と理由。

## 開発ループ

```
backlog から 1 項目 → 受け入れ条件を確認 → テストを書く → 実装 → make check
     ↑                                                              │
     └── backlog を更新 ← PR（テンプレート） ← ドキュメント更新 ←──────┘
```

### 1. 項目を選ぶ

- `docs/product/backlog.md` の **Now** を上から見て、最初の `Ready` を選ぶ。
- 選んだら状態を `In progress` にする（同じ PR に含める）。
- 受け入れ条件が曖昧なら、**最も保守的な解釈** を選んで PR 本文に書く。解釈次第で成果物が大きく変わるなら、実装せずに質問する。

### 2. テストを先に書く

| レイヤー | テスト手段 | 例 |
| --- | --- | --- |
| `internal/music` | テーブル駆動の単体テスト。既知の理論的事実をデータとして列挙 | C major の I〜vii° が `C, Dm, Em, F, G, Am, Bdim` |
| ユースケース | fake LLM の台本 + 入出力の検証 | 「雨の日曜」→ minor の候補が含まれる |
| `httpapi` | `httptest` で JSON の形とステータスを検証 | 400 時のエラーエンベロープ |
| フロント（ロジック） | Vitest。hook/純粋関数は DOM なしで | tick → 小節・拍の変換 |
| フロント（UI） | Testing Library。role/text で検証、スナップショット禁止 | 再生ボタンを押すと status が更新される |

詳細は `testing.md`。

### 3. 実装する

- 配置は `docs/architecture/overview.md`。迷ったら既存コードの隣に置く。
- 既存ファイルは全体書き換えではなく差分編集。
- 依存は追加しない（必要なら ADR）。

### 4. `make check`

```bash
make check
```

通らなければ完了ではない。lint を無効化・テストをスキップして通すことは禁止。

### 5. ドキュメントを更新する

| 変更したもの | 更新するもの |
| --- | --- |
| HTTP ハンドラ | `docs/api/openapi.yaml`、`frontend/src/shared/api` |
| ドメインの語彙・規約 | `docs/architecture/domain-model.md`、`docs/glossary.md` |
| 設計判断・依存 | `docs/architecture/adr/NNNN-*.md` と索引 |
| 環境変数 | `backend/.env.example`、`internal/config` のテスト |
| 完了・分割・追加 | `docs/product/backlog.md` |

更新し終えたら `docs-sync` の自己点検（`docs-hygiene.md`）を行う。差分に出てくる識別子・パス・BL-ID を docs から逆引きし、言及が古くなっていないかを確かめる手順。

### 6. コミットと PR

- Conventional Commits。1 コミットは 1 つの意図。
- PR は `.github/pull_request_template.md` を埋める。**「スコープ外」欄が空の PR は疑わしい**（本当に何も思いつかなかったか、黙って実装したか）。
- PR は小さく。目安は 400 行差分以内（生成物・lock ファイル除く）。超えるなら分割し、backlog に分割後の項目を書く。
- 1 ブランチ = 1 項目。ブランチ名は `feat/BL-NNN-short-name` のように backlog ID を含める。

## 完了の定義（Definition of Done）

- [ ] 受け入れ条件をすべて満たす
- [ ] `make check` がローカルで緑（CI も同じコマンドなので緑になる）
- [ ] 振る舞いにテストがある（バグ修正は再現テストから）
- [ ] 影響するドキュメントを更新した
- [ ] backlog の状態を更新した（`Done` に移動、または分割）
- [ ] PR に「スコープ外」を書いた
- [ ] 新しい秘密情報・依存・API 破壊的変更を含まない（含むなら ADR と明示）

## ガードレールの一覧と理由

| ガードレール | 理由 |
| --- | --- |
| backlog にないものは作らない | 価値の低い「あったほうがいい」機能で複雑さが増えるのを防ぐ |
| LLM に音符を書かせない（ADR-0003） | 再現性・検証可能性・解説との一致 |
| 依存追加には ADR | 依存は将来のコスト。判断を可視化する |
| `make check` 必須 | 局所最適な「動いた」を防ぐ。CI と同一 |
| テストの削除・スキップ禁止 | 緑にする最短経路がテスト破壊になるのを防ぐ |
| PR サイズ上限 | 人間がレビューできる範囲に保つ |
| `.env` を読まない・書かない | 秘密情報の漏洩防止。`.claude/settings.json` でも deny |
| 鍵のあるシェルで、値を表示するコマンド（`env`、`printenv`、`docker compose config`、`docker inspect`）を実行しない | 事故による鍵の読み取りを防ぐ。実 LLM での起動は人が行い、エージェントは起動済みのサーバーに触る（ADR-0011） |
| `main` 直 push・force push 禁止 | 履歴の保護 |

## エージェントのセッション開始チェックリスト

1. `AGENTS.md` を読む（Claude Code は `CLAUDE.md` 経由で自動）。
2. `git status` / `git log --oneline -10` で現在地を確認。
3. `docs/product/backlog.md` の Now を確認し、`In progress` が残っていればそれを引き継ぐ。
4. `make doctor` で版を確認し、`make check` を一度回して出発点が緑であることを確認する。
5. 着手。

## 人間の役割

- backlog の優先順位の最終決定（Now に入れる・外す）。
- 受け入れ条件の曖昧さへの回答。
- PR レビューとマージ。
- ADR の承認。

エージェントはこれらを **代行しない**。提案はできる（backlog の Ideas、ADR の Proposed）。
