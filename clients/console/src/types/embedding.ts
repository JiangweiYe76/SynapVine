export interface EmbeddingProvider {
  id: string
  name: string
  base_url: string
  model: string
  dimensions: number
  is_default: boolean
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface EmbeddingProviderCreateRequest {
  name: string
  base_url: string
  api_key: string
  model: string
  dimensions?: number
  is_default?: boolean
}

export interface EmbeddingProviderUpdateRequest {
  name?: string
  base_url?: string
  api_key?: string
  model?: string
  dimensions?: number
  is_default?: boolean
  is_enabled?: boolean
}

export interface EmbeddingProviderListResponse {
  providers: EmbeddingProvider[]
  total: number
}

export interface EmbeddingTestResponse {
  ok: boolean
  dimensions?: number
  latency_ms?: number
  error?: string
}
