#!/usr/bin/env bash
# Prints a Markdown report of documentation that deserves a human look.
# Heuristics only; an empty report means "nothing stands out".
#
#   - backlog items left "In progress"
#   - backlog items waiting in Next/Later/Ideas for a long time (default 90 days)
#   - docs not touched for a long time (default 120 days)
#   - words that usually mark a stale statement ("予定", "現在 NN", "TODO", "まだ");
#     lines quoting the word itself (「予定」) are rule text, not statements
#
# Usage: scripts/docs-staleness.sh [max-doc-age-days] [max-backlog-age-days]
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

max_age="${1:-120}"
max_backlog_age="${2:-90}"
now=$(date +%s)

section() { printf '\n### %s\n\n' "$1"; }

in_progress=$(grep -E '^\| BL-[0-9]{3} .*\| In progress \|' docs/product/backlog.md | grep -oE 'BL-[0-9]{3}' || true)
if [ -n "$in_progress" ]; then
  section "In progress のまま残っている backlog 項目"
  while IFS= read -r id; do printf -- '- %s\n' "$id"; done <<<"$in_progress"
fi

# Backlog rows that have sat unchanged in a waiting section. git blame gives
# the last change of each line; awk tracks which "## " section a line is in.
aging=$(git blame --line-porcelain docs/product/backlog.md 2>/dev/null | awk -v now="$now" -v max="$max_backlog_age" '
  /^author-time / { t = $2; next }
  /^\t/ {
    line = substr($0, 2)
    if (line ~ /^## /) { section = line; next }
    if (section ~ /^## (Next|Later|Ideas)/ && line ~ /^\| BL-[0-9]+ /) {
      age = int((now - t) / 86400)
      if (age >= max) { split(line, f, "|"); gsub(/^ +| +$/, "", f[2]); printf "- %s（%d 日、%s）\n", f[2], age, section }
    }
  }')
if [ -n "$aging" ]; then
  section "${max_backlog_age} 日以上動いていない backlog 項目（Now に上げるか、Ideas に落とすか、消す）"
  printf '%s\n' "$aging"
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
  # shellcheck disable=SC2016 # the backticks are Markdown, not a command
  printf '%s\n' "$markers" | sed -E 's/^/- `/; s/$/`/'
fi
