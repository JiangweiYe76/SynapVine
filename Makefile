.PHONY: help \
	console-backend console-backend-build \
	console-frontend console-frontend-mock console-frontend-build \
	console-test-backend console-test-frontend \
	console-deps console-deps-backend console-deps-frontend \
	console-clean console-dev

help:
	@echo "AI-Graph Console Development Commands"
	@echo ""
	@echo "Backend (Go):"
	@echo "  console-backend        Run console backend server (port 8001)"
	@echo "  console-backend-build  Build console backend binary"
	@echo "  console-test-backend   Run backend tests"
	@echo "  console-deps-backend   Install/update Go dependencies"
	@echo ""
	@echo "Frontend (Vue 3):"
	@echo "  console-frontend       Run frontend dev server (port 5174)"
	@echo "  console-frontend-mock  Run frontend with mock API"
	@echo "  console-frontend-build Build frontend for production"
	@echo "  console-test-frontend  Run frontend tests (if configured)"
	@echo "  console-deps-frontend  Install npm dependencies"
	@echo ""
	@echo "Combined:"
	@echo "  console-dev            Run both backend and frontend dev servers"
	@echo "  console-deps           Install all dependencies"
	@echo "  console-clean          Clean build artifacts"

console-backend:
	cd services/console && go run main.go

console-backend-build:
	cd services/console && go build -o bin/console-server main.go

console-test-backend:
	cd services/console && go test ./...

console-deps-backend:
	cd services/console && go mod tidy

console-frontend:
	cd clients/console && bun run dev

console-frontend-mock:
	cd clients/console && VITE_USE_MOCK=true bun run dev

console-frontend-build:
	cd clients/console && bun run build

console-test-frontend:
	cd clients/console && bun test

console-deps-frontend:
	cd clients/console && bun install

console-deps: console-deps-backend console-deps-frontend

console-dev:
	@echo "Starting backend (port 8001) and frontend (port 5174)..."
	@cd services/console && go run main.go & \
	cd clients/console && bun run dev & \
	wait

console-clean:
	rm -rf clients/console/dist
	rm -rf clients/console/node_modules/.vite
	rm -rf services/console/bin
