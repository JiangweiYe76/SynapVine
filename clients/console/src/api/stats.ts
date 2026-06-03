import { fetchAPI } from './client'
import type { StatsResponse } from '../types/graph'

export const statsAPI = {
  get: () => fetchAPI<StatsResponse>('/stats'),
}
