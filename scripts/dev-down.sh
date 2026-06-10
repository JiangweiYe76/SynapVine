#!/usr/bin/env bash
# Tear down the dev stack started by `make dev`.
#
# Args:
#   $1 PID_DIR
#   $2 COMPOSE_PROJECT
set -euo pipefail

PID_DIR="$1"
COMPOSE_PROJECT="$2"

echo "==> Stopping dev processes..."
if [ -d "$PID_DIR" ]; then
  for pidfile in "$PID_DIR"/*.pid; do
    [ -f "$pidfile" ] || continue
    pid=$(cat "$pidfile")
    name=$(basename "$pidfile" .pid)
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      echo "    stopped $name (pid $pid)"
    fi
  done
  rm -rf "$PID_DIR"
fi

echo "==> Stopping Neo4j (docker-compose)..."
(cd services/infra && COMPOSE_PROJECT_NAME="$COMPOSE_PROJECT" docker-compose down) || true

echo "Done."
