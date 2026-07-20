#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
  echo "Usage: $0 <test-command> [args...]" >&2
  exit 2
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "${SCRIPT_DIR}/.." && pwd)
cd "${REPO_ROOT}"

echo "Building Go binary..."
mkdir -p bin
go build -o bin/quantlete ./cmd/quantlete

TEST_DATA_DIR=$(mktemp -d "${TMPDIR:-/tmp}/quantlete-e2e.XXXXXX")
SERVER_PID=""

cleanup() {
  local exit_code=$?

  if [ -n "${SERVER_PID}" ]; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  rm -rf -- "${TEST_DATA_DIR}"

  exit "${exit_code}"
}
trap cleanup EXIT

# port is overridable so local runs don't collide with a developer's own
# server instance on 8081 (CI uses the default)
E2E_PORT="${E2E_PORT:-8081}"

echo "Generating demo database..."
QUANTLETE_STORAGE_DATA_DIR="${TEST_DATA_DIR}" QUANTLETE_STORAGE_DB_FILE=test.db \
  ./bin/quantlete demo --activities=50 --months=6 --athlete="E2E Test User"

echo "Starting Go server on port ${E2E_PORT}..."
QUANTLETE_STORAGE_DATA_DIR="${TEST_DATA_DIR}" QUANTLETE_STORAGE_DB_FILE=test.db \
QUANTLETE_SERVER_PORT="${E2E_PORT}" QUANTLETE_SERVER_DEV_MODE=true \
  ./bin/quantlete serve &
SERVER_PID=$!

echo "Waiting for server..."
for attempt in {1..30}; do
  if curl --fail --silent "http://localhost:${E2E_PORT}/api/v1/auth/status" >/dev/null 2>&1; then
    echo "Server ready"
    break
  fi
  if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
    echo "Server exited before becoming ready" >&2
    exit 1
  fi
  if [ "${attempt}" -eq 30 ]; then
    echo "Server failed to start" >&2
    exit 1
  fi
  sleep 1
done

"$@"
