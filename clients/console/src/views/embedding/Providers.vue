<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  useVueTable,
  createColumnHelper,
  getCoreRowModel,
  FlexRender,
} from '@tanstack/vue-table'
import Layout from '../../components/Layout.vue'
import {
  Waves,
  Pencil,
  Trash2,
  Zap,
  CheckCircle2,
  XCircle,
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

import { embeddingAPI } from '@/api/embedding'
import { useAuthStore } from '@/stores/auth'
import type { EmbeddingProvider } from '@/types/embedding'
import EmbeddingProviderFormDialog from '@/components/EmbeddingProviderFormDialog.vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const authStore = useAuthStore()
const providers = ref<EmbeddingProvider[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const formDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const selectedProvider = ref<EmbeddingProvider | null>(null)

const testingId = ref<string | null>(null)
const testResult = ref<{ id: string; ok: boolean; message: string } | null>(null)

const deleting = ref(false)
const deleteError = ref<string | null>(null)

async function handleDelete() {
  if (!selectedProvider.value) return
  deleting.value = true
  deleteError.value = null
  try {
    await embeddingAPI.delete(selectedProvider.value.id)
    deleteDialogOpen.value = false
    fetchProviders()
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : 'Failed to delete provider'
  } finally {
    deleting.value = false
  }
}

function openCreateDialog() {
  selectedProvider.value = null
  formDialogOpen.value = true
}

function openEditDialog(provider: EmbeddingProvider) {
  selectedProvider.value = provider
  formDialogOpen.value = true
}

function openDeleteDialog(provider: EmbeddingProvider) {
  selectedProvider.value = provider
  deleteDialogOpen.value = true
}

async function handleTest(provider: EmbeddingProvider) {
  testingId.value = provider.id
  testResult.value = null
  try {
    const res = await embeddingAPI.test(provider.id)
    testResult.value = {
      id: provider.id,
      ok: res.ok,
      message: res.ok
        ? `Connected (${res.latency_ms}ms): ${res.dimensions}d`
        : `Failed: ${res.error}`,
    }
  } catch (e) {
    testResult.value = {
      id: provider.id,
      ok: false,
      message: e instanceof Error ? e.message : 'Test failed',
    }
  } finally {
    testingId.value = null
    setTimeout(() => {
      testResult.value = null
    }, 5000)
  }
}

async function handleToggleDefault(provider: EmbeddingProvider) {
  try {
    await embeddingAPI.update(provider.id, { is_default: !provider.is_default })
    await fetchProviders()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to update provider'
  }
}

const columnHelper = createColumnHelper<EmbeddingProvider>()

const columns = [
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (info) =>
      h('div', { class: 'flex items-center gap-2' }, [
        h(Waves, { class: 'h-4 w-4 text-muted-foreground' }),
        h('span', { class: 'font-medium' }, info.getValue()),
      ]),
  }),
  columnHelper.accessor('model', {
    header: 'Model',
    cell: (info) =>
      h('span', { class: 'font-mono text-xs' }, info.getValue()),
  }),
  columnHelper.accessor('dimensions', {
    header: 'Dimensions',
    cell: (info) =>
      h('span', { class: 'font-mono text-xs' }, String(info.getValue())),
  }),
  columnHelper.accessor('base_url', {
    header: 'Base URL',
    cell: (info) =>
      h('span', {
        class: 'font-mono text-xs text-muted-foreground max-w-[200px] truncate block',
        title: info.getValue(),
      }, info.getValue()),
  }),
  columnHelper.accessor('is_default', {
    header: 'Default',
    cell: (info) => {
      const isDefault = info.getValue()
      return h('button', {
        class: `inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors cursor-pointer ${
          isDefault
            ? 'bg-primary text-primary-foreground'
            : 'border border-input bg-background hover:bg-accent hover:text-accent-foreground'
        }`,
        onClick: () => handleToggleDefault(info.row.original),
      }, isDefault ? 'Default' : 'Set default')
    },
  }),
  columnHelper.accessor('is_enabled', {
    header: 'Status',
    cell: (info) => {
      const enabled = info.getValue()
      return h('div', { class: 'flex items-center gap-1.5' }, [
        enabled
          ? h(CheckCircle2, { class: 'h-4 w-4 text-green-500' })
          : h(XCircle, { class: 'h-4 w-4 text-muted-foreground' }),
        h('span', { class: 'text-sm' }, enabled ? 'Enabled' : 'Disabled'),
      ])
    },
  }),
  columnHelper.display({
    id: 'actions',
    header: '',
    cell: (info) => {
      const provider = info.row.original
      if (!authStore.isEditor) return null

      const isTesting = testingId.value === provider.id
      const result = testResult.value?.id === provider.id ? testResult.value : null

      return h('div', { class: 'flex items-center justify-end gap-1' }, [
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          disabled: isTesting,
          title: 'Test connectivity',
          onClick: () => handleTest(provider),
        }, () => isTesting
          ? h('div', { class: 'animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full' })
          : h(Zap, { class: `h-4 w-4 ${result?.ok ? 'text-green-500' : result && !result.ok ? 'text-destructive' : ''}` })
        ),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openEditDialog(provider),
        }, () => h(Pencil, { class: 'h-4 w-4' })),
        h(Button, {
          variant: 'ghost',
          size: 'icon-sm',
          onClick: () => openDeleteDialog(provider),
        }, () => h(Trash2, { class: 'h-4 w-4 text-destructive' })),
      ])
    },
  }),
]

const table = useVueTable({
  get data() {
    return providers.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})

async function fetchProviders() {
  loading.value = true
  error.value = null
  try {
    const res = await embeddingAPI.list()
    providers.value = res.providers
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load providers'
  } finally {
    loading.value = false
  }
}

function handleSaved() {
  fetchProviders()
}

onMounted(fetchProviders)
</script>

<template>
  <Layout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Embedding Providers</h2>
          <p class="text-muted-foreground">
            Manage embedding model providers for node similarity and search
          </p>
        </div>
        <Button v-if="authStore.isEditor" @click="openCreateDialog">Add Provider</Button>
      </div>

      <!-- Test result toast -->
      <div
        v-if="testResult"
        class="rounded-lg border p-4 text-sm"
        :class="testResult.ok ? 'bg-green-50 border-green-200 text-green-800 dark:bg-green-950 dark:border-green-800 dark:text-green-200' : 'bg-destructive/10 border-destructive/20 text-destructive'"
      >
        {{ testResult.message }}
      </div>

      <Card>
        <CardContent class="pt-6">
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>

          <div v-else-if="error" class="flex flex-col items-center justify-center py-12 text-center">
            <Waves class="h-12 w-12 text-destructive/50 mb-4" />
            <h3 class="text-lg font-medium">Error loading providers</h3>
            <p class="text-sm text-muted-foreground mt-1">{{ error }}</p>
            <Button variant="outline" class="mt-4" @click="fetchProviders">Retry</Button>
          </div>

          <div v-else-if="table.getRowModel().rows.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
            <Waves class="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-medium">No embedding providers configured</h3>
            <p class="text-sm text-muted-foreground mt-1">
              Add a provider to enable embedding-based features
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
                      'w-[100px]': header.id === 'is_default' || header.id === 'is_enabled',
                      'w-[140px]': header.id === 'actions',
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

    <EmbeddingProviderFormDialog
      :open="formDialogOpen"
      :provider="selectedProvider"
      @update:open="formDialogOpen = $event"
      @saved="handleSaved"
    />

    <Dialog :open="deleteDialogOpen" @update:open="deleteDialogOpen = $event">
      <DialogContent class="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Delete Provider</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete <strong>{{ selectedProvider?.name }}</strong>?
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <p v-if="deleteError" class="text-sm text-destructive">{{ deleteError }}</p>

        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">Cancel</Button>
          <Button variant="destructive" :disabled="deleting" @click="handleDelete">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </Layout>
</template>
