# middy developer harness.
#
# `make check` is the single gate every change must pass before commit.
# CI runs exactly the same targets, so a green `make check` locally means a
# green CI. Keep this file boring: no hidden behaviour, no network at check time.

.DEFAULT_GOAL := help
SHELL := /bin/bash

BACKEND  := backend
FRONTEND := frontend
# Override when the local toolchain lacks the coverage tools (e.g. an auto-downloaded
# toolchain module): make check GOTEST_FLAGS=-race
GOTEST_FLAGS ?= -race -cover
GO_VERSION   := $(shell awk '$$1 == "golang" { print $$2 }' .tool-versions)
NODE_VERSION := $(shell awk '$$1 == "nodejs" { print $$2 }' .tool-versions)
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: help setup check test lint fmt fmt-check build docs-check deps-check fixtures test-report doctor e2e smoke eval dev-backend dev-frontend run image up down \
        backend-test backend-lint backend-fmt backend-fmt-check backend-build \
        frontend-test frontend-lint shell-lint actions-lint dockerfile-lint frontend-fmt frontend-fmt-check frontend-build frontend-typecheck \
        clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

doctor: ## Compare local Go / Node / golangci-lint with .tool-versions
	bash scripts/doctor.sh

setup: ## Install all dependencies (Go modules, npm packages)
	cd $(BACKEND) && go mod download
	cd $(FRONTEND) && npm ci

check: fmt-check lint test build docs-check ## Run every gate (format, lint, test, build, docs) for both apps

test: backend-test frontend-test ## Run all tests
test-report: ## Slowest tests, totals and time budgets for both suites (scripts/test-report.sh)
	bash scripts/test-report.sh

eval: ## Run the prompt evaluation set against the real API (needs ANTHROPIC_API_KEY; prints a Markdown table)
	cd $(BACKEND) && go test -count=1 -tags eval -run TestLiveEval -v ./internal/ideas/evals/

e2e: ## Playwright scenarios against a running stack (E2E_BASE_URL, default http://localhost:8080; see `make up`)
	cd $(FRONTEND) && npx playwright test

smoke: ## Smoke-test a running stack (SMOKE_URL, default http://localhost:8080; see `make up`)
	bash scripts/smoke.sh $(or $(SMOKE_URL),http://localhost:8080)
lint: backend-lint frontend-lint shell-lint actions-lint ## Run all linters
fmt: backend-fmt frontend-fmt ## Format all code in place
fmt-check: backend-fmt-check frontend-fmt-check ## Verify formatting without changing files
build: frontend-build backend-build ## Build the frontend (frontend/dist) and the API server (backend/bin/server)

dev-backend: ## Run the API server; settings come only from the environment (ADR-0011)
	cd $(BACKEND) && go run ./cmd/server

dev-frontend: ## Run the Vite dev server (proxies /api to API_UPSTREAM, default http://localhost:8080)
	cd $(FRONTEND) && npm run dev

run: ## Run the built API server on :8080 (the API only; `make up` serves the UI too); build first with `make build`
	$(BACKEND)/bin/server

image: ## Build the two production images (middy-api:local, middy-web:local; ADR-0027) with the toolchain from .tool-versions
	docker build --target api --build-arg VERSION=$(VERSION) --build-arg GO_VERSION=$(GO_VERSION) --build-arg NODE_VERSION=$(NODE_VERSION) -t middy-api:local .
	docker build --target web --build-arg NODE_VERSION=$(NODE_VERSION) -t middy-web:local .

up: ## Build and run the production-like stack (nginx + API) with docker compose (http://localhost:8080; settings from the environment)
	VERSION=$(VERSION) GO_VERSION=$(GO_VERSION) NODE_VERSION=$(NODE_VERSION) docker compose up --build

down: ## Stop the compose stack
	docker compose down

# --- backend -------------------------------------------------------------

backend-test: ## Go tests with race detector and coverage (GOTEST_FLAGS)
	cd $(BACKEND) && go test $(GOTEST_FLAGS) ./...

shell-lint: ## shellcheck the scripts (install: apt install shellcheck, or brew install shellcheck)
	@command -v shellcheck >/dev/null || { echo "shellcheck not found: apt install shellcheck, or brew install shellcheck"; exit 1; }
	shellcheck scripts/*.sh

actions-lint: ## actionlint the GitHub workflows (install: go install github.com/rhysd/actionlint/cmd/actionlint@v<version in .tool-versions>)
	@command -v actionlint >/dev/null || { echo "actionlint not found: go install github.com/rhysd/actionlint/cmd/actionlint@v$$(awk '$$1 == "actionlint" { print $$2 }' .tool-versions)"; exit 1; }
	actionlint

dockerfile-lint: ## hadolint the Dockerfile, with a local hadolint or the pinned image (not in `check`: it needs one of them; CI runs it)
	@v=$$(awk '$$1 == "hadolint" { print $$2 }' .tool-versions); \
	if command -v hadolint >/dev/null; then hadolint Dockerfile; \
	elif command -v docker >/dev/null; then docker run --rm -i "ghcr.io/hadolint/hadolint:v$$v" < Dockerfile; \
	else echo "dockerfile-lint: needs hadolint $$v or docker"; exit 1; fi

backend-lint: ## go vet + golangci-lint (the config is verified first, as CI does)
	cd $(BACKEND) && go vet ./... && golangci-lint config verify && golangci-lint run ./...

backend-fmt: ## gofmt in place
	cd $(BACKEND) && gofmt -w .

backend-fmt-check: ## Fail if any Go file is not gofmt-ed
	@cd $(BACKEND) && out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

backend-build: ## Compile the API server to backend/bin/server
	cd $(BACKEND) && go build -ldflags "-X github.com/hassaku63/middy/backend/internal/httpapi.Version=$(VERSION)" -o bin/server ./cmd/server

# --- frontend ------------------------------------------------------------

frontend-typecheck: ## tsc -b, after checking the generated API types are current (ADR-0009)
	bash scripts/check-api-types.sh
	cd $(FRONTEND) && npm run typecheck

frontend-test: ## vitest
	cd $(FRONTEND) && npm test

frontend-lint: frontend-typecheck ## eslint and knip (unused exports, files, dependencies; ADR-0007), after typecheck
	cd $(FRONTEND) && npm run lint && npm run knip

frontend-fmt: ## prettier --write
	cd $(FRONTEND) && npm run format

frontend-fmt-check: ## prettier --check
	cd $(FRONTEND) && npm run format:check

frontend-build: ## vite production build (includes PWA assets) + bundle size budget
	cd $(FRONTEND) && npm run build
	bash scripts/check-bundle-size.sh

# --- docs ----------------------------------------------------------------

docs-check: ## Verify links, paths, index and backlog ids in Markdown (scripts/check-docs.sh)
	bash scripts/check-docs.sh

fixtures: ## Regenerate the cross-language fixtures the frontend tests check against (then commit them)
	cd $(BACKEND) && go run ./cmd/fixtures -out ../$(FRONTEND)/src/shared/api/fixtures

deps-check: ## go.mod tidy check + npm audit of production deps (needs network; CI runs them in the backend and frontend jobs)
	cd $(BACKEND) && go mod tidy -diff
	cd $(FRONTEND) && npm audit --omit=dev --audit-level=high

clean: ## Remove build outputs (backend/bin, frontend/dist, dev-dist, coverage)
	rm -rf backend/bin frontend/dist frontend/dev-dist frontend/coverage
