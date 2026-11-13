#!/usr/bin/env bash
# Lists the E2E tests that failed and then passed on Playwright's retry
# (#501): the run is green, so without this nothing shows they wobble.
# Reads Playwright's JSON report; prints a GitHub warning per test and,
# in Actions, a list in the job summary. No report, or none flaky: quiet.
#
# Usage: scripts/e2e-flaky.sh <results.json>
set -euo pipefail

report="${1:?results.json required}"
[ -f "$report" ] || exit 0

flaky=$(jq -r '
  [.. | objects | select(has("specs")) | .specs[]
   | select(any(.tests[]; .status == "flaky"))
   | "\(.file)\t\(.line)\t\(.title)"] | unique | .[]' "$report")
[ -n "$flaky" ] || exit 0

{
  echo "### Flaky E2E tests (passed on retry)"
  echo
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}"
while IFS=$'\t' read -r file line title; do
  echo "::warning file=frontend/e2e/${file},line=${line}::flaky (passed on retry): ${title}"
  echo "- \`e2e/${file}:${line}\` ${title}" >> "${GITHUB_STEP_SUMMARY:-/dev/null}"
done <<<"$flaky"
