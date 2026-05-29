<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, inject, type Ref } from 'vue'
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

interface SavedCamera {
  pos: { x: number; y: number; z: number }
  lookAt: { x: number; y: number; z: number }
}
const savedCamera = ref<SavedCamera | null>(null)

function getNodeColor(node: any) {
  return PALETTE[node.community_id % PALETTE.length]
}

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
  const size = (score - 5.5) * 1.5

  const group = new THREE.Group()

  const sphereGeo = new THREE.SphereGeometry(size, 24, 24)
  const sphereMat = new THREE.MeshPhongMaterial({
    color: color,
    shininess: 80,
    transparent: true,
    opacity: 0.95,
    emissive: isSelected ? color : '#000000',
    emissiveIntensity: isSelected ? 0.6 : 0,
  })
  const sphere = new THREE.Mesh(sphereGeo, sphereMat)
  group.add(sphere)

  if (isSelected) {
    const outlineGeo = new THREE.SphereGeometry(size * 1.18, 24, 24)
    const outlineMat = new THREE.MeshBasicMaterial({
      color: '#58a6ff',
      side: THREE.BackSide,
      transparent: false,
    })
    const outline = new THREE.Mesh(outlineGeo, outlineMat)
    group.add(outline)
  }

  return group
}

function refreshNodes() {
  if (!graph) return
  graph.nodeThreeObject((n: any) => createNodeObject(n))
}

function getBackgroundColor(): string {
  return getComputedStyle(document.documentElement).getPropertyValue('--color-bg-primary').trim() || '#0d1117'
}

function relationColor(relation: string): string {
  const colors: Record<string, string> = {
    // 架构/继承 — 柔蓝灰
    '架构基础': '#5a7a8a',
    '核心机制': '#5a7a8a',
    '机制组成': '#5a7a8a',
    '基础架构': '#5a7a8a',
    '骨架网络': '#5a7a8a',
    '仅用编码器': '#5a7a8a',
    '仅用解码器': '#5a7a8a',
    '变体基础': '#5a7a8a',
    '基础模型': '#5a7a8a',
    '取代': '#5a7a8a',
    '统一框架': '#5a7a8a',
    '理论基础': '#5a7a8a',
    '子领域': '#5a7a8a',
    '范式': '#5a7a8a',
    // 训练/优化 — 柔绿
    '训练': '#6a8a6a',
    '优化': '#6a8a6a',
    '正则化': '#6a8a6a',
    '加速推理': '#6a8a6a',
    '可组合': '#6a8a6a',
    '训练方式': '#6a8a6a',
    '超参数': '#6a8a6a',
    // 演进/改进 — 柔紫
    '改进': '#7a6a8a',
    '升级迭代': '#7a6a8a',
    '优化改进': '#7a6a8a',
    '变体': '#7a6a8a',
    '扩展': '#7a6a8a',
    '简化变体': '#7a6a8a',
    '增强': '#7a6a8a',
    '替代方案': '#7a6a8a',
    '提升推理': '#7a6a8a',
    '知识增强': '#7a6a8a',
    '推理增强': '#7a6a8a',
    '被超越': '#7a6a8a',
    // 同领域
    '同领域': '#5a6a6a',
    // 跨领域
    '跨领域': '#4a4a5a',
  }
  return colors[relation] || '#5a6a6a'
}

function applyCommunityFilter(_communityIds: number[]) {
  refreshNodes()
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

  graph = ForceGraph3D({ controlType: 'orbit' })(containerRef.value)
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
  refreshNodes()

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
  if (themeObserver) {
    themeObserver.disconnect()
    themeObserver = null
  }
})
</script>

<template>
  <div ref="containerRef" class="w-full h-full" />
</template>