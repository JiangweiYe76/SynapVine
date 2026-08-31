# AGENTS.md — SynapVine (AI Knowledge Graph)

Project overview, build commands, and conventions for AI coding agents.

---

## Project Overview

SynapVine is an interactive 3D knowledge graph visualization of AI concepts. Nodes represent AI concepts (Transformer, BERT, GAN, etc.) and edges represent relationships. Communities are managed and detected in the core service; the portal visualizes the community hierarchy returned by core.

The project is organized as a microservice monorepo:

- **`services/portal`** — Public read-only graph API (sources all data from core)
- **`services/core`** — Neo4j data service, migration tooling, community management & detection
- **`services/console`** — Authentication & management console API
- **`clients/portal`** — 3D knowledge graph visualization
- **`clients/console`** — Management console UI

---

## Tech Stack

| Layer | Tech | Version |
|-------|------|---------|
| Backend | Go | 1.26.3 |
| Backend Framework | Fiber | v2.52.13 |
| Backend Graph Lib | gonum | v0.17.0 |
| Graph DB | Neo4j | 5.23 (via `services/infra`) |
| Auth DB | MySQL | 8.0 (via `services/infra`, console service only) |
| Password Hashing | argon2id (`golang.org/x/crypto/argon2`) | — |
| Frontend Framework | Vue | 3.5.32 |
| Frontend Language | TypeScript | ~6.0.2 |
| Build Tool | Vite | 8.0.10 |
| CSS | Tailwind CSS | 4.3.0 |
| 3D Rendering | 3d-force-graph (Three.js) | ^1.80.0 |
| i18n (Portal only) | vue-i18n | ^11.4.2 |
| Icons (Portal) | lucide-vue-next | ^1.0.0 |
| Icons (Console) | @lucide/vue | ^1.17.0 |
| State Management (Console) | pinia | ^3.0.2 |
| Routing (Console) | vue-router | ^4.5.1 |
| Package Manager | Bun | >= 1.0 |

---

## Development Commands

### Backend — Portal

```bash
cd services/portal
go mod tidy
go run main.go
```

- Server starts on `http://localhost:8000`
- Uses `slog` with JSON output
- Loads graph data from the core service (via `CORE_URL`, default
  `http://localhost:8001`). The portal requires core to be reachable;
  there is no local fallback.

### Backend — Core

```bash
cd services/core
go mod tidy
go run main.go
```

- Connects to Neo4j (configure via `NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`)
- Runs migrations via `cd cmd/migrate && go run main.go`
- Seeds data via `cd cmd/seed && go run main.go`

### Backend — Console

```bash
cd services/console
go mod tidy
go run main.go
```

- Server starts on `http://localhost:8002`
- Provides JWT-based authentication: `/api/auth/login`, `/api/auth/refresh`, `/api/auth/logout`, `/api/me`
- Users, refresh tokens, and audit events are persisted in MySQL (see `internal/store`). Migrations are applied automatically on startup; run `go run ./cmd/seed` once to bootstrap the first admin (the dev script `make dev` does this automatically).
- Mutation routes (`POST/PUT/DELETE` on nodes, edges, communities) require `admin` or `editor` role via the `RequireRole` middleware in `internal/handler/rbac.go`. `viewer` can read.
- Passwords are hashed with **argon2id** (see `internal/auth/password.go`).

### Frontend — Portal

```bash
cd clients/portal
bun install

# Dev server with mock API (no Go server needed)
bun run dev

# Production build
bun run build

# Preview production build
bun run preview
```

- Dev server starts on `http://localhost:5173`
- Vite proxies `/api` to `http://localhost:8000`

### Frontend — Console

```bash
cd clients/console
bun install
bun run dev
```

- Dev server starts on `http://localhost:5174`
- Vite proxies `/api` to `http://localhost:8002`

### Infrastructure (Neo4j + MySQL)

```bash
cd services/infra
docker-compose up -d
```

- Neo4j Browser: `http://localhost:7474`

---

## Environment Variables

### `services/portal`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | Server port |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | CORS allowed origin |
| `CORE_URL` | `http://localhost:8001` | Base URL of the core service. The portal requires core to be reachable. |
| `SERVICE_TOKEN` | *(none — required by core)* | The portal's service token, presented to core via the `X-Service-Token` header (read-tier access). |

### `services/core`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | Server port (reserved for future gRPC/HTTP) |
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j username |
| `NEO4J_PASSWORD` | `synapvine123` | Neo4j password |
| `SERVICE_TOKENS` | *(none — required)* | Service-to-service tokens in `portal=<token>,console=<token>,discovery=<token>` format. Core refuses to start without it (fail-closed). |

### `services/console`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8002` | Server port |
| `ALLOWED_ORIGIN` | `http://localhost:5174` | CORS allowed origin |
| `CORE_URL` | *(none — required)* | Base URL of the core service |
| `JWT_SECRET` | *(none — required)* | JWT signing secret |
| `COOKIE_SECURE` | `true` | Whether the refresh-token cookie gets the `Secure` attribute. Set to `false` for dev over plain HTTP (`make dev` does this automatically). |
| `MYSQL_DSN` | *(none — required)* | DSN for the console auth DB (e.g. `synapvine:synapvine123@tcp(localhost:3306)/synapvine_console?parseTime=true`) |
| `ADMIN_USERNAME` | *(required for seed)* | Username of the bootstrap admin created by `cmd/seed` |
| `ADMIN_PASSWORD` | *(required for seed)* | Plaintext password of the bootstrap admin |
| `SERVICE_TOKEN` | *(none — required by core)* | The console's service token, presented to core (write-tier) and discovery via the `X-Service-Token` header. |

The console service is **stateless across restarts** for users: all users, refresh tokens, and audit events live in MySQL. Migrations are applied automatically on startup. The first admin is created by running `cd services/console && go run ./cmd/seed` (the dev script `make dev` does this for you).

### `services/discovery`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8003` | Server port |
| `CORE_URL` | `http://localhost:8001` | Base URL of the core service |
| `ALLOWED_ORIGIN` | `http://localhost:5174` | CORS allowed origin (console frontend) |
| `SERVICE_TOKEN` | *(none — required by core)* | The discovery service's token, presented to core (write + internal tier) via the `X-Service-Token` header. |
| `SERVICE_TOKENS` | *(none — required)* | Tokens accepted on `/api/analyze`, in `console=<token>` format. Discovery refuses to start without it (fail-closed). |

### `clients/portal`

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_USE_MOCK` | `false` | If `true`, all API calls are intercepted by the built-in mock server |

---

## Language Policy

All development artifacts must be written in English:

- **Code comments** — All Go doc comments, inline `//` comments, and Vue/TS JSDoc must be in English.
- **Log messages** — All log output messages and human-readable attribute values must be in English.
- **API error messages** — The `message` field in all HTTP error responses must be in English.
- **Commit messages & PRs** — All Git commit messages and pull request descriptions must be in English.

---

## Code Conventions

### Go (Backend)

- Use `log/slog` with structured logging. Prefer `slog.Info("event_name", slog.String("key", val))` over `fmt.Printf`.
- Group imports: (1) standard library, (2) internal packages, (3) third-party packages. Separate groups with blank lines.
- All exported types/functions must have doc comments.
- Layered architecture: `handler` (HTTP) -> `service` (business logic) -> `coreclient`/`db` (data).
- Error responses follow the JSON shape: `{"error": "error_code", "message": "human readable"}`.
- All `slog` messages and attribute values must be in English.
- All doc comments on exported identifiers must be in English.

### Vue / TypeScript (Frontend)

- Use `<script setup lang="ts">` for all SFCs.
- Use Composition API and `ref`/`computed` from Vue. Avoid Options API.
- Stateful logic belongs in `composables/` (e.g., `useGraph.ts`, `useTheme.ts`). Console uses `stores/` (Pinia) for auth state.
- Shared types live in `types/graph.ts` (Portal) or `types/auth.ts` (Console). Keep them in sync with Go models.
- API calls are centralized in `api/graph.ts` (Portal) or `api/auth.ts` (Console). Do not call `fetch` directly from components.
- Components use PascalCase file names (e.g., `GraphCanvas.vue`).
- Use Tailwind CSS for styling. CSS variables for theme colors (e.g., `--color-bg-primary`). Scoped `<style>` blocks are reserved for transitions/animations only.
- All code comments, JSDoc, and inline annotations must be in English.

### Internationalization (i18n)

- **Portal only**: All user-visible strings must go through `vue-i18n`. Use `t('key')` from `useI18n()`.
- **Portal only**: `en-US.json` is the **source of truth** for development. All new UI text keys and English values are added there first.
- **Portal only**: `zh-CN.json` is maintained as a translation layer. Do not write Chinese text directly into components or code.
- Console currently does not use `vue-i18n`; all UI strings are English hardcoded.

---

## Architecture Decisions

### Mock-First Development (Portal)

The Portal frontend can run entirely without the Go backend. When `VITE_USE_MOCK=true`, `api/graph.ts` routes all calls to `mock/server.ts` using `mock/data.ts`. This enables offline frontend development. **Do not** let mock logic leak into the production API layer.

The Console frontend does not currently have a mock mode.

### Security Layers (Portal)

The Portal API has the following defenses for `/api/graph/*` endpoints:

1. Rate limiting (60 req/min per IP)
2. CORS with origin/referer validation
3. Token-based auth (temporary tokens, 5 min TTL)

HMAC request signing and AES-GCM response encryption modules exist in `services/portal/internal/security/` but are **not currently enabled** on any route. If you enable them, you **must** update the frontend `api/graph.ts` signing/encryption protocol to stay compatible.

### Community Detection

- Communities are managed through `services/core` (create, update, delete, assign nodes).
- Louvain detection runs in `services/core` via `POST /api/communities/detect`.
- `services/portal` loads the community tree from core at startup and holds it in memory (static data).

### Data Model

- The core service (which fronts Neo4j) is the authoritative source of
  truth for nodes, edges, and statistics. Portal and console both proxy
  through core.
- If you change the JSON schema, update all of the following:
  - `services/portal/internal/model/graph.go`
  - `clients/portal/src/types/graph.ts`
  - `clients/portal/src/mock/data.ts`

---

## API Endpoints

For the authoritative, up-to-date list of HTTP endpoints, run the services
and inspect the live routes (each Fiber app prints them at startup). Inline
endpoint tables are intentionally not maintained in this document.

---

## Testing

- **Backend:** Use standard `go test ./...` from the relevant `services/<name>/` directory. No test framework beyond the Go standard library is currently configured.
- **Frontend:** No test runner is configured yet. If adding tests, prefer **Vitest** (aligns with Vite). Run via `bun test`.
- Always verify the dev server (`bun run dev`) and the production build (`bun run build && bun run preview`) both work after significant changes.
- If you modify the security code or the community data flow between core and portal, manually verify the Portal server starts correctly and the API returns expected data.

---

## File Boundaries & "Do Not Touch" Zones

| Zone | Rule |
|------|------|
| `clients/portal/src/mock/` | Dev-only. Never import mock data into production components or API layer. |
| `services/portal/internal/security/` | Crypto/signature logic. Changes require coordinated frontend updates. |
| `clients/portal/src/locales/` | Always keep `en-US.json` and `zh-CN.json` in sync. Do not hardcode strings in components. |

---

## Commit & PR Guidelines

- **NEVER** run `git commit` or `git push` unless the user explicitly asks you to. Making code changes does NOT mean you should commit them. Always wait for the user to review and explicitly request a commit or push.
- Keep commits focused. A single commit should address one concern (feature, fix, or refactor).
- If you modify both Go and Vue code, mention both in the commit message (e.g., `feat: add node filtering to API and UI`).
- Run `go mod tidy` in the affected `services/<name>/` directory before committing Go changes.
- Run `bun run build` in the affected `clients/<name>/` directory before committing frontend changes to ensure TypeScript compiles cleanly.
- All commit messages and PR descriptions must be written in English.
