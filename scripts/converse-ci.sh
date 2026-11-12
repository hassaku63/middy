#!/usr/bin/env bash
# Measure conversations on the real LLM, for the evaluation labels
# (ADR-0063). Starts the API server here with LLM_PROVIDER=anthropic, sends
# conversations to it with cmd/converse, and writes the report to
# <report.md>. The report is for reading, not a gate.
#
#   converse   backend/cmd/converse/turns-447.json, three times (#453,
#              needs-eval-converse, BL-313)
#   scenarios  every script in backend/cmd/converse/scenarios, once each,
#              side by side, then the conversation reviewer (#473,
#              needs-eval-scenarios, BL-316, ADR-0064)
#
# The key reaches the server and the reviewer (converse judge) only: the
# client that sends the conversations and the report run without it.
#
# Usage: ANTHROPIC_API_KEY=... scripts/converse-ci.sh converse|scenarios <report.md>
#        LLM_PROVIDER=fake scripts/converse-ci.sh scenarios <report.md>   # a dry run without a key
set -euo pipefail

mode="${1:?usage: scripts/converse-ci.sh converse|scenarios <report.md>}"
report="$(realpath -m "${2:?usage: scripts/converse-ci.sh converse|scenarios <report.md>}")"
case "$mode" in
  converse | scenarios) ;;
  *)
    echo "::error::unknown mode: $mode"
    exit 1
    ;;
esac
work="$(mktemp -d)"
port=18080
base="http://localhost:$port"
provider="${LLM_PROVIDER:-anthropic}"

cd "$(dirname "$0")/../backend"
go build -o "$work/server" ./cmd/server
go build -o "$work/converse" ./cmd/converse

PORT="$port" LLM_PROVIDER="$provider" "$work/server" > "$work/server.log" 2>&1 &
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

mkdir -p "$(dirname "$report")"
if [ "$mode" = converse ]; then
  env -u ANTHROPIC_API_KEY "$work/converse" collect -base "$base" \
    -turns cmd/converse/turns-447.json -runs 3 -out "$work/out"
  {
    echo "会話をまたいだ計測（cmd/converse、turns-447 を 3 回。報告だけで、合否には使わない）"
    echo
    env -u ANTHROPIC_API_KEY "$work/converse" report "$work"/out/*.json
  } > "$report"
else
  # One run of each script, all at once: a conversation is sent turn by
  # turn, but the scripts do not wait for each other.
  pids=()
  for script in cmd/converse/scenarios/*.json; do
    name="$(basename "$script" .json)"
    env -u ANTHROPIC_API_KEY "$work/converse" collect -base "$base" \
      -turns "$script" -runs 1 -out "$work/out/$name" &
    pids+=("$!")
  done
  for pid in "${pids[@]}"; do
    wait "$pid"
  done
  {
    echo "会話のシナリオ（cmd/converse、scenarios の台本を 1 回ずつ。報告だけで、合否には使わない）"
    echo
    env -u ANTHROPIC_API_KEY "$work/converse" report "$work"/out/*/run-1.json
    LLM_PROVIDER="$provider" "$work/converse" judge "$work"/out/*/run-1.json
  } > "$report"
fi
cat "$report"
