#!/usr/bin/env bash
# Prints a Markdown report of documentation that deserves a human look.
# Heuristics only; an empty report means "nothing stands out".
#
#   - backlog items left "In progress"
#   - docs not touched for a long time (default 120 days)
#   - words that usually mark a stale statement ("予定", "現在 NN", "TODO", "まだ");
#     lines quoting the word itself (「予定」) are rule text, not statements
#
# Usage: scripts/docs-staleness.sh [max-age-days]
set -uo pipefail
cd "$(dirname "$0")/.."

max_age="${1:-120}"
now=$(date +%s)

section() { printf '\n### %s\n\n' "$1"; }

in_progress=$(grep -E '^\| BL-[0-9]{3} .*\| In progress \|' docs/product/backlog.md | grep -oE 'BL-[0-9]{3}' || true)
if [ -n "$in_progress" ]; then
  section "In progress のまま残っている backlog 項目"
  printf -- '- %s\n' $in_progress
fi

old=""
while IFS= read -r file; do
  last=$(git log -1 --format=%ct -- "$file" 2>/dev/null)
  [ -z "$last" ] && last=$now # uncommitted file
  age=$(( (now - last) / 86400 ))
  if [ "$age" -ge "$max_age" ]; then
    old="$old
- \`$file\` (${age} 日)"
  fi
done < <(find docs AGENTS.md README.md -type f -name '*.md' | sort)
if [ -n "$old" ]; then
  section "${max_age} 日以上更新されていない文書（内容が現状と合っているか確認）"
  printf '%s\n' "$old" | sed '/^$/d'
fi

# ADRs are historical records and may legitimately talk about the future.
markers=$(grep -nE '予定|現在 ?[0-9]|TODO|まだ(実装|存在)' --include='*.md' -r docs AGENTS.md README.md 2>/dev/null | grep -vE '^docs/(dev/docs-hygiene\.md|architecture/adr/)' | grep -v '「予定」' || true)
if [ -n "$markers" ]; then
  section "腐りやすい言い回しを含む行（実態と照合する）"
  printf '%s\n' "$markers" | sed -E 's/^/- `/; s/$/`/'
fi
