# middy developer harness.
#
# `make check` is the single gate every change must pass before commit.
# CI runs exactly the same targets, so a green `make check` locally means a
# green CI. Keep this file boring: no hidden behaviour, no network at check time.

.DEFAULT_GOAL := help
SHELL := /bin/bash

BACKEND  := backend
FRONTEND := frontend
EMBED_DIR := $(BACKEND)/internal/web/dist
# Override when the local toolchain lacks the coverage tools (e.g. an auto-downloaded
# toolchain module): make check GOTEST_FLAGS=-race
GOTEST_FLAGS ?= -race -cover
GO_VERSION   := $(shell awk '$$1 == "golang" { print $$2 }' .tool-versions)
NODE_VERSION := $(shell awk '$$1 == "nodejs" { print $$2 }' .tool-versions)
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: help setup check test lint fmt fmt-check build docs-check deps-check fixtures test-report doctor e2e dev-backend dev-frontend run image up down \
        backend-test backend-lint backend-fmt backend-fmt-check backend-build backend-embed \
        frontend-test frontend-lint frontend-fmt frontend-fmt-check frontend-build frontend-typecheck \
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

e2e: ## Playwright scenarios against a running server (E2E_BASE_URL, default http://localhost:8080; see `make run` / `make up`)
	cd $(FRONTEND) && npx playwright test
lint: backend-lint frontend-lint ## Run all linters
fmt: backend-fmt frontend-fmt ## Format all code in place
fmt-check: backend-fmt-check frontend-fmt-check ## Verify formatting without changing files
build: frontend-build backend-embed backend-build ## Build the single binary: vite build -> embed -> go build (ADR-0008)

dev-backend: ## Run the API server with live config from backend/.env (if present)
	cd $(BACKEND) && set -a && { [ -f .env ] && source .env || true; } && set +a && go run ./cmd/server

dev-frontend: ## Run the Vite dev server (proxies /api to :8080)
	cd $(FRONTEND) && npm run dev

run: ## Run the built single binary (UI + API on :8080); build first with `make build`
	$(BACKEND)/bin/server

image: ## Build the production container image (middy:local) with the toolchain from .tool-versions
	docker build --build-arg VERSION=$(VERSION) --build-arg GO_VERSION=$(GO_VERSION) --build-arg NODE_VERSION=$(NODE_VERSION) -t middy:local .

up: ## Build and run the production-like stack with docker compose (http://localhost:8080)
	VERSION=$(VERSION) GO_VERSION=$(GO_VERSION) NODE_VERSION=$(NODE_VERSION) docker compose up --build

down: ## Stop the compose stack
	docker compose down

# --- backend -------------------------------------------------------------

backend-test: ## Go tests with race detector and coverage (GOTEST_FLAGS)
	cd $(BACKEND) && go test $(GOTEST_FLAGS) ./...

backend-lint: ## go vet + golangci-lint
	cd $(BACKEND) && go vet ./... && golangci-lint run ./...

backend-fmt: ## gofmt in place
	cd $(BACKEND) && gofmt -w .

backend-fmt-check: ## Fail if any Go file is not gofmt-ed
	@cd $(BACKEND) && out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

backend-embed: ## Copy frontend/dist into the Go embed directory
	rm -rf $(EMBED_DIR)
	mkdir -p $(EMBED_DIR)
	cp -R $(FRONTEND)/dist/. $(EMBED_DIR)/
	touch $(EMBED_DIR)/.gitkeep

backend-build: ## Compile the server binary to backend/bin/server (embeds whatever is in the embed directory)
	cd $(BACKEND) && go build -ldflags "-X github.com/hassaku63/middy/backend/internal/httpapi.Version=$(VERSION)" -o bin/server ./cmd/server

# --- frontend ------------------------------------------------------------

frontend-typecheck: ## tsc -b
	cd $(FRONTEND) && npm run typecheck

frontend-test: ## vitest
	cd $(FRONTEND) && npm test

frontend-lint: frontend-typecheck ## eslint (after typecheck)
	cd $(FRONTEND) && npm run lint

frontend-fmt: ## prettier --write
	cd $(FRONTEND) && npm run format

frontend-fmt-check: ## prettier --check
	cd $(FRONTEND) && npm run format:check

frontend-build: ## vite production build (includes PWA assets)
	cd $(FRONTEND) && npm run build

# --- docs ----------------------------------------------------------------

docs-check: ## Verify links, paths, index and backlog ids in Markdown (scripts/check-docs.sh)
	bash scripts/check-docs.sh

fixtures: ## Regenerate the cross-language fixtures the frontend tests check against (then commit them)
	cd $(BACKEND) && go run ./cmd/fixtures -out ../$(FRONTEND)/src/shared/api/fixtures

deps-check: ## go.mod tidy check + npm audit of production deps (needs network; CI `deps` job runs this)
	cd $(BACKEND) && go mod tidy -diff
	cd $(FRONTEND) && npm audit --omit=dev --audit-level=high

clean: ## Remove build outputs (backend/bin, embedded frontend, frontend/dist, dev-dist, coverage)
	rm -rf backend/bin frontend/dist frontend/dev-dist frontend/coverage
	rm -rf backend/internal/web/dist && mkdir -p backend/internal/web/dist && touch backend/internal/web/dist/.gitkeep
