<script setup lang="ts">
import { ref, onMounted, onUnmounted, provide, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GraphCanvas from './components/GraphCanvas.vue'
import SearchBar from './components/SearchBar.vue'
import CommunityLegend from './components/CommunityLegend.vue'
import NodeDetail from './components/NodeDetail.vue'
import StatusBar from './components/StatusBar.vue'
import TimelineControl from './components/TimelineControl.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import { useGraph } from './composables/useGraph'
import { useTimeline } from './composables/useTimeline'
import { useTheme } from './composables/useTheme'
import type { SearchResult, NodeDetail as NodeDetailType } from './types/graph'
import { searchNodes, getNodeDetail } from './api/graph'
import { Zap, Sun, Moon, Clock, Settings, X } from 'lucide-vue-next'

const {
  nodes,
  edges,
  communities,
  stats,
  selectedNode,
  highlightedCommunity,
  loading,
  error,
  loadInitial,
  selectNode,
  highlightCommunity,
  clearError,
} = useGraph()

const theme = useTheme()
provide('theme', theme)

const { t, locale } = useI18n()

const timeline = useTimeline(nodes, edges)
provide('timeline', timeline)

const searchResults = ref<SearchResult[]>([])
const nodeDetail = ref<NodeDetailType | null>(null)
const showSettings = ref(false)
const timelineVisible = ref(true)

const isInitialLoading = computed(() => loading.value && nodes.value.length === 0)

function toggleLocale() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && nodeDetail.value) {
    handleCloseDetail()
  }
}

onMounted(() => {
  loadInitial()
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

async function handleSearch(query: string) {
  const response = await searchNodes(query, 10)
  searchResults.value = response.results
}

async function handleSelectNode(id: string) {
  await selectNode(id)
  const detail = await getNodeDetail(id)
  nodeDetail.value = detail
}

async function handleNodeClick(id: string) {
  await selectNode(id)
  const detail = await getNodeDetail(id)
  setTimeout(() => {
    nodeDetail.value = detail
  }, 0)
}

function handleBackgroundClick() {
  selectNode(null)
  nodeDetail.value = null
}

function handleCloseDetail() {
  selectNode(null)
  nodeDetail.value = null
}

async function handleNeighborClick(id: string) {
  await handleSelectNode(id)
}

function handleCommunityHighlight(communityId: number | null) {
  highlightCommunity(communityId)
}
</script>

<template>
  <div class="flex flex-col h-screen bg-(--color-bg-primary) transition-colors duration-300">
    <header class="h-16 bg-(--color-bg-secondary) border-b border-(--color-border-default) flex items-center justify-between px-8 shrink-0 transition-colors duration-300">
      <div class="flex items-center gap-4">
        <Zap class="w-7 h-7 text-[#58a6ff]" />
        <h1 class="text-xl font-semibold text-(--color-text-primary)">{{ t('app.title') }}</h1>
      </div>
      <div class="flex items-center gap-4">
        <SearchBar
          :on-search="handleSearch"
          :on-select="handleSelectNode"
          :results="searchResults"
          :loading="loading"
        />
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="theme.toggleTheme"
          :title="theme.theme.value === 'dark' ? t('theme.switchToLight') : t('theme.switchToDark')"
        >
          <Sun v-if="theme.theme.value === 'dark'" class="w-5 h-5" />
          <Moon v-else class="w-5 h-5" />
        </button>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) text-xs font-semibold transition-colors"
          @click="toggleLocale"
          :title="locale === 'zh-CN' ? 'Switch to English' : '切换到中文'"
        >
          {{ locale === 'zh-CN' ? 'EN' : '中' }}
        </button>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="timelineVisible = !timelineVisible"
          :title="timelineVisible ? t('timeline.hide') : t('timeline.show')"
        >
          <Clock v-if="timelineVisible" class="w-5 h-5" />
          <Clock v-else class="w-5 h-5 opacity-40" />
        </button>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="showSettings = true"
          :title="t('settings.title')"
        >
          <Settings class="w-5 h-5" />
        </button>
      </div>
    </header>

    <div class="flex-1 flex relative overflow-hidden">
      <CommunityLegend
        :communities="communities"
        :highlighted-community="highlightedCommunity"
        @highlight="handleCommunityHighlight"
      />

      <GraphCanvas
        :nodes="timelineVisible ? timeline.visibleNodes.value : nodes"
        :edges="timelineVisible ? timeline.visibleEdges.value : edges"
        :communities="communities"
        :selected-node="selectedNode"
        :highlighted-community="highlightedCommunity"
        :highlighted-neighbors="new Set()"
        @node-click="handleNodeClick"
        @background-click="handleBackgroundClick"
      />

      <NodeDetail
        :detail="nodeDetail"
        :communities="communities"
        @close="handleCloseDetail"
        @neighbor-click="handleNeighborClick"
      />

      <div
        v-if="isInitialLoading"
        class="absolute inset-0 bg-(--color-bg-primary)/80 flex items-center justify-center z-50"
      >
        <div class="flex flex-col items-center gap-4">
          <div class="w-12 h-12 border-4 border-[#58a6ff] border-t-transparent rounded-full animate-spin" />
          <span class="text-sm text-(--color-text-secondary)">{{ t('app.loading') }}</span>
        </div>
      </div>

      <Transition name="fade">
        <div
          v-if="error"
          class="absolute bottom-4 left-1/2 -translate-x-1/2 bg-red-500/90 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-3 z-50"
        >
          <span class="text-sm">{{ error }}</span>
          <button
            class="text-white/80 hover:text-white transition-colors"
            @click="clearError"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </Transition>
    </div>

    <Transition name="timeline-slide">
      <TimelineControl v-if="timelineVisible" />
    </Transition>
    <StatusBar :stats="stats" :loading="loading" />
    <SettingsDialog v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.timeline-slide-enter-active,
.timeline-slide-leave-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.timeline-slide-enter-from,
.timeline-slide-leave-to {
  opacity: 0;
  transform: translateY(100%);
}
</style>