<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GraphStats } from '../types/graph'
const { t, locale } = useI18n()

const props = defineProps<{
  stats: GraphStats | null
  loading: boolean
}>()

// Format the last-updated timestamp using the active UI locale; the
// em dash placeholder covers both "loading" and "never mutated".
const lastUpdated = computed(() => {
  const raw = props.stats?.last_updated
  if (!raw) return '—'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
})
</script>

<template>
  <div class="h-14 bg-(--color-bg-secondary) border-t border-(--color-border-default) flex items-center px-8 gap-12 transition-colors duration-300">
    <div class="flex items-center gap-3">
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.nodes') }}:</span>
      <span class="text-base text-(--color-text-primary) font-mono tabular-nums font-medium">{{ stats?.total_nodes || 0 }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.edges') }}:</span>
      <span class="text-base text-(--color-text-primary) font-mono tabular-nums font-medium">{{ stats?.total_edges || 0 }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.communities') }}:</span>
      <span class="text-base text-(--color-text-primary) font-mono tabular-nums font-medium">{{ stats?.community_count || 0 }}</span>
    </div>
    <div class="flex items-center gap-3">
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.updated') }}:</span>
      <span class="text-base text-(--color-text-primary) font-mono tabular-nums font-medium">{{ lastUpdated }}</span>
    </div>
    <div v-if="loading" class="flex items-center gap-3 ml-auto">
      <div class="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.loading') }}</span>
    </div>
  </div>
</template>
