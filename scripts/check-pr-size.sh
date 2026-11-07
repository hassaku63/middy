#!/usr/bin/env bash
# Measures the size of a change against a base ref, excluding generated
# and lock files, and enforces the PR budget from AGENTS.md (ADR-0007):
#   > 600 changed lines: fail    > 400: warn
#
# Usage: scripts/check-pr-size.sh <base-ref> [head-ref]
set -uo pipefail
cd "$(dirname "$0")/.."

base="${1:?base ref required}"
head="${2:-HEAD}"
warn_at=400
fail_at=600

# Paths that do not count: lock files, binary assets, generated fixtures.
excludes=(
  ':!**/package-lock.json' ':!**/go.sum' ':!**/*.png' ':!**/*.svg' ':!**/*.mid'
  ':!backend/internal/web/dist/**'
)

total=$(git diff --numstat "$base...$head" -- . "${excludes[@]}" \
  | awk '{ if ($1 != "-") s += $1 + $2 } END { print s + 0 }')

echo "pr-size: $total changed lines (excluding lock files, binaries, generated)"
if [ "$total" -gt "$fail_at" ]; then
  echo "::error::PR touches $total lines (> $fail_at). Split it into slices (docs/dev/workflow.md)." >&2
  exit 1
fi
if [ "$total" -gt "$warn_at" ]; then
  echo "::warning::PR touches $total lines (> $warn_at). Consider splitting." >&2
fi
