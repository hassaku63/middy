# CLAUDE.md

@AGENTS.md

## Claude Code 固有の補足

- 許可済みコマンドは `.claude/settings.json` に定義済み（`make`, `go test`, `npm run` など）。個人的な追加は `.claude/settings.local.json`（gitignore 済み）に書く。
- 作業開始時は `docs/product/backlog.md` の「Now」を読み、1 項目だけ選ぶ。完了時に同ファイルのステータスを更新する。
- `make check` が通るまで「完了」と報告しない。失敗したら出力をそのまま示す。
- 新しい設計判断をしたら `docs/architecture/adr/` にテンプレート（`0000-template.md`）から 1 枚追加する。
