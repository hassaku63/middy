#!/usr/bin/env bash
# Compares the local toolchain with .tool-versions and explains mismatches.
# Major.minor must match for Go and Node (patch may differ); golangci-lint
# must match major.minor too because lint rules change between minors.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

want() { awk -v tool="$1" '$1 == tool { print $2 }' .tool-versions; }
minor() { echo "$1" | sed -E 's/^v//; s/^([0-9]+\.[0-9]+).*/\1/'; }

status=0
report() { # tool wanted found
  if [ -z "$3" ]; then
    echo "✗ $1: not installed (want $2)"; status=1
  elif [ "$(minor "$2")" != "$(minor "$3")" ]; then
    echo "✗ $1: found $3, want $2 (major.minor must match; CI uses $2)"; status=1
  else
    echo "✓ $1: $3 (want $2)"
  fi
}

# Run inside backend/ so the go.mod toolchain directive applies (Go may auto-download it).
report golang "$(want golang)" "$(cd backend && go version 2>/dev/null | awk '{ sub(/^go/, "", $3); print $3 }')"
report nodejs "$(want nodejs)" "$(node --version 2>/dev/null)"
report golangci-lint "$(want golangci-lint)" "$(golangci-lint version 2>/dev/null | sed -nE 's/.*version v?([0-9]+\.[0-9]+\.[0-9]+).*/\1/p')"

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  echo "✓ docker: $(docker version --format '{{.Server.Version}}' 2>/dev/null) (optional; needed for make up / make image)"
else
  echo "- docker: not available (optional; needed for make up / make image)"
fi

if [ "$status" -ne 0 ]; then
  echo
  echo "Install the versions in .tool-versions (asdf: \`asdf install\`, mise: \`mise install\`), or expect CI to differ." >&2
fi
exit $status
