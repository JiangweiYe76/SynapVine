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

const props = defineProps<{
  open: boolean
  title?: string
  description?: string
  deleteFn: () => Promise<void>
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'deleted'): void
}>()

const deleting = ref(false)
const deleteError = ref<string | null>(null)

async function handleDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await props.deleteFn()
    emit('deleted')
    emit('update:open', false)
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : 'Failed to delete'
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ title || 'Confirm Delete' }}</DialogTitle>
        <DialogDescription>
          <slot>
            {{ description || 'Are you sure you want to delete this item? This action cannot be undone.' }}
          </slot>
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
