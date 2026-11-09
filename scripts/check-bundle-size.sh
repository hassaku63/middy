#!/usr/bin/env bash
# Bundle size budget for the PWA (ADR-0007 "detect", backlog BL-052).
# Runs after `vite build`; measures gzip -9 sizes, which approximate what
# the browser downloads.
#
#   largest JS chunk: 120 kB gzip    all JS: 160 kB gzip    all CSS: 20 kB gzip
#
# Usage: scripts/check-bundle-size.sh [dist-dir]
set -uo pipefail
dist="${1:-$(dirname "$0")/../frontend/dist}"
max_chunk_kb=120
max_js_kb=160
max_css_kb=20

if [ ! -d "$dist/assets" ]; then
  echo "bundle-size: $dist/assets not found; run \`make frontend-build\` first" >&2
  exit 1
fi

gz() { gzip -9c "$1" | wc -c | tr -d ' '; }
kb() { awk -v b="$1" 'BEGIN { printf "%.1f", b / 1024 }'; }

largest=0; largest_name=""; js_total=0; css_total=0
for f in "$dist"/assets/*.js; do
  [ -e "$f" ] || continue
  s=$(gz "$f"); js_total=$((js_total + s))
  if [ "$s" -gt "$largest" ]; then largest=$s; largest_name=$(basename "$f"); fi
done
for f in "$dist"/assets/*.css; do
  [ -e "$f" ] || continue
  css_total=$((css_total + $(gz "$f")))
done

echo "bundle-size: largest JS $(kb "$largest") kB ($largest_name), all JS $(kb "$js_total") kB, CSS $(kb "$css_total") kB (gzip)"
status=0
check() { # name actual-bytes budget-kb
  if [ "$2" -gt $(( $3 * 1024 )) ]; then
    echo "::error::$1 is $(kb "$2") kB gzip, over the $3 kB budget (docs/dev/harness.md)" >&2
    status=1
  fi
}
check "largest JS chunk" "$largest" "$max_chunk_kb"
check "total JS" "$js_total" "$max_js_kb"
check "total CSS" "$css_total" "$max_css_kb"
exit $status
