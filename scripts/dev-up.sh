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
#   $7 DISCOVERY_PORT
#   $8 CORE_URL
#   $9 PID_DIR
#   $10 COMPOSE_PROJECT
set -euo pipefail

STACK="$1"
CORE_PORT="$2"
CONSOLE_PORT="$3"
CONSOLE_FE_PORT="$4"
PORTAL_PORT="$5"
PORTAL_FE_PORT="$6"
DISCOVERY_PORT="$7"
CORE_URL="$8"
PID_DIR="$9"
COMPOSE_PROJECT="${10}"

mkdir -p "$PID_DIR"

cleanup() {
  echo
  echo "==> Shutting down dev processes... (Neo4j + MySQL left running; use make dev-down to stop them)"
  for pidfile in "$PID_DIR"/*.pid; do
    [ -f "$pidfile" ] || continue
    pid=$(cat "$pidfile")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  done
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

start_seed() {
  # Args: name dir [KEY=VAL ...]
  local name="$1" dir="$2"
  shift 2
  echo "==> Running $name..."
  ( cd "$dir" && env "$@" go run ./cmd/seed ) &
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

# 1. Neo4j + MySQL
echo "==> Starting Neo4j + MySQL (docker-compose)..."
(cd services/infra && COMPOSE_PROJECT_NAME="$COMPOSE_PROJECT" docker-compose up -d)
echo -n "    Waiting for Neo4j to become healthy"
for i in $(seq 1 90); do
  status=$(docker inspect -f '{{.State.Health.Status}}' synapvine-neo4j 2>/dev/null || echo "starting")
  if [ "$status" = "healthy" ]; then echo " healthy"; break; fi
  echo -n "."
  sleep 1
  if [ "$i" -eq 90 ]; then echo " TIMEOUT (status=$status)"; exit 1; fi
done
echo -n "    Waiting for MySQL to become healthy"
for i in $(seq 1 90); do
  status=$(docker inspect -f '{{.State.Health.Status}}' synapvine-mysql 2>/dev/null || echo "starting")
  if [ "$status" = "healthy" ]; then echo " healthy"; break; fi
  echo -n "."
  sleep 1
  if [ "$i" -eq 90 ]; then echo " TIMEOUT (status=$status)"; exit 1; fi
done

# 2. Core (always started; both portals and consoles need it)
start_backend core services/core "$CORE_PORT" \
  PORT="$CORE_PORT" \
  MYSQL_DSN="synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true"
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
  # Seed the console MySQL with the dev admin (admin / admin123) the
  # first time the stack comes up. The seed tool is idempotent and
  # exits 0 when users already exist, so re-running is safe.
  #
  # Important: `wait` without an argument waits for *every* background
  # job, including the long-running core backend started above, which
  # would hang the script forever. We wait only for the seed (its PID
  # is $!) and then drop the seed's pidfile so cleanup doesn't try to
  # kill a stale process later.
  start_seed console-seed services/console \
    MYSQL_DSN="synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true" \
    ADMIN_USERNAME="admin" \
    ADMIN_PASSWORD="admin123"
  wait $!
  rm -f "$PID_DIR/console-seed.pid"
  start_backend console services/console "$CONSOLE_PORT" \
    CORE_URL="$CORE_URL" \
    PORT="$CONSOLE_PORT" \
    MYSQL_DSN="synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true" \
    JWT_SECRET="console-dev-secret-key-change-in-production"
  start_frontend console-fe clients/console "$CONSOLE_FE_PORT"
  # Discovery depends on core (papers, review queue) and console (LLM
  # provider config), so it starts after both are up. It only ships
  # with the console stack; the portal-only stack does not start it.
  start_backend discovery services/discovery "$DISCOVERY_PORT" \
    PORT="$DISCOVERY_PORT" \
    CORE_URL="$CORE_URL" \
    CONSOLE_URL="http://localhost:$CONSOLE_PORT"
  wait_for "http://localhost:$DISCOVERY_PORT/health" "healthy" 30 || exit 1
fi

if $need_portal; then
  start_backend portal services/portal "$PORTAL_PORT" CORE_URL="$CORE_URL" PORT="$PORTAL_PORT"
  start_frontend portal-fe clients/portal "$PORTAL_FE_PORT"
fi

echo
echo "Stack is up:"
echo "  Neo4j           bolt://localhost:7687  browser: http://localhost:7474"
echo "  MySQL           localhost:3306  (db: synapvine_console, user: synapvine)"
echo "  Core            $CORE_URL"
if $need_console; then
echo "  Console API     http://localhost:$CONSOLE_PORT  (dev login: admin / admin123)"
echo "  Console UI      http://localhost:$CONSOLE_FE_PORT"
echo "  Discovery API   http://localhost:$DISCOVERY_PORT  (POST /api/analyze)"
fi
if $need_portal; then
echo "  Portal API      http://localhost:$PORTAL_PORT"
echo "  Portal UI       http://localhost:$PORTAL_FE_PORT"
fi
echo
echo "Press Ctrl+C to stop everything."

wait
