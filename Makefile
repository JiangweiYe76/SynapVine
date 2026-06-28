.PHONY: help dev dev-down

help:
	@echo "AI-Graph Development"
	@echo ""
	@echo "  make dev         Bring up Neo4j + MySQL + core + portal + console + both frontends"
	@echo "  make dev-down    Stop everything started by any of the above"

CORE_PORT        ?= 8001
CONSOLE_PORT     ?= 8002
PORTAL_PORT      ?= 8000
DISCOVERY_PORT   ?= 8003
CONSOLE_FE_PORT  ?= 5174
PORTAL_FE_PORT   ?= 5173
CORE_URL         ?= http://localhost:$(CORE_PORT)
PID_DIR          := .dev-pids
COMPOSE_PROJECT  ?= ai-graph

dev:
	@mkdir -p $(PID_DIR)
	@bash ./scripts/dev-up.sh \
		"all" \
		"$(CORE_PORT)" "$(CONSOLE_PORT)" "$(CONSOLE_FE_PORT)" \
		"$(PORTAL_PORT)" "$(PORTAL_FE_PORT)" \
		"$(DISCOVERY_PORT)" \
		"$(CORE_URL)" "$(PID_DIR)" "$(COMPOSE_PROJECT)"

dev-down:
	@bash ./scripts/dev-down.sh "$(PID_DIR)" "$(COMPOSE_PROJECT)"
