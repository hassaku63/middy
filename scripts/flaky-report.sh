#!/usr/bin/env bash
# Prints a Markdown list of CI workflow runs that only passed after a
# re-run within the last N days (default 30): the cheapest signal of a
# flaky test or job (ADR-0007, backlog BL-054). Needs the gh CLI with
# read access to Actions. An empty output means nothing suspicious.
#
# Usage: scripts/flaky-report.sh <owner/repo> [days]
set -uo pipefail

repo="${1:?owner/repo required}"
days="${2:-30}"
since=$(date -u -d "-${days} days" +%Y-%m-%d 2>/dev/null || date -u -v-"${days}"d +%Y-%m-%d)

# Page explicitly (at most 10 x 100 runs) rather than with --paginate: some
# proxies reject the numeric-id URLs in the Link header.
runs=""
for page in $(seq 1 10); do
  chunk=$(gh api "repos/$repo/actions/runs?created=%3E%3D${since}&per_page=100&page=${page}" \
    --jq '.workflow_runs[] | select(.run_attempt > 1 and .conclusion == "success")
          | "- \(.name) `\(.head_branch)` attempt \(.run_attempt): \(.html_url)"' 2>/dev/null) || {
    echo "flaky-report: could not read workflow runs for $repo" >&2
    exit 1
  }
  count=$(gh api "repos/$repo/actions/runs?created=%3E%3D${since}&per_page=100&page=${page}" --jq '.workflow_runs | length' 2>/dev/null || echo 0)
  [ -n "$chunk" ] && runs="${runs}${chunk}
"
  [ "$count" -lt 100 ] && break
done
runs=$(printf '%s' "$runs" | sed '/^$/d')
if [ -n "$runs" ]; then
  printf '\n### 再実行で通った CI（直近 %s 日、flaky の疑い）\n\n' "$days"
  printf '%s\n' "$runs"
fi
