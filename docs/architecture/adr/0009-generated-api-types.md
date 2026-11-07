# 0009. TypeScript の API 型を OpenAPI から生成する

- 日付: 2026-11-07
- 状態: Accepted

## 文脈

ADR-0004 は「エンドポイントが 5 を超えたら（ここでは 5 本に達したら）`openapi-typescript` で型生成に切り替える」と決めていた。`health`、`timelines/from-progression`、`patterns`、`export/midi`、`ideas` の 5 本になり、手書きの型はすでに一部が契約と食い違っていた（`Accompaniment.hits` の欠落など）。

## 決定

- dev 依存に `openapi-typescript` を追加し、`frontend/src/shared/api/schema.gen.ts` を `docs/api/openapi.yaml` から生成する（`npm run gen:api`）。生成物はコミットする（ビルドにネットワークや生成手順を要求しないため）。
- `client.ts` は生成された `components['schemas']` に短い名前を付けるだけにし、API の形を手で書かない。
- `scripts/check-api-types.sh` が生成し直して差分を検査する。`make check`（`frontend-typecheck` の前）と CI `frontend` が実行する。生成に失敗する契約（YAML の不正など）もここで落ちる。
- `default` を持つ項目は任意のまま生成する（`--default-non-nullable=false`）。クライアントは省略して送るため。
- 生成物は ESLint と PR サイズ予算の対象外。

## 結果

- Go（整合性テスト）と TypeScript（鮮度検査）の両側で契約との一致が機械的に保証される。
- 導入時に、YAML として不正な `description`（コロンを含む未クォート文字列）が見つかった。行単位の自前パーサでは検出できていなかった。
- 再検討の条件: 生成器が OpenAPI 3.1 の機能（`$ref` の兄弟キーなど）を誤って扱い、型を手で補う必要が出たとき。
