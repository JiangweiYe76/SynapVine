import { fetchAPI } from './client'
import type {
  LLMProvider,
  LLMProviderCreateRequest,
  LLMProviderUpdateRequest,
  LLMProviderListResponse,
  LLMTestResponse,
  LLMUsageSummary,
} from '../types/llm'

export const llmAPI = {
  list: () => fetchAPI<LLMProviderListResponse>('/llm/providers'),
  get: (id: string) => fetchAPI<LLMProvider>(`/llm/providers/${id}`),

  getUsageSummary: (from?: string, to?: string) => {
    const params = new URLSearchParams()
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    const qs = params.toString()
    return fetchAPI<LLMUsageSummary>(`/llm/usage/summary${qs ? `?${qs}` : ''}`)
  },

  getDefault: () => fetchAPI<LLMProvider>('/llm/providers/default'),

  create: (data: LLMProviderCreateRequest) =>
    fetchAPI<LLMProvider>('/llm/providers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: string, data: LLMProviderUpdateRequest) =>
    fetchAPI<LLMProvider>(`/llm/providers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/llm/providers/${id}`, {
      method: 'DELETE',
    }),

  test: (id: string) =>
    fetchAPI<LLMTestResponse>(`/llm/providers/${id}/test`, {
      method: 'POST',
    }),
}
