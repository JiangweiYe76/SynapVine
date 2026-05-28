<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SearchResult } from '../types/graph'
import { Search as SearchIcon } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'

const { t } = useI18n()

const props = defineProps<{
  onSearch: (query: string) => Promise<void>
  onSelect: (id: string) => void
  results: SearchResult[]
  loading: boolean
}>()

const query = ref('')
const showResults = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(query, (newQuery) => {
  if (debounceTimer) clearTimeout(debounceTimer)

  if (!newQuery.trim()) {
    showResults.value = false
    return
  }

  debounceTimer = setTimeout(async () => {
    await props.onSearch(newQuery)
    showResults.value = true
  }, 300)
})

function handleSelect(result: SearchResult) {
  query.value = result.name
  showResults.value = false
  props.onSelect(result.id)
}

function handleBlur() {
  setTimeout(() => {
    showResults.value = false
  }, 200)
}
</script>

<template>
  <div class="relative w-96">
    <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10 pointer-events-none" />
    <Input
      v-model="query"
      :placeholder="t('search.placeholder')"
      class="pl-10 pr-10"
      @blur="handleBlur"
    />
    <div v-if="loading" class="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />

    <div
      v-if="showResults && results.length > 0"
      class="absolute top-full left-0 right-0 mt-2 bg-(--color-bg-secondary) border-(--color-border-default) rounded-xl shadow-2xl max-h-72 overflow-y-auto z-50 transition-colors duration-300"
    >
      <div
        v-for="result in results"
        :key="result.id"
        class="px-5 py-4 hover:bg-(--color-bg-tertiary) cursor-pointer border-b border-(--color-border-muted) last:border-b-0 transition-colors"
        @mousedown.prevent
        @click="handleSelect(result)"
      >
        <div class="text-base font-medium text-(--color-text-primary)">{{ result.name }}</div>
        <div class="text-sm text-(--color-text-secondary) mt-1">{{ result.highlight }}</div>
      </div>
    </div>

    <div
      v-else-if="showResults && results.length === 0 && query.trim()"
      class="absolute top-full left-0 right-0 mt-2 bg-(--color-bg-secondary) border-(--color-border-default) rounded-xl shadow-2xl p-6 z-50 transition-colors duration-300"
    >
      <div class="text-base text-(--color-text-secondary) text-center">{{ t('search.noResults') }}</div>
    </div>
  </div>
</template>
