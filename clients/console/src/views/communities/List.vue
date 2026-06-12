<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'
import Layout from '../../components/Layout.vue'
import { Layers, ChevronRight, ChevronDown, Pencil, Trash2, Plus } from '@lucide/vue'
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
import { Badge } from '@/components/ui/badge'
import { communitiesAPI } from '@/api/communities'
import type { HierarchicalCommunity } from '@/types/graph'
import CommunityFormDialog from '@/components/CommunityFormDialog.vue'
import CommunityDeleteConfirmDialog from '@/components/CommunityDeleteConfirmDialog.vue'

interface FlatRow {
  id: string
  name: string
  color: string
  level: number
  parentId: string | null
  nodeCount: number
  depth: number
  hasChildren: boolean
  expanded: boolean
  children: FlatRow[]
}

const tree = ref<HierarchicalCommunity[]>([])
const flatRows = ref<FlatRow[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const formDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const selectedCommunityId = ref<string | null>(null)

const expandedIds = ref(new Set<string>())

function buildFlatRows(
  nodes: HierarchicalCommunity[],
  depth = 0,
  out: FlatRow[] = [],
  parentStack: string[] = [],
): FlatRow[] {
  for (const n of nodes) {
    const flat: FlatRow = {
      id: n.id,
      name: n.name,
      color: n.color,
      level: n.level,
      parentId: n.parent_id,
      nodeCount: n.node_count,
      depth,
      hasChildren: (n.children?.length ?? 0) > 0,
      expanded: expandedIds.value.has(n.id) || depth === 0,
      children: [],
    }
    out.push(flat)
    if (flat.expanded && n.children?.length) {
      buildFlatRows(n.children, depth + 1, out, [...parentStack, n.id])
    }
  }
  return out
}

function recomputeFlat() {
  flatRows.value = buildFlatRows(tree.value)
}

function toggleExpand(id: string) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
  expandedIds.value = new Set(expandedIds.value)
  recomputeFlat()
}

async function fetchTree() {
  loading.value = true
  error.value = null
  try {
    const res = await communitiesAPI.tree()
    tree.value = res.communities
    recomputeFlat()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load communities'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  selectedCommunityId.value = null
  formDialogOpen.value = true
}

function openEditDialog(id: string) {
  selectedCommunityId.value = id
  formDialogOpen.value = true
}

function openDeleteDialog(id: string) {
  selectedCommunityId.value = id
  deleteDialogOpen.value = true
}

const selectedCommunity = computed(() => {
  if (selectedCommunityId.value === null) return null
  return findById(tree.value, selectedCommunityId.value)
})

function findById(nodes: HierarchicalCommunity[], id: string): HierarchicalCommunity | null {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children?.length) {
      const r = findById(n.children, id)
      if (r) return r
    }
  }
  return null
}

const columnHelper = createColumnHelper<FlatRow>()

const columns = [
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (info) => {
      const row = info.row.original
      return h('div', { class: 'flex items-center gap-2', style: { paddingLeft: `${row.depth * 20}px` } }, [
        h(
          'button',
          {
            type: 'button',
            class: 'inline-flex h-5 w-5 items-center justify-center text-muted-foreground hover:text-foreground',
            onClick: () => row.hasChildren && toggleExpand(row.id),
            disabled: !row.hasChildren,
          },
          row.hasChildren
            ? h(row.expanded ? ChevronDown : ChevronRight, { class: 'h-4 w-4' })
            : h('span', { class: 'block h-4 w-4' }),
        ),
        h('span', {
          class: 'inline-block h-3 w-3 rounded-sm border',
          style: { backgroundColor: row.color },
        }),
        h('span', { class: 'font-medium' }, row.name),
      ])
    },
  }),
  columnHelper.accessor('id', {
    header: 'ID',
    cell: (info) => h('span', { class: 'font-mono text-xs' }, String(info.getValue())),
  }),
  columnHelper.accessor('level', {
    header: 'Level',
    cell: (info) => {
      const lvl = info.getValue()
      return h(Badge, { variant: 'outline' }, () => `L${lvl}`)
    },
  }),
  columnHelper.accessor('parentId', {
    header: 'Parent',
    cell: (info) => {
      const v = info.getValue()
      return v === null
        ? h('span', { class: 'text-muted-foreground' }, '—')
        : h('span', { class: 'font-mono text-xs' }, String(v))
    },
  }),
  columnHelper.accessor('nodeCount', {
    header: () => h('span', { class: 'block text-right' }, 'Nodes'),
    cell: (info) => h('span', { class: 'block text-right' }, String(info.getValue())),
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const row = info.row.original
      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'icon-sm',
            onClick: () => openEditDialog(row.id),
          },
          () => h(Pencil, { class: 'h-4 w-4' }),
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'icon-sm',
            onClick: () => openDeleteDialog(row.id),
          },
          () => h(Trash2, { class: 'h-4 w-4 text-destructive' }),
        ),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return flatRows.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

function handleSaved() {
  fetchTree()
}

function handleDeleted() {
  fetchTree()
}

onMounted(fetchTree)
</script>

<template>
  <Layout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Communities</h2>
          <p class="text-muted-foreground">Manage hierarchical community structure</p>
        </div>
        <Button @click="openCreateDialog">
          <Plus class="h-4 w-4 mr-1" />
          Add Community
        </Button>
      </div>

      <Card>
        <CardContent class="pt-6">
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <Layers class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading communities</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchTree">Retry</Button>
          </div>

          <div v-else-if="flatRows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <Layers class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No communities found</h3>
            <p class="text-sm text-muted-foreground mt-1">Create your first community to get started</p>
          </div>

          <div v-else class="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
                  <TableHead
                    v-for="header in headerGroup.headers"
                    :key="header.id"
                    :class="{
                      'w-[100px]': header.id === 'id',
                      'w-[80px]': header.id === 'level' || header.id === 'parentId' || header.id === 'nodeCount' || header.id === 'actions',
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
        </CardContent>
      </Card>
    </div>

    <CommunityFormDialog
      :open="formDialogOpen"
      :community="selectedCommunity"
      :tree="tree"
      @update:open="formDialogOpen = $event"
      @saved="handleSaved"
    />

    <CommunityDeleteConfirmDialog
      :open="deleteDialogOpen"
      :community-id="selectedCommunityId"
      :community-name="selectedCommunity?.name"
      @update:open="deleteDialogOpen = $event"
      @deleted="handleDeleted"
    />
  </Layout>
</template>
