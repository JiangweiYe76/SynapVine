<script setup lang="ts">
import { ref, onMounted, provide, computed } from 'vue'
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

onMounted(() => {
  loadInitial()
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
        <svg class="w-7 h-7 text-[#58a6ff]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
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
          <svg v-if="theme.theme.value === 'dark'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
          </svg>
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
          <svg v-if="timelineVisible" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            <line stroke-linecap="round" stroke-width="2" x1="2" y1="2" x2="22" y2="22" />
          </svg>
        </button>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="showSettings = true"
          :title="t('settings.title')"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
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
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
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