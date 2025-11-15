#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-local}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
ADDR="${ADDR:-:8080}"
DATA_DIR="${DATA_DIR:-$ROOT_DIR/data}"
FRONTEND_DIST="${FRONTEND_DIR:-$ROOT_DIR/frontend/dist}"
DATABASE_URL="${DATABASE_URL:-}"
AUTH_SECRET="${AUTH_SECRET:-}"
POSTGRES_USER="${POSTGRES_USER:-codecopy}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-codecopy}"
POSTGRES_DB="${POSTGRES_DB:-codecopybook}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    head -c 32 /dev/urandom | xxd -p
  fi
}

wait_for_postgres() {
  echo ">>> Waiting for PostgreSQL to become ready..."
  for _ in $(seq 1 30); do
    if docker compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "PostgreSQL did not become ready in time" >&2
  exit 1
}

ensure_local_postgres() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "Docker is required to bootstrap a local PostgreSQL instance. Please install Docker or set DATABASE_URL manually." >&2
    exit 1
  fi
  echo ">>> Starting local PostgreSQL via docker compose"
  (cd "$ROOT_DIR" && docker compose up -d postgres)
  wait_for_postgres
  DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
  echo ">>> DATABASE_URL set to $DATABASE_URL"
}

case "$MODE" in
  local)
    if [[ -z "$DATABASE_URL" ]]; then
      ensure_local_postgres
    fi
    if [[ -z "$AUTH_SECRET" ]]; then
      AUTH_SECRET="$(generate_secret)"
      echo ">>> Generated AUTH_SECRET automatically"
    fi
    echo ">>> Building frontend assets"
    (cd "$ROOT_DIR/frontend" && npm install && npm run build)

    echo ">>> Building backend binary"
    mkdir -p "$BIN_DIR"
    (cd "$ROOT_DIR/backend" && go build -o "$BIN_DIR/codecopybook" ./cmd/server)

    echo ">>> Ensuring data directory at $DATA_DIR"
    mkdir -p "$DATA_DIR"

    echo ">>> Launching server on $ADDR"
    (cd "$ROOT_DIR" && DATA_DIR="$DATA_DIR" FRONTEND_DIR="$FRONTEND_DIST" DATABASE_URL="$DATABASE_URL" AUTH_SECRET="$AUTH_SECRET" "$BIN_DIR/codecopybook" -addr "$ADDR")
    ;;
  docker)
    echo ">>> Building Docker image"
    (cd "$ROOT_DIR" && docker compose build)
    echo ">>> Starting Docker Compose stack"
    (cd "$ROOT_DIR" && docker compose up -d)
    ;;
  *)
    echo "Usage: scripts/deploy.sh [local|docker]" >&2
    exit 1
    ;;
esac
