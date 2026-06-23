import { fetchAPI } from './client'
import type {
  ReviewQueueItem,
  ReviewQueueListResponse,
} from '../types/paper'

export const reviewAPI = {
  list: (offset = 0, limit = 20, status = '') => {
    const params = new URLSearchParams({ offset: String(offset), limit: String(limit) })
    if (status) params.set('status', status)
    return fetchAPI<ReviewQueueListResponse>(`/review-queue?${params.toString()}`)
  },

  get: (id: string) => fetchAPI<ReviewQueueItem>(`/review-queue/${id}`),

  approve: (id: string, reviewerId: string, notes = '') =>
    fetchAPI<ReviewQueueItem>(`/review-queue/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({ reviewer_id: reviewerId, review_notes: notes }),
    }),

  reject: (id: string, reviewerId: string, notes = '') =>
    fetchAPI<void>(`/review-queue/${id}/reject`, {
      method: 'POST',
      body: JSON.stringify({ reviewer_id: reviewerId, review_notes: notes }),
    }),
}
