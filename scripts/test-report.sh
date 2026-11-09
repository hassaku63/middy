#!/usr/bin/env bash
# Runs both test suites and prints the slowest tests and the totals, then
# warns when a suite exceeds its time budget (ADR-0007, "detect").
#
#   backend unit tests: 60 s   frontend unit tests: 30 s
#
# Usage: scripts/test-report.sh [backend|frontend|all]   (default: all)
#
# CI already runs each suite once with a JSON reporter and passes the file
# in, so the report does not run the tests a second time (ADR-0016):
#   GO_TEST_JSON=<go test -json output>   VITEST_JSON=<vitest json report>
# The step that produced the file owns pass / fail; the report only reads it.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

backend_budget=60
frontend_budget=30
top=10
status=0
which="${1:-all}"

report_backend() {
echo "## backend (go test -json)"
go_json=$(mktemp); trap 'rm -f "$go_json"' EXIT
if [ -n "${GO_TEST_JSON:-}" ]; then
  cp "$GO_TEST_JSON" "$go_json"; go_status=0
else
  (cd backend && go test -json ./... > "$go_json"); go_status=$?
fi
# One line per finished test: "<seconds> <package>.<Test>"
awk -F'"' '
  /"Action":"(pass|fail)"/ && /"Test":"/ {
    pkg=""; test=""; elapsed=0
    for (i = 1; i <= NF; i++) {
      if ($i == "Package") pkg = $(i+2)
      if ($i == "Test") test = $(i+2)
      if ($i == "Elapsed") { split($(i+1), a, "[:,}]"); elapsed = a[2] }
    }
    if (test !~ /\//) printf "%8.3f %s.%s\n", elapsed, pkg, test
  }' "$go_json" | sort -rn > "$go_json.tests"
echo "slowest $top:"; head -n "$top" "$go_json.tests"
go_total=$(awk '{ s += $1 } END { printf "%.1f", s }' "$go_json.tests")
go_count=$(wc -l < "$go_json.tests" | tr -d ' ')
echo "total: ${go_total}s across ${go_count} tests (budget ${backend_budget}s)"
rm -f "$go_json.tests"
[ "$go_status" -ne 0 ] && { echo "::error::go test failed"; status=1; }
echo "per package:"
awk -F'"' '/"Action":"(pass|fail)"/ && !/"Test":"/ {
    pkg=""; elapsed=0
    for (i = 1; i <= NF; i++) {
      if ($i == "Package") pkg = $(i+2)
      if ($i == "Elapsed") { split($(i+1), a, "[:,}]"); elapsed = a[2] }
    }
    printf "%8.2f %s\n", elapsed, pkg
  }' "$go_json" | sort -rn
awk -v t="$go_total" -v b="$backend_budget" 'BEGIN { exit !(t > b) }' && { echo "::warning::backend tests take ${go_total}s, over the ${backend_budget}s budget"; }
}

report_frontend() {
echo "## frontend (vitest json reporter)"
vt_json=$(mktemp)
if [ -n "${VITEST_JSON:-}" ]; then
  cp "$VITEST_JSON" "$vt_json"; vt_status=0
else
  (cd frontend && npx vitest run --reporter=json --outputFile="$vt_json" > /dev/null 2>&1); vt_status=$?
fi
python3 - "$vt_json" "$top" "$frontend_budget" <<'PY'
import json, sys
path, top, budget = sys.argv[1], int(sys.argv[2]), float(sys.argv[3])
report = json.load(open(path))
tests = []
for f in report.get("testResults", []):
    for a in f.get("assertionResults", []):
        name = f["name"].split("/src/")[-1] + " > " + a["fullName"]
        tests.append(((a.get("duration") or 0) / 1000.0, name))
tests.sort(reverse=True)
print(f"slowest {top}:")
for d, name in tests[:top]:
    print(f"{d:8.3f} {name}")
total = sum(d for d, _ in tests)
wall = sum((f.get("endTime", 0) - f.get("startTime", 0)) for f in report.get("testResults", [])) / 1000.0
print(f"total: {total:.1f}s in tests, {wall:.1f}s per-file wall across {len(tests)} tests (budget {budget:.0f}s)")
if wall > budget:
    print(f"::warning::frontend tests take {wall:.1f}s, over the {budget:.0f}s budget")
PY
rm -f "$vt_json"
[ "$vt_status" -ne 0 ] && { echo "::error::vitest failed"; status=1; }
}

case "$which" in
  backend) report_backend ;;
  frontend) report_frontend ;;
  all) report_backend; echo; report_frontend ;;
  *) echo "usage: $0 [backend|frontend|all]" >&2; exit 2 ;;
esac
exit $status
