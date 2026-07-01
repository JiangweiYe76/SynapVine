<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'

import {
  FileText,
  ChevronLeft,
  ChevronRight,
  Pencil,
  Trash2,
  Eye,
} from '@lucide/vue'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { papersAPI } from '@/api/papers'
import { useAuthStore } from '@/stores/auth'
import { usePagination } from '@/composables/usePagination'
import type { Paper } from '@/types/paper'
import PaperFormDialog from '@/components/PaperFormDialog.vue'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'

const authStore = useAuthStore()
const papers = ref<Paper[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const formDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const selectedPaper = ref<Paper | null>(null)

const statusColors: Record<string, string> = {
  uploaded: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  analyzing: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  analyzed: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  reviewing: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
  merged: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200',
}

async function fetchPapers() {
  loading.value = true
  error.value = null
  try {
    const res = await papersAPI.list(pagination.offset.value, pagination.pageSize)
    papers.value = res.papers
    pagination.setTotalItems(res.total)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load papers'
  } finally {
    loading.value = false
  }
}

const pagination = usePagination({
  fetchFn: fetchPapers,
})

function openCreateDialog() {
  selectedPaper.value = null
  formDialogOpen.value = true
}

function openEditDialog(paper: Paper) {
  selectedPaper.value = paper
  formDialogOpen.value = true
}

function viewPaperPDF(paper: Paper) {
  window.open(papersAPI.pdfURL(paper.id), '_blank')
}

function openDeleteDialog(paper: Paper) {
  selectedPaper.value = paper
  deleteDialogOpen.value = true
}

const columnHelper = createColumnHelper<Paper>()

const columns = [
  columnHelper.accessor('title', {
    header: 'Title',
    cell: (info) =>
      h('span', { class: 'font-medium max-w-[300px] truncate block', title: info.getValue() }, info.getValue()),
  }),
  columnHelper.accessor('authors', {
    header: 'Authors',
    cell: (info) =>
      h('span', { class: 'text-sm text-muted-foreground max-w-[200px] truncate block', title: info.getValue() }, info.getValue() || '—'),
  }),
  columnHelper.accessor('status', {
    header: 'Status',
    cell: (info) => {
      const status = info.getValue()
      return h('span', {
        class: `inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[status] || ''}`,
      }, status)
    },
  }),
  columnHelper.accessor('created_at', {
    header: () => h('span', { class: 'block text-right' }, 'Created'),
    cell: (info) =>
      h('span', { class: 'block text-right text-sm text-muted-foreground' },
        new Date(info.getValue()).toLocaleDateString()),
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const paper = info.row.original
      if (!authStore.isEditor) return null
      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => viewPaperPDF(paper),
        }, () => h(Eye, { class: 'h-4 w-4' })),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openEditDialog(paper),
        }, () => h(Pencil, { class: 'h-4 w-4' })),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openDeleteDialog(paper),
        }, () => h(Trash2, { class: 'h-4 w-4 text-destructive' })),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return papers.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

function handleSaved() {
  fetchPapers()
}

onMounted(fetchPapers)
</script>

<template>

    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Papers</h2>
          <p class="text-muted-foreground">
            Upload and manage research papers for AI concept extraction
          </p>
        </div>
        <Button v-if="authStore.isEditor" @click="openCreateDialog">Upload Paper</Button>
      </div>

      <Card>
        <CardContent class="pt-6">
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <FileText class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading papers</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchPapers">Retry</Button>
          </div>

          <div v-else-if="table.getRowModel().rows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <FileText class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No papers uploaded</h3>
            <p class="text-sm text-muted-foreground mt-1">
              Upload a research paper to start AI concept extraction
            </p>
          </div>

          <div v-else class="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
                  <TableHead
                    v-for="header in headerGroup.headers"
                    :key="header.id"
                    :class="{
                      'w-[100px]': header.id === 'status' || header.id === 'actions',
                      'w-[120px]': header.id === 'created_at',
                    }"
                  >
                    <FlexRender
                      v-if="!header.isPlaceholder"
                      :render="header.column.columnDef.header"
                      :props="header.getContext()"
                    />
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="row in table.getRowModel().rows" :key="row.id">
                  <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                    <FlexRender
                      :render="cell.column.columnDef.cell"
                      :props="cell.getContext()"
                    />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <div v-if="!loading && !error && table.getRowModel().rows.length > 0" class="flex items-center justify-between mt-4">
            <p class="text-sm text-muted-foreground">
              {{ pagination.totalItems.value }} papers total — Page {{ pagination.currentPage.value + 1 }} of {{ pagination.totalPages }}
            </p>
            <div class="flex items-center gap-2">
              <Button variant="outline" size="sm" :disabled="pagination.currentPage.value === 0" @click="pagination.prevPage">
                <ChevronLeft class="h-4 w-4 mr-1" /> Previous
              </Button>
              <Button variant="outline" size="sm" :disabled="pagination.currentPage.value >= pagination.totalPages.value - 1" @click="pagination.nextPage">
                Next <ChevronRight class="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <PaperFormDialog
      :open="formDialogOpen"
      :paper="selectedPaper"
      @update:open="formDialogOpen = $event"
      @saved="handleSaved"
    />

    <DeleteConfirmDialog
      :open="deleteDialogOpen"
      title="Delete Paper"
      :description="`Are you sure you want to delete ${selectedPaper?.title}? This action cannot be undone.`"
      :delete-fn="async () => { if (selectedPaper) await papersAPI.delete(selectedPaper.id) }"
      @update:open="deleteDialogOpen = $event"
      @deleted="fetchPapers"
    />

</template>
