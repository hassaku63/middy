#!/usr/bin/env bash
# Fails when frontend/src/shared/api/schema.gen.ts is not what
# docs/api/openapi.yaml generates (ADR-0009). Fix with `npm run gen:api`
# in frontend/ and commit the result. Also catches an OpenAPI document
# that no longer parses.
set -uo pipefail
cd "$(dirname "$0")/../frontend"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
out="$tmp/schema.gen.ts"

if ! npx --no-install openapi-typescript ../docs/api/openapi.yaml --default-non-nullable=false -o "$out" >"$tmp/log" 2>&1; then
  cat "$tmp/log" >&2
  echo "api-types: docs/api/openapi.yaml could not be turned into types (invalid document?)" >&2
  exit 1
fi
npx --no-install prettier --stdin-filepath src/shared/api/schema.gen.ts < "$out" > "$tmp/formatted.ts"

if ! diff -q "$tmp/formatted.ts" src/shared/api/schema.gen.ts >/dev/null; then
  diff -u src/shared/api/schema.gen.ts "$tmp/formatted.ts" | head -40 >&2
  echo "api-types: schema.gen.ts is stale; run \`npm run gen:api\` in frontend/ and commit it" >&2
  exit 1
fi
echo "api-types: ok"
