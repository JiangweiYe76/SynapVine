<script setup lang="ts">
import { ref } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { nodesAPI } from '@/api/nodes'
import type { Node } from '@/types/graph'

const props = defineProps<{
  open: boolean
  node?: Node | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'deleted'): void
}>()

const deleting = ref(false)
const deleteError = ref<string | null>(null)

async function handleDelete() {
  if (!props.node) return

  deleting.value = true
  deleteError.value = null

  try {
    await nodesAPI.delete(props.node.id)
    emit('deleted')
    emit('update:open', false)
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : 'Failed to delete node'
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>Delete Node</DialogTitle>
        <DialogDescription>
          Are you sure you want to delete <strong>{{ node?.name }}</strong>?
          This action cannot be undone.
        </DialogDescription>
      </DialogHeader>

      <p v-if="deleteError" class="text-sm text-destructive">{{ deleteError }}</p>

      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">Cancel</Button>
        <Button variant="destructive" :disabled="deleting" @click="handleDelete">
          {{ deleting ? 'Deleting...' : 'Delete' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
