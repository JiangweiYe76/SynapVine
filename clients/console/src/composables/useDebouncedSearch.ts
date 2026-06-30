import { ref, watch, onUnmounted } from 'vue'

export interface UseDebouncedSearchOptions {
  debounceMs?: number
  onSearch: () => void
  onResetPage?: () => void
}

export function useDebouncedSearch(options: UseDebouncedSearchOptions) {
  const { debounceMs = 300, onSearch, onResetPage } = options

  const searchQuery = ref('')
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  watch(searchQuery, () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      onResetPage?.()
      onSearch()
    }, debounceMs)
  })

  onUnmounted(() => {
    if (searchTimer) {
      clearTimeout(searchTimer)
      searchTimer = null
    }
  })

  return {
    searchQuery,
  }
}
