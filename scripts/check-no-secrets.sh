#!/usr/bin/env bash
# Fails when a file holds something shaped like a secret (#502): an
# Anthropic or GitHub token, an AWS access key or a private key. Used on
# the evaluation's report before it leaves the job: GitHub masks secrets
# in logs, not in comments, summaries or artifacts. It catches mistakes,
# not someone set on leaking; it prints only where, never the match.
#
# Usage: scripts/check-no-secrets.sh <file>
set -euo pipefail

file="${1:?file required}"
[ -f "$file" ] || exit 0

pattern='sk-ant-[A-Za-z0-9_-]{8,}|gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|AKIA[0-9A-Z]{16}|-----BEGIN [A-Z ]*PRIVATE KEY-----'
lines=$(grep -nE "$pattern" "$file" | cut -d: -f1 | paste -sd, -) || true
if [ -n "$lines" ]; then
  echo "::error::$file holds a secret-shaped string (lines $lines); it is not posted"
  exit 1
fi
