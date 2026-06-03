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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { edgesAPI } from '@/api/edges'
import { nodesAPI } from '@/api/nodes'
import type { Edge, EdgeCreateRequest, EdgeUpdateRequest, Node } from '@/types/graph'

const props = defineProps<{
  open: boolean
  edge?: Edge | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isEdit = computed(() => !!props.edge)

const nodes = ref<Node[]>([])
const nodesLoading = ref(false)

const form = ref({
  source: '',
  target: '',
  relation: '',
  weight: 0.5,
})

const saving = ref(false)
const saveError = ref<string | null>(null)

async function fetchNodes() {
  nodesLoading.value = true
  try {
    const res = await nodesAPI.list(0, 1000)
    nodes.value = res.nodes
  } catch (e) {
    // silently fail, user can still type manually
  } finally {
    nodesLoading.value = false
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    fetchNodes()
    if (props.edge) {
      form.value = {
        source: props.edge.source,
        target: props.edge.target,
        relation: props.edge.relation,
        weight: props.edge.weight,
      }
    } else {
      form.value = {
        source: '',
        target: '',
        relation: '',
        weight: 0.5,
      }
    }
  }
})

async function handleSave() {
  if (!form.value.source.trim() || !form.value.target.trim()) {
    saveError.value = 'Source and target are required'
    return
  }

  if (form.value.source === form.value.target) {
    saveError.value = 'Source and target cannot be the same'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    if (isEdit.value) {
      const update: EdgeUpdateRequest = {}
      if (form.value.relation !== props.edge!.relation) update.relation = form.value.relation
      if (form.value.weight !== props.edge!.weight) update.weight = form.value.weight
      await edgesAPI.update(props.edge!.source, props.edge!.target, update)
    } else {
      const create: EdgeCreateRequest = {
        source: form.value.source.trim(),
        target: form.value.target.trim(),
        relation: form.value.relation,
        weight: form.value.weight,
      }
      await edgesAPI.create(create)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save edge'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? 'Edit Edge' : 'Add Edge' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the edge details below.' : 'Connect two nodes with a relationship.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label>Source</Label>
            <Select v-model="form.source" :disabled="isEdit">
              <SelectTrigger>
                <SelectValue placeholder="Select source" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="n in nodes" :key="n.id" :value="n.id">
                  {{ n.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-2">
            <Label>Target</Label>
            <Select v-model="form.target" :disabled="isEdit">
              <SelectTrigger>
                <SelectValue placeholder="Select target" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="n in nodes" :key="n.id" :value="n.id">
                  {{ n.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="edge-relation">Relation</Label>
          <Input
            id="edge-relation"
            v-model="form.relation"
            placeholder="e.g. based_on, evolved_to"
          />
        </div>

        <div class="space-y-2">
          <Label for="edge-weight">Weight</Label>
          <Input
            id="edge-weight"
            v-model.number="form.weight"
            type="number"
            min="0"
            max="1"
            step="0.01"
          />
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
