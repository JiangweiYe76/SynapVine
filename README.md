# SynapVine — AI Knowledge Graph

An interactive 3D knowledge graph visualization of AI concepts, with a management console for graph operations. Nodes represent AI concepts (Transformer, BERT, GAN, etc.) and edges represent relationships between them. Communities are managed and detected in the core service; the portal visualizes the community hierarchy returned by core.

***

## Features

### 3D Force-Directed Graph (Portal)

- WebGL-rendered 3D graph with force-directed layout
- Node size reflects influence score, edge width reflects relationship weight
- Drag, pan, zoom, and rotate the canvas

### Community Visualization

- Hierarchical communities loaded from the core service
- Multi-level community trees with color-coded groups
- Interactive legend panel for browsing communities

### Search & Navigation

- Search bar with real-time fuzzy matching on node names and descriptions
- Select a result to fly the camera to the target node
- Highlight the target node and its 1-hop neighbors

### Node Details

- Click any node to open a detail panel
- View name, community, influence score, description, and connected neighbors
- Click a neighbor to navigate to it

### Timeline Mode

- Watch the evolution of AI concepts from 1956 to present
- Play, pause, and scrub through history with a timeline slider
- Speed control (0.5x, 1x, 2x, 5x)
- Milestone event popups at key years (e.g., 2017 Transformer paper, 2022 ChatGPT)

### Management Console

- Dashboard with graph statistics
- Node and edge management interfaces
- JWT-based authentication

***

## Architecture

```
                              ┌──────────────────────┐
                              │      clients/        │
                              │  ┌───────┐ ┌───────┐ │
                              │  │portal │ │console│ │
                              │  └───────┘ └───────┘ │
                              └──────────┬─────┬─────┘
                                         │     │
                              ┌──────────┘     └──────────┐
                              │                           │
                              ▼                           ▼
                    ┌─────────────────┐        ┌─────────────────┐
                    │  services/      │        │  services/      │
                    │  portal         │        │  console        │
                    │                 │        │                 │
                    │  Public read-   │        │  Auth + console │
                    │  only graph     │        │  management     │
                    └────────┬────────┘        └────────┬────────┘
                             │                          │
                             └────────────┬─────────────┘
                                          │
                                          ▼
                               ┌──────────────────────┐
                               │        core          │
                               │   (internal)         │
                               │   Neo4j CRUD         │
                               │   Community CRUD     │
                               │   Community detect   │
                               │   Review queue       │
                               └──────────────────────┘
                                          │
                                          ▼
                               ┌──────────────────────┐
                               │     discovery        │
                               │   (internal)         │
                               │   arXiv / social     │
                               │   LLM pipeline       │
                               └──────────────────────┘
```

***


## Getting Started

### Prerequisites

- **Go** >= 1.26
- **Bun** >= 1.0 (or npm)
- **Node.js** >= 18
- **Docker** & **Docker Compose** (for Neo4j)

### Infrastructure

```bash
cd services/infra
docker-compose up -d
```

- Neo4j Browser: `http://localhost:7474`
- MySQL: `localhost:3306` (console auth DB)

### Backend — Portal

```bash
cd services/portal
go mod tidy
go run main.go
```

**Environment variables:**

| Variable         | Default                   | Description                       |
| ---------------- | ------------------------- | --------------------------------- |
| `PORT`           | `8000`                    | Server port                       |
| `ALLOWED_ORIGIN` | `http://localhost:5173`   | CORS allowed origin               |
| `CORE_URL`       | `http://localhost:8001`   | Core service URL (mandatory)      |

### Backend — Core

```bash
cd services/core
go mod tidy
go run main.go
```

**Environment variables:**

| Variable         | Default                 | Description                |
| ---------------- | ----------------------- | -------------------------- |
| `PORT`           | `8001`                  | Server port                |
| `NEO4J_URI`      | `bolt://localhost:7687` | Neo4j Bolt URI             |
| `NEO4J_USER`     | `neo4j`                 | Neo4j username             |
| `NEO4J_PASSWORD` | `synapvine123`          | Neo4j password             |

### Frontend — Portal

```bash
cd clients/portal
bun install
bun run dev
```

Dev server proxies `/api` to Portal backend.

### Frontend — Console

```bash
cd clients/console
bun install
bun run dev
```

Dev server proxies `/api` to Console backend.

### Backend — Console

```bash
cd services/console
go mod tidy
go run main.go
```

**Environment variables:**

| Variable         | Default                 | Description                       |
| ---------------- | ----------------------- | --------------------------------- |
| `PORT`           | `8002`                  | Server port                       |
| `ALLOWED_ORIGIN` | `http://localhost:5174` | CORS allowed origin               |
| `CORE_URL`       | *(none — required)*     | Base URL of the core service      |
| `JWT_SECRET`     | *(none — required)*     | JWT signing secret                |
| `MYSQL_DSN`      | *(none — required)*     | DSN for the console auth DB       |

***

## API Endpoints

### Portal (Public Graph API)

| Method | Path                         | Auth | Description                                 |
| ------ | ---------------------------- | ---- | ------------------------------------------- |
| GET    | `/api/token`                 | No   | Obtain a temporary access token             |
| GET    | `/api/graph/summary`         | Yes  | Graph overview (communities, stats, top 20) |
| GET    | `/api/graph/nodes`           | Yes  | Paginated node list (with filters)          |
| GET    | `/api/graph/nodes/:id`       | Yes  | Node detail + neighbors                     |
| GET    | `/api/graph/nodes/:id/edges` | Yes  | Edges connected to a node                   |
| GET    | `/api/graph/search`          | Yes  | Search nodes by name/description            |
| GET    | `/api/graph/expand`          | Yes  | Expand graph by loading neighbors & edges   |
| GET    | `/api/graph/timeline`        | Yes  | Min/max year range of `first_appeared`      |

All `/api/graph/*` endpoints require a valid token (`?token=xxx`).

### Console (Auth)

| Method | Path              | Auth | Description          |
| ------ | ----------------- | ---- | -------------------- |
| POST   | `/api/auth/login` | No   | User login           |
| GET    | `/api/me`         | JWT  | Current user profile |

***

## Tech Stack

| Layer | Tech | Version |
|-------|------|---------|
| Portal Frontend | Vue + TypeScript | 3.5.32 |
| Console Frontend | Vue + TypeScript | 3.5.32 |
| Build Tool | Vite | 8.0.10 |
| CSS | Tailwind CSS | 4.3.0 |
| 3D Rendering | 3d-force-graph (Three.js) | ^1.80.0 |
| i18n | vue-i18n | ^11.4.2 |
| Portal API | Go + Fiber | 1.26.3 / v2.52.13 |
| Auth / Core | Go + Fiber | 1.26.3 / v2.52.13 |
| Graph DB | Neo4j | 5.23 |
| Graph Lib | gonum | v0.17.0 |
| Package Manager | Bun | >= 1.0 |

***

## Roadmap

### Phase 1 — Static Graph

- [x] Graph JSON data source
- [x] 3D force-directed graph rendering (Portal)
- [x] Community management and detection in core
- [x] Community visualization in portal
- [x] Search, timeline, dark theme
- [x] Management console (Dashboard, login)
- [x] Neo4j integration (core)
- [x] Node/edge CRUD in console
- [ ] Review queue for discovery pipeline

### Phase 2 — Dynamic Graph

- [x] `services/console` (auth + console management)
- [ ] `services/discovery` (TypeScript) — arXiv / social media ingestion
- [ ] LLM pipeline for node/edge generation
- [ ] Automated community re-detection
- [ ] Live data freshness status

***

## Security (Portal Layer)

- Rate limiting (60 req/min per IP)
- Token-based authentication (temporary tokens, 5 min TTL)
- HMAC request signing with nonce replay protection (module exists but not enabled)
- AES-GCM response encryption (module exists but not enabled)
- CORS with origin/referer validation

***

## License

Apache License, Version 2.0
