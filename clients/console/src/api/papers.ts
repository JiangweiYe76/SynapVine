import { fetchAPI } from './client'
import type {
  Paper,
  PaperCreateRequest,
  PaperUpdateRequest,
  PapersListResponse,
} from '../types/paper'

export const papersAPI = {
  list: (offset = 0, limit = 20) => {
    const params = new URLSearchParams({ offset: String(offset), limit: String(limit) })
    return fetchAPI<PapersListResponse>(`/papers?${params.toString()}`)
  },

  get: (id: string) => fetchAPI<Paper>(`/papers/${id}`),

  create: (data: PaperCreateRequest) =>
    fetchAPI<Paper>('/papers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  createWithPDF: (formData: FormData) =>
    fetchAPI<Paper>('/papers', {
      method: 'POST',
      body: formData,
      // Do NOT set Content-Type; the browser will set the correct
      // multipart/form-data boundary automatically.
      headers: {},
    }),

  update: (id: string, data: PaperUpdateRequest) =>
    fetchAPI<Paper>(`/papers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchAPI<void>(`/papers/${id}`, {
      method: 'DELETE',
    }),

  analyze: (id: string) =>
    fetchAPI<{ status: string; message: string }>(`/papers/${id}/analyze`, {
      method: 'POST',
    }),

  downloadPDF: async (id: string): Promise<void> => {
    const token = localStorage.getItem('token') || ''
    const res = await fetch(`/api/papers/${id}/pdf`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!res.ok) throw new Error('Failed to download PDF')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    window.open(url, '_blank')
    URL.revokeObjectURL(url)
  },
}
