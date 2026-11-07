## 概要

<!-- 何を・なぜ。関連する backlog ID / Issue / ADR を書く（例: closes #12, refs BL-NNN） -->

## リスク区分

<!-- 1 つ選ぶ: docs のみ / テストのみ / 振る舞いの変更 / 契約（API・データ形式）の変更 / ハーネス（CI・lint・ビルド）の変更 -->

## レビューの見どころ

<!-- レビュアーが最初に読むべきファイルと、判断してほしい点を 1〜3 行で。自己レビュー（/code-review 等）の要約があればここに -->

## 変更内容

-

## 確認したこと

- [ ] `make check` がローカルで通る
- [ ] 振る舞いの変更に対応するテストを追加・更新した
- [ ] API の形を変えた場合 `docs/api/openapi.yaml` を更新した
- [ ] 設計判断をした場合 ADR を追加した（`docs/architecture/adr/`）
- [ ] `docs/product/backlog.md` のステータスを更新した
- [ ] UI を変えた場合、`docs/design/ux-principles.md` の手順（目的〜確かめ方）を数行で書いた
- [ ] `docs-sync` の自己点検（`docs/dev/docs-hygiene.md`）を行い、言及が変わる文書を直した

## スコープ外（やらなかったこと・次に回したこと）

<!-- 「あったほうがいい」と思ったが今回は実装しなかったものを backlog に追記した上でここに列挙する -->
