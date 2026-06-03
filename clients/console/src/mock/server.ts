import { mockUser, mockCredentials, mockNodes } from './data'
import type { LoginRequest, LoginResponse, User } from '../types/auth'
import type {
  Node,
  NodesListResponse,
  NodeCreateRequest,
  NodeUpdateRequest,
  StatsResponse,
} from '../types/graph'

export function createMockServer() {
  let token: string | null = null
  let nodes: Node[] = [...mockNodes]

  return {
    login(credentials: LoginRequest): LoginResponse {
      if (
        credentials.username === mockCredentials.username &&
        credentials.password === mockCredentials.password
      ) {
        token = 'mock_token_' + Date.now()
        return { token, user: mockUser }
      }
      throw new Error('Invalid username or password')
    },

    me(_authHeader?: string): User {
      return mockUser
    },

    listNodes(offset = 0, limit = 20, search = ''): NodesListResponse {
      let filtered = nodes
      if (search) {
        const q = search.toLowerCase()
        filtered = nodes.filter(
          (n) =>
            n.id.toLowerCase().includes(q) ||
            n.name.toLowerCase().includes(q) ||
            n.category.toLowerCase().includes(q) ||
            n.description.toLowerCase().includes(q)
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
      if (nodes.some((n) => n.id === data.id)) {
        throw new Error('Node already exists')
      }
      const node: Node = {
        id: data.id,
        name: data.name,
        category: data.category,
        description: data.description,
        influence_score: data.influence_score,
        first_appeared: data.first_appeared,
        milestones: data.milestones,
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
    },

    getStats(): StatsResponse {
      const categories: Record<string, number> = {}
      let totalInfluence = 0
      for (const n of nodes) {
        categories[n.category] = (categories[n.category] ?? 0) + 1
        totalInfluence += n.influence_score
      }
      return {
        total_nodes: nodes.length,
        total_edges: 0,
        category_count: Object.keys(categories).length,
        categories,
        avg_influence: nodes.length > 0 ? totalInfluence / nodes.length : 0,
      }
    },
  }
}
