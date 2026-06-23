export interface Paper {
  id: string
  title: string
  authors: string
  source_url: string
  raw_text: string
  status: PaperStatus
  created_at: string
  updated_at: string
}

export type PaperStatus = 'uploaded' | 'analyzing' | 'analyzed' | 'reviewing' | 'merged'

export interface PaperCreateRequest {
  title: string
  authors: string
  source_url?: string
  raw_text: string
}

export interface PaperUpdateRequest {
  title?: string
  authors?: string
  source_url?: string
  status?: string
}

export interface PapersListResponse {
  papers: Paper[]
  total: number
}

export interface ExtractedNode {
  name: string
  description: string
  relevance: number
}

export interface ExtractedEdge {
  source: string
  target: string
  relation: string
  weight: number
}

export interface ReviewQueueItem {
  id: string
  paper_id: string
  extracted_nodes: ExtractedNode[]
  extracted_edges: ExtractedEdge[]
  status: ReviewStatus
  reviewer_id: string | null
  review_notes: string | null
  created_at: string
  reviewed_at: string | null
}

export type ReviewStatus = 'pending' | 'approved' | 'rejected'

export interface ReviewQueueListResponse {
  items: ReviewQueueItem[]
  total: number
}
