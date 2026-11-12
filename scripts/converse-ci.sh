#!/usr/bin/env bash
# Measure what survives a conversation (#453) on the real LLM, for the
# needs-eval-converse label (BL-313). Starts the API server here with
# LLM_PROVIDER=anthropic, sends backend/cmd/converse/turns-447.json three
# times with cmd/converse, and writes the report's table to <report.md>.
# The report is for reading, not a gate.
#
# The key reaches the server's environment only: cmd/converse is a client
# of the server and runs without it.
#
# Usage: ANTHROPIC_API_KEY=... scripts/converse-ci.sh <report.md>
#        LLM_PROVIDER=fake scripts/converse-ci.sh <report.md>   # a dry run without a key
set -euo pipefail

report="$(realpath -m "${1:?usage: scripts/converse-ci.sh <report.md>}")"
work="$(mktemp -d)"
port=18080
base="http://localhost:$port"

cd "$(dirname "$0")/../backend"
go build -o "$work/server" ./cmd/server
go build -o "$work/converse" ./cmd/converse

PORT="$port" LLM_PROVIDER="${LLM_PROVIDER:-anthropic}" "$work/server" > "$work/server.log" 2>&1 &
server=$!
trap 'kill "$server" 2> /dev/null || true; rm -rf "$work"' EXIT

up=false
for _ in $(seq 1 30); do
  if curl -fsS "$base/api/v1/health" > /dev/null 2>&1; then
    up=true
    break
  fi
  sleep 1
done
if [ "$up" != true ]; then
  echo "::error::the server did not answer its health check"
  cat "$work/server.log"
  exit 1
fi

env -u ANTHROPIC_API_KEY "$work/converse" collect -base "$base" \
  -turns cmd/converse/turns-447.json -runs 3 -out "$work/out"
mkdir -p "$(dirname "$report")"
{
  echo "会話をまたいだ計測（cmd/converse、turns-447 を 3 回。報告だけで、合否には使わない）"
  echo
  env -u ANTHROPIC_API_KEY "$work/converse" report "$work"/out/*.json
} > "$report"
cat "$report"
