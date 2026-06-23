export interface LLMProvider {
  id: string
  name: string
  base_url: string
  model: string
  max_tokens: number
  temperature: number
  is_default: boolean
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface LLMProviderCreateRequest {
  name: string
  base_url: string
  api_key: string
  model: string
  max_tokens?: number
  temperature?: number
  is_default?: boolean
}

export interface LLMProviderUpdateRequest {
  name?: string
  base_url?: string
  api_key?: string
  model?: string
  max_tokens?: number
  temperature?: number
  is_default?: boolean
  is_enabled?: boolean
}

export interface LLMProviderListResponse {
  providers: LLMProvider[]
  total: number
}

export interface LLMTestResponse {
  ok: boolean
  model?: string
  latency_ms?: number
  error?: string
}
