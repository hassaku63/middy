# 問題の種類（problem types）

API（`/api` の下）が返すエラーの種類の一覧。前段の nginx が API の代わりに答えるものも含む。アプリのファイルと画面の URL のエラーは平文で、ここには載せない（ADR-0033）。エラー応答は RFC 9457 の problem detail（`application/problem+json`）で、`type` はこの文書の各見出しを指す（ADR-0030）。

```json
{
  "type": "https://github.com/hassaku63/middy/blob/main/docs/api/problems.md#invalid_request",
  "title": "Invalid request",
  "status": 400,
  "detail": "progression must not be empty",
  "code": "invalid_request"
}
```

- `title` は種類ごとに決まっている。`detail` は、その出来事に固有の開発者向けの説明で、画面には出さない。
- `code` は拡張メンバーで、見出しと同じ識別子。画面の文言は `code` から選ぶ。
- 種類を足すときは、ここに見出しを足す。Go のテスト（`backend/internal/httpapi`）が、Go の使う種類とこの文書の見出しの一致を確かめる。

## invalid_json

- status: 400、title: `Invalid JSON`
- 要求の本文が JSON として読めない。未知のフィールド、後ろに続く余計な値、本文の大きさの上限を超えたものも含む。

## invalid_request

- status: 400、title: `Invalid request`
- JSON としては読めるが、値が契約を満たさない（空の進行、範囲外の値など）。`detail` に理由がある。

## not_found

- status: 404、title: `Not found`
- そのパスに API のエンドポイントがない。

## method_not_allowed

- status: 405、title: `Method not allowed`
- パスはあるが、そのメソッドを受け付けない。`Allow` ヘッダーに受け付けるメソッドがある。

## request_too_large

- status: 413、title: `Request too large`
- 要求の本文が上限を超えた。上限は経路ごとに API が持つ（`/timelines/from-progression` は 64 KiB、`/ideas` は契約の最大の会話が入る大きさ、`/export/midi` は 4 MiB）。4 MiB を超える本文は、前段の nginx が先に答える。

## llm_declined

- status: 422、title: `Model declined`
- 言語モデルが、その会話への答えを断った。言い換えると答えられることがある。

## internal_error

- status: 500、title: `Internal error`
- サーバーの中の予期しない失敗。

## llm_unavailable

- status: 502、title: `Model unavailable`
- 言語モデルの API に届かない、または API が失敗を返した。

## llm_invalid_output

- status: 502、title: `Model output invalid`
- 言語モデルの答えが、再試行のあとも契約を満たさなかった。

## api_unavailable

- status: 503、title: `API unavailable`
- nginx が API のコンテナに届かない（起動中、停止中、待ちの時間切れ）。nginx だけが答える。
