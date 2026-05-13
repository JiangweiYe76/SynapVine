import type {
  GraphSummary,
  PaginatedResponse,
  GraphNode,
  NodeDetail,
  NodeEdgesResponse,
  SearchResponse,
  ExpandResponse,
} from '../types/graph'
import { mockServer } from '../mock/server'

const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

let token: string | null = null

async function mockFetch(path: string, params?: Record<string, string>): Promise<unknown> {
  const cleanPath = path.replace(/^\//, '')

  switch (cleanPath) {
    case 'summary':
      return mockServer.getSummary()

    case 'nodes': {
      const p: Parameters<typeof mockServer.getNodes>[0] = {}
      if (params?.offset) p.offset = parseInt(params.offset)
      if (params?.limit) p.limit = parseInt(params.limit)
      if (params?.sort) p.sort = params.sort
      if (params?.community_id) p.community_id = parseInt(params.community_id)
      if (params?.ids) p.ids = params.ids
      return mockServer.getNodes(p)
    }

    case 'search':
      return mockServer.search(params?.q || '', params?.limit ? parseInt(params.limit) : 20)

    case 'expand': {
      const expandParams: Parameters<typeof mockServer.expand>[0] = { ids: params?.ids || '' }
      if (params?.include_edges === 'true' || params?.include_edges === true) expandParams.include_edges = true
      if (params?.include_neighbors === 'true' || params?.include_neighbors === true) expandParams.include_neighbors = true
      return mockServer.expand(expandParams)
    }

    default: {
      const nodeDetailMatch = cleanPath.match(/^nodes\/(.+)$/)
      if (nodeDetailMatch) {
        const nodeId = nodeDetailMatch[1]
        if (cleanPath.endsWith('/edges')) {
          return mockServer.getNodeEdges(nodeId.replace('/edges', ''))
        }
        return mockServer.getNodeDetail(nodeId)
      }
      throw new Error(`Unknown endpoint: ${path}`)
    }
  }
}

async function fetchAPI<T>(path: string, params?: Record<string, string>): Promise<T> {
  if (USE_MOCK) {
    return mockFetch(path, params) as T
  }

  if (!token && path !== '/token') {
    await getToken()
  }

  const queryParams = new URLSearchParams(params)
  if (token) queryParams.set('token', token)

  const url = `/api/graph${path}?${queryParams.toString()}`
  const response = await fetch(url)

  if (response.status === 401) {
    token = null
    await getToken()
    return fetchAPI<T>(path, params)
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({}))
    throw new Error(body.message || `API error: ${response.status}`)
  }

  return response.json()
}

export async function getToken(): Promise<string> {
  if (USE_MOCK) {
    token = mockServer.getToken()
    return token
  }
  const response = await fetch('/api/token')
  if (!response.ok) {
    throw new Error('Failed to get token')
  }
  const data = await response.json()
  token = data.token
  return token
}

export async function getSummary(): Promise<GraphSummary> {
  return fetchAPI<GraphSummary>('/summary')
}

export async function getNodes(params?: {
  offset?: number
  limit?: number
  sort?: string
  community_id?: number
  ids?: string
}): Promise<PaginatedResponse<GraphNode>> {
  return fetchAPI<PaginatedResponse<GraphNode>>('/nodes', params as Record<string, string>)
}

export async function getNodeDetail(id: string): Promise<NodeDetail | null> {
  return fetchAPI<NodeDetail>(`/nodes/${id}`)
}

export async function getNodeEdges(id: string, direction?: string): Promise<NodeEdgesResponse | null> {
  const params: Record<string, string> = {}
  if (direction) params.direction = direction
  return fetchAPI<NodeEdgesResponse>(`/nodes/${id}/edges`, params)
}

export async function searchNodes(query: string, limit?: number): Promise<SearchResponse> {
  return fetchAPI<SearchResponse>('/search', { q: query, limit: limit?.toString() })
}

export async function expandNodes(params: {
  ids: string
  include_edges?: boolean
  include_neighbors?: boolean
}): Promise<ExpandResponse> {
  return fetchAPI<ExpandResponse>('/expand', params as Record<string, string>)
}
