import { ref, shallowRef, triggerRef, computed, provide, inject, type InjectionKey, type Ref } from 'vue'
import type {
  GraphNode,
  GraphEdge,
  HierarchicalCommunity,
  GraphStats,
  TimelineRange,
} from '../types/graph'
import {
  getSummary,
  getNodes,
  getNodeDetail,
  searchNodes,
  expandNodes,
  getTimelineRange,
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
  // Server-computed range of `first_appeared` across the full graph.
  // Independent of the (partial) nodes ref, so the timeline slider can
  // show the full extent of the dataset, not just the loaded window.
  timelineRange: Ref<TimelineRange>
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

// Maximum number of node IDs sent per expandNodes request. Keeps the
// URL query string well below server limits (~550 chars per batch).
const EXPAND_BATCH_SIZE = 50

export function useGraph(): GraphComposable {
  // Shallow refs: node/edge arrays can grow large, so deep reactivity
  // (traversed on every render dependency check) is not worth its cost.
  // Mutations in place must be followed by triggerRef to notify watchers.
  const nodes = shallowRef<GraphNode[]>([])
  const edges = shallowRef<GraphEdge[]>([])
  const communities = ref<HierarchicalCommunity[]>([])
  const stats = ref<GraphStats | null>(null)
  const selectedNode = ref<GraphNode | null>(null)
  const highlightedCommunity = ref<number[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  // Fall back to the current calendar year until the server-computed
  // range arrives, so the timeline slider has a sane single-point
  // default instead of NaN/Infinity.
  const fallbackYear = new Date().getFullYear()
  const timelineRange = ref<TimelineRange>({ min_year: fallbackYear, max_year: fallbackYear })

  const nodeMap = computed(() => {
    const map = new Map<string, GraphNode>()
    nodes.value.forEach(n => map.set(n.id, n))
    return map
  })

  /** Merge new edges into the existing edge list, skipping duplicates. */
  function mergeEdges(existing: GraphEdge[], incoming: GraphEdge[]) {
    const seen = new Set(existing.map(e => `${e.source}-${e.target}`))
    for (const e of incoming) {
      const key = `${e.source}-${e.target}`
      const rev = `${e.target}-${e.source}`
      if (!seen.has(key) && !seen.has(rev)) {
        existing.push(e)
        seen.add(key)
      }
    }
  }

  /**
   * Fetch edges for the given node IDs in batches to avoid URL length
   * limits. Returns all unique edges across batches.
   */
  async function expandNodesBatched(
    ids: string[],
    opts: { include_edges?: boolean; include_neighbors?: boolean } = {},
  ): Promise<{ nodes: GraphNode[]; edges: GraphEdge[] }> {
    const allNodes: GraphNode[] = []
    const allEdges: GraphEdge[] = []
    const seenEdgeKeys = new Set<string>()

    for (let i = 0; i < ids.length; i += EXPAND_BATCH_SIZE) {
      const batch = ids.slice(i, i + EXPAND_BATCH_SIZE)
      const resp = await expandNodes({ ids: batch.join(','), ...opts })
      allNodes.push(...resp.nodes)
      for (const e of resp.edges) {
        const key = `${e.source}-${e.target}`
        if (!seenEdgeKeys.has(key)) {
          seenEdgeKeys.add(key)
          allEdges.push(e)
        }
      }
    }
    return { nodes: allNodes, edges: allEdges }
  }

  async function loadInitial() {
    loading.value = true
    error.value = null
    try {
      const [summary, range] = await Promise.all([
        getSummary(),
        getTimelineRange(),
      ])
      communities.value = summary.communities
      stats.value = summary.stats
      timelineRange.value = range

      const allNodes = await getNodes({ limit: 200 })
      nodes.value = allNodes.nodes

      if (nodes.value.length > 0) {
        const ids = nodes.value.map(n => n.id)
        const { edges: fetchedEdges } = await expandNodesBatched(ids, { include_edges: true })
        edges.value = fetchedEdges
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

      if (newNodes.length > 0) {
        nodes.value.push(...newNodes)
        triggerRef(nodes)

        const newIds = newNodes.map(n => n.id)
        const { edges: newEdges } = await expandNodesBatched(newIds, { include_edges: true })
        mergeEdges(edges.value, newEdges)
        triggerRef(edges)
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

      if (newNodes.length > 0) {
        nodes.value.push(...newNodes)
        triggerRef(nodes)

        const newIds = newNodes.map(n => n.id)
        const { edges: newEdges } = await expandNodesBatched(newIds, { include_edges: true })
        mergeEdges(edges.value, newEdges)
        triggerRef(edges)
      }
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
        if (newNodes.length > 0) {
          nodes.value.push(...newNodes)
          triggerRef(nodes)
        }
        mergeEdges(edges.value, expandResponse.edges)
        triggerRef(edges)
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
          triggerRef(nodes)
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
          triggerRef(nodes)
          edges.value.push(...expandResponse.edges)
          triggerRef(edges)
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
            collectAllDescendants(c.children)
          }
          return true
        }
        if (c.children && collectDescendants(c.children, targetId)) {
          return true
        }
      }
      return false
    }
    // Recursively collect every descendant id, not just direct children.
    function collectAllDescendants(comms: HierarchicalCommunity[]) {
      for (const c of comms) {
        allIds.add(c.id)
        if (c.children) collectAllDescendants(c.children)
      }
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
    timelineRange,
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
  const timeline = useTimeline(
    graph.nodes as Ref<GraphNode[]>,
    graph.edges as Ref<GraphEdge[]>,
    graph.timelineRange,
  )

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
