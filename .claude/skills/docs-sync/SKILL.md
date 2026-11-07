---
name: docs-sync
description: PR を出す前に、差分で言及が変わる文書を逆引きして古い記述を直す自己点検。docs/dev/docs-hygiene.md の手順を実行する。
---

# docs-sync

`docs/dev/docs-hygiene.md` の「自己点検 `docs-sync`」の手順 1〜7 をそのまま実行する。手順の本体はその文書だけに書いてあり、ここには書かない（二重管理を避けるため）。

要点:

1. `git diff main...HEAD` から追加・削除・改名した名前（型、関数、ファイル、環境変数、エンドポイント、BL-ID）を拾う。
2. `grep -rn --include='*.md' -e '<name>' docs AGENTS.md README.md CLAUDE.md .github` で逆引きする。
3. 古い記述は直す。経緯の説明は現在形に書き直す。腐る値（数値・日付・「予定」）は削除するか BL-ID に置き換える。
4. `make docs-check` と `bash scripts/docs-staleness.sh` を実行し、自分の変更に関係する指摘を潰す。
5. 直すべき箇所がなければ、その旨を PR 本文に 1 行書く。

サブエージェントに委任する場合は、`docs/dev/docs-hygiene.md` の手順と差分を渡し、「ファイル:行、現状、修正案」の一覧だけを返させる。
