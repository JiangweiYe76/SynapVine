# AI Knowledge Graph

An interactive 3D knowledge graph visualization of AI concepts, built with Vue 3 and Go. Nodes represent AI concepts (Transformer, BERT, GAN, etc.) and edges represent relationships between them. Communities are automatically detected using the Louvain algorithm, with multi-level hierarchical grouping.

***

## Features

### 3D Force-Directed Graph

- WebGL-rendered 3D graph with force-directed layout
- Node size reflects influence score, edge width reflects relationship weight
- Drag, pan, zoom, and rotate the canvas

### Community Detection

- Automatic community detection via Louvain algorithm
- Multi-level hierarchical communities (up to 3 levels)
- Color-coded communities with an interactive legend panel

### Search & Navigation

- Search bar with real-time fuzzy matching on node names and descriptions
- Select a result to fly the camera to the target nodGe
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

### Dark Theme

- Toggle between dark and light themes
- Persistent preference via localStorage

### Security (Anti-Scraping)

- Rate limiting (60 req/min per IP)
- Token-based authentication (temporary tokens, 5 min TTL)
- HMAC request signing with nonce replay protection
- AES-GCM response encryption (DevTools sees only binary)
- CORS with origin/referer validation

***

## Architecture

```
┌──────────────────────────────────────────────────┐
│  Vue 3 + TypeScript + Vite                        │
│  ┌──────────┐ ┌───────────┐ ┌──────────────────┐ │
│  │ SearchBar │ │ GraphCanvas│  │   NodeDetail     │ │
│  └──────────┘ │ (3d-force  │  └──────────────────┘ │
│  ┌──────────┐ │  -graph)  │                       │
│  │Community  │ │           │                       │
│  │ Legend    │ └───────────┘                       │
│  └──────────┘ ┌──────────────┐ ┌───────────────┐  │
│  ┌──────────┐ │ Timeline     │ │  StatusBar     │  │
│  │ Theme    │ │ Control      │ └───────────────┘  │
│  │ Toggle    │ └──────────────┘                   │
│  └──────────┘                                      │
└──────────────────────────────────────────────────┘
                    │
       HTTP/API (token + HMAC + AES)
                    │
┌──────────────────────────────────────────────────┐
│  Go + Fiber (REST API)                            │
│  ┌──────────┐ ┌───────────┐ ┌──────────────────┐ │
│  │ Handler  │ │  Service   │ │  Community        │ │
│  │ (router) │ │ (business) │ │  (Louvain +       │ │
│  └──────────┘ └───────────┘ │   Hierarchical)    │ │
│  ┌──────────┐ ┌───────────┐ └──────────────────┘ │
│  │ Security │ │  Loader    │                      │
│  │ (token,  │ │ (graph.json)                      │
│  │  signer, │ └───────────┘                      │
│  │  cipher) │                                      │
│  └──────────┘                                      │
└──────────────────────────────────────────────────┘
                    │
            data/graph.json
```

***

## Project Structure

```
AI-Graph/
├── PRD.md                         # Product requirements (Chinese)
├── TechDesign.md                  # Technical design (Chinese)
├── data/
│   └── graph.json                 # Static graph data
├── server/
│   ├── main.go                    # Entry point, Fiber app
│   ├── go.mod / go.sum
│   └── internal/
│       ├── config/config.go       # Configuration
│       ├── model/graph.go         # Data structures
│       ├── loader/loader.go       # JSON data loading
│       ├── community/             # Louvain & hierarchical detection
│       ├── service/graph.go       # Business logic
│       ├── handler/graph.go       # HTTP handlers
│       └── security/              # Token, HMAC, AES, nonce
├── web/
│   ├── package.json / bun.lock
│   ├── vite.config.ts
│   └── src/
│       ├── main.ts                # App entry
│       ├── App.vue                 # Root component
│       ├── api/graph.ts           # API client
│       ├── mock/                  # Mock API for dev
│       ├── types/graph.ts         # TypeScript types
│       ├── composables/           # State management
│       │   ├── useGraph.ts
│       │   ├── useTheme.ts
│       │   └── useTimeline.ts
│       └── components/            # Vue components
│           ├── GraphCanvas.vue    # 3D graph renderer
│           ├── SearchBar.vue
│           ├── CommunityLegend.vue
│           ├── NodeDetail.vue
│           ├── StatusBar.vue
│           └── TimelineControl.vue
```

***

## Getting Started

### Prerequisites

- **Go** ≥ 1.21
- **Bun** ≥ 1.0 (or npm)
- **Node.js** ≥ 18

### Server

```bash
cd server
go mod tidy
go run main.go
```

The server starts on `http://localhost:8000`.

**Environment variables:**

| Variable         | Default                 | Description          |
| ---------------- | ----------------------- | -------------------- |
| `PORT`           | `8000`                  | Server port          |
| `DATA_PATH`      | `../data/graph.json`    | Graph data file path |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | CORS allowed origin  |

### Web

```bash
cd web
bun install

# Mock mode (no server needed)
bun run dev

# Production build
bun run build
bun run preview
```

The dev server starts on `http://localhost:5173`. In mock mode, all API calls are intercepted by the built-in mock server.

***

## API Endpoints

| Method | Path                         | Description                                 |
| ------ | ---------------------------- | ------------------------------------------- |
| GET    | `/api/token`                 | Obtain a temporary access token             |
| GET    | `/api/graph/summary`         | Graph overview (communities, stats, top 20) |
| GET    | `/api/graph/nodes`           | Paginated node list (with filters)          |
| GET    | `/api/graph/nodes/:id`       | Node detail + neighbors                     |
| GET    | `/api/graph/nodes/:id/edges` | Edges connected to a node                   |
| GET    | `/api/graph/search`          | Search nodes by name/description            |
| GET    | `/api/graph/expand`          | Expand graph by loading neighbors & edges   |

All `/api/graph/*` endpoints require a valid token (`?token=xxx`).

***

## Roadmap

### Phase 1 (Current) — Static Graph MVP

- [x] Graph JSON data
- [x] Server API with community detection
- [x] 3D force-directed graph rendering
- [x] Node interaction (hover, click, drag)
- [x] Search & community filtering
- [x] Timeline playback
- [x] Dark theme
- [ ] Edge hover tooltip
- [ ] Data quality review

### Phase 2 (Planned) — Dynamic Data Pipeline

- [ ] Neo4j graph database integration
- [ ] ArXiv paper ingestion pipeline
- [ ] Reddit / Hacker News social media crawler
- [ ] Automated keyword extraction & new concept discovery
- [ ] Scheduled community re-detection
- [ ] Live data freshness status

***

## Technical Highlights

- **Louvain Community Detection**: Pure Go implementation using `gonum/graph`, with recursive hierarchical partitioning for up to 3 levels of communities.
- **WebGL 3D Rendering**: `3d-force-graph` (Three.js) for GPU-accelerated rendering, smooth at thousands of nodes.
- **Defense in Depth**: 5-layer API protection (rate limiting → CORS → token auth → HMAC signing → AES response encryption).
- **Mock-First Development**: Web frontend runs independently with a full mock API server, enabling offline development without the Go server.

***

## License

Apache License, Version 2.0
