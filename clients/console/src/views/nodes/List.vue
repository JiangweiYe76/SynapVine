<script setup lang="ts">
import { ref, onMounted, computed, watch, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'
import Layout from '../../components/Layout.vue'
import { CircleDot, Search, ChevronLeft, ChevronRight, Pencil, Trash2 } from '@lucide/vue'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { nodesAPI } from '@/api/nodes'
import type { Node } from '@/types/graph'
import NodeFormDialog from '@/components/NodeFormDialog.vue'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'

const nodes = ref<Node[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')
const currentPage = ref(0)
const pageSize = 20
const totalNodes = ref(0)

const formDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const selectedNode = ref<Node | null>(null)

const totalPages = computed(() => Math.ceil(totalNodes.value / pageSize))

function openCreateDialog() {
  selectedNode.value = null
  formDialogOpen.value = true
}

function openEditDialog(node: Node) {
  selectedNode.value = node
  formDialogOpen.value = true
}

function openDeleteDialog(node: Node) {
  selectedNode.value = node
  deleteDialogOpen.value = true
}

const columnHelper = createColumnHelper<Node>()

const columns = [
  columnHelper.accessor('id', {
    header: 'ID',
    cell: (info) =>
      h('span', { class: 'font-mono text-xs' }, info.getValue()),
  }),
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (info) =>
      h('span', { class: 'font-medium' }, info.getValue()),
  }),
  columnHelper.accessor('community_id', {
    header: 'Community',
    cell: (info) => {
      const value = info.getValue()
      return h('span', { class: 'font-mono text-xs text-muted-foreground' }, value ?? '—')
    },
  }),
  columnHelper.accessor('influence_score', {
    header: () => h('span', { class: 'block text-right' }, 'Score'),
    cell: (info) =>
      h('span', { class: 'block text-right' }, info.getValue().toFixed(1)),
  }),
  columnHelper.accessor('first_appeared', {
    header: () => h('span', { class: 'block text-right' }, 'First Appeared'),
    cell: (info) =>
      h('span', { class: 'block text-right' }, String(info.getValue())),
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const node = info.row.original
      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openEditDialog(node),
        }, () => h(Pencil, { class: 'h-4 w-4' })),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openDeleteDialog(node),
        }, () => h(Trash2, { class: 'h-4 w-4 text-destructive' })),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return nodes.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 0
    fetchNodes()
  }, 300)
})

async function fetchNodes() {
  loading.value = true
  error.value = null
  try {
    const res = await nodesAPI.list(currentPage.value * pageSize, pageSize, searchQuery.value)
    nodes.value = res.nodes
    totalNodes.value = res.pagination.total
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load nodes'
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (currentPage.value > 0) {
    currentPage.value--
    fetchNodes()
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value - 1) {
    currentPage.value++
    fetchNodes()
  }
}

function handleSaved() {
  fetchNodes()
}

function handleDeleted() {
  fetchNodes()
}

onMounted(fetchNodes)
</script>

<template>
  <Layout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Nodes</h2>
          <p class="text-muted-foreground">
            Manage knowledge graph nodes
          </p>
        </div>
        <Button @click="openCreateDialog">Add Node</Button>
      </div>

      <Card>
        <CardContent class="pt-6">
          <div class="flex items-center gap-4 mb-4">
            <div class="relative flex-1 max-w-sm">
              <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                placeholder="Search nodes..."
                class="pl-8"
              />
            </div>
            <p class="text-sm text-muted-foreground">
              {{ totalNodes }} nodes total
            </p>
          </div>

          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <CircleDot class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading nodes</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchNodes">Retry</Button>
          </div>

          <div v-else-if="table.getRowModel().rows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <CircleDot class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No nodes found</h3>
            <p class="text-sm text-muted-foreground mt-1">
              {{ searchQuery ? 'Try a different search term' : 'No nodes available' }}
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
                      'w-[120px]': header.id === 'id',
                      'w-[100px]': header.id === 'influence_score' || header.id === 'first_appeared' || header.id === 'actions',
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

    <NodeFormDialog
      :open="formDialogOpen"
      :node="selectedNode"
      @update:open="formDialogOpen = $event"
      @saved="handleSaved"
    />

    <DeleteConfirmDialog
      :open="deleteDialogOpen"
      :node="selectedNode"
      @update:open="deleteDialogOpen = $event"
      @deleted="handleDeleted"
    />
  </Layout>
</template>
