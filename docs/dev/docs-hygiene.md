# ドキュメント鮮度の維持

目的: 古くなった文書や散在する過去の経緯が、AI と人間の両方にとって **信頼できないノイズ** になるのを防ぐ。方針の根拠は ADR-0006。

## 原則

1. **現在の状態だけを書く。** 「以前は〜だった」「〜に変更した」は書かない。経緯は git log と ADR にだけ残す。
2. **事実は 1 箇所。** 同じ事実を複数の文書にコピーしない。他の文書はリンクか BL-ID で参照する。
   - 進捗・予定 → `docs/product/backlog.md` だけ
   - API の形 → `docs/api/openapi.yaml` だけ
   - 設定項目 → `backend/.env.example` と `internal/config` だけ
3. **腐る値を書かない。** カバレッジの数値、依存のバージョン、日付、「まだ実装されていない」の列挙。必要なら「`make backend-test` の出力を見る」のように取得方法を書く。
4. **将来の話は BL-ID で指す。** 「予定」「後で」ではなく「BL-010」。ID が backlog から消えれば検査が落ち、文書を直す合図になる。
5. **文書の種類を分ける。** `docs/product`, `docs/architecture/*.md`, `docs/dev` は「現在」。`docs/architecture/adr` だけが「経緯」で、追記のみ。ADR を Superseded にするときは、影響する「現在」の文書を同じ PR で直す。
6. **消す勇気。** 役目を終えた文書・段落・backlog 行は削除する。git に残る。

## 機械的な検査（`make check` に含まれる）

| 検査 | 場所 | 落ちる条件 |
| --- | --- | --- |
| Markdown リンク | `scripts/check-docs.sh` | 相対リンク先のファイルが存在しない |
| パス参照 | 同上 | バッククォート内の `backend/…` `frontend/…` `docs/…` `scripts/…` `.github/…` `.claude/…` が存在しない（gitignore 済みのパスは除く） |
| 索引 | 同上 | `docs/` 配下の文書が `docs/README.md` に、ADR が `docs/architecture/adr/README.md` に載っていない |
| backlog ID | 同上 | 文書・ワークフロー内の `BL-NNN` が backlog に存在しない |
| API 契約 | `backend/internal/httpapi/openapi_test.go` | 登録ルートと `docs/api/openapi.yaml` の操作が一致しない |

単体で走らせる: `make docs-check`、`cd backend && go test ./internal/httpapi/ -run TestRoutesMatchOpenAPI`。

検査を黙らせるためにパスをバッククォートから外したり、ID を伏せたりしない。落ちたら文書を現状に合わせる。

## 自己点検 `docs-sync`（PR を出す前に毎回）

ツールを問わず同じ手順。Claude Code では `/docs-sync` スキルがこの手順を読み込む。

1. 差分を取る: `git diff main...HEAD --stat` と `git diff main...HEAD`。
2. 差分から **言及されうる名前** を拾う: 追加・削除・改名した型/関数/ファイル/ディレクトリ/環境変数/エンドポイント/BL-ID。
3. それぞれを docs から逆引きする:
   ```bash
   grep -rn --include='*.md' -e '<name>' docs AGENTS.md README.md CLAUDE.md .github
   ```
4. ヒットした箇所ごとに判定する: (a) 現状と一致 → 何もしない、(b) 古い → 直す、(c) 経緯の説明になっている → 削除して現在形に。
5. 変更が次に該当するなら、対応する文書を直したか確認する:
   - ハンドラ → `docs/api/openapi.yaml`（テストが落ちるので忘れにくい）
   - ドメインの型・規約 → `docs/architecture/domain-model.md`, `docs/glossary.md`
   - ディレクトリ構成 → `docs/architecture/overview.md`, `README.md` のツリー, `AGENTS.md` の表
   - 環境変数 → `backend/.env.example`
   - 依存追加・設計判断 → ADR
   - スライスの完了・分割 → `docs/product/backlog.md`
6. `make docs-check` と `bash scripts/docs-staleness.sh` を走らせ、自分の変更に関係する指摘を潰す。
7. PR 本文の「確認したこと」にチェックを入れる。

サブエージェントに任せる場合は、手順 1〜6 をそのまま渡し、「直すべき箇所の一覧（ファイル:行、現状、修正案）」だけを返させる。本体の編集は PR の作者が行う。

## 定期の庭仕事（月次）

`.github/workflows/docs-gardening.yml` が毎月 1 日に既定ブランチで走る（手動実行も可）。

1. `scripts/check-docs.sh` を実行。
2. `scripts/docs-staleness.sh` で見直し候補を出す: `In progress` のまま残った backlog 項目、120 日以上更新のない文書、「予定」「現在 NN」「TODO」などの腐りやすい言い回し。
3. 指摘があれば `docs-gardening` ラベルの Issue を 1 つ作る（既にあれば更新する）。

Issue を受けた人（または AI エージェント）は、1 スライスとして文書を現状に合わせる PR を出す。指摘が誤検知なら、検査スクリプトの方を直す。

LLM による深い監査（「この説明は実装と意味的に一致しているか」）は、必要なときに `docs-sync` の手順をそのまま新しいセッションに渡して行う。定期実行に組み込むかは、月次の指摘量を見て判断する（検討は backlog で扱う）。

## よくある腐り方と対処

| 症状 | 対処 |
| --- | --- |
| 「〜予定」「後で〜する」 | BL-ID に置き換える。backlog になければ Ideas に 1 行追加する |
| 数値（カバレッジ、行数、件数） | 削除し、取得コマンドを書く |
| 「この PR で」「今回」 | PR 本文に移し、文書は現在形にする |
| 同じ表が 2 箇所にある | 片方を削除し、リンクにする |
| ADR に「現在の実装は〜」 | ADR は決定時点の記録。現在の実装は overview / domain-model に書く |
| 完了した backlog 行が増え続ける | 直前のマイルストーンより古い Done 行を削除する |
