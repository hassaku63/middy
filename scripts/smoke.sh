#!/usr/bin/env bash
# Smoke-test a running middy against what serving it must do: the API's
# health, the SPA and its cache headers, the 404 rules (ADR-0022) and the
# body limit. The same checks run against `make up`, the CI image job, the
# e2e workflow and whatever is deployed later (#206), so a change of
# topology changes how the app is started, not what is expected.
#
# Usage: scripts/smoke.sh <base-url> [expected-version]
set -uo pipefail

base="${1:?usage: scripts/smoke.sh <base-url> [expected-version]}"
base="${base%/}"
want_version="${2:-}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
failures=0

check() { # check <description> <command...>
  local what="$1"
  shift
  if "$@"; then
    echo "ok    $what"
  else
    echo "FAIL  $what"
    failures=$((failures + 1))
  fi
}

# fetch <name> <path> [curl args...]: body to $tmp/<name>, headers to
# $tmp/<name>.h, status to $tmp/<name>.s.
fetch() {
  local name="$1" path="$2"
  shift 2
  curl -sS -o "$tmp/$name" -D "$tmp/$name.h" -w '%{http_code}' "$@" "$base$path" >"$tmp/$name.s" 2>/dev/null || echo 000 >"$tmp/$name.s"
}
status() { [ "$(cat "$tmp/$1.s")" = "$2" ]; }
header() { grep -qi "^$2:.*$3" "$tmp/$1.h"; }
body() { grep -q "$2" "$tmp/$1"; }

# The server may still be starting.
for _ in $(seq 1 30); do
  fetch health /api/v1/health
  status health 200 && break
  sleep 1
done
check "health answers 200" status health 200
if [ -n "$want_version" ]; then
  check "health reports version $want_version" body health "\"version\":\"$want_version\""
fi

fetch index /
check "/ is HTML" header index content-type text/html
check "/ holds the app root" body index '<div id="root">'
check "/ is revalidated" header index cache-control no-cache

fetch sw /sw.js
check "sw.js is revalidated" header sw cache-control no-cache
fetch manifest /manifest.webmanifest
check "the manifest is revalidated" header manifest cache-control no-cache

asset="$(grep -o '/assets/[^"]*\.js' "$tmp/index" | head -1)"
if [ -n "$asset" ]; then
  fetch asset "$asset"
  check "$asset is immutable" header asset cache-control immutable
else
  check "/ links a script under /assets/" false
fi

fetch missing /assets/does-not-exist.js
check "a missing file is 404" status missing 404
check "a missing file says not_found" body missing '"not_found"'
check "a missing file is not cached" header missing cache-control no-cache

fetch assetsdir /assets
check "/assets itself is 404" status assetsdir 404

fetch post /some/screen -X POST
check "a POST to a screen URL is 405" status post 405
check "a POST to a screen URL is JSON" header post content-type application/json

fetch screen /some/screen
check "a screen URL is 200" status screen 200
check "a screen URL is HTML" header screen content-type text/html

fetch api404 /api/v1/nope
check "an unknown API path is 404" status api404 404
check "an unknown API path is a problem detail" header api404 content-type application/problem+json
check "an unknown API path says not_found" body api404 '"code":"not_found"'

head -c 70000 /dev/zero | tr '\0' 'a' >"$tmp/big.json"
fetch big /api/v1/ideas -X POST -H 'Content-Type: application/json' --data-binary "@$tmp/big.json"
check "a 70 KB request is refused with 400" status big 400
check "a 70 KB request says invalid_json" body big '"invalid_json"'

if [ "$failures" -gt 0 ]; then
  echo "smoke: $failures check(s) failed against $base"
  exit 1
fi
echo "smoke: all checks passed against $base"
