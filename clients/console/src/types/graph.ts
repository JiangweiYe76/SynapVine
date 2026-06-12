export interface Node {
  id: string
  name: string
  description: string
  influence_score: number
  first_appeared: string
  milestones?: string[]
  community_id?: string | null
}

export interface Edge {
  source: string
  target: string
  weight: number
  relation: string
}

export interface Community {
  id: string
  name: string
  color: string
  level: number
  domain: string
  parent_id: string | null
  node_count: number
}

export interface HierarchicalCommunity {
  id: string
  parent_id: string | null
  name: string
  color: string
  level: number
  domain: string
  node_count: number
  children?: HierarchicalCommunity[]
}

export interface CommunitiesListResponse {
  communities: Community[]
}

export interface CommunitiesTreeResponse {
  communities: HierarchicalCommunity[]
}

export interface CommunityCreateRequest {
  // Optional: when omitted, the backend mints a fresh UUID and returns it
  // as part of the created resource.
  id?: string
  name: string
  color: string
  domain: string
  parent_id?: string | null
}

export interface CommunityUpdateRequest {
  name?: string
  color?: string
  domain?: string
  parent_id?: string | null
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
  // Optional: when omitted, the backend mints a fresh UUID and returns it
  // as part of the created resource.
  id?: string
  name: string
  description: string
  influence_score: number
  first_appeared: string
  milestones?: string[]
  community_id?: string | null
}

export interface NodeUpdateRequest {
  name?: string
  description?: string
  influence_score?: number
  first_appeared?: string
  milestones?: string[]
  /**
   * Community assignment using a tri-state convention:
   * - `undefined` (field absent) -> leave the assignment unchanged
   * - `null`                    -> remove the node from its community
   * - `string`                  -> assign the node to the given community
   *
   * The console handler forwards this to the core service, which performs
   * the actual reconciliation against the BELONGS_TO relationship.
   */
  community_id?: string | null | undefined
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
  avg_influence: number
}
