# 0033. problem detail は API の応答にだけ使い、アプリのファイルと画面のエラーは短い平文にする

- 日付: 2026-11-09
- 状態: Accepted（所有者の指摘は #267 のあとの会話。ADR-0030 の文脈の「nginx が自分で答えるエラーも API と同じ形」を変える）

## 文脈

ADR-0030 は、API のエラー応答を RFC 9457 の problem detail にした。BL-164 では、nginx が自分で答える 4 種類（404、405、413、503）もすべて problem detail にした。そのうち 404 と 405 は、存在しないファイル（ADR-0022）や画面の URL への POST で、`/api` の外で起きる。

RFC 9457 は HTTP API のエラーの形を決める仕様で、アプリのファイルや画面を配る応答のためのものではない。`/api` の外の応答を受け取るのはブラウザと Service Worker で、本文を読む API の利用者ではない。所有者は、フロントのコンテンツに対する応答は HTTP API の話と分けて考えるべきだと指摘した。

## 決定

1. **problem detail は `/api` の下の応答にだけ使う。** API が返すもの（Go の 404・405 を含む）と、nginx が API の代わりに答えるもの（413 の本文の上限、503 の API に届かない）。種類の一覧は `docs/api/problems.md`。
2. **アプリのファイルと画面の URL で nginx が答えるエラー（404、405、413）は、短い `text/plain` にする。** ステータスと、今のヘッダー（404 の `Cache-Control: no-cache`、405 の `Allow: GET, HEAD`）は変えない。nginx の既定の HTML のページは返さない。
3. **nginx の設定では、`/api` の location に `error_page` を置いて分ける。** サーバー全体の `error_page` は平文を返し、`/api` の location だけが problem detail の名前つき location を指す。

## 検討した選択肢

- A: `/api` だけを problem detail にし、ほかは平文（採用）。利点: 仕様の適用範囲に合う。`problems.md` は API の契約だけを書けばよい。
- B: すべてを problem detail のまま（BL-164 の形）。欠点: API でない応答に API のエラーの型を付け、`problems.md` に API でない場面が混ざる。
- C: アプリのファイルと画面のエラーは nginx の既定の HTML。欠点: nginx の名前と版の手がかりを含むページになり、壊れたファイルの読み込みに HTML を返す。

## 結果

- フロントの API クライアントが読むのは `/api` の応答だけなので、画面の振る舞いは変わらない。
- `scripts/smoke.sh` は、存在しないファイルと画面の URL への POST が平文であること、`/api` の 404 と 413 が problem detail であることを確かめる。
- `/api` の外に API を足すときは、その location にも `/api` と同じ `error_page` を置く。
