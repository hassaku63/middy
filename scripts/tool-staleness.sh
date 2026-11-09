#!/usr/bin/env bash
# Prints a Markdown report of the tools and dependencies that deserve a
# look (ADR-0029, decision 5). Facts from the package registries only; what
# to do about them is the quarterly review's call, not this script's.
#
#   - direct npm and Go dependencies marked deprecated
#   - direct dependencies with no release for a long time (default 365 days)
#   - direct npm dependencies a major version behind
#   - golangci-lint behind its latest release (minor versions too)
#   - the checklist for the review itself
#
# Usage: scripts/tool-staleness.sh [max-quiet-days]
# Needs node/npm, go, jq and curl, and network access to the registries.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

max_quiet="${1:-365}"
cutoff=$(date -u -d "-${max_quiet} days" +%Y-%m-%d 2>/dev/null || date -u -v-"${max_quiet}"d +%Y-%m-%d)

section() { printf '\n### %s\n\n' "$1"; }
major() { printf '%s' "${1%%.*}"; }

deprecated=""
quiet=""
behind=""
tools=""

# --- npm (frontend) ---------------------------------------------------------
pkgs=$(jq -r '(.dependencies // {}) + (.devDependencies // {}) | keys[]' frontend/package.json)
outdated=$(npm --prefix frontend outdated --json 2>/dev/null || true)
[ -n "$outdated" ] || outdated='{}'
for pkg in $pkgs; do
  installed=$(jq -r --arg p "$pkg" '.packages["node_modules/" + $p].version // empty' frontend/package-lock.json)
  [ -n "$installed" ] || continue
  meta=$(npm --prefix frontend view "$pkg@$installed" deprecated --json 2>/dev/null || true)
  if [ -n "$meta" ] && [ "$meta" != '""' ]; then
    deprecated+="- npm \`$pkg@$installed\`: $(jq -r '.' <<<"$meta")"$'\n'
  fi
  latest_time=$(npm --prefix frontend view "$pkg" --json time dist-tags 2>/dev/null \
    | jq -r '.time[.["dist-tags"].latest] // empty' | cut -c1-10)
  if [ -n "$latest_time" ] && [[ "$latest_time" < "$cutoff" ]]; then
    quiet+="- npm \`$pkg\`: the latest release is from $latest_time"$'\n'
  fi
  latest=$(jq -r --arg p "$pkg" '.[$p].latest // empty' <<<"$outdated")
  if [ -n "$latest" ] && [ "$(major "$latest")" != "$(major "$installed")" ]; then
    behind+="- npm \`$pkg\`: $installed installed, $latest is the latest"$'\n'
  fi
done

# --- Go (backend) -----------------------------------------------------------
gomods=$(cd backend && go list -m -u -json all 2>/dev/null \
  | jq -c 'select((.Main // false) == false and (.Indirect // false) == false)')
while IFS= read -r mod; do
  [ -n "$mod" ] || continue
  path=$(jq -r '.Path' <<<"$mod")
  dep=$(jq -r '.Deprecated // empty' <<<"$mod")
  [ -z "$dep" ] || deprecated+="- Go \`$path\`: $dep"$'\n'
  newest=$(jq -r '(.Update.Time // .Time // "")[0:10]' <<<"$mod")
  if [ -n "$newest" ] && [[ "$newest" < "$cutoff" ]]; then
    quiet+="- Go \`$path\`: the latest release is from $newest"$'\n'
  fi
done <<<"$gomods"

# --- golangci-lint ----------------------------------------------------------
pinned=$(awk '$1 == "golangci-lint" { print $2 }' .tool-versions)
newest_lint=$(curl -fsS --retry 3 https://proxy.golang.org/github.com/golangci/golangci-lint/v2/@latest 2>/dev/null \
  | jq -r '.Version // empty' | sed 's/^v//')
if [ -n "$newest_lint" ] && [ "$newest_lint" != "$pinned" ]; then
  tools+="- golangci-lint: $pinned in \`.tool-versions\`, $newest_lint is the latest"$'\n'
fi

if [ -n "$deprecated" ]; then section "非推奨の依存"; printf '%s' "$deprecated"; fi
if [ -n "$quiet" ]; then section "${max_quiet} 日以上公開のない依存"; printf '%s' "$quiet"; fi
if [ -n "$behind" ]; then section "大きな版が遅れている依存"; printf '%s' "$behind"; fi
if [ -n "$tools" ]; then section "最新でない道具"; printf '%s' "$tools"; fi

section "四半期の見直し（ADR-0029 の決定 5）"
cat <<'EOF'
- [ ] `docs/dev/harness.md` の各検査は、今も役に立っているか。捕まえた実績のない lint は削除の候補にする（ADR-0007）
- [ ] 上の依存と道具に、後継やより一般的な代わりが出ていないか（一次情報で確かめる）
- [ ] 検討の結果は企画の Issue にする（`/propose`）。実装はしない
EOF
