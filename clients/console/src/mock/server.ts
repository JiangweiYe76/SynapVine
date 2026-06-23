import { mockUser, mockCredentials, mockNodes, mockEdges, mockPapers, mockReviewItems, mockLLMProviders } from './data'
import type {
  LoginRequest,
  LogoutRequest,
  RefreshRequest,
  SessionResponse,
  User,
} from '../types/auth'
import type {
  Node,
  NodesListResponse,
  NodeCreateRequest,
  NodeUpdateRequest,
  Edge,
  EdgesListResponse,
  EdgeCreateRequest,
  EdgeUpdateRequest,
  StatsResponse,
} from '../types/graph'
import type {
  Paper,
  PapersListResponse,
  PaperCreateRequest,
  ReviewQueueItem,
  ReviewQueueListResponse,
} from '../types/paper'
import type {
  LLMProvider,
  LLMProviderCreateRequest,
  LLMProviderListResponse,
  LLMTestResponse,
} from '../types/llm'

// 24h access TTL, 7d refresh TTL — same as the real backend constants.
const ACCESS_TTL_MS = 24 * 60 * 60 * 1000
const REFRESH_TTL_MS = 7 * 24 * 60 * 60 * 1000

interface MockToken {
  value: string
  expiresAt: number
}

interface MockRefreshToken {
  value: string
  expiresAt: number
  revoked: boolean
}

export function createMockServer() {
  let access: MockToken | null = null
  let refresh: MockRefreshToken | null = null
  let nodes: Node[] = [...mockNodes]
  let edges: Edge[] = [...mockEdges]
  let papers: Paper[] = [...mockPapers]
  let reviewItems: ReviewQueueItem[] = [...mockReviewItems]
  let llmProviders: LLMProvider[] = [...mockLLMProviders]

  function newToken(prefix: string): string {
    return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
  }

  function issueSession(user: User): SessionResponse {
    const now = Date.now()
    access = { value: newToken('mock_access'), expiresAt: now + ACCESS_TTL_MS }
    refresh = { value: newToken('mock_refresh'), expiresAt: now + REFRESH_TTL_MS, revoked: false }
    return {
      token: access.value,
      refresh_token: refresh.value,
      expires_at: new Date(access.expiresAt).toISOString(),
      user,
    }
  }

  return {
    login(credentials: LoginRequest): SessionResponse {
      if (
        credentials.username === mockCredentials.username &&
        credentials.password === mockCredentials.password
      ) {
        return issueSession(mockUser)
      }
      throw new Error('Invalid username or password')
    },

    refresh(_req: RefreshRequest): SessionResponse {
      if (!refresh || refresh.revoked || refresh.expiresAt < Date.now()) {
        refresh = null
        access = null
        throw new Error('Refresh token is invalid or has been revoked')
      }
      // Rotate: invalidate the old refresh token, mint a new pair.
      const old = refresh
      const session = issueSession(mockUser)
      old.revoked = true
      return session
    },

    logout(_req: LogoutRequest): void {
      if (refresh) refresh.revoked = true
      access = null
      refresh = null
    },

    me(_authHeader?: string): User {
      if (!access || access.expiresAt < Date.now()) {
        throw new Error('Session expired')
      }
      return mockUser
    },

    listNodes(offset = 0, limit = 20, search = ''): NodesListResponse {
      let filtered = nodes
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter(
          (n) =>
            n.id.toLowerCase().includes(q) ||
            n.name.toLowerCase().includes(q) ||
            n.description.toLowerCase().includes(q),
        )
      }
      const total = filtered.length
      const sliced = filtered.slice(offset, offset + limit)
      return {
        nodes: sliced,
        pagination: {
          offset,
          limit,
          total,
          has_more: offset + limit < total,
        },
      }
    },

    getNode(id: string): Node {
      const node = nodes.find((n) => n.id === id)
      if (!node) throw new Error('Node not found')
      return node
    },

    createNode(data: NodeCreateRequest): Node {
      // Mirror the core service: when no id is supplied the mock mints a
      // fresh one too, so the frontend never has to fabricate an id.
      const id = data.id && data.id.length > 0 ? data.id : cryptoRandomId()
      if (nodes.some((n) => n.id === id)) {
        throw new Error('Node already exists')
      }
      const node: Node = {
        id,
        name: data.name,
        description: data.description,
        influence_score: data.influence_score,
        first_appeared: data.first_appeared,
        milestones: data.milestones,
        community_id: data.community_id ?? null,
      }
      nodes.push(node)
      return node
    },

    updateNode(id: string, data: NodeUpdateRequest): Node {
      const idx = nodes.findIndex((n) => n.id === id)
      if (idx === -1) throw new Error('Node not found')
      nodes[idx] = { ...nodes[idx], ...data }
      return nodes[idx]
    },

    deleteNode(id: string): void {
      const idx = nodes.findIndex((n) => n.id === id)
      if (idx === -1) throw new Error('Node not found')
      nodes.splice(idx, 1)
      // Also delete related edges
      edges = edges.filter((e) => e.source !== id && e.target !== id)
    },

    listEdges(offset = 0, limit = 20, search = ''): EdgesListResponse {
      let filtered = edges
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter(
          (e) =>
            e.source.toLowerCase().includes(q) ||
            e.target.toLowerCase().includes(q) ||
            e.relation.toLowerCase().includes(q),
        )
      }
      const total = filtered.length
      const sliced = filtered.slice(offset, offset + limit)
      return {
        edges: sliced,
        pagination: {
          offset,
          limit,
          total,
          has_more: offset + limit < total,
        },
      }
    },

    getEdge(source: string, target: string): Edge {
      const edge = edges.find((e) => e.source === source && e.target === target)
      if (!edge) throw new Error('Edge not found')
      return edge
    },

    createEdge(data: EdgeCreateRequest): Edge {
      if (!nodes.some((n) => n.id === data.source)) {
        throw new Error('Source node does not exist')
      }
      if (!nodes.some((n) => n.id === data.target)) {
        throw new Error('Target node does not exist')
      }
      if (edges.some((e) => e.source === data.source && e.target === data.target)) {
        throw new Error('Edge already exists')
      }
      const edge: Edge = {
        source: data.source,
        target: data.target,
        weight: data.weight,
        relation: data.relation,
      }
      edges.push(edge)
      return edge
    },

    updateEdge(source: string, target: string, data: EdgeUpdateRequest): Edge {
      const idx = edges.findIndex((e) => e.source === source && e.target === target)
      if (idx === -1) throw new Error('Edge not found')
      edges[idx] = { ...edges[idx], ...data }
      return edges[idx]
    },

    deleteEdge(source: string, target: string): void {
      const idx = edges.findIndex((e) => e.source === source && e.target === target)
      if (idx === -1) throw new Error('Edge not found')
      edges.splice(idx, 1)
    },

    getStats(): StatsResponse {
      let totalInfluence = 0
      for (const n of nodes) {
        totalInfluence += n.influence_score
      }
      return {
        total_nodes: nodes.length,
        total_edges: edges.length,
        avg_influence: nodes.length > 0 ? totalInfluence / nodes.length : 0,
      }
    },

    // --- Papers ---

    listPapers(offset = 0, limit = 20): PapersListResponse {
      const total = papers.length
      return { papers: papers.slice(offset, offset + limit), total }
    },

    getPaper(id: string): Paper {
      const p = papers.find((pp) => pp.id === id)
      if (!p) throw new Error('Paper not found')
      return p
    },

    createPaper(data: PaperCreateRequest): Paper {
      const paper: Paper = {
        id: cryptoRandomId(),
        title: data.title,
        authors: data.authors,
        source_url: data.source_url || '',
        raw_text: data.raw_text,
        status: 'uploaded',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      papers.unshift(paper)
      return paper
    },

    updatePaper(id: string, data: Partial<Paper>): Paper {
      const idx = papers.findIndex((p) => p.id === id)
      if (idx === -1) throw new Error('Paper not found')
      papers[idx] = { ...papers[idx], ...data, updated_at: new Date().toISOString() }
      return papers[idx]
    },

    deletePaper(id: string): void {
      const idx = papers.findIndex((p) => p.id === id)
      if (idx === -1) throw new Error('Paper not found')
      papers.splice(idx, 1)
    },

    // --- Review Queue ---

    listReviewItems(offset = 0, limit = 20, status = ''): ReviewQueueListResponse {
      let filtered = reviewItems
      if (status) filtered = filtered.filter((r) => r.status === status)
      const total = filtered.length
      return { items: filtered.slice(offset, offset + limit), total }
    },

    getReviewItem(id: string): ReviewQueueItem {
      const item = reviewItems.find((r) => r.id === id)
      if (!item) throw new Error('Review item not found')
      return item
    },

    approveReviewItem(id: string, reviewerId: string, notes: string): ReviewQueueItem {
      const idx = reviewItems.findIndex((r) => r.id === id)
      if (idx === -1) throw new Error('Review item not found')
      reviewItems[idx] = {
        ...reviewItems[idx],
        status: 'approved',
        reviewer_id: reviewerId,
        review_notes: notes,
        reviewed_at: new Date().toISOString(),
      }
      return reviewItems[idx]
    },

    rejectReviewItem(id: string, reviewerId: string, notes: string): void {
      const idx = reviewItems.findIndex((r) => r.id === id)
      if (idx === -1) throw new Error('Review item not found')
      reviewItems[idx] = {
        ...reviewItems[idx],
        status: 'rejected',
        reviewer_id: reviewerId,
        review_notes: notes,
        reviewed_at: new Date().toISOString(),
      }
    },

    // --- LLM Providers ---

    listLLMProviders(): LLMProviderListResponse {
      return { providers: [...llmProviders], total: llmProviders.length }
    },

    getLLMProvider(id: string): LLMProvider {
      const p = llmProviders.find((pp) => pp.id === id)
      if (!p) throw new Error('LLM provider not found')
      return p
    },

    getDefaultLLMProvider(): LLMProvider {
      const p = llmProviders.find((pp) => pp.is_default)
      if (!p) throw new Error('No default provider')
      return p
    },

    createLLMProvider(data: LLMProviderCreateRequest): LLMProvider {
      if (data.is_default) llmProviders.forEach((p) => (p.is_default = false))
      const provider: LLMProvider = {
        id: cryptoRandomId(),
        name: data.name,
        base_url: data.base_url,
        model: data.model,
        max_tokens: data.max_tokens || 4096,
        temperature: data.temperature || 0.7,
        is_default: data.is_default || false,
        is_enabled: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      llmProviders.push(provider)
      return provider
    },

    updateLLMProvider(id: string, data: Partial<LLMProvider>): LLMProvider {
      const idx = llmProviders.findIndex((p) => p.id === id)
      if (idx === -1) throw new Error('LLM provider not found')
      if (data.is_default) llmProviders.forEach((p) => (p.is_default = false))
      llmProviders[idx] = { ...llmProviders[idx], ...data, updated_at: new Date().toISOString() }
      return llmProviders[idx]
    },

    deleteLLMProvider(id: string): void {
      const idx = llmProviders.findIndex((p) => p.id === id)
      if (idx === -1) throw new Error('LLM provider not found')
      llmProviders.splice(idx, 1)
    },

    testLLMProvider(_id: string): LLMTestResponse {
      return { ok: true, model: 'OK', latency_ms: 150 }
    },
  }
}

// cryptoRandomId generates a UUID-shaped string. The mock layer does not
// have access to crypto.randomUUID in every test environment, so we
// roll our own using Math.random and assert the v4 layout.
function cryptoRandomId(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}
