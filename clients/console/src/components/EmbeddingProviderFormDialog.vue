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
import { embeddingAPI } from '@/api/embedding'
import type {
  EmbeddingProvider,
  EmbeddingProviderCreateRequest,
  EmbeddingProviderUpdateRequest,
} from '@/types/embedding'

const props = defineProps<{
  open: boolean
  provider?: EmbeddingProvider | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isEdit = computed(() => !!props.provider)

const form = ref({
  name: '',
  base_url: '',
  api_key: '',
  model: '',
  dimensions: 1536,
  is_default: false,
})

const saving = ref(false)
const saveError = ref<string | null>(null)

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    if (props.provider) {
      form.value = {
        name: props.provider.name,
        base_url: props.provider.base_url,
        api_key: '',
        model: props.provider.model,
        dimensions: props.provider.dimensions,
        is_default: props.provider.is_default,
      }
    } else {
      form.value = {
        name: '',
        base_url: '',
        api_key: '',
        model: '',
        dimensions: 1536,
        is_default: false,
      }
    }
  }
})

async function handleSave() {
  if (!form.value.name.trim()) {
    saveError.value = 'Name is required'
    return
  }
  if (!form.value.base_url.trim()) {
    saveError.value = 'Base URL is required'
    return
  }
  if (!isEdit.value && !form.value.api_key.trim()) {
    saveError.value = 'API key is required'
    return
  }
  if (!form.value.model.trim()) {
    saveError.value = 'Model is required'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    if (isEdit.value) {
      const update: EmbeddingProviderUpdateRequest = {}
      if (form.value.name !== props.provider!.name) update.name = form.value.name
      if (form.value.base_url !== props.provider!.base_url) update.base_url = form.value.base_url
      if (form.value.api_key) update.api_key = form.value.api_key
      if (form.value.model !== props.provider!.model) update.model = form.value.model
      if (form.value.dimensions !== props.provider!.dimensions) update.dimensions = form.value.dimensions
      if (form.value.is_default !== props.provider!.is_default) update.is_default = form.value.is_default

      if (Object.keys(update).length > 0) {
        await embeddingAPI.update(props.provider!.id, update)
      }
    } else {
      const create: EmbeddingProviderCreateRequest = {
        name: form.value.name.trim(),
        base_url: form.value.base_url.trim(),
        api_key: form.value.api_key.trim(),
        model: form.value.model.trim(),
        dimensions: form.value.dimensions,
        is_default: form.value.is_default,
      }
      await embeddingAPI.create(create)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save provider'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? 'Edit Embedding Provider' : 'Add Embedding Provider' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the embedding provider configuration.' : 'Configure a new OpenAI-compatible embedding provider.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2 max-h-[70vh] overflow-y-auto px-2">
        <div class="space-y-2">
          <Label for="emb-name">Name</Label>
          <Input
            id="emb-name"
            v-model="form.name"
            placeholder="e.g. openai-embedding"
          />
        </div>

        <div class="space-y-2">
          <Label for="emb-url">Base URL</Label>
          <Input
            id="emb-url"
            v-model="form.base_url"
            placeholder="e.g. https://api.openai.com/v1"
          />
          <p class="text-xs text-muted-foreground">
            OpenAI-compatible endpoint. The /embeddings path is appended automatically.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="emb-key">API Key</Label>
          <Input
            id="emb-key"
            v-model="form.api_key"
            type="password"
            :placeholder="isEdit ? 'Leave empty to keep current key' : 'sk-...'"
          />
        </div>

        <div class="space-y-2">
          <Label for="emb-model">Model</Label>
          <Input
            id="emb-model"
            v-model="form.model"
            placeholder="e.g. text-embedding-3-small"
          />
        </div>

        <div class="space-y-2">
          <Label for="emb-dims">Dimensions</Label>
          <Input
            id="emb-dims"
            v-model.number="form.dimensions"
            type="number"
            min="1"
            max="65536"
          />
        </div>

        <div class="flex items-center gap-2">
          <input
            id="emb-default"
            v-model="form.is_default"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300"
          />
          <Label for="emb-default" class="text-sm font-normal">
            Set as default provider
          </Label>
        </div>

        <p v-if="saveError" class="text-sm text-destructive">{{ saveError }}</p>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">Cancel</Button>
        <Button :disabled="saving" @click="handleSave">
          {{ saving ? 'Saving...' : (isEdit ? 'Update' : 'Create') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
