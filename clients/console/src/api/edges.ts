import { fetchAPI } from './client'
import type {
  Edge,
  EdgesListResponse,
  EdgeCreateRequest,
  EdgeUpdateRequest,
} from '../types/graph'

export const edgesAPI = {
  list: (offset = 0, limit = 20, search = '') => {
    const params = new URLSearchParams({ offset: String(offset), limit: String(limit) })
    if (search) params.set('search', search)
    return fetchAPI<EdgesListResponse>(`/edges?${params.toString()}`)
  },

  get: (source: string, target: string) => fetchAPI<Edge>(`/edges/${source}/${target}`),

  create: (data: EdgeCreateRequest) =>
    fetchAPI<Edge>('/edges', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (source: string, target: string, data: EdgeUpdateRequest) =>
    fetchAPI<Edge>(`/edges/${source}/${target}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (source: string, target: string) =>
    fetchAPI<void>(`/edges/${source}/${target}`, {
      method: 'DELETE',
    }),
}
