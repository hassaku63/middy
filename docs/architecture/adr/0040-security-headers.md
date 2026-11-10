# 0040. web のコンテナがセキュリティヘッダと CSP を付ける

- 日付: 2026-11-10
- 状態: Accepted（所有者の判断は #206）

## 文脈

一般公開に向けて（#206）、配信の層にセキュリティヘッダも CSP もなかった。配信の規則は nginx が持ち（ADR-0027）、API は Go が答える。フロントのビルドには、インラインの `<script>`・`<style>` も、外部のホストもない（#206 の調査）。CodeQL の警告の見立て（#280）では、API の応答に `X-Content-Type-Options: nosniff` を付ける強化を勧めていた。

## 決定

1. **nginx がすべての応答に付ける**:
   - `X-Content-Type-Options: nosniff`
   - `Referrer-Policy: strict-origin-when-cross-origin`
   - `X-Frame-Options: DENY`
   - `Permissions-Policy`（camera・microphone・geolocation を空に）
   - `Strict-Transport-Security`（HTTP では無視されるので常に付ける）
   - CSP: `default-src 'self'`。スクリプトもスタイルも同じオリジンだけで、`frame-ancestors 'none'`、`object-src 'none'` も付ける。
2. **ヘッダは 1 つの snippet（`deploy/nginx/snippets/security-headers.conf`）に置く。** server と、自分で `add_header` を持つ各 location が include する。nginx は、location に `add_header` があると server の分を引き継がないため。`always` で、エラーの応答にも付ける。
3. **Go の API も `nosniff` を付ける**（`httpapi.noSniff`）。nginx を通らずに API へ届いた場合にも効くようにするため。nginx は、上流からの同じヘッダを隠し、自分のものを 1 回だけ付ける。
4. **CSP は、最初から強制する**（Report-Only の段階は踏まない。報告の送り先がないため）。適合しているかは、主な操作で違反が 1 件も出ないことを E2E（`e2e/csp.spec.ts`）で確かめる。ヘッダが付いていることは `scripts/smoke.sh` で確かめる。

## 検討した選択肢

- A: nginx の snippet と、Go の `nosniff`（採用）。
- B: Go がすべてのヘッダを付ける。欠点：静的ファイルは nginx が返すので、届かない。
- C: CSP を Report-Only から始める。欠点：報告の送り先がなく、違反に気づけない。E2E で確かめるほうが確実。

## 結果

- インラインのスクリプトやスタイル、外部のホストを使うと、CSP に止められる。使うときは、この ADR と snippet を同時に直す。
- React が CSSOM 越しに設定するスタイルは、インラインのスタイルに当たらない。
- 読み取り専用のファイルシステム、イメージの走査、乱用の対策は、#206 の A-6 以降で決める。
