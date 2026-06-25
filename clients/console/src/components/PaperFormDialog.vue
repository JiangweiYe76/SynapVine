<script setup lang="ts">
import { ref, watch } from 'vue'
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
import { papersAPI } from '@/api/papers'
import type { Paper, PaperUpdateRequest } from '@/types/paper'
import { FileText, X } from '@lucide/vue'

const props = defineProps<{
  open: boolean
  paper?: Paper | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isEdit = ref(false)

const form = ref({
  title: '',
  authors: '',
  source_url: '',
})

const pdfFile = ref<File | null>(null)
const saving = ref(false)
const saveError = ref<string | null>(null)

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    pdfFile.value = null
    isEdit.value = !!props.paper
    if (props.paper) {
      form.value = {
        title: props.paper.title,
        authors: props.paper.authors,
        source_url: props.paper.source_url,
      }
    } else {
      form.value = { title: '', authors: '', source_url: '' }
    }
  }
})

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (!file.name.toLowerCase().endsWith('.pdf')) {
    saveError.value = 'Only PDF files are supported'
    target.value = ''
    return
  }
  pdfFile.value = file
  saveError.value = null
}

function clearFile() {
  pdfFile.value = null
  const input = document.getElementById('paper-pdf') as HTMLInputElement
  if (input) input.value = ''
}

async function handleSave() {
  if (!form.value.title.trim()) {
    saveError.value = 'Title is required'
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
      if (!pdfFile.value) {
        saveError.value = 'Please select a PDF file'
        saving.value = false
        return
      }
      const formData = new FormData()
      formData.append('pdf', pdfFile.value)
      if (form.value.title.trim()) formData.append('title', form.value.title.trim())
      if (form.value.authors.trim()) formData.append('authors', form.value.authors.trim())
      if (form.value.source_url.trim()) formData.append('source_url', form.value.source_url.trim())
      await papersAPI.createWithPDF(formData)
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
          {{ isEdit ? 'Update the paper details.' : 'Upload a PDF file for AI concept extraction.' }}
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
          <Label for="paper-pdf">PDF File</Label>
          <div v-if="!pdfFile" class="flex items-center gap-2">
            <Input
              id="paper-pdf"
              type="file"
              accept=".pdf"
              class="cursor-pointer"
              @change="onFileChange"
            />
          </div>
          <div v-else class="flex items-center gap-2 rounded-md border p-2">
            <FileText class="h-4 w-4 text-muted-foreground shrink-0" />
            <span class="text-sm truncate flex-1">{{ pdfFile.name }}</span>
            <span class="text-xs text-muted-foreground whitespace-nowrap">
              {{ (pdfFile.size / 1024).toFixed(0) }} KB
            </span>
            <Button variant="ghost" size="sm" class="h-6 w-6 p-0" @click="clearFile">
              <X class="h-4 w-4" />
            </Button>
          </div>
          <p class="text-xs text-muted-foreground">
            Text will be extracted from the PDF automatically. Title is auto-filled from the filename if left empty.
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
