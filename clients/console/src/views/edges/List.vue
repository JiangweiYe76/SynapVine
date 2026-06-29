<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'

import { GitBranch, Search, ChevronLeft, ChevronRight, Pencil, Trash2 } from '@lucide/vue'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { edgesAPI } from '@/api/edges'
import { nodesAPI } from '@/api/nodes'
import { useAuthStore } from '@/stores/auth'
import type { Edge, Node } from '@/types/graph'
import EdgeFormDialog from '@/components/EdgeFormDialog.vue'
import EdgeDeleteConfirmDialog from '@/components/EdgeDeleteConfirmDialog.vue'

const authStore = useAuthStore()
const edges = ref<Edge[]>([])
const nodes = ref<Record<string, Node>>({})
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')
const currentPage = ref(0)
const pageSize = 20
const totalEdges = ref(0)

const formDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const selectedEdge = ref<Edge | null>(null)

const totalPages = computed(() => Math.ceil(totalEdges.value / pageSize))

function openCreateDialog() {
  selectedEdge.value = null
  formDialogOpen.value = true
}

function openEditDialog(edge: Edge) {
  selectedEdge.value = edge
  formDialogOpen.value = true
}

function openDeleteDialog(edge: Edge) {
  selectedEdge.value = edge
  deleteDialogOpen.value = true
}

async function loadNodes() {
  try {
    const res = await nodesAPI.list(0, 1000)
    const map: Record<string, Node> = {}
    for (const n of res.nodes) {
      map[n.id] = n
    }
    nodes.value = map
  } catch (e) {
    // silently ignore
  }
}

function nodeName(id: string): string {
  return nodes.value[id]?.name ?? id
}

const columnHelper = createColumnHelper<Edge>()

const columns = [
  columnHelper.accessor('source', {
    header: 'Source',
    cell: (info) =>
      h('span', { class: 'font-medium' }, nodeName(info.getValue())),
  }),
  columnHelper.accessor('target', {
    header: 'Target',
    cell: (info) =>
      h('span', { class: 'font-medium' }, nodeName(info.getValue())),
  }),
  columnHelper.accessor('relation', {
    header: 'Relation',
    cell: (info) =>
      h(Badge, { variant: 'outline' }, () => info.getValue()),
  }),
  columnHelper.accessor('weight', {
    header: () => h('span', { class: 'block text-right' }, 'Weight'),
    cell: (info) =>
      h('span', { class: 'block text-right' }, info.getValue().toFixed(2)),
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const edge = info.row.original
      // Mutation actions are gated on the role so a viewer never even
      // sees the buttons; the server enforces it again.
      if (!authStore.isEditor) return null
      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openEditDialog(edge),
        }, () => h(Pencil, { class: 'h-4 w-4' })),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openDeleteDialog(edge),
        }, () => h(Trash2, { class: 'h-4 w-4 text-destructive' })),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return edges.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 0
    fetchEdges()
  }, 300)
})

async function fetchEdges() {
  loading.value = true
  error.value = null
  try {
    const res = await edgesAPI.list(currentPage.value * pageSize, pageSize, searchQuery.value)
    edges.value = res.edges
    totalEdges.value = res.pagination.total
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load edges'
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (currentPage.value > 0) {
    currentPage.value--
    fetchEdges()
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value - 1) {
    currentPage.value++
    fetchEdges()
  }
}

function handleSaved() {
  fetchEdges()
}

function handleDeleted() {
  fetchEdges()
}

onMounted(() => {
  loadNodes()
  fetchEdges()
})
</script>

<template>

    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Edges</h2>
          <p class="text-muted-foreground">
            Manage knowledge graph relationships
          </p>
        </div>
        <Button v-if="authStore.isEditor" @click="openCreateDialog">Add Edge</Button>
      </div>

      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4 mb-4">
            <div class="relative flex-1 max-w-sm">
              <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                placeholder="Search edges..."
                class="pl-8"
              />
            </div>
            <p class="text-sm text-muted-foreground">
              {{ totalEdges }} edges total
            </p>
          </div>

          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <GitBranch class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading edges</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchEdges">Retry</Button>
          </div>

          <div v-else-if="table.getRowModel().rows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <GitBranch class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No edges found</h3>
            <p class="text-sm text-muted-foreground mt-1">
              {{ searchQuery ? 'Try a different search term' : 'No edges available' }}
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
                      'w-[100px]': header.id === 'weight' || header.id === 'actions',
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
              Page {{ currentPage + 1 }} of {{ totalPages }}
            </p>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                :disabled="currentPage === 0"
                @click="prevPage"
              >
                <ChevronLeft class="h-4 w-4 mr-1" />
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                :disabled="currentPage >= totalPages - 1"
                @click="nextPage"
              >
                Next
                <ChevronRight class="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <EdgeFormDialog
      :open="formDialogOpen"
      :edge="selectedEdge"
      @update:open="formDialogOpen = $event"
      @saved="handleSaved"
    />

    <EdgeDeleteConfirmDialog
      :open="deleteDialogOpen"
      :edge="selectedEdge"
      @update:open="deleteDialogOpen = $event"
      @deleted="handleDeleted"
    />

</template>
