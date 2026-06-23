<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { papersAPI } from '@/api/papers'
import type { Paper, PaperCreateRequest, PaperUpdateRequest } from '@/types/paper'

const props = defineProps<{
  open: boolean
  paper?: Paper | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isEdit = computed(() => !!props.paper)

const form = ref({
  title: '',
  authors: '',
  source_url: '',
  raw_text: '',
})

const saving = ref(false)
const saveError = ref<string | null>(null)

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    if (props.paper) {
      form.value = {
        title: props.paper.title,
        authors: props.paper.authors,
        source_url: props.paper.source_url,
        raw_text: props.paper.raw_text,
      }
    } else {
      form.value = {
        title: '',
        authors: '',
        source_url: '',
        raw_text: '',
      }
    }
  }
})

async function handleSave() {
  if (!form.value.title.trim()) {
    saveError.value = 'Title is required'
    return
  }
  if (!isEdit.value && !form.value.raw_text.trim()) {
    saveError.value = 'Paper text is required'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    if (isEdit.value) {
      const update: PaperUpdateRequest = {}
      if (form.value.title !== props.paper!.title) update.title = form.value.title
      if (form.value.authors !== props.paper!.authors) update.authors = form.value.authors
      if (form.value.source_url !== props.paper!.source_url) update.source_url = form.value.source_url
      if (Object.keys(update).length > 0) {
        await papersAPI.update(props.paper!.id, update)
      }
    } else {
      const create: PaperCreateRequest = {
        title: form.value.title.trim(),
        authors: form.value.authors.trim(),
        source_url: form.value.source_url.trim() || undefined,
        raw_text: form.value.raw_text.trim(),
      }
      await papersAPI.create(create)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save paper'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-2xl">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? 'Edit Paper' : 'Upload Paper' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the paper details.' : 'Paste the paper text for AI concept extraction.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2 max-h-[70vh] overflow-y-auto px-2">
        <div class="space-y-2">
          <Label for="paper-title">Title</Label>
          <Input
            id="paper-title"
            v-model="form.title"
            placeholder="e.g. Attention Is All You Need"
          />
        </div>

        <div class="space-y-2">
          <Label for="paper-authors">Authors</Label>
          <Input
            id="paper-authors"
            v-model="form.authors"
            placeholder="e.g. Vaswani et al."
          />
        </div>

        <div class="space-y-2">
          <Label for="paper-url">Source URL (optional)</Label>
          <Input
            id="paper-url"
            v-model="form.source_url"
            placeholder="e.g. https://arxiv.org/abs/1706.03762"
          />
        </div>

        <div v-if="!isEdit" class="space-y-2">
          <Label for="paper-text">Paper Text</Label>
          <Textarea
            id="paper-text"
            v-model="form.raw_text"
            placeholder="Paste the full paper text here..."
            rows="12"
            class="font-mono text-sm"
          />
          <p class="text-xs text-muted-foreground">
            The LLM will extract AI concepts and relationships from this text.
          </p>
        </div>

        <p v-if="saveError" class="text-sm text-destructive">{{ saveError }}</p>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">Cancel</Button>
        <Button :disabled="saving" @click="handleSave">
          {{ saving ? 'Saving...' : (isEdit ? 'Update' : 'Upload') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
