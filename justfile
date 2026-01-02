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

# Generate all code (SQL queries, schema, TypeScript types)
generate: generate-sql generate-schema generate-ts-types

# Generate SQL query code (Go + TypeScript)
generate-sql:
    go run ./scripts/generate-sql

# Generate schema for browser WASM mode
generate-schema:
    go run ./scripts/generate-schema

# Generate TypeScript types from Go WASM bridge structs
generate-ts-types:
    go run ./scripts/generate-ts-types

# ============================================================================
# Go WASM Storage Layer
# ============================================================================

# Build Go storage layer to WASM
build-go-wasm:
    #!/usr/bin/env bash
    set -e
    mkdir -p web/public/wasm
    echo "Building Go WASM..."
    GOOS=js GOARCH=wasm go build -o web/public/wasm/quantlete.wasm ./cmd/wasm/
    echo "Copying wasm_exec.js..."
    # Go 1.24+ uses lib/wasm, older versions use misc/wasm
    if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
        cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/public/wasm/
    else
        cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/public/wasm/
    fi
    echo "Copying sql-wasm.wasm..."
    cp web/node_modules/sql.js/dist/sql-wasm.wasm web/public/wasm/ 2>/dev/null || true
    ls -lh web/public/wasm/quantlete.wasm

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

# Show deployment checklist
deploy-checklist:
    @echo "Cloudflare Deployment Checklist"
    @echo "================================"
    @echo ""
    @echo "Prerequisites:"
    @echo "  1. Add quantlete.fit zone to Cloudflare"
    @echo "  2. Set CLOUDFLARE_ACCOUNT_ID in .env"
    @echo "  3. Login: wrangler login (use correct account)"
    @echo "  4. Verify: wrangler whoami"
    @echo ""
    @echo "Deploy Worker (proxy.quantlete.fit):"
    @echo "  just deploy-worker"
    @echo ""
    @echo "Deploy Pages (app.quantlete.fit):"
    @echo "  just deploy-wasm"
    @echo "  Then add custom domain in CF dashboard:"
    @echo "    Pages > quantlete > Custom domains > Add > app.quantlete.fit"
    @echo ""
    @echo "Deploy both:"
    @echo "  just deploy-cf"

# Deploy WASM mode to Cloudflare Pages
deploy-wasm: build-web-wasm
    npx wrangler pages deploy web/dist-wasm --project-name=quantlete --branch=main --commit-dirty=true

# Deploy worker to Cloudflare Workers
deploy-worker:
    cd worker && npm run deploy

# Deploy both worker and pages
deploy-cf: deploy-worker deploy-wasm
    @echo ""
    @echo "Deployed! Next steps:"
    @echo "  - Worker: Verify proxy.quantlete.fit is working"
    @echo "  - Pages: Add custom domain app.quantlete.fit in CF dashboard"
    @echo "    Pages > quantlete > Custom domains > Add domain"

# Preview WASM mode locally
preview-wasm:
    cd web && yarn preview:wasm

# ============================================================================
# Setup
# ============================================================================

# Install all dependencies
setup: setup-go setup-web setup-worker setup-landing check-tinygo

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
# Landing Page
# ============================================================================

# Run landing page dev server
dev-landing:
    cd landing && yarn dev

# Build landing page
build-landing:
    cd landing && yarn build

# Install landing page dependencies
setup-landing:
    cd landing && yarn install

# Preview landing page build
preview-landing:
    cd landing && yarn preview

# ============================================================================
# Documentation Site
# ============================================================================

# Serve docs locally with hot reload
dev-docs:
    npx docsify-cli serve docs

# Deploy docs to Cloudflare Pages
deploy-docs:
    npx wrangler pages deploy docs --project-name=quantlete-docs --branch=main --commit-dirty=true

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

# ============================================================================
# macOS App
# ============================================================================

# Build complete macOS app (Go binary + Swift app)
build-macos: build-macos-go build-macos-swift
    #!/usr/bin/env bash
    set -e
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    if [ -z "$APP_PATH" ]; then
        echo "Error: Quantlete.app not found in DerivedData"
        exit 1
    fi
    mkdir -p "$APP_PATH/Contents/Resources"
    cp bin/quantlete "$APP_PATH/Contents/Resources/"
    echo "Built: $APP_PATH"

# Build Go binary for macOS (universal binary)
build-macos-go: generate build-web
    #!/usr/bin/env bash
    set -e
    mkdir -p bin
    echo "Building for arm64..."
    GOOS=darwin GOARCH=arm64 go build -o bin/quantlete-darwin-arm64 ./cmd/quantlete
    echo "Building for amd64..."
    GOOS=darwin GOARCH=amd64 go build -o bin/quantlete-darwin-amd64 ./cmd/quantlete
    echo "Creating universal binary..."
    lipo -create -output bin/quantlete bin/quantlete-darwin-arm64 bin/quantlete-darwin-amd64
    rm bin/quantlete-darwin-arm64 bin/quantlete-darwin-amd64
    echo "Built: bin/quantlete (universal)"

# Build Swift macOS app
build-macos-swift:
    xcodebuild -project macos/Quantlete.xcodeproj -scheme Quantlete -configuration Debug build

# Build macOS app for release (signed)
build-macos-release: build-macos-go
    xcodebuild -project macos/Quantlete.xcodeproj -scheme Quantlete -configuration Release build
    #!/usr/bin/env bash
    set -e
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Release -name "Quantlete.app" -type d 2>/dev/null | head -1)
    mkdir -p "$APP_PATH/Contents/Resources"
    cp bin/quantlete "$APP_PATH/Contents/Resources/"
    echo "Built: $APP_PATH"

# Run macOS app
run-macos: build-macos
    #!/usr/bin/env bash
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    open "$APP_PATH"

# Clean macOS build artifacts
clean-macos:
    rm -rf ~/Library/Developer/Xcode/DerivedData/Quantlete-*

# Code signing identity (Developer ID for distribution outside App Store)
MACOS_SIGN_IDENTITY := "Developer ID Application: Ameba Labs, LLC (X93LWC49WV)"
MACOS_KEYCHAIN_PROFILE := "notarytool-kefir"

# Sign macOS binary and app bundle
sign-macos: build-macos-go
    #!/usr/bin/env bash
    set -e
    echo "Code signing Go binary..."
    codesign --force --options runtime --sign "{{MACOS_SIGN_IDENTITY}}" --timestamp bin/quantlete
    echo "Verifying binary signature..."
    codesign -dv --verbose=2 bin/quantlete

# Build and sign complete macOS app for distribution
build-macos-signed: sign-macos build-macos-swift
    #!/usr/bin/env bash
    set -e
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    if [ -z "$APP_PATH" ]; then
        echo "Error: Quantlete.app not found in DerivedData"
        exit 1
    fi
    mkdir -p "$APP_PATH/Contents/Resources"
    cp bin/quantlete "$APP_PATH/Contents/Resources/"

    echo "Code signing app bundle..."
    codesign --force --deep --options runtime --sign "{{MACOS_SIGN_IDENTITY}}" --timestamp "$APP_PATH"

    echo "Verifying app signature..."
    codesign -dv --verbose=2 "$APP_PATH"
    echo "Built and signed: $APP_PATH"

# Create zip archive for notarization
package-macos: build-macos-signed
    #!/usr/bin/env bash
    set -e
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    echo "Creating zip archive for notarization..."
    rm -f Quantlete.zip
    ditto -c -k --keepParent "$APP_PATH" Quantlete.zip
    echo "Archive created: Quantlete.zip"

# Submit for notarization
notarize-macos: package-macos
    #!/usr/bin/env bash
    set -e
    echo "Submitting for notarization..."
    xcrun notarytool submit Quantlete.zip \
        --keychain-profile "{{MACOS_KEYCHAIN_PROFILE}}" \
        --wait

    echo "Stapling notarization ticket to app..."
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    xcrun stapler staple "$APP_PATH"

    echo "Verifying notarization..."
    spctl -a -vvv -t exec "$APP_PATH" 2>&1 || true
    echo "App is notarized and ready for distribution!"

# Create distribution DMG
dist-macos: notarize-macos
    #!/usr/bin/env bash
    set -e
    APP_PATH=$(find ~/Library/Developer/Xcode/DerivedData/Quantlete-*/Build/Products/Debug -name "Quantlete.app" -type d 2>/dev/null | head -1)
    mkdir -p dist
    rm -f dist/Quantlete.dmg

    echo "Creating DMG..."
    hdiutil create -volname Quantlete -srcfolder "$APP_PATH" -ov -format UDZO dist/Quantlete.dmg

    echo "Signing DMG..."
    codesign --force --sign "{{MACOS_SIGN_IDENTITY}}" --timestamp dist/Quantlete.dmg

    echo "Notarizing DMG..."
    xcrun notarytool submit dist/Quantlete.dmg \
        --keychain-profile "{{MACOS_KEYCHAIN_PROFILE}}" \
        --wait
    xcrun stapler staple dist/Quantlete.dmg

    echo "Creating checksum..."
    cd dist && shasum -a 256 Quantlete.dmg > Quantlete.dmg.sha256

    echo "Distribution ready: dist/Quantlete.dmg"
