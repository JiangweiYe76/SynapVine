<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, defineAsyncComponent, h } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchBar from './components/SearchBar.vue'
import CommunityLegend from './components/CommunityLegend.vue'
import NodeDetail from './components/NodeDetail.vue'
import StatusBar from './components/StatusBar.vue'
import TimelineControl from './components/TimelineControl.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import { provideGraph } from './composables/useGraph'
import { provideTheme } from './composables/useTheme'
import type { SearchResult, NodeDetail as NodeDetailType } from './types/graph'
import { searchNodes, getNodeDetail } from './api/graph'
import { Zap, Sun, Moon, Settings, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

const theme = provideTheme()

// Lazy-load the 3D graph canvas so three.js and 3d-force-graph are split
// into a separate chunk that does not block first paint.
const GraphCanvas = defineAsyncComponent({
  loader: () => import('./components/GraphCanvas.vue'),
  delay: 200,
  loadingComponent: () =>
    h('div', { class: 'flex-1 flex items-center justify-center' }, [
      h('div', { class: 'w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin' }),
    ]),
})

const { t, locale } = useI18n()

const {
  nodes,
  edges,
  communities,
  stats,
  selectedNode,
  highlightedCommunity,
  loading,
  error,
  timeline,
  loadInitial,
  selectNode,
  highlightCommunity,
  clearError,
} = provideGraph()

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
  <TooltipProvider>
    <div class="flex flex-col h-screen bg-(--color-bg-primary) transition-colors duration-300">
    <header class="h-16 bg-(--color-bg-secondary) border-b border-(--color-border-default) flex items-center justify-between px-8 shrink-0 transition-colors duration-300">
      <div class="flex items-center gap-4">
        <Zap class="w-7 h-7 text-primary" />
        <h1 class="text-xl font-semibold text-(--color-text-primary)">{{ t('app.title') }}</h1>
      </div>
      <div class="flex items-center gap-4">
        <SearchBar
          :on-search="handleSearch"
          :on-select="handleSelectNode"
          :results="searchResults"
          :loading="loading"
        />
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              @click="theme.toggleTheme"
            >
              <Sun v-if="theme.theme.value === 'dark'" class="w-5 h-5" />
              <Moon v-else class="w-5 h-5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{{ theme.theme.value === 'dark' ? t('theme.switchToLight') : t('theme.switchToDark') }}</p>
          </TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              class="w-10 h-10 text-xs font-semibold"
              @click="toggleLocale"
            >
              {{ locale === 'zh-CN' ? 'EN' : '中' }}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{{ locale === 'zh-CN' ? 'Switch to English' : '切换到中文' }}</p>
          </TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              @click="showSettings = true"
            >
              <Settings class="w-5 h-5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{{ t('settings.title') }}</p>
          </TooltipContent>
        </Tooltip>
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
          <div class="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          <span class="text-sm text-(--color-text-secondary)">{{ t('app.loading') }}</span>
        </div>
      </div>

      <Transition name="fade">
        <div
          v-if="error"
          class="absolute bottom-4 left-1/2 -translate-x-1/2 bg-red-500/90 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-3 z-50"
        >
          <span class="text-sm">{{ error }}</span>
          <Button
            variant="ghost"
            class="text-white/80 hover:text-white h-auto w-auto p-0"
            @click="clearError"
          >
            <X class="w-4 h-4" />
          </Button>
        </div>
      </Transition>
    </div>

    <TimelineControl v-model="timelineVisible" />
    <StatusBar :stats="stats" :loading="loading" />
    <SettingsDialog v-model:open="showSettings" />
  </div>
  </TooltipProvider>
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


</style>