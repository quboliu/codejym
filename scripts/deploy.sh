#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-local}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
ADDR="${ADDR:-:8080}"
DATA_DIR="${DATA_DIR:-$ROOT_DIR/data}"
FRONTEND_DIST="${FRONTEND_DIR:-$ROOT_DIR/frontend/dist}"

case "$MODE" in
  local)
    echo ">>> Building frontend assets"
    (cd "$ROOT_DIR/frontend" && npm install && npm run build)

    echo ">>> Building backend binary"
    mkdir -p "$BIN_DIR"
    (cd "$ROOT_DIR/backend" && go build -o "$BIN_DIR/codecopybook" ./cmd/server)

    echo ">>> Ensuring data directory at $DATA_DIR"
    mkdir -p "$DATA_DIR"

    echo ">>> Launching server on $ADDR"
    (cd "$ROOT_DIR" && DATA_DIR="$DATA_DIR" FRONTEND_DIR="$FRONTEND_DIST" "$BIN_DIR/codecopybook" -addr "$ADDR")
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
