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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { nodesAPI } from '@/api/nodes'
import type { Node, NodeCreateRequest, NodeUpdateRequest } from '@/types/graph'

const props = defineProps<{
  open: boolean
  node?: Node | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const categories = [
  'dl_arch', 'dl_mechanism', 'dl_technique',
  'nlp_model', 'nlp_technique',
  'cv_model',
  'gen_model',
  'multimodal',
  'speech_model',
  'rl_algorithm',
  'gnn',
  'optimizer',
  'alignment',
  'application',
  'framework',
  'infrastructure',
  'organization',
  'platform',
]

const isEdit = computed(() => !!props.node)

const form = ref({
  id: '',
  name: '',
  category: '',
  description: '',
  influence_score: 5,
  first_appeared: 2024,
})

const saving = ref(false)
const saveError = ref<string | null>(null)

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    if (props.node) {
      form.value = {
        id: props.node.id,
        name: props.node.name,
        category: props.node.category,
        description: props.node.description,
        influence_score: props.node.influence_score,
        first_appeared: props.node.first_appeared,
      }
    } else {
      form.value = {
        id: '',
        name: '',
        category: '',
        description: '',
        influence_score: 5,
        first_appeared: 2024,
      }
    }
  }
})

async function handleSave() {
  if (!form.value.id.trim() || !form.value.name.trim()) {
    saveError.value = 'ID and Name are required'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    if (isEdit.value) {
      const update: NodeUpdateRequest = {}
      if (form.value.name !== props.node!.name) update.name = form.value.name
      if (form.value.category !== props.node!.category) update.category = form.value.category
      if (form.value.description !== props.node!.description) update.description = form.value.description
      if (form.value.influence_score !== props.node!.influence_score) update.influence_score = form.value.influence_score
      if (form.value.first_appeared !== props.node!.first_appeared) update.first_appeared = form.value.first_appeared
      await nodesAPI.update(props.node!.id, update)
    } else {
      const create: NodeCreateRequest = {
        id: form.value.id.trim(),
        name: form.value.name.trim(),
        category: form.value.category,
        description: form.value.description,
        influence_score: form.value.influence_score,
        first_appeared: form.value.first_appeared,
      }
      await nodesAPI.create(create)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save node'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? 'Edit Node' : 'Add Node' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the node details below.' : 'Fill in the details for the new node.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2">
        <div class="space-y-2">
          <Label for="node-id">ID</Label>
          <Input
            id="node-id"
            v-model="form.id"
            placeholder="e.g. transformer"
            :disabled="isEdit"
          />
        </div>

        <div class="space-y-2">
          <Label for="node-name">Name</Label>
          <Input
            id="node-name"
            v-model="form.name"
            placeholder="e.g. Transformer"
          />
        </div>

        <div class="space-y-2">
          <Label>Category</Label>
          <Select v-model="form.category">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Select a category" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="cat in categories" :key="cat" :value="cat">
                {{ cat }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="space-y-2">
          <Label for="node-desc">Description</Label>
          <Textarea
            id="node-desc"
            v-model="form.description"
            placeholder="Brief description of the node"
            rows="3"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="node-score">Influence Score</Label>
            <Input
              id="node-score"
              v-model.number="form.influence_score"
              type="number"
              min="0"
              max="10"
              step="0.1"
            />
          </div>
          <div class="space-y-2">
            <Label for="node-year">First Appeared</Label>
            <Input
              id="node-year"
              v-model.number="form.first_appeared"
              type="number"
              min="1900"
              max="2030"
            />
          </div>
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
