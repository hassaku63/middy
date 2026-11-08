#!/usr/bin/env bash
# Installs golangci-lint <version> into <dir> (ADR-0016). CI caches <dir>
# by version, so this runs only on a cache miss. It tries the release
# binary first (checksum verified, with retries). When GitHub's release
# downloads fail (they returned 504 for hours, #97), it builds the same
# version from the Go module proxy instead; that path needs Go on PATH and
# takes about a minute, which is acceptable once per version. Run it from
# backend/ so the fallback builds with the project's Go.
#
# Usage: scripts/install-golangci-lint.sh v2.5.0 "$HOME/.local/golangci-lint"
set -euo pipefail

version="${1:?version like v2.5.0}"
dest="${2:?destination directory}"
v="${version#v}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "install-golangci-lint: unsupported architecture $(uname -m)" >&2; exit 1 ;;
esac
name="golangci-lint-${v}-${os}-${arch}"
base="https://github.com/golangci/golangci-lint/releases/download/${version}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$dest"

fetch() { curl -fsSL --retry 4 --retry-all-errors --retry-delay 3 -o "$2" "$1"; }
sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | awk '{ print $1 }'; }

if fetch "$base/${name}.tar.gz" "$tmp/lint.tgz" && fetch "$base/golangci-lint-${v}-checksums.txt" "$tmp/sums.txt"; then
  want=$(awk -v f="${name}.tar.gz" '$2 == f { print $1 }' "$tmp/sums.txt")
  got=$(sha256 "$tmp/lint.tgz")
  if [ -z "$want" ] || [ "$want" != "$got" ]; then
    echo "install-golangci-lint: checksum mismatch for ${name}.tar.gz (want '${want}', got '${got}')" >&2
    exit 1
  fi
  tar -xzf "$tmp/lint.tgz" -C "$tmp"
  mv "$tmp/$name/golangci-lint" "$dest/golangci-lint"
else
  echo "::warning::golangci-lint release download failed; building ${version} from the Go module proxy"
  # Build with the Go this directory uses (run it from backend/). Otherwise
  # Go may pick an older toolchain, and that binary refuses to lint a module
  # that needs a newer Go.
  GOTOOLCHAIN="$(go env GOVERSION)" GOBIN="$dest" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${version}"
fi
"$dest/golangci-lint" version
