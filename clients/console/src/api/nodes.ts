import { fetchAPI } from './client'
import type {
  Node,
  NodesListResponse,
  NodeCreateRequest,
  NodeUpdateRequest,
} from '../types/graph'

export const nodesAPI = {
  list: (offset = 0, limit = 20, search = '') => {
    const params = new URLSearchParams({ offset: String(offset), limit: String(limit) })
    if (search) params.set('search', search)
    return fetchAPI<NodesListResponse>(`/nodes?${params.toString()}`)
  },

  get: (id: string) => fetchAPI<Node>(`/nodes/${id}`),

  create: (data: NodeCreateRequest) =>
    fetchAPI<Node>('/nodes', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: NodeUpdateRequest) =>
    fetchAPI<Node>(`/nodes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/nodes/${id}`, {
      method: 'DELETE',
    }),
}
