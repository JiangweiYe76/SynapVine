# AGENTS.md — SynapVine (AI Knowledge Graph)

Project overview, build commands, and conventions for AI coding agents.

---

## Project Overview

SynapVine is an interactive 3D knowledge graph visualization of AI concepts. Nodes represent AI concepts (Transformer, BERT, GAN, etc.) and edges represent relationships. Communities are detected automatically via the Louvain algorithm with multi-level hierarchical grouping.

The project is organized as a microservice monorepo:

- **`services/portal`** — Public read-only graph API
- **`services/core`** — Neo4j data service & migration tooling
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
- Loads static data from `DATA_PATH` (default `../data/graph.json`)

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

- Server starts on `http://localhost:8001`
- Provides JWT-based authentication (`/api/auth/login`, `/api/me`)

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
- Vite proxies `/api` to `http://localhost:8001`

### Infrastructure (Neo4j)

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
| `DATA_PATH` | `../data/graph.json` | Static graph data file path |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | CORS allowed origin |

### `services/core`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | Server port (reserved for future gRPC/HTTP) |
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j username |
| `NEO4J_PASSWORD` | `synapvine123` | Neo4j password |

### `services/console`

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | Server port |
| `ALLOWED_ORIGIN` | `http://localhost:5174` | CORS allowed origin |
| `JWT_SECRET` | *(dev key)* | JWT signing secret |

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
- Layered architecture: `handler` (HTTP) -> `service` (business logic) -> `community`/`loader`/`db` (data).
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

- Flat communities: Louvain algorithm via `gonum/graph`.
- Hierarchical communities: recursive partitioning up to 3 levels (`MaxLevels: 3`, `MinCommunitySize: 3`).
- Computed at `services/portal` startup and held in memory (static data).

### Data Model

- Current source of truth for Portal: `data/graph.json` (or path specified by `DATA_PATH`).
- Core service targets Neo4j as the long-term source of truth (work in progress).
- If you change the JSON schema, update all of the following:
  - `services/portal/internal/model/graph.go`
  - `clients/portal/src/types/graph.ts`
  - `clients/portal/src/mock/data.ts`

---

## API Endpoints

### Portal (Public Graph API)

| Method | Path | Auth Required | Description |
|--------|------|---------------|-------------|
| GET | `/api/token` | No | Obtain a temporary access token |
| GET | `/api/graph/summary` | Yes | Communities, stats, top 20 nodes |
| GET | `/api/graph/nodes` | Yes | Paginated node list with filters |
| GET | `/api/graph/nodes/:id` | Yes | Node detail + neighbors |
| GET | `/api/graph/nodes/:id/edges` | Yes | Edges connected to a node |
| GET | `/api/graph/search` | Yes | Fuzzy search by name/description |
| GET | `/api/graph/expand` | Yes | Expand graph by loading neighbors |

All `/api/graph/*` endpoints require a valid token (`?token=xxx`).

### Console (Auth)

| Method | Path | Auth Required | Description |
|--------|------|---------------|-------------|
| POST | `/api/auth/login` | No | User login |
| GET | `/api/me` | JWT | Current user profile |

---

## Testing

- **Backend:** Use standard `go test ./...` from the relevant `services/<name>/` directory. No test framework beyond the Go standard library is currently configured.
- **Frontend:** No test runner is configured yet. If adding tests, prefer **Vitest** (aligns with Vite). Run via `bun test`.
- Always verify the dev server (`bun run dev`) and the production build (`bun run build && bun run preview`) both work after significant changes.
- If you modify the security or community detection code, manually verify the Portal server starts correctly and the API returns expected data.

---

## File Boundaries & "Do Not Touch" Zones

| Zone | Rule |
|------|------|
| `clients/portal/src/mock/` | Dev-only. Never import mock data into production components or API layer. |
| `services/portal/internal/security/` | Crypto/signature logic. Changes require coordinated frontend updates. |
| `data/graph.json` | Static data source. Schema changes must be reflected in Go models, TS types, and mock data. |
| `clients/portal/src/locales/` | Always keep `en-US.json` and `zh-CN.json` in sync. Do not hardcode strings in components. |

---

## Commit & PR Guidelines

- Keep commits focused. A single commit should address one concern (feature, fix, or refactor).
- If you modify both Go and Vue code, mention both in the commit message (e.g., `feat: add node filtering to API and UI`).
- Run `go mod tidy` in the affected `services/<name>/` directory before committing Go changes.
- Run `bun run build` in the affected `clients/<name>/` directory before committing frontend changes to ensure TypeScript compiles cleanly.
- All commit messages and PR descriptions must be written in English.
