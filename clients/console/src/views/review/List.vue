<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'

import {
  ClipboardCheck,
  ChevronLeft,
  ChevronRight,
  CheckCircle2,
  XCircle,
  Clock,
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { reviewAPI } from '@/api/review'
import { useAuthStore } from '@/stores/auth'
import { usePagination } from '@/composables/usePagination'
import type { ReviewQueueItem, ReviewStatus } from '@/types/paper'

const authStore = useAuthStore()
const items = ref<ReviewQueueItem[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const statusFilter = ref<ReviewStatus | ''>('')

// Detail dialog
const detailOpen = ref(false)
const selectedItem = ref<ReviewQueueItem | null>(null)

// Action dialog
const actionDialogOpen = ref(false)
const actionType = ref<'approve' | 'reject'>('approve')
const actionNotes = ref('')
const actionLoading = ref(false)
const actionError = ref<string | null>(null)

const statusIcons: Record<string, any> = {
  pending: Clock,
  approved: CheckCircle2,
  rejected: XCircle,
}

const statusColors: Record<string, string> = {
  pending: 'text-yellow-500',
  approved: 'text-green-500',
  rejected: 'text-destructive',
}

function openDetail(item: ReviewQueueItem) {
  selectedItem.value = item
  detailOpen.value = true
}

function openActionDialog(item: ReviewQueueItem, type: 'approve' | 'reject') {
  selectedItem.value = item
  actionType.value = type
  actionNotes.value = ''
  actionError.value = null
  actionDialogOpen.value = true
}

async function handleAction() {
  if (!selectedItem.value) return
  actionLoading.value = true
  actionError.value = null
  try {
    const userId = authStore.user?.id || 'unknown'
    if (actionType.value === 'approve') {
      await reviewAPI.approve(selectedItem.value.id, userId, actionNotes.value)
    } else {
      await reviewAPI.reject(selectedItem.value.id, userId, actionNotes.value)
    }
    actionDialogOpen.value = false
    detailOpen.value = false
    fetchItems()
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : 'Action failed'
  } finally {
    actionLoading.value = false
  }
}

async function fetchItems() {
  loading.value = true
  error.value = null
  try {
    const res = await reviewAPI.list(pagination.offset.value, pagination.pageSize, statusFilter.value)
    items.value = res.items
    pagination.setTotalItems(res.total)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load review queue'
  } finally {
    loading.value = false
  }
}

const pagination = usePagination({
  fetchFn: fetchItems,
})

function handleStatusFilterChange(status: ReviewStatus | '') {
  statusFilter.value = status
  pagination.resetPage()
  fetchItems()
}

const columnHelper = createColumnHelper<ReviewQueueItem>()

const columns = [
  columnHelper.accessor('paper_id', {
    header: 'Paper ID',
    cell: (info) =>
      h('span', { class: 'font-mono text-xs' }, info.getValue().slice(0, 8) + '...'),
  }),
  columnHelper.accessor('created_at', {
    header: 'Submitted',
    cell: (info) =>
      h('span', { class: 'text-sm text-muted-foreground' },
        new Date(info.getValue()).toLocaleDateString()),
  }),
  columnHelper.accessor('status', {
    header: 'Status',
    cell: (info) => {
      const status = info.getValue()
      const icon = statusIcons[status] || Clock
      const color = statusColors[status] || ''
      return h('div', { class: 'flex items-center gap-1.5' }, [
        h(icon, { class: `h-4 w-4 ${color}` }),
        h('span', { class: 'text-sm capitalize' }, status),
      ])
    },
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const item = info.row.original
      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          title: 'View details',
          onClick: () => openDetail(item),
        }, () => h(Eye, { class: 'h-4 w-4' })),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return items.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

function parseNodes(item: ReviewQueueItem) {
  if (Array.isArray(item.extracted_nodes)) return item.extracted_nodes
  try { return JSON.parse(item.extracted_nodes as any) } catch { return [] }
}

function parseEdges(item: ReviewQueueItem) {
  if (Array.isArray(item.extracted_edges)) return item.extracted_edges
  try { return JSON.parse(item.extracted_edges as any) } catch { return [] }
}

onMounted(fetchItems)
</script>

<template>

    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Review Queue</h2>
          <p class="text-muted-foreground">
            Review AI-extracted concepts before merging into the graph
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            v-for="opt in ['', 'pending', 'approved', 'rejected']"
            :key="opt"
            class="inline-flex items-center rounded-full px-3 py-1 text-xs font-medium transition-colors cursor-pointer"
            :class="statusFilter === opt ? 'bg-primary text-primary-foreground' : 'border border-input bg-background hover:bg-accent hover:text-accent-foreground'"
            @click="handleStatusFilterChange(opt as ReviewStatus | '')"
          >
            {{ opt || 'All' }}
          </button>
        </div>
      </div>

      <Card>
        <CardContent class="pt-6">
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <ClipboardCheck class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading review queue</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchItems">Retry</Button>
          </div>

          <div v-else-if="table.getRowModel().rows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <ClipboardCheck class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No items in queue</h3>
            <p class="text-sm text-muted-foreground mt-1">
              {{ statusFilter ? `No ${statusFilter} items` : 'The review queue is empty' }}
            </p>
          </div>

          <div v-else class="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
                  <TableHead v-for="header in headerGroup.headers" :key="header.id">
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
              {{ pagination.totalItems.value }} items — Page {{ pagination.currentPage.value + 1 }} of {{ pagination.totalPages }}
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

    <!-- Detail Dialog -->
    <Dialog :open="detailOpen" @update:open="detailOpen = $event">
      <DialogContent class="sm:max-w-3xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Review Item</DialogTitle>
          <DialogDescription>
            Paper ID: {{ selectedItem?.paper_id }}
          </DialogDescription>
        </DialogHeader>

        <div v-if="selectedItem" class="space-y-4">
          <!-- Extracted Nodes -->
          <div>
            <h4 class="text-sm font-medium mb-2">Extracted Nodes ({{ parseNodes(selectedItem).length }})</h4>
            <div class="rounded-md border overflow-x-auto">
              <table class="w-full text-sm">
                <thead class="bg-muted">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium">Name</th>
                    <th class="px-3 py-2 text-left font-medium">Description</th>
                    <th class="px-3 py-2 text-right font-medium">Relevance</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(node, i) in parseNodes(selectedItem)" :key="i" class="border-t">
                    <td class="px-3 py-2 font-medium">{{ node.name }}</td>
                    <td class="px-3 py-2 text-muted-foreground">{{ node.description }}</td>
                    <td class="px-3 py-2 text-right">{{ node.relevance }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Extracted Edges -->
          <div>
            <h4 class="text-sm font-medium mb-2">Extracted Edges ({{ parseEdges(selectedItem).length }})</h4>
            <div class="rounded-md border overflow-x-auto">
              <table class="w-full text-sm">
                <thead class="bg-muted">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium">Source</th>
                    <th class="px-3 py-2 text-left font-medium">Target</th>
                    <th class="px-3 py-2 text-left font-medium">Relation</th>
                    <th class="px-3 py-2 text-right font-medium">Weight</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(edge, i) in parseEdges(selectedItem)" :key="i" class="border-t">
                    <td class="px-3 py-2 font-medium">{{ edge.source }}</td>
                    <td class="px-3 py-2 font-medium">{{ edge.target }}</td>
                    <td class="px-3 py-2 text-muted-foreground">{{ edge.relation }}</td>
                    <td class="px-3 py-2 text-right">{{ edge.weight }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <DialogFooter v-if="selectedItem?.status === 'pending' && authStore.isEditor">
          <Button variant="outline" @click="detailOpen = false">Close</Button>
          <Button variant="destructive" @click="openActionDialog(selectedItem!, 'reject')">Reject</Button>
          <Button @click="openActionDialog(selectedItem!, 'approve')">Approve</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Action Dialog (approve/reject) -->
    <Dialog :open="actionDialogOpen" @update:open="actionDialogOpen = $event">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ actionType === 'approve' ? 'Approve' : 'Reject' }} Review Item</DialogTitle>
          <DialogDescription>
            {{ actionType === 'approve'
              ? 'The extracted nodes and edges will be merged into the graph.'
              : 'This extraction will be discarded.' }}
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-2">
          <Label for="action-notes">Notes (optional)</Label>
          <Textarea
            id="action-notes"
            v-model="actionNotes"
            placeholder="Add review notes..."
            rows="3"
          />
        </div>

        <p v-if="actionError" class="text-sm text-destructive">{{ actionError }}</p>

        <DialogFooter>
          <Button variant="outline" @click="actionDialogOpen = false">Cancel</Button>
          <Button
            :variant="actionType === 'approve' ? 'default' : 'destructive'"
            :disabled="actionLoading"
            @click="handleAction"
          >
            {{ actionLoading ? 'Processing...' : (actionType === 'approve' ? 'Approve' : 'Reject') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

</template>
