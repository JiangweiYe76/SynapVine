import type {
  GraphSummary,
  PaginatedResponse,
  GraphNode,
  GraphEdge,
  NodeDetail,
  NodeEdgesResponse,
  SearchResponse,
  ExpandResponse,
} from '../types/graph'
import { generateMockData } from './data'

export function createMockServer() {
  console.log('Generating mock data...')
  const mockData = generateMockData()
  console.log('Mock data generated, nodes:', mockData.nodes.length, 'edges:', mockData.edges.length)
  const nodeMap = new Map(mockData.nodes.map(n => [n.id, n]))

  function getDescendantCommunityIds(communityId: number): number[] {
    return mockData.allCommunityIds.get(communityId) || [communityId]
  }

  return {
    getToken(): string {
      return 'mock_token_' + Date.now()
    },

    getSummary(): GraphSummary {
      const topNodes = [...mockData.nodes]
        .sort((a, b) => b.influence_score - a.influence_score)
        .slice(0, 20)

      return {
        communities: mockData.hierarchicalCommunities,
        stats: {
          total_nodes: mockData.nodes.length,
          total_edges: mockData.edges.length,
          community_count: mockData.communities.length,
          max_level: 2,
        },
        top_nodes: topNodes,
      }
    },

    getNodes(params: {
      offset?: number
      limit?: number
      sort?: string
      community_id?: number
      ids?: string
    }): PaginatedResponse<GraphNode> {
      let result = [...mockData.nodes]

      if (params.ids) {
        const idSet = new Set(params.ids.split(','))
        result = result.filter(n => idSet.has(n.id))
      } else if (params.community_id !== undefined && params.community_id > 0) {
        const descendantIds = getDescendantCommunityIds(params.community_id)
        const idSet = new Set(descendantIds)
        result = result.filter(n => idSet.has(n.community_id))
      }

      if (params.sort === 'name') {
        result.sort((a, b) => a.name.localeCompare(b.name))
      } else {
        result.sort((a, b) => b.influence_score - a.influence_score)
      }

      const offset = params.offset || 0
      const limit = Math.min(params.limit || 100, 500)
      const paginated = result.slice(offset, offset + limit)

      return {
        nodes: paginated,
        pagination: {
          offset,
          limit,
          total: result.length,
          has_more: offset + limit < result.length,
        },
      }
    },

    getNodeDetail(id: string): NodeDetail | null {
      const node = nodeMap.get(id)
      if (!node) return null

      const neighbors = mockData.edges
        .filter(e => e.source === id || e.target === id)
        .map(e => {
          const neighborId = e.source === id ? e.target : e.source
          const neighbor = nodeMap.get(neighborId)
          if (!neighbor) return null
          return {
            id: neighbor.id,
            name: neighbor.name,
            community_id: neighbor.community_id,
            influence_score: neighbor.influence_score,
            weight: e.weight,
            relation: e.relation,
          }
        })
        .filter(Boolean) as NodeDetail['neighbors']

      neighbors.sort((a, b) => b.weight - a.weight)

      return { node, neighbors }
    },

    getNodeEdges(id: string): NodeEdgesResponse | null {
      const node = nodeMap.get(id)
      if (!node) return null

      const edges = mockData.edges.filter(e => e.source === id || e.target === id)
      return { node_id: id, edges }
    },

    search(query: string, limit = 20): SearchResponse {
      const q = query.toLowerCase()
      const results = mockData.nodes
        .filter(n =>
          n.name.toLowerCase().includes(q) ||
          n.description.toLowerCase().includes(q)
        )
        .map(n => ({
          id: n.id,
          name: n.name,
          community_id: n.community_id,
          influence_score: n.influence_score,
          highlight: n.description,
        }))
        .slice(0, limit)

      return { query, results }
    },

    expand(params: {
      ids: string
      include_edges?: boolean
      include_neighbors?: boolean
    }): ExpandResponse {
      const idSet = new Set(params.ids.split(','))
      const nodes: GraphNode[] = []
      const edges: GraphEdge[] = []

      idSet.forEach(id => {
        const node = nodeMap.get(id)
        if (node) nodes.push(node)
      })

      if (params.include_edges !== false) {
        mockData.edges.forEach(e => {
          if (idSet.has(e.source) && idSet.has(e.target)) {
            edges.push(e)
          }
        })
      }

      if (params.include_neighbors) {
        mockData.edges.forEach(e => {
          if (idSet.has(e.source) && !idSet.has(e.target)) {
            const neighbor = nodeMap.get(e.target)
            if (neighbor) nodes.push(neighbor)
          }
          if (idSet.has(e.target) && !idSet.has(e.source)) {
            const neighbor = nodeMap.get(e.source)
            if (neighbor) nodes.push(neighbor)
          }
        })
      }

      return { nodes, edges: edges as unknown as import('../types/graph').GraphEdge[] }
    },
  }
}
