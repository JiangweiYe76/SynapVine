import { fetchAPI } from './client'
import type {
  EmbeddingProvider,
  EmbeddingProviderCreateRequest,
  EmbeddingProviderUpdateRequest,
  EmbeddingProviderListResponse,
  EmbeddingTestResponse,
} from '../types/embedding'

export const embeddingAPI = {
  list: () => fetchAPI<EmbeddingProviderListResponse>('/embedding/providers'),

  get: (id: string) => fetchAPI<EmbeddingProvider>(`/embedding/providers/${id}`),

  getDefault: () => fetchAPI<EmbeddingProvider>('/embedding/providers/default'),

  create: (data: EmbeddingProviderCreateRequest) =>
    fetchAPI<EmbeddingProvider>('/embedding/providers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: EmbeddingProviderUpdateRequest) =>
    fetchAPI<EmbeddingProvider>(`/embedding/providers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/embedding/providers/${id}`, {
      method: 'DELETE',
    }),

  test: (id: string) =>
    fetchAPI<EmbeddingTestResponse>(`/embedding/providers/${id}/test`, {
      method: 'POST',
    }),
}
