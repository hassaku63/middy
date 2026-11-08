#!/usr/bin/env bash
# Deterministic documentation checks. Part of `make check` and CI.
#
# What it enforces (see docs/dev/docs-hygiene.md for the why):
#   1. Every relative Markdown link resolves to an existing file.
#   2. Every repository path mentioned in backticks exists.
#   3. Every file under docs/ is listed in docs/README.md (ADRs via their index).
#   4. Every backlog ID (BL-NNN) mentioned anywhere exists in the backlog,
#      except in ADRs, which are historical records (ADR-0006).
#   5. Agent instruction files stay within their line budget (ADR-0007).
#   6. No merge conflict markers are left in any Markdown file.
#
# No dependencies beyond bash, grep, sed, find. Must stay portable (macOS/Linux).
set -uo pipefail

cd "$(dirname "$0")/.."

problems=$(mktemp)
trap 'rm -f "$problems"' EXIT
problem() { printf 'docs-check: %s\n' "$*" | tee -a "$problems" >&2; }

# Markdown files we own (skip dependencies and build output).
md_files() {
  find . -type f -name '*.md' \
    -not -path './frontend/node_modules/*' \
    -not -path './frontend/dist/*' \
    -not -path './.git/*' | sort
}

# --- 1. relative links ------------------------------------------------------
while IFS= read -r file; do
  dir=$(dirname "$file")
  # [text](target) where target is not a URL and not a pure anchor.
  grep -oE '\]\([^)]+\)' "$file" 2>/dev/null | sed -E 's/^\]\(//; s/\)$//' | while IFS= read -r target; do
    case "$target" in
      http://*|https://*|mailto:*|'#'*) continue ;;
    esac
    path="${target%%#*}"
    [ -z "$path" ] && continue
    if [ ! -e "$dir/$path" ]; then
      problem "$file: broken link -> $target"
    fi
  done
done < <(md_files)

# --- 2. backticked repository paths ----------------------------------------
# Matches `backend/...`, `frontend/...`, `docs/...`, `scripts/...`, `.github/...`, `.claude/...`.
# Placeholders containing < > * … are skipped. A trailing / means "directory".
# Paths that are gitignored (backend/.env, .claude/settings.local.json) are
# intentionally absent from the checkout and therefore allowed.
while IFS= read -r file; do
  grep -oE '`(backend|frontend|docs|scripts|\.github|\.claude)/[^` ]*`' "$file" 2>/dev/null \
    | tr -d '`' | sort -u | while IFS= read -r path; do
    case "$path" in
      *'<'*|*'>'*|*'*'*|*'…'*) continue ;;
    esac
    if [ ! -e "$path" ] && ! git check-ignore -q -- "$path"; then
      problem "$file: path does not exist -> $path"
    fi
  done
done < <(md_files)

# --- 3. docs index covers every doc ------------------------------------------
index=docs/README.md
while IFS= read -r doc; do
  rel="${doc#./docs/}"
  case "$rel" in
    README.md) continue ;;
    architecture/adr/*) # ADRs are indexed by their own README
      adr_rel="${rel#architecture/adr/}"
      [ "$adr_rel" = "README.md" ] && continue
      grep -qF "($adr_rel)" docs/architecture/adr/README.md || problem "docs/architecture/adr/README.md: ADR not indexed -> $rel"
      continue ;;
  esac
  grep -qF "($rel)" "$index" || problem "$index: document not indexed -> docs/$rel"
done < <(find ./docs -type f -name '*.md' | sort)

# --- 4. backlog IDs resolve ----------------------------------------------------
backlog=docs/product/backlog.md
known=$(grep -oE '^\| BL-[0-9]{3} ' "$backlog" | grep -oE 'BL-[0-9]{3}' | sort -u)
while IFS= read -r file; do
  [ "$file" = "./$backlog" ] && continue
  case "$file" in ./docs/architecture/adr/*) continue ;; esac
  grep -oE 'BL-[0-9]{3}' "$file" 2>/dev/null | sort -u | while IFS= read -r id; do
    printf '%s\n' "$known" | grep -qx "$id" || problem "$file: unknown backlog id -> $id"
  done
done < <(md_files; find .github -type f \( -name '*.yml' -o -name '*.yaml' \) | sort)

# --- 5. instruction files stay short ---------------------------------------
budget=150
for f in AGENTS.md CLAUDE.md; do
  lines=$(wc -l < "$f" | tr -d ' ')
  if [ "$lines" -gt "$budget" ]; then
    problem "$f: $lines lines exceeds the $budget-line budget; move detail into docs/ and link it"
  fi
done

# --- 6. no leftover merge conflict markers ------------------------------------
while IFS= read -r file; do
  # A conflict always has both ends; a lone "=======" is a setext heading.
  if grep -qE '^<<<<<<< ' "$file" && grep -qE '^>>>>>>> ' "$file"; then
    problem "$file: merge conflict markers left in the file"
  fi
done < <(md_files)

if [ -s "$problems" ]; then
  echo "docs-check: FAILED ($(wc -l < "$problems" | tr -d ' ') problems)" >&2
  exit 1
fi
echo "docs-check: ok"
