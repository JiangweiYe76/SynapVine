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
import { communitiesAPI } from '@/api/communities'

const props = defineProps<{
  open: boolean
  communityId: string | null
  communityName?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'deleted'): void
}>()

const deleting = ref(false)
const deleteError = ref<string | null>(null)

async function handleDelete() {
  if (props.communityId === null) return

  deleting.value = true
  deleteError.value = null

  try {
    await communitiesAPI.delete(props.communityId)
    emit('deleted')
    emit('update:open', false)
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : 'Failed to delete community'
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>Delete Community</DialogTitle>
        <DialogDescription>
          Are you sure you want to delete
          <strong v-if="communityName">{{ communityName }}</strong>
          <strong v-else>this community</strong>?
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
