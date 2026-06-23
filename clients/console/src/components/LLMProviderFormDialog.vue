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
import { llmAPI } from '@/api/llm'
import type {
  LLMProvider,
  LLMProviderCreateRequest,
  LLMProviderUpdateRequest,
} from '@/types/llm'

const props = defineProps<{
  open: boolean
  provider?: LLMProvider | null
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
  max_tokens: 4096,
  temperature: 0.7,
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
        api_key: '', // Never pre-fill API key
        model: props.provider.model,
        max_tokens: props.provider.max_tokens,
        temperature: props.provider.temperature,
        is_default: props.provider.is_default,
      }
    } else {
      form.value = {
        name: '',
        base_url: '',
        api_key: '',
        model: '',
        max_tokens: 4096,
        temperature: 0.7,
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
      const update: LLMProviderUpdateRequest = {}
      if (form.value.name !== props.provider!.name) update.name = form.value.name
      if (form.value.base_url !== props.provider!.base_url) update.base_url = form.value.base_url
      if (form.value.api_key) update.api_key = form.value.api_key
      if (form.value.model !== props.provider!.model) update.model = form.value.model
      if (form.value.max_tokens !== props.provider!.max_tokens) update.max_tokens = form.value.max_tokens
      if (form.value.temperature !== props.provider!.temperature) update.temperature = form.value.temperature
      if (form.value.is_default !== props.provider!.is_default) update.is_default = form.value.is_default

      if (Object.keys(update).length > 0) {
        await llmAPI.update(props.provider!.id, update)
      }
    } else {
      const create: LLMProviderCreateRequest = {
        name: form.value.name.trim(),
        base_url: form.value.base_url.trim(),
        api_key: form.value.api_key.trim(),
        model: form.value.model.trim(),
        max_tokens: form.value.max_tokens,
        temperature: form.value.temperature,
        is_default: form.value.is_default,
      }
      await llmAPI.create(create)
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
        <DialogTitle>{{ isEdit ? 'Edit Provider' : 'Add Provider' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the provider configuration.' : 'Configure a new OpenAI-compatible LLM provider.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2 max-h-[70vh] overflow-y-auto px-2">
        <div class="space-y-2">
          <Label for="provider-name">Name</Label>
          <Input
            id="provider-name"
            v-model="form.name"
            placeholder="e.g. deepseek-chat"
          />
        </div>

        <div class="space-y-2">
          <Label for="provider-url">Base URL</Label>
          <Input
            id="provider-url"
            v-model="form.base_url"
            placeholder="e.g. https://api.deepseek.com/v1"
          />
          <p class="text-xs text-muted-foreground">
            OpenAI-compatible endpoint (e.g. /v1). The /chat/completions path is appended automatically.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="provider-key">API Key</Label>
          <Input
            id="provider-key"
            v-model="form.api_key"
            type="password"
            :placeholder="isEdit ? 'Leave empty to keep current key' : 'sk-...'"
          />
        </div>

        <div class="space-y-2">
          <Label for="provider-model">Model</Label>
          <Input
            id="provider-model"
            v-model="form.model"
            placeholder="e.g. deepseek-chat, gpt-4o"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="provider-tokens">Max Tokens</Label>
            <Input
              id="provider-tokens"
              v-model.number="form.max_tokens"
              type="number"
              min="1"
              max="128000"
            />
          </div>
          <div class="space-y-2">
            <Label for="provider-temp">Temperature</Label>
            <Input
              id="provider-temp"
              v-model.number="form.temperature"
              type="number"
              min="0"
              max="2"
              step="0.1"
            />
          </div>
        </div>

        <div class="flex items-center gap-2">
          <input
            id="provider-default"
            v-model="form.is_default"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300"
          />
          <Label for="provider-default" class="text-sm font-normal">
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
