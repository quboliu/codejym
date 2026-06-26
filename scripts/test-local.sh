#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-gate}"

log() {
  printf '\n==> %s\n' "$*"
}

ensure_frontend_deps() {
  if [[ "${CI:-}" == "true" || ! -d "$ROOT_DIR/frontend/node_modules" ]]; then
    log "Installing frontend dependencies"
    (cd "$ROOT_DIR/frontend" && npm ci)
  fi
}

backend_gate() {
  log "Checking Go formatting"
  local gofmt_out
  gofmt_out="$(find "$ROOT_DIR/backend" -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 gofmt -l)"
  if [[ -n "$gofmt_out" ]]; then
    printf '%s\n' "$gofmt_out"
    printf 'Go files need formatting. Run: gofmt -w <files>\n' >&2
    return 1
  fi

  log "Running Go tests"
  (cd "$ROOT_DIR/backend" && go test ./...)

  log "Running go vet"
  (cd "$ROOT_DIR/backend" && go vet ./...)
}

frontend_gate() {
  ensure_frontend_deps

  log "Running frontend tests"
  (cd "$ROOT_DIR/frontend" && npm test)

  log "Building frontend"
  (cd "$ROOT_DIR/frontend" && npm run build)

  log "Running frontend lint"
  (cd "$ROOT_DIR/frontend" && npm run lint -- --quiet)
}

wait_for_health() {
  local base_url="$1"
  log "Waiting for ${base_url}/healthz"
  for _ in $(seq 1 80); do
    if curl -fsS "${base_url}/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  curl -v "${base_url}/healthz" || true
  return 1
}

smoke_gate() {
  SMOKE_PROJECT="${COMPOSE_PROJECT_NAME:-codejym-smoke}"
  SMOKE_PORT="${PORT:-19080}"
  SMOKE_POSTGRES_PORT="${POSTGRES_PORT:-15433}"
  SMOKE_BASE_URL="${BASE_URL:-http://127.0.0.1:${SMOKE_PORT}}"

  cleanup() {
    if [[ "${KEEP_SERVICES:-0}" != "1" ]]; then
      log "Stopping smoke stack"
      PORT="$SMOKE_PORT" POSTGRES_PORT="$SMOKE_POSTGRES_PORT" docker compose -p "$SMOKE_PROJECT" -f "$ROOT_DIR/config/docker-compose.yml" down -v
    fi
  }
  trap cleanup EXIT

  log "Starting smoke stack on ${SMOKE_BASE_URL}"
  PORT="$SMOKE_PORT" POSTGRES_PORT="$SMOKE_POSTGRES_PORT" docker compose -p "$SMOKE_PROJECT" -f "$ROOT_DIR/config/docker-compose.yml" up -d --build
  wait_for_health "$SMOKE_BASE_URL"

  log "Running deployment smoke test"
  node "$ROOT_DIR/scripts/smoke-deploy.mjs" --base-url "$SMOKE_BASE_URL"
}

case "$MODE" in
  backend)
    backend_gate
    ;;
  frontend)
    frontend_gate
    ;;
  gate)
    backend_gate
    frontend_gate
    ;;
  smoke)
    smoke_gate
    ;;
  release)
    backend_gate
    frontend_gate
    smoke_gate
    ;;
  *)
    printf 'Unknown mode: %s\n' "$MODE" >&2
    printf 'Usage: %s [backend|frontend|gate|smoke|release]\n' "$0" >&2
    exit 2
    ;;
esac
