<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import ForceGraph3D from '3d-force-graph'
import * as THREE from 'three'
import type { GraphNode, GraphEdge, Community } from '../types/graph'
import { PALETTE } from '../types/graph'

const props = defineProps<{
  nodes: GraphNode[]
  edges: GraphEdge[]
  communities: Community[]
  selectedNode: GraphNode | null
  highlightedCommunity: number[]
  highlightedNeighbors: Set<string>
}>()

const emit = defineEmits<{
  nodeClick: [nodeId: string]
  backgroundClick: []
}>()

const containerRef = ref<HTMLDivElement>()
let graph: any = null

// Cache for Three.js geometries and materials to avoid GPU memory leaks
const geometryCache = new Map<string, any>()
const materialCache = new Map<string, any>()
const nodeObjects = new Map<string, any>()

function getCachedGeometry(radius: string, widthSegments = 24, heightSegments = 24): any {
  const key = `sphere_${radius}_${widthSegments}_${heightSegments}`
  if (!geometryCache.has(key)) {
    geometryCache.set(key, new THREE.SphereGeometry(parseFloat(radius), widthSegments, heightSegments))
  }
  return geometryCache.get(key)!
}

function getCachedMaterial(color: string, isSelected: boolean, isOutline = false): any {
  const key = `mat_${color}_${isSelected}_${isOutline}`
  if (!materialCache.has(key)) {
    if (isOutline) {
      materialCache.set(key, new THREE.MeshBasicMaterial({
        color: '#58a6ff',
        side: THREE.BackSide,
        transparent: false,
      }))
    } else {
      // getFilteredColor appends an alpha suffix (#RRGGBBAA) to dim
      // non-highlighted community nodes. Three.js Color ignores the alpha
      // channel, so parse it and map to material opacity instead.
      const hasAlpha = color.length === 9
      const baseColor = hasAlpha ? color.slice(0, 7) : color
      const opacity = hasAlpha ? parseInt(color.slice(7, 9), 16) / 255 : 0.95
      materialCache.set(key, new THREE.MeshPhongMaterial({
        color: baseColor,
        shininess: 80,
        transparent: true,
        opacity,
        emissive: isSelected ? baseColor : '#000000',
        emissiveIntensity: isSelected ? 0.6 : 0,
      }))
    }
  }
  return materialCache.get(key)!
}

interface SavedCamera {
  pos: { x: number; y: number; z: number }
  lookAt: { x: number; y: number; z: number }
}
const savedCamera = ref<SavedCamera | null>(null)

function getFilteredColor(node: any, communityIds: number[]) {
  const base = PALETTE[node.community_id % PALETTE.length]
  if (communityIds.length === 0) return base
  if (communityIds.includes(node.community_id)) return base
  return `${base}1A`
}

function createNodeObject(node: any) {
  const communityIds = props.highlightedCommunity
  const isSelected = props.selectedNode?.id === node.id
  const color = getFilteredColor(node, communityIds)
  const score = node.influence_score || 5
  const size = Math.max(0.5, (score - 5) * 1.5)

  const group = new THREE.Group()

  const sphereGeo = getCachedGeometry(size.toString())
  const sphereMat = getCachedMaterial(color, isSelected)
  const sphere = new THREE.Mesh(sphereGeo, sphereMat)
  group.add(sphere)

  if (isSelected) {
    const outlineGeo = getCachedGeometry((size * 1.18).toString())
    const outlineMat = getCachedMaterial(color, true, true)
    const outline = new THREE.Mesh(outlineGeo, outlineMat)
    group.add(outline)
  }

  // Store reference for later updates
  nodeObjects.set(node.id, group)

  return group
}

function updateSelectedNodeVisuals() {
  const selectedId = props.selectedNode?.id
  nodeObjects.forEach((group, nodeId) => {
    const isSelected = nodeId === selectedId
    const sphere = group.children[0] as any
    if (sphere && sphere.material) {
      sphere.material.emissive = new THREE.Color(isSelected ? sphere.material.color : '#000000')
      sphere.material.emissiveIntensity = isSelected ? 0.6 : 0
    }
  })
}

// Re-apply community-filtered colors to all existing node objects.
// Swaps each sphere's material reference to the appropriate cached
// material so dimmed (non-highlighted) nodes become semi-transparent.
function updateNodeColors() {
  if (!graph) return
  const communityIds = props.highlightedCommunity
  const graphNodes = graph.graphData()?.nodes || []
  const nodeMap = new Map<string, any>(graphNodes.map((n: any) => [n.id, n]))

  nodeObjects.forEach((group, nodeId) => {
    const sphere = group.children[0] as any
    if (!sphere) return
    const node = nodeMap.get(nodeId)
    if (!node) return

    const isSelected = props.selectedNode?.id === nodeId
    const color = getFilteredColor(node, communityIds)
    sphere.material = getCachedMaterial(color, isSelected)
  })
}

function getBackgroundColor(): string {
  return getComputedStyle(document.documentElement).getPropertyValue('--color-bg-primary').trim() || '#0d1117'
}

function relationColor(relation: string): string {
  const colors: Record<string, string> = {
    // Architecture / inheritance — soft blue-gray
    'architecture foundation': '#5a7a8a',
    'core mechanism': '#5a7a8a',
    'component': '#5a7a8a',
    'base architecture': '#5a7a8a',
    'backbone network': '#5a7a8a',
    'encoder only': '#5a7a8a',
    'decoder only': '#5a7a8a',
    'variant foundation': '#5a7a8a',
    'foundation model': '#5a7a8a',
    'replaces': '#5a7a8a',
    'unified framework': '#5a7a8a',
    'theoretical basis': '#5a7a8a',
    'subfield': '#5a7a8a',
    'paradigm': '#5a7a8a',
    'implementation': '#5a7a8a',
    'depends on': '#5a7a8a',
    'applied to CV': '#5a7a8a',
    'used for': '#5a7a8a',
    'image encoder': '#5a7a8a',
    'uses': '#5a7a8a',
    'uses CNN': '#5a7a8a',
    'retrieval dependency': '#5a7a8a',
    'challenge': '#5a7a8a',
    'based on': '#5a7a8a',
    // Training / optimization — soft green
    'training': '#6a8a6a',
    'optimization': '#6a8a6a',
    'regularization': '#6a8a6a',
    'accelerates inference': '#6a8a6a',
    'composable': '#6a8a6a',
    'composable with': '#6a8a6a',
    'training method': '#6a8a6a',
    'hyperparameter': '#6a8a6a',
    'optimizes': '#6a8a6a',
    // Evolution / improvement — soft purple
    'improvement': '#7a6a8a',
    'iteration': '#7a6a8a',
    'optimized variant': '#7a6a8a',
    'variant': '#7a6a8a',
    'extends': '#7a6a8a',
    'simplified variant': '#7a6a8a',
    'enhancement': '#7a6a8a',
    'enhances': '#7a6a8a',
    'alternative': '#7a6a8a',
    'enables': '#7a6a8a',
    'improves': '#7a6a8a',
    'improves reasoning': '#7a6a8a',
    'reasoning enhancement': '#7a6a8a',
    'knowledge enhancement': '#7a6a8a',
    'surpassed by': '#7a6a8a',
    // Same domain
    'same domain': '#5a6a6a',
    // Cross domain
    'cross domain': '#4a4a5a',
  }
  return colors[relation] || '#5a6a6a'
}

function applyCommunityFilter(_communityIds: number[]) {
  updateNodeColors()
}

function collectPositions() {
  if (!graph) return new Map()
  const existing = graph.graphData()
  const map = new Map<string, { x: number; y: number; z: number }>()
  if (existing?.nodes) {
    for (const n of existing.nodes) {
      if (n.x != null) map.set(n.id, { x: n.x, y: n.y, z: n.z })
    }
  }
  return map
}

function initGraph() {
  if (!containerRef.value || graph) return

  const nodeData = props.nodes
  const linkData = props.edges.map(e => ({
    source: e.source,
    target: e.target,
    weight: e.weight,
    relation: e.relation
  }))

  graph = (ForceGraph3D as any)({ controlType: 'orbit' })(containerRef.value)
    .graphData({ nodes: nodeData, links: linkData })
    .nodeId('id')
    .nodeVal('influence_score')
    .nodeColor((n: any) => PALETTE[n.community_id % PALETTE.length])
    .nodeLabel((n: any) => `${n.name} (${n.influence_score})`)
    .nodeThreeObject((n: any) => createNodeObject(n))
    .linkWidth((link: any) => link.weight * 3)
    .linkOpacity(0.45)
    .linkColor((link: any) => relationColor(link.relation))
    .onNodeClick((node: any) => emit('nodeClick', node.id))
    .onBackgroundClick(() => emit('backgroundClick'))
    .enableNavigationControls(true)
    .showNavInfo(false)
    .backgroundColor(getBackgroundColor())

  setTimeout(() => {
    graph?.zoomToFit(600, 60)
  }, 800)
}

function updateGraphTheme() {
  if (graph) {
    graph.backgroundColor(getBackgroundColor())
  }
}

let themeObserver: MutationObserver | null = null

onMounted(() => {
  setTimeout(initGraph, 100)

  themeObserver = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.attributeName === 'data-theme') {
        updateGraphTheme()
      }
    }
  })

  themeObserver.observe(document.documentElement, { attributes: true })
})

watch([() => props.nodes, () => props.edges], ([newNodes, newEdges]) => {
  if (!graph && newNodes?.length > 0) {
    initGraph()
  } else if (graph) {
    const posMap = collectPositions()
    const nodeData = newNodes.map(n => {
      const pos = posMap.get(n.id)
      return pos ? { ...n, x: pos.x, y: pos.y, z: pos.z } : { ...n }
    })
    graph.graphData({
      nodes: nodeData,
      links: newEdges.map(e => ({ source: e.source, target: e.target, weight: e.weight, relation: e.relation }))
    })
    graph.backgroundColor(getBackgroundColor())
    applyCommunityFilter(props.highlightedCommunity)
  }
}, { deep: true })

watch(() => props.highlightedCommunity, (communityIds) => {
  applyCommunityFilter(communityIds)
})

watch(() => props.selectedNode, (node, prevNode) => {
  // Only update visuals for selected node instead of rebuilding all nodes
  if (graph) {
    updateSelectedNodeVisuals()
  }

  if (!graph) return

  if (!node) {
    if (savedCamera.value) {
      graph.cameraPosition(
        savedCamera.value.pos,
        savedCamera.value.lookAt,
        800
      )
      savedCamera.value = null
    }
    return
  }

  if (!prevNode) {
    const cp = graph.cameraPosition()
    savedCamera.value = {
      pos: { x: cp.x, y: cp.y, z: cp.z },
      lookAt: { x: cp.lookAt.x, y: cp.lookAt.y, z: cp.lookAt.z },
    }
  }

  const tryLookAt = () => {
    const graphNodes = graph.graphData()?.nodes || []
    const graphNode = graphNodes.find((n: any) => n.id === node.id)

    if (!graphNode || graphNode.x == null) {
      setTimeout(tryLookAt, 16)
      return
    }

    const distance = 160
    const distRatio = 1 + distance / Math.hypot(graphNode.x, graphNode.y, graphNode.z)
    graph.cameraPosition(
      { x: graphNode.x * distRatio, y: graphNode.y * distRatio, z: graphNode.z * distRatio },
      graphNode,
      800
    )
  }

  tryLookAt()
})

onUnmounted(() => {
  if (graph) {
    graph._destructor()
    graph = null
  }
  
  // Dispose all cached geometries
  geometryCache.forEach((geometry) => {
    geometry.dispose()
  })
  geometryCache.clear()
  
  // Dispose all cached materials
  materialCache.forEach((material) => {
    material.dispose()
  })
  materialCache.clear()
  
  // Clear node objects references
  nodeObjects.clear()
  
  if (themeObserver) {
    themeObserver.disconnect()
    themeObserver = null
  }
})
</script>

<template>
  <div ref="containerRef" class="w-full h-full" />
</template>