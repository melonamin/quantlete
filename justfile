# Stata - Statistics for Strava
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
# Building
# ============================================================================

# Build everything
build: build-web build-go

# Build React frontend
build-web:
    cd web && yarn build

# Build Go binary (requires web to be built first)
build-go:
    go build -o bin/stata ./cmd/stata

# Build with version info
build-release version="dev":
    go build -ldflags "-X main.version={{version}} -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/stata ./cmd/stata

# Clean build artifacts
clean:
    rm -rf bin/
    rm -rf web/dist/

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
    docker build -t stata:latest .

# Run Docker container
docker-run:
    docker run -p 8080:8080 -v stata-data:/data stata:latest

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
# Setup
# ============================================================================

# Install all dependencies
setup: setup-go setup-web

# Install Go dependencies
setup-go:
    go mod download

# Install React dependencies
setup-web:
    cd web && yarn install

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
    @./bin/stata version 2>/dev/null || go run ./cmd/stata version

# Run the server
run: build
    ./bin/stata serve

# Import activities from Strava
import:
    go run ./cmd/stata import
