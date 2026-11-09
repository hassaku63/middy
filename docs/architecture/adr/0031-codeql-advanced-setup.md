# 0031. CodeQL はワークフローで動かし、文書だけの変更では走らせない

- 日付: 2026-11-09
- 状態: Accepted（所有者の判断は #266 の C）

## 文脈

CodeQL は default setup（GitHub の設定画面で有効にする、リポジトリにファイルのない形）で動いている。PR と `main` への push のたびに 3 言語（actions、go、javascript-typescript）を解析する。default setup は走らせる条件を変えられないので、文書だけの PR でも PR 側と push 側の両方で走る。1 本あたり、切り上げで約 10 分の実行になる（#266 の測定）。文書だけの PR では、最後に終わるチェックが CodeQL になり、待ちが延びる。

CodeQL は必須チェックではない（ADR-0028 の決定 5）。

## 決定

1. **CodeQL を advanced setup（`.github/workflows/codeql.yml`）に移す。** 言語とビルドの形は default setup と同じにする（actions と javascript-typescript は `none`、go は `autobuild`）。Go は `.tool-versions` の版を使う（ADR-0016）。
2. **`docs/**` と `**/*.md` だけを変える PR と push では走らせない**（`paths-ignore`）。これらは解析の対象のコードを含まない。
3. **週に 1 回、すべてを解析する**（`schedule`）。文書だけの変更が続いても、新しいクエリでの解析が止まらない。
4. **PR では新しい push が古い実行を取り消す。`main` では走っている実行を止めない。** 待っている実行は、より新しい実行に置き換わってよい。解析は毎回コード全体なので、最新の実行が前の分も含む。
5. **default setup は所有者が設定画面で切る。** 両方が有効だと、advanced setup の結果の取り込みが拒まれる。

## 検討した選択肢

- A: advanced setup に移し、文書だけの変更を除く（採用）。欠点: ワークフローが 1 本増え、actionlint と Dependabot の対象になる。
- B: default setup のまま。欠点: 走らせる条件を変えられない。
- C: CodeQL をやめる。欠点: 公開リポジトリでは無料で、依存とコードの検査を 1 つ失う。

## 結果

- 文書だけの PR では CodeQL が走らない。コードの PR、`main` への push、週 1 回では 3 言語とも走る。
- 解析の条件や言語を変えるときは `codeql.yml` を直す。設定画面ではない。
- 解析の対象に Markdown 以外の文書の形式を足したら、`paths-ignore` を見直す。
