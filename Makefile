# middy developer harness.
#
# `make check` is the single gate every change must pass before commit.
# CI runs exactly the same targets, so a green `make check` locally means a
# green CI. Keep this file boring: no hidden behaviour, no network at check time.

.DEFAULT_GOAL := help
SHELL := /bin/bash

BACKEND  := backend
FRONTEND := frontend

.PHONY: help setup check test lint fmt fmt-check build dev-backend dev-frontend \
        backend-test backend-lint backend-fmt backend-fmt-check backend-build \
        frontend-test frontend-lint frontend-fmt frontend-fmt-check frontend-build frontend-typecheck \
        clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install all dependencies (Go modules, npm packages)
	cd $(BACKEND) && go mod download
	cd $(FRONTEND) && npm ci

check: fmt-check lint test build ## Run every gate (format, lint, test, build) for both apps

test: backend-test frontend-test ## Run all tests
lint: backend-lint frontend-lint ## Run all linters
fmt: backend-fmt frontend-fmt ## Format all code in place
fmt-check: backend-fmt-check frontend-fmt-check ## Verify formatting without changing files
build: backend-build frontend-build ## Build both apps

dev-backend: ## Run the API server with live config from backend/.env (if present)
	cd $(BACKEND) && set -a && { [ -f .env ] && source .env || true; } && set +a && go run ./cmd/server

dev-frontend: ## Run the Vite dev server (proxies /api to :8080)
	cd $(FRONTEND) && npm run dev

# --- backend -------------------------------------------------------------

backend-test: ## Go tests with race detector and coverage
	cd $(BACKEND) && go test -race -cover ./...

backend-lint: ## go vet + golangci-lint
	cd $(BACKEND) && go vet ./... && golangci-lint run ./...

backend-fmt: ## gofmt in place
	cd $(BACKEND) && gofmt -w .

backend-fmt-check: ## Fail if any Go file is not gofmt-ed
	@cd $(BACKEND) && out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

backend-build: ## Compile the server binary to backend/bin/server
	cd $(BACKEND) && go build -o bin/server ./cmd/server

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

clean: ## Remove build outputs (backend/bin, frontend/dist, dev-dist, coverage)
	rm -rf backend/bin frontend/dist frontend/dev-dist frontend/coverage
