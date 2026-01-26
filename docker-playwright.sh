#!/usr/bin/env bash
set -euo pipefail

# docker-playwright.sh - run Playwright tests in Docker container
# adapted from pondpilot pattern for Quantlete E2E testing

IMAGE="mcr.microsoft.com/playwright:v1.58.0-noble"
MODE="${1:-test}"
PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-4}"
TTY_FLAGS="-i"

if [ -t 1 ]; then
  TTY_FLAGS="-it"
fi

# allow connecting to host services (Go server on 8081)
NETWORK_FLAGS="--add-host=host.docker.internal:host-gateway"

run_in_container() {
  local command="$1"
  docker run --rm --init ${TTY_FLAGS} ${NETWORK_FLAGS} \
    -e PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS}" \
    -e PLAYWRIGHT_TIMEOUT \
    -e CI="${CI:-true}" \
    -e E2E_IN_DOCKER=true \
    -v "${PWD}/web":/work \
    -w /work \
    -u "$(id -u):$(id -g)" \
    "$IMAGE" \
    bash -lc "export HOME=/tmp && corepack yarn install --immutable && corepack yarn ${command}"
}

case "$MODE" in
  test)
    run_in_container "test:e2e"
    ;;
  test-ui)
    # for local debugging with headed browser (requires X11 forwarding)
    run_in_container "test:e2e --headed"
    ;;
  shell)
    docker run --rm ${TTY_FLAGS} ${NETWORK_FLAGS} \
      -v "${PWD}/web":/work \
      -w /work \
      -u "$(id -u):$(id -g)" \
      "$IMAGE" \
      bash
    ;;
  *)
    echo "Usage: $0 {test|test-ui|shell}"
    exit 1
    ;;
esac
