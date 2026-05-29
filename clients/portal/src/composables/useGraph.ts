import { ref, computed, provide, inject, watch, type InjectionKey, type Ref } from 'vue'
import type {
  GraphNode,
  GraphEdge,
  Community,
  HierarchicalCommunity,
  GraphStats,
} from '../types/graph'
import {
  getSummary,
  getNodes,
  getNodeDetail,
  searchNodes,
  expandNodes,
} from '../api/graph'
import { useTimeline, type TimelineComposable } from './useTimeline'

export interface GraphState {
  nodes: Ref<GraphNode[]>
  edges: Ref<GraphEdge[]>
  communities: Ref<HierarchicalCommunity[]>
  stats: Ref<GraphStats | null>
  selectedNode: Ref<GraphNode | null>
  highlightedCommunity: Ref<number[]>
  loading: Ref<boolean>
  error: Ref<string | null>
}

export interface GraphActions {
  loadInitial: () => Promise<void>
  loadMore: () => Promise<void>
  loadCommunity: (communityId: number) => Promise<void>
  search: (query: string) => Promise<void>
  selectNode: (id: string | null) => Promise<void>
  expandNode: (id: string) => Promise<void>
  highlightCommunity: (id: number | null) => void
  clearError: () => void
}

export type GraphComposable = GraphState & GraphActions

export const GraphKey: InjectionKey<GraphComposable> = Symbol('graph')
export const TimelineKey: InjectionKey<TimelineComposable> = Symbol('timeline')

export function useGraph(): GraphComposable {
  const nodes = ref<GraphNode[]>([])
  const edges = ref<GraphEdge[]>([])
  const communities = ref<HierarchicalCommunity[]>([])
  const stats = ref<GraphStats | null>(null)
  const selectedNode = ref<GraphNode | null>(null)
  const highlightedCommunity = ref<number[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const nodeMap = computed(() => {
    const map = new Map<string, GraphNode>()
    nodes.value.forEach(n => map.set(n.id, n))
    return map
  })

  async function loadInitial() {
    loading.value = true
    error.value = null
    try {
      const summary = await getSummary()
      communities.value = summary.communities
      stats.value = summary.stats

      const allNodes = await getNodes({ limit: 500 })
      nodes.value = allNodes.nodes

      if (nodes.value.length > 0) {
        const expandResponse = await expandNodes({
          ids: nodes.value.map(n => n.id).join(','),
          include_edges: true,
        })
        edges.value = expandResponse.edges
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load graph'
    } finally {
      loading.value = false
    }
  }

  async function loadMore() {
    if (loading.value) return
    loading.value = true
    try {
      const offset = nodes.value.length
      const response = await getNodes({ offset, limit: 50 })
      const newNodes = response.nodes.filter(n => !nodeMap.value.has(n.id))
      nodes.value.push(...newNodes)

      if (newNodes.length > 0) {
        const expandResponse = await expandNodes({
          ids: newNodes.map(n => n.id).join(','),
          include_edges: true,
        })
        const existingEdges = new Set(edges.value.map(e => `${e.source}-${e.target}`))
        const uniqueNewEdges = expandResponse.edges.filter(e =>
          !existingEdges.has(`${e.source}-${e.target}`) &&
          !existingEdges.has(`${e.target}-${e.source}`)
        )
        edges.value.push(...uniqueNewEdges)
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load more nodes'
    } finally {
      loading.value = false
    }
  }

  async function loadCommunity(communityId: number) {
    loading.value = true
    try {
      const response = await getNodes({ community_id: communityId, limit: 200 })
      const newNodes = response.nodes.filter(n => !nodeMap.value.has(n.id))
      nodes.value.push(...newNodes)

      const allIds = nodes.value.map(n => n.id)
      const expandResponse = await expandNodes({
        ids: allIds.join(','),
        include_edges: true,
      })
      edges.value = expandResponse.edges
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load community'
    } finally {
      loading.value = false
    }
  }

  async function search(query: string) {
    if (!query.trim()) return
    loading.value = true
    try {
      const response = await searchNodes(query, 10)
      if (response.results.length > 0) {
        const resultIds = response.results.map(r => r.id)
        const expandResponse = await expandNodes({
          ids: resultIds.join(','),
          include_neighbors: true,
        })

        const newNodes = expandResponse.nodes.filter(n => !nodeMap.value.has(n.id))
        nodes.value.push(...newNodes)

        const existingEdges = new Set(edges.value.map(e => `${e.source}-${e.target}`))
        const uniqueNewEdges = expandResponse.edges.filter(e =>
          !existingEdges.has(`${e.source}-${e.target}`) &&
          !existingEdges.has(`${e.target}-${e.source}`)
        )
        edges.value.push(...uniqueNewEdges)
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Search failed'
    } finally {
      loading.value = false
    }
  }

  async function selectNode(id: string | null) {
    if (!id) {
      selectedNode.value = null
      return
    }

    const existing = nodeMap.value.get(id)
    if (existing) {
      selectedNode.value = existing
      return
    }

    loading.value = true
    try {
      const detail = await getNodeDetail(id)
      if (detail) {
        if (!nodeMap.value.has(detail.node.id)) {
          nodes.value.push(detail.node)
        }
        selectedNode.value = detail.node

        const neighborIds = detail.neighbors.map(n => n.id)
        const newNeighborIds = neighborIds.filter(nid => !nodeMap.value.has(nid))

        if (newNeighborIds.length > 0) {
          const expandResponse = await expandNodes({
            ids: [...newNeighborIds, id].join(','),
            include_edges: true,
          })
          nodes.value.push(...expandResponse.nodes.filter(n => !nodeMap.value.has(n.id)))
          edges.value.push(...expandResponse.edges)
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load node detail'
    } finally {
      loading.value = false
    }
  }

  async function expandNode(id: string) {
    await selectNode(id)
  }

  async function highlightCommunity(id: number | null) {
    if (id === null) {
      highlightedCommunity.value = []
      return
    }

    const allIds = new Set<number>()
    function collectDescendants(comms: HierarchicalCommunity[], targetId: number): boolean {
      for (const c of comms) {
        if (c.id === targetId) {
          allIds.add(c.id)
          if (c.children) {
            for (const child of c.children) {
              allIds.add(child.id)
            }
          }
          return true
        }
        if (c.children && collectDescendants(c.children, targetId)) {
          return true
        }
      }
      return false
    }
    collectDescendants(communities.value, id)

    highlightedCommunity.value = [...allIds]
    await loadCommunity(id)
  }

  function clearError() {
    error.value = null
  }

  return {
    nodes,
    edges,
    communities,
    stats,
    selectedNode,
    highlightedCommunity,
    loading,
    error,
    loadInitial,
    loadMore,
    loadCommunity,
    search,
    selectNode,
    expandNode,
    highlightCommunity,
    clearError,
  }
}

export function useGraphWithTimeline(): GraphComposable & { timeline: TimelineComposable } {
  const graph = useGraph()
  const timeline = useTimeline(graph.nodes as Ref<GraphNode[]>, graph.edges as Ref<GraphEdge[]>)

  return {
    ...graph,
    timeline,
  }
}

export function provideGraph() {
  const graphWithTimeline = useGraphWithTimeline()
  provide(GraphKey, graphWithTimeline)
  provide(TimelineKey, graphWithTimeline.timeline)
  return graphWithTimeline
}

export function injectGraph(): GraphComposable {
  const graph = inject(GraphKey)
  if (!graph) {
    throw new Error('useGraph must be used within a provider')
  }
  return graph
}

export function injectTimeline(): TimelineComposable {
  const timeline = inject(TimelineKey)
  if (!timeline) {
    throw new Error('useTimeline must be used within a provider')
  }
  return timeline
}
