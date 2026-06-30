import { ref, computed } from 'vue'

export interface UsePaginationOptions {
  pageSize?: number
  fetchFn: () => Promise<void>
}

export function usePagination(options: UsePaginationOptions) {
  const { pageSize = 20, fetchFn } = options

  const currentPage = ref(0)
  const totalItems = ref(0)

  const totalPages = computed(() => Math.ceil(totalItems.value / pageSize))

  function prevPage() {
    if (currentPage.value > 0) {
      currentPage.value--
      fetchFn()
    }
  }

  function nextPage() {
    if (currentPage.value < totalPages.value - 1) {
      currentPage.value++
      fetchFn()
    }
  }

  function resetPage() {
    currentPage.value = 0
  }

  function setTotalItems(total: number) {
    totalItems.value = total
  }

  const offset = computed(() => currentPage.value * pageSize)

  return {
    currentPage,
    totalItems,
    totalPages,
    pageSize,
    offset,
    prevPage,
    nextPage,
    resetPage,
    setTotalItems,
  }
}
