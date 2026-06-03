import { mockUser, mockCredentials, mockNodes, mockEdges } from './data'
import type { LoginRequest, LoginResponse, User } from '../types/auth'
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

export function createMockServer() {
  let token: string | null = null
  let nodes: Node[] = [...mockNodes]
  let edges: Edge[] = [...mockEdges]

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
      // Also delete related edges
      edges = edges.filter((e) => e.source !== id && e.target !== id)
    },

    listEdges(offset = 0, limit = 20, search = ''): EdgesListResponse {
      let filtered = edges
      if (search) {
        const q = search.toLowerCase()
        filtered = edges.filter(
          (e) =>
            e.source.toLowerCase().includes(q) ||
            e.target.toLowerCase().includes(q) ||
            e.relation.toLowerCase().includes(q)
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
      const categories: Record<string, number> = {}
      let totalInfluence = 0
      for (const n of nodes) {
        categories[n.category] = (categories[n.category] ?? 0) + 1
        totalInfluence += n.influence_score
      }
      return {
        total_nodes: nodes.length,
        total_edges: edges.length,
        category_count: Object.keys(categories).length,
        categories,
        avg_influence: nodes.length > 0 ? totalInfluence / nodes.length : 0,
      }
    },
  }
}
