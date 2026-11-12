#!/usr/bin/env bash
# Decides which CI jobs a change needs (ADR-0016, ADR-0032), from the files
# changed between <base-ref> and <head-ref>: a pull request against its base
# branch, a push to main against the previous head. Prints
# `<job>=true|false` lines for $GITHUB_OUTPUT. This file is the one place
# that knows which paths feed
# which job; when a job starts reading a new path, add it here.
#
# Usage: scripts/ci-changes.sh <base-ref> [head-ref]
set -euo pipefail
cd "$(dirname "$0")/.."

base="${1:?base ref required}"
head="${2:-HEAD}"
changed=$(git diff --name-only "$base...$head")

# Paths that can change what every job does: the workflow, the pinned
# toolchains, the Makefile and the scripts the jobs call.
all='^(\.github/workflows/ci\.yml|\.tool-versions|Makefile|scripts/)'
# backend tests also read the API contract (route parity) and the
# committed fixtures (staleness check).
backend='^(backend/|docs/api/openapi\.yaml|frontend/src/shared/api/fixtures/)'
# frontend checks the generated API types against the contract; the Go
# side generates the fixtures the frontend tests compare against.
frontend='^(frontend/|docs/api/openapi\.yaml|backend/internal/fixtures/|backend/cmd/fixtures/)'
# the image builds both apps.
image='^(backend/|frontend/|Dockerfile|\.dockerignore|compose\.yaml|deploy/|scripts/smoke\.sh)'
# the real-API evaluation (eval.yml, ADR-0018) only matters when what the
# model is told, or how its answer is read and rendered, may have changed.
# It costs money, so it does not follow the "everything" paths above.
eval='^(backend/internal/(ideas|intent|llm|music)/|backend/cmd/converse/|backend/go\.(mod|sum)|\.github/workflows/eval\.yml|scripts/converse-ci\.sh|Makefile)'

touches() { grep -qE "$1" <<<"$changed"; }
job() { if touches "$all" || touches "$2"; then echo "$1=true"; else echo "$1=false"; fi; }
job backend "$backend"
job frontend "$frontend"
job image "$image"
if touches "$eval"; then echo "eval=true"; else echo "eval=false"; fi
