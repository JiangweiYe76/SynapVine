#!/usr/bin/env bash
# Bring up the dev stack.
#
# Args (passed by the Makefile):
#   $1 STACK           "console" | "portal" | "all"
#   $2 CORE_PORT
#   $3 CONSOLE_PORT
#   $4 CONSOLE_FE_PORT
#   $5 PORTAL_PORT
#   $6 PORTAL_FE_PORT
#   $7 CORE_URL
#   $8 PID_DIR
#   $9 COMPOSE_PROJECT
set -euo pipefail

STACK="$1"
CORE_PORT="$2"
CONSOLE_PORT="$3"
CONSOLE_FE_PORT="$4"
PORTAL_PORT="$5"
PORTAL_FE_PORT="$6"
CORE_URL="$7"
PID_DIR="$8"
COMPOSE_PROJECT="$9"

mkdir -p "$PID_DIR"

cleanup() {
  echo
  echo "==> Shutting down..."
  for pidfile in "$PID_DIR"/*.pid; do
    [ -f "$pidfile" ] || continue
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  done
  (cd services/infra && COMPOSE_PROJECT_NAME="$COMPOSE_PROJECT" docker-compose down) || true
  rm -rf "$PID_DIR"
  echo "Done."
}
trap cleanup EXIT INT TERM

start_backend() {
  # Args: name dir port [KEY=VAL ...]
  # KEY=VAL args are env-var assignments. We pass them through env(1) so
  # that bash's quoting does not mangle them.
  local name="$1" dir="$2" port="$3"
  shift 3
  echo "==> Starting $name backend on :$port..."
  ( cd "$dir" && env PORT="$port" "$@" go run main.go ) &
  echo $! > "$PID_DIR/$name.pid"
}

start_frontend() {
  local name="$1" dir="$2" port="$3"
  echo "==> Starting $name frontend on :$port..."
  (cd "$dir" && bun run dev) &
  echo $! > "$PID_DIR/$name-fe.pid"
}

wait_for() {
  local url="$1" label="$2" max="${3:-30}"
  echo -n "    Waiting for $url"
  for i in $(seq 1 "$max"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo " $label"; return 0
    fi
    echo -n "."
    sleep 1
    if [ "$i" -eq "$max" ]; then echo " TIMEOUT"; return 1; fi
  done
}

# 1. Neo4j
echo "==> Starting Neo4j (docker-compose)..."
(cd services/infra && COMPOSE_PROJECT_NAME="$COMPOSE_PROJECT" docker-compose up -d)
echo -n "    Waiting for Neo4j to become healthy"
for i in $(seq 1 90); do
  status=$(docker inspect -f '{{.State.Health.Status}}' synapvine-neo4j 2>/dev/null || echo "starting")
  if [ "$status" = "healthy" ]; then echo " healthy"; break; fi
  echo -n "."
  sleep 1
  if [ "$i" -eq 90 ]; then echo " TIMEOUT (status=$status)"; exit 1; fi
done

# 2. Core (always started; both portals and consoles need it)
start_backend core services/core "$CORE_PORT" PORT="$CORE_PORT"
wait_for "$CORE_URL/health" "healthy" 30 || exit 1

# 3. Frontends and (optional) extra backends per stack
need_console=false
need_portal=false
case "$STACK" in
  console) need_console=true ;;
  portal)  need_portal=true ;;
  all)     need_console=true; need_portal=true ;;
  *) echo "Unknown STACK: $STACK" >&2; exit 1 ;;
esac

if $need_console; then
  start_backend console services/console "$CONSOLE_PORT" CORE_URL="$CORE_URL" PORT="$CONSOLE_PORT"
  start_frontend console-fe clients/console "$CONSOLE_FE_PORT"
fi

if $need_portal; then
  start_backend portal services/portal "$PORTAL_PORT" CORE_URL="$CORE_URL" PORT="$PORTAL_PORT"
  start_frontend portal-fe clients/portal "$PORTAL_FE_PORT"
fi

echo
echo "Stack is up:"
echo "  Neo4j           bolt://localhost:7687  browser: http://localhost:7474"
echo "  Core            $CORE_URL"
if $need_console; then
echo "  Console API     http://localhost:$CONSOLE_PORT"
echo "  Console UI      http://localhost:$CONSOLE_FE_PORT"
fi
if $need_portal; then
echo "  Portal API      http://localhost:$PORTAL_PORT"
echo "  Portal UI       http://localhost:$PORTAL_FE_PORT"
fi
echo
echo "Press Ctrl+C to stop everything."

wait
