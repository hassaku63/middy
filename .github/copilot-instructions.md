# GitHub Copilot instructions for middy

The canonical, tool-agnostic instructions live in [`AGENTS.md`](../AGENTS.md) at the repository root. Read it first; this file only restates the rules that matter most when suggesting code.

- Backend is Go (`backend/`, standard library `net/http`, no framework). Frontend is React + TypeScript with Vite (`frontend/`).
- Music theory logic is deterministic and lives in `backend/internal/music`. LLM calls only interpret user intent into structured parameters; they never produce notes.
- Every behaviour change ships with tests. Go: table-driven `_test.go` next to the code. Frontend: Vitest + Testing Library in `*.test.ts(x)`.
- `make check` must pass (gofmt, golangci-lint, eslint, prettier, vitest, go test, builds).
- API shapes are the contract in `docs/api/openapi.yaml`; update it with any handler change.
- Do not add dependencies, skip tests, or disable lint rules to make something pass.
- Commit messages follow Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `test:`, `refactor:`).
