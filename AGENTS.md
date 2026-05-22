# AGENTS.md — SynapVine (AI Knowledge Graph)

Project overview, build commands, and conventions for AI coding agents.

---

## Project Overview

SynapVine is an interactive 3D knowledge graph visualization of AI concepts. Nodes represent AI concepts (Transformer, BERT, GAN, etc.) and edges represent relationships. Communities are detected automatically via the Louvain algorithm with multi-level hierarchical grouping.

---

## Tech Stack

| Layer | Tech | Version |
|-------|------|---------|
| Backend | Go | 1.26.3 |
| Backend Framework | Fiber | v2.52.13 |
| Backend Graph Lib | gonum | v0.17.0 |
| Frontend Framework | Vue | 3.5.32 |
| Frontend Language | TypeScript | ~6.0.2 |
| Build Tool | Vite | 8.0.10 |
| CSS | Tailwind CSS | 4.3.0 |
| 3D Rendering | 3d-force-graph (Three.js) | ^1.80.0 |
| i18n | vue-i18n | ^11.4.2 |
| Icons | lucide-vue-next | ^1.0.0 |
| Package Manager | Bun | >= 1.0 |

---

## Development Commands

### Backend

```bash
cd server
go mod tidy
go run main.go
```

- Server starts on `http://localhost:8000`
- Uses `slog` with JSON output

### Frontend

```bash
cd web
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

---

## Environment Variables

### Backend (`server/`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | Server port |
| `DATA_PATH` | `../data/graph.json` | Static graph data file path |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | CORS allowed origin |

### Frontend (`web/`)

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_USE_MOCK` | `false` | If `true`, all API calls are intercepted by the built-in mock server |

---

## Code Conventions

### Go (Backend)

- Use `log/slog` with structured logging. Prefer `slog.Info("event_name", slog.String("key", val))` over `fmt.Printf`.
- Group imports: (1) standard library, (2) internal packages, (3) third-party packages. Separate groups with blank lines.
- All exported types/functions must have doc comments.
- Layered architecture: `handler` (HTTP) -> `service` (business logic) -> `community`/`loader` (data).
- Error responses follow the JSON shape: `{"error": "error_code", "message": "human readable"}`.

### Vue / TypeScript (Frontend)

- Use `<script setup lang="ts">` for all SFCs.
- Use Composition API and `ref`/`computed` from Vue. Avoid Options API.
- Stateful logic belongs in `composables/` (e.g., `useGraph.ts`, `useTheme.ts`).
- Shared types live in `types/graph.ts`. Keep them in sync with Go `internal/model/graph.go`.
- API calls are centralized in `api/graph.ts`. Do not call `fetch` directly from components.
- Components use PascalCase file names (e.g., `GraphCanvas.vue`).
- Use Tailwind CSS for styling. CSS variables for theme colors (e.g., `--color-bg-primary`). Scoped `<style>` blocks are reserved for transitions/animations only.

### Internationalization (i18n)

- All user-visible strings must go through `vue-i18n`. Use `t('key')` from `useI18n()`.
- Translation files: `web/src/locales/en-US.json`, `web/src/locales/zh-CN.json`.
- When adding new UI text, add keys to **both** locale files.

---

## Architecture Decisions

### Mock-First Development

The frontend can run entirely without the Go backend. When `VITE_USE_MOCK=true`, `api/graph.ts` routes all calls to `mock/server.ts` using `mock/data.ts`. This enables offline frontend development. **Do not** let mock logic leak into the production API layer.

### Security Layers (Backend)

The API has a 5-layer defense stack. All `/api/graph/*` endpoints require a valid token:

1. Rate limiting (60 req/min per IP)
2. CORS with origin/referer validation
3. Token-based auth (temporary tokens, 5 min TTL)
4. HMAC request signing with nonce replay protection
5. AES-GCM response encryption

If you modify anything in `server/internal/security/`, you **must** ensure the frontend `api/graph.ts` signing/encryption protocol stays compatible.

### Community Detection

- Flat communities: Louvain algorithm via `gonum/graph`.
- Hierarchical communities: recursive partitioning up to 3 levels (`MaxLevels: 3`, `MinCommunitySize: 3`).
- Computed at server startup and held in memory (static data).

### Data Model

- Current source of truth: `data/graph.json` (or path specified by `DATA_PATH`).
- If you change the JSON schema, update both:
  - `server/internal/model/graph.go`
  - `web/src/types/graph.ts`
  - `web/src/mock/data.ts`

---

## API Endpoints

| Method | Path | Auth Required | Description |
|--------|------|---------------|-------------|
| GET | `/api/token` | No | Obtain a temporary access token |
| GET | `/api/graph/summary` | Yes | Communities, stats, top 20 nodes |
| GET | `/api/graph/nodes` | Yes | Paginated node list with filters |
| GET | `/api/graph/nodes/:id` | Yes | Node detail + neighbors |
| GET | `/api/graph/nodes/:id/edges` | Yes | Edges connected to a node |
| GET | `/api/graph/search` | Yes | Fuzzy search by name/description |
| GET | `/api/graph/expand` | Yes | Expand graph by loading neighbors |

---

## Testing

- **Backend:** Use standard `go test ./...` from the `server/` directory. No test framework beyond the Go standard library is currently configured.
- **Frontend:** No test runner is configured yet. If adding tests, prefer **Vitest** (aligns with Vite). Run via `bun test`.
- Always verify the dev server (`bun run dev`) and the production build (`bun run build && bun run preview`) both work after significant changes.
- If you modify the security or community detection code, manually verify the server starts correctly and the API returns expected data.

---

## File Boundaries & "Do Not Touch" Zones

| Zone | Rule |
|------|------|
| `web/src/mock/` | Dev-only. Never import mock data into production components or API layer. |
| `server/internal/security/` | Crypto/signature logic. Changes require coordinated frontend updates. |
| `data/graph.json` | Static data source. Schema changes must be reflected in Go models, TS types, and mock data. |
| `web/src/locales/` | Always keep `en-US.json` and `zh-CN.json` in sync. Do not hardcode strings in components. |

---

## Commit & PR Guidelines

- Keep commits focused. A single commit should address one concern (feature, fix, or refactor).
- If you modify both Go and Vue code, mention both in the commit message (e.g., `feat: add node filtering to API and UI`).
- Run `go mod tidy` in `server/` before committing Go changes.
- Run `bun run build` in `web/` before committing frontend changes to ensure TypeScript compiles cleanly.
