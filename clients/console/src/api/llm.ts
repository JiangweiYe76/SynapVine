import { fetchAPI } from './client'
import type {
  LLMProvider,
  LLMProviderCreateRequest,
  LLMProviderUpdateRequest,
  LLMProviderListResponse,
  LLMTestResponse,
} from '../types/llm'

export const llmAPI = {
  list: () => fetchAPI<LLMProviderListResponse>('/llm/providers'),

  get: (id: string) => fetchAPI<LLMProvider>(`/llm/providers/${id}`),

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
