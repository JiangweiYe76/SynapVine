export interface Node {
  id: string
  name: string
  category: string
  description: string
  influence_score: number
  first_appeared: string
  milestones?: string[]
}

export interface Edge {
  source: string
  target: string
  weight: number
  relation: string
}

export interface Pagination {
  offset: number
  limit: number
  total: number
  has_more: boolean
}

export interface NodesListResponse {
  nodes: Node[]
  pagination: Pagination
}

export interface NodeCreateRequest {
  id: string
  name: string
  category: string
  description: string
  influence_score: number
  first_appeared: string
  milestones?: string[]
}

export interface NodeUpdateRequest {
  name?: string
  category?: string
  description?: string
  influence_score?: number
  first_appeared?: string
  milestones?: string[]
}

export interface EdgeCreateRequest {
  source: string
  target: string
  weight: number
  relation: string
}

export interface EdgeUpdateRequest {
  weight?: number
  relation?: string
}

export interface EdgesListResponse {
  edges: Edge[]
  pagination: Pagination
}

export interface StatsResponse {
  total_nodes: number
  total_edges: number
  category_count: number
  categories: Record<string, number>
  avg_influence: number
}
