export interface GraphNode {
  id: string
  name: string
  community_id: number
  influence_score: number
  description: string
  degree: number
  first_appeared?: string    // 首次出现年月 (YYYY-MM)
  milestones?: Milestone[]   // 关键里程碑
  x?: number
  y?: number
  z?: number
}

export interface Milestone {
  year: number
  event: string
  impact?: number
}

export interface TimelineRange {
  minYear: number
  maxYear: number
}

export interface GraphEdge {
  source: string
  target: string
  weight: number
  relation: string
}

export interface Community {
  id: number
  name: string
  color: string
  node_count: number
}

export interface HierarchicalCommunity {
  id: number
  parent_id: number | null
  name: string
  color: string
  level: number
  node_count: number
  children?: HierarchicalCommunity[]
}

export interface GraphStats {
  total_nodes: number
  total_edges: number
  community_count: number
  max_level: number
}

export interface GraphSummary {
  communities: HierarchicalCommunity[]
  stats: GraphStats
  top_nodes: GraphNode[]
}

export interface PaginatedResponse<T> {
  nodes: T[]
  pagination: {
    offset: number
    limit: number
    total: number
    has_more: boolean
  }
}

export interface NodeDetail {
  node: GraphNode
  neighbors: NeighborNode[]
}

export interface NeighborNode {
  id: string
  name: string
  community_id: number
  influence_score: number
  weight: number
  relation: string
}

export interface NodeEdgesResponse {
  node_id: string
  edges: GraphEdge[]
}

export interface SearchResult {
  id: string
  name: string
  community_id: number
  influence_score: number
  highlight: string
}

export interface SearchResponse {
  query: string
  results: SearchResult[]
}

export interface ExpandResponse {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

export const PALETTE: string[] = [
  '#4C78A8', '#F58518', '#E45756', '#72B7B2',
  '#54A24B', '#EECA3B', '#B279A2', '#FF9DA6',
  '#9D755D', '#BAB0AC',
]

export const LEVEL_PALETTES: Record<number, string[]> = {
  0: ['#666666'],
  1: ['#4C78A8', '#F58518', '#E45756', '#72B7B2', '#54A24B', '#EECA3B', '#B279A2'],
  2: ['#5B8CC7', '#7A9FD4', '#5291C7', '#F79E26', '#F7A64A', '#E67070', '#E97E7E', '#8AC9C4', '#A3D5D1', '#69B35A', '#85C278', '#F2D355', '#F5DE7A', '#C592B0', '#D4ACC5', '#FFB3BA', '#FFCDD1', '#B39483', '#CDB5A7'],
}
