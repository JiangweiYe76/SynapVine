<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()

const props = defineProps<{
  stats: { total_nodes: number; total_edges: number; community_count: number } | null
  loading: boolean
}>()
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
    <div v-if="loading" class="flex items-center gap-3 ml-auto">
      <div class="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
      <span class="text-sm text-(--color-text-secondary)">{{ t('statusBar.loading') }}</span>
    </div>
  </div>
</template>
