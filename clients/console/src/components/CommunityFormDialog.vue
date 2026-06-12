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
import { Copy } from '@lucide/vue'
import { communitiesAPI } from '@/api/communities'
import type {
  Community,
  CommunityCreateRequest,
  CommunityUpdateRequest,
  HierarchicalCommunity,
} from '@/types/graph'

const props = defineProps<{
  open: boolean
  community?: Community | null
  tree: HierarchicalCommunity[]
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const presetColors = [
  '#5a7a8a', '#6a8a6a', '#7a6a8a', '#8a6a5a',
  '#5a6a8a', '#6a5a8a', '#8a7a5a', '#5a8a7a',
  '#7a8a5a', '#8a5a6a', '#6a7a8a', '#7a5a6a',
  '#5a6a6a', '#4a5a6a', '#a8a8a8', '#cc8a5a',
]
const isEdit = computed(() => !!props.community)

// `form.id` is populated only when editing. On add the server mints a
// fresh UUID and the dialog never asks the user to fabricate one. The
// read-only ID block + copy button is therefore hidden in create mode.
const form = ref({
  id: '',
  name: '',
  color: presetColors[0],
  domain: 'ai',
  parentId: null as string | null,
  useCustomColor: false,
})

const saving = ref(false)
const saveError = ref<string | null>(null)
const copyState = ref<'idle' | 'copied' | 'error'>('idle')

interface FlatOption {
  id: string
  name: string
  depth: number
}

function flattenTree(nodes: HierarchicalCommunity[], depth = 0, out: FlatOption[] = []): FlatOption[] {
  for (const n of nodes) {
    out.push({ id: n.id, name: n.name, depth })
    if (n.children?.length) flattenTree(n.children, depth + 1, out)
  }
  return out
}

// When editing, exclude the community itself and its descendants from the
// parent options, since setting any of them as a parent would create a cycle.
function collectDescendantIds(root: HierarchicalCommunity): Set<string> {
  const ids = new Set<string>()
  const walk = (n: HierarchicalCommunity) => {
    ids.add(n.id)
    n.children?.forEach(walk)
  }
  walk(root)
  return ids
}

const parentOptions = computed<FlatOption[]>(() => {
  const flat = flattenTree(props.tree)
  if (!props.community) return flat
  const excluded = new Set<string>()
  const findSubtree = (nodes: HierarchicalCommunity[]): HierarchicalCommunity | null => {
    for (const n of nodes) {
      if (n.id === props.community!.id) return n
      if (n.children?.length) {
        const r = findSubtree(n.children)
        if (r) return r
      }
    }
    return null
  }
  const subtree = findSubtree(props.tree)
  if (subtree) {
    collectDescendantIds(subtree).forEach((id) => excluded.add(id))
  }
  return flat.filter((o) => !excluded.has(o.id))
})

async function copyId() {
  if (!form.value.id) return
  try {
    await navigator.clipboard.writeText(form.value.id)
    copyState.value = 'copied'
    setTimeout(() => {
      copyState.value = 'idle'
    }, 1500)
  } catch {
    copyState.value = 'error'
    setTimeout(() => {
      copyState.value = 'idle'
    }, 1500)
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    copyState.value = 'idle'
    if (props.community) {
      const isPreset = presetColors.includes(props.community.color)
      form.value = {
        id: props.community.id,
        name: props.community.name,
        color: isPreset ? props.community.color : presetColors[0],
        domain: props.community.domain,
        parentId: props.community.parent_id,
        useCustomColor: !isPreset,
      }
    } else {
      form.value = {
        id: '',
        name: '',
        color: presetColors[0],
        domain: 'ai',
        parentId: null,
        useCustomColor: false,
      }
    }
  }
})

async function handleSave() {
  if (!form.value.name.trim()) {
    saveError.value = 'Name is required'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    const finalColor = form.value.color

    if (isEdit.value) {
      const update: CommunityUpdateRequest = {}
      if (form.value.name !== props.community!.name) update.name = form.value.name
      if (finalColor !== props.community!.color) update.color = finalColor
      if (form.value.domain !== props.community!.domain) update.domain = form.value.domain
      const origParent = props.community!.parent_id ?? null
      if (form.value.parentId !== origParent) update.parent_id = form.value.parentId
      await communitiesAPI.update(props.community!.id, update)
    } else {
      const create: CommunityCreateRequest = {
        name: form.value.name.trim(),
        color: finalColor,
        domain: form.value.domain,
        parent_id: form.value.parentId,
      }
      await communitiesAPI.create(create)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save community'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? 'Edit Community' : 'Add Community' }}</DialogTitle>
        <DialogDescription>
          {{ isEdit ? 'Update the community details below.' : 'Fill in the details for the new community.' }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2">
        <!-- ID is only shown when editing: the backend mints a fresh
             UUID on create and returns it in the response. -->
        <div v-if="isEdit" class="space-y-2">
          <Label>ID</Label>
          <div class="flex items-center gap-2">
            <code
              class="flex-1 min-w-0 truncate rounded border bg-muted px-2 py-1.5 text-xs font-mono"
              :title="form.id"
            >{{ form.id }}</code>
            <Button
              type="button"
              size="icon"
              variant="outline"
              class="h-8 w-8 shrink-0"
              :aria-label="copyState === 'copied' ? 'Copied' : 'Copy ID'"
              @click="copyId"
            >
              <Copy v-if="copyState === 'idle'" class="h-3.5 w-3.5" />
              <span v-else-if="copyState === 'copied'" class="text-[10px] font-medium text-primary">✓</span>
              <span v-else class="text-[10px] font-medium text-destructive">!</span>
            </Button>
          </div>
          <p class="text-xs text-muted-foreground">
            Generated by the server. Copy and paste it into relationships or other systems.
          </p>
        </div>

        <div class="space-y-2">
          <Label for="comm-name">Name</Label>
          <Input
            id="comm-name"
            v-model="form.name"
            placeholder="e.g. Natural Language Processing"
          />
        </div>

        <div class="space-y-2">
          <Label>Color</Label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="c in presetColors"
              :key="c"
              type="button"
              class="h-7 w-7 rounded border-2 transition-transform hover:scale-110"
              :class="!form.useCustomColor && form.color === c ? 'border-foreground' : 'border-transparent'"
              :style="{ backgroundColor: c }"
              :aria-label="`Color ${c}`"
              @click="form.useCustomColor = false; form.color = c"
            />
            <button
              type="button"
              class="h-7 px-2 rounded border text-xs"
              :class="form.useCustomColor ? 'border-foreground bg-accent' : 'border-input'"
              @click="form.useCustomColor = true"
            >
              Custom
            </button>
          </div>
          <Input
            v-if="form.useCustomColor"
            v-model="form.color"
            placeholder="#5a7a8a"
            class="font-mono"
          />
        </div>

        <div class="space-y-2">
          <Label for="comm-domain">Domain</Label>
          <Input
            id="comm-domain"
            v-model="form.domain"
            placeholder="e.g. ai"
          />
        </div>

        <div class="space-y-2">
          <Label>Parent</Label>
          <Select
            :model-value="form.parentId === null ? '__none__' : form.parentId"
            @update:model-value="(v) => form.parentId = v === '__none__' ? null : String(v)"
          >
            <SelectTrigger class="w-full">
              <SelectValue placeholder="No parent (root)" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__none__">No parent (root)</SelectItem>
              <SelectItem
                v-for="opt in parentOptions"
                :key="opt.id"
                :value="opt.id"
              >
                <span :style="{ paddingLeft: `${opt.depth * 12}px` }">
                  {{ '— '.repeat(opt.depth) }}{{ opt.name }}
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
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
