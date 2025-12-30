# Quantlete - Statistics for Strava
# Task runner using just (https://github.com/casey/just)

# Load .env file if it exists
set dotenv-load

# Default recipe - show available commands
default:
    @just --list

# ============================================================================
# Development
# ============================================================================

# Run both Go API and React dev server concurrently
dev:
    #!/usr/bin/env bash
    set -a && source .env 2>/dev/null && set +a
    lsof -ti:8081 | xargs -r kill -9 2>/dev/null || true
    trap 'kill 0' EXIT
    stdbuf -oL air 2>&1 | stdbuf -oL sed 's/^/[api] /' &
    (cd web && yarn dev 2>&1) | sed 's/^/[web] /' &
    wait

# Run Go API server in dev mode with hot reload
dev-api:
    #!/usr/bin/env bash
    set -a && source .env 2>/dev/null && set +a && exec air

# Run React dev server
dev-web:
    cd web && yarn dev

# ============================================================================
# Code Generation
# ============================================================================

# Generate all code (SQL queries, schema)
generate: generate-sql generate-schema

# Generate SQL query code (Go + TypeScript)
generate-sql:
    go run ./scripts/generate-sql

# Generate schema for browser WASM mode
generate-schema:
    go run ./scripts/generate-schema

# ============================================================================
# WASM Algorithms (TinyGo)
# ============================================================================

# Build shared WASM algorithms with TinyGo
build-algorithms:
    #!/usr/bin/env bash
    set -e
    which tinygo > /dev/null || (echo "Error: TinyGo not installed. Run: yay -S tinygo-bin" && exit 1)
    mkdir -p algorithms/build web/public/wasm
    tinygo build -o algorithms/build/algorithms.wasm -target wasm -opt 2 -no-debug ./algorithms/go/wasm
    cp algorithms/build/algorithms.wasm web/public/wasm/
    cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" web/public/wasm/

# Build algorithms for development (with debug info, larger binary)
build-algorithms-debug:
    #!/usr/bin/env bash
    set -e
    which tinygo > /dev/null || (echo "Error: TinyGo not installed. Run: yay -S tinygo-bin" && exit 1)
    mkdir -p algorithms/build web/public/wasm
    tinygo build -o algorithms/build/algorithms.wasm -target wasm ./algorithms/go/wasm
    cp algorithms/build/algorithms.wasm web/public/wasm/
    cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" web/public/wasm/

# Test algorithms (native Go)
test-algorithms:
    go test -v ./algorithms/go/...

# Check TinyGo installation
check-tinygo:
    @which tinygo > /dev/null && tinygo version || echo "TinyGo not installed. Run: yay -S tinygo-bin"

# ============================================================================
# Building
# ============================================================================

# Build everything
build: generate build-web build-go

# Build React frontend (server mode, for Go embedding)
build-web:
    cd web && yarn build:server

# Build React frontend (WASM mode, for static hosting)
build-web-wasm:
    cd web && yarn build:wasm

# Build Go binary (requires web to be built first)
build-go:
    go build -o bin/quantlete ./cmd/quantlete

# Build with version info
build-release version="dev":
    go build -ldflags "-X main.version={{version}} -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/quantlete ./cmd/quantlete

# Clean build artifacts
clean:
    rm -rf bin/
    rm -rf web/dist/
    rm -rf web/dist-wasm/

# ============================================================================
# Testing
# ============================================================================

# Run all tests
test: test-go test-web

# Run Go tests
test-go:
    go test -v ./cmd/... ./internal/...

# Run Go tests with coverage
test-go-cover:
    go test -v -coverprofile=coverage.out ./cmd/... ./internal/...
    go tool cover -html=coverage.out -o coverage.html

# Run React unit tests
test-web:
    cd web && yarn test:unit

# Run Playwright E2E tests
test-e2e:
    cd web && yarn test

# ============================================================================
# Linting & Formatting
# ============================================================================

# Run all linters
lint: lint-go lint-web

# Lint Go code
lint-go:
    golangci-lint run ./cmd/... ./internal/...

# Lint React code
lint-web:
    cd web && yarn lint && yarn typecheck

# Fix linting issues
lint-fix: lint-fix-go lint-fix-web

# Fix Go linting issues
lint-fix-go:
    golangci-lint run --fix ./cmd/... ./internal/...

# Fix React linting issues
lint-fix-web:
    cd web && yarn lint:fix

# Format code
fmt: fmt-go fmt-web

# Format Go code
fmt-go:
    gofmt -w cmd/ internal/
    goimports -w cmd/ internal/

# Format React code
fmt-web:
    cd web && yarn prettier:write

# Check formatting
fmt-check: fmt-check-go fmt-check-web

# Check Go formatting
fmt-check-go:
    @test -z "$(gofmt -l cmd/ internal/)" || (echo "Go files need formatting:" && gofmt -l cmd/ internal/ && exit 1)

# Check React formatting
fmt-check-web:
    cd web && yarn prettier

# ============================================================================
# Database
# ============================================================================

# Run database migrations (placeholder)
db-migrate:
    @echo "Database migrations not yet implemented"

# ============================================================================
# Docker
# ============================================================================

# Build Docker image
docker-build:
    docker build -t quantlete:latest .

# Run Docker container
docker-run:
    docker run -p 8080:8080 -v quantlete-data:/data quantlete:latest

# ============================================================================
# Release
# ============================================================================

# Create a release with GoReleaser
release:
    goreleaser release --clean

# Create a snapshot release (for testing)
release-snapshot:
    goreleaser release --snapshot --clean

# ============================================================================
# Cloudflare Deployment (WASM mode)
# ============================================================================

# Deploy WASM mode to Cloudflare Pages
deploy-wasm: build-web-wasm
    cd web && npx wrangler pages deploy dist-wasm --project-name=quantlete

# Deploy worker to Cloudflare Workers
deploy-worker:
    cd worker && npm run deploy

# Deploy both worker and pages
deploy-cf: deploy-worker deploy-wasm

# Preview WASM mode locally
preview-wasm:
    cd web && yarn preview:wasm

# ============================================================================
# Setup
# ============================================================================

# Install all dependencies
setup: setup-go setup-web setup-worker check-tinygo

# Install Go dependencies
setup-go:
    go mod download

# Install React dependencies
setup-web:
    cd web && yarn install

# Install Worker dependencies (for WASM mode OAuth proxy)
setup-worker:
    cd worker && npm install

# Install development tools
setup-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/goreleaser/goreleaser/v2@latest
    go install github.com/air-verse/air@latest

# ============================================================================
# Utilities
# ============================================================================

# Show version
version:
    @./bin/quantlete version 2>/dev/null || go run ./cmd/quantlete version

# Run the server
run: build
    ./bin/quantlete serve

# Import activities from Strava
import:
    go run ./cmd/quantlete import
