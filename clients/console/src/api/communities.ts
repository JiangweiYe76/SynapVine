import { fetchAPI } from './client'
import type {
  Community,
  CommunitiesListResponse,
  CommunitiesTreeResponse,
  CommunityCreateRequest,
  CommunityUpdateRequest,
} from '../types/graph'

export const communitiesAPI = {
  list: () => fetchAPI<CommunitiesListResponse>('/communities'),

  tree: () => fetchAPI<CommunitiesTreeResponse>('/communities/tree'),

  get: (id: string) => fetchAPI<Community>(`/communities/${id}`),

  create: (data: CommunityCreateRequest) =>
    fetchAPI<Community>('/communities', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: CommunityUpdateRequest) =>
    fetchAPI<Community>(`/communities/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/communities/${id}`, {
      method: 'DELETE',
    }),
}
