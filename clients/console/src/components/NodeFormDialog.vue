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
import { Copy } from '@lucide/vue'
import { nodesAPI } from '@/api/nodes'
import { communitiesAPI } from '@/api/communities'
import type {
  Node,
  NodeCreateRequest,
  NodeUpdateRequest,
  HierarchicalCommunity,
} from '@/types/graph'

const props = defineProps<{
  open: boolean
  node?: Node | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const isEdit = computed(() => !!props.node)

// Returns the current month in YYYY-MM format, which is the wire format
// the backend expects for `first_appeared` and the value format used by
// <input type="month">. Computed once at module load so the dialog
// initialises to a stable "this month" value.
function currentYearMonth(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

// YYYY-MM regex: exactly four digits, a dash, then two digits.
const YEAR_MONTH_RE = /^\d{4}-(0[1-9]|1[0-2])$/

// Form state. The id is only relevant when editing: the backend mints a
// fresh UUID on create and returns it in the response, so the dialog
// never has to fabricate an identifier.
//
// communityId uses the tri-state semantics of NodeUpdateRequest:
//   - null   -> user chose "no community"
//   - string -> user picked a specific community
//   - the third state "leave unchanged" is encoded by not including the
//     field in the update payload when editing.
const form = ref({
  id: '',
  name: '',
  description: '',
  influence_score: 5 as number | null,
  first_appeared: currentYearMonth(),
  communityId: null as string | null,
})

const saving = ref(false)
const saveError = ref<string | null>(null)
const copyState = ref<'idle' | 'copied' | 'error'>('idle')

const communityTree = ref<HierarchicalCommunity[]>([])
const communitiesLoading = ref(false)

interface FlatCommunityOption {
  id: string
  name: string
  color: string
  depth: number
}

function flattenTree(nodes: HierarchicalCommunity[], depth = 0, out: FlatCommunityOption[] = []): FlatCommunityOption[] {
  for (const n of nodes) {
    out.push({ id: n.id, name: n.name, color: n.color, depth })
    if (n.children?.length) flattenTree(n.children, depth + 1, out)
  }
  return out
}

const communityOptions = computed<FlatCommunityOption[]>(() => flattenTree(communityTree.value))

async function loadCommunities() {
  communitiesLoading.value = true
  try {
    const res = await communitiesAPI.tree()
    communityTree.value = res.communities
  } catch (e) {
    // Non-fatal: the selector will simply show only "no community".
    communityTree.value = []
    saveError.value = e instanceof Error ? e.message : 'Failed to load communities'
  } finally {
    communitiesLoading.value = false
  }
}

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

// Sanitize the YYYY-MM input as the user types. Strips anything other than
// digits and dashes, and clamps the length to 7 characters so the field can
// never hold a value the regex would reject for length reasons. Real
// validation still happens in handleSave.
function sanitizeYearMonth(value: string): string {
  return value.replace(/[^0-9-]/g, '').slice(0, 7)
}

// Sanitize the influence score input. Strips anything other than digits
// and a single decimal point. Returns a number, or null when the field is
// empty so v-model can store it without coercing to 0 silently.
function sanitizeScore(value: string): number | null {
  const cleaned = value.replace(/[^0-9.]/g, '')
  // Collapse multiple dots to a single one (e.g. "1.2.3" -> "1.23")
  const normalized = cleaned.replace(/(\..*)\./g, '$1')
  if (normalized === '' || normalized === '.') return null
  const n = Number(normalized)
  return Number.isFinite(n) ? n : null
}

// Block any keystroke that would introduce a non-numeric, non-dot
// character. Returning false from a keypress handler prevents the
// character from being inserted into the input in the first place, so the
// user never sees letters appear. Pasted content and IME events are
// still covered by the input-level sanitizer.
function onScoreKeypress(e: KeyboardEvent) {
  // Always allow control / navigation / editing keys.
  if (e.ctrlKey || e.metaKey || e.altKey) return
  if (e.key.length !== 1) return // Backspace, Arrow keys, etc.
  if (/[0-9.]/.test(e.key)) return
  e.preventDefault()
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    saveError.value = null
    copyState.value = 'idle'
    if (props.node) {
      form.value = {
        id: props.node.id,
        name: props.node.name,
        description: props.node.description,
        influence_score: props.node.influence_score,
        first_appeared: props.node.first_appeared,
        communityId: props.node.community_id ?? null,
      }
    } else {
      form.value = {
        id: '',
        name: '',
        description: '',
        influence_score: 5 as number | null,
        first_appeared: currentYearMonth(),
        communityId: null,
      }
    }
    // Lazy-load communities the first time the dialog opens and on subsequent
    // opens so the picker reflects the latest tree.
    if (communityTree.value.length === 0) {
      loadCommunities()
    }
  }
})

function buildUpdatePayload(): NodeUpdateRequest {
  const update: NodeUpdateRequest = {}
  if (form.value.name !== props.node!.name) update.name = form.value.name
  if (form.value.description !== props.node!.description) update.description = form.value.description
  if (form.value.influence_score !== props.node!.influence_score) {
    update.influence_score = form.value.influence_score ?? undefined
  }
  if (form.value.first_appeared !== props.node!.first_appeared) update.first_appeared = form.value.first_appeared
  if (form.value.communityId !== (props.node!.community_id ?? null)) {
    update.community_id = form.value.communityId
  }
  return update
}

function isPayloadEmpty(p: NodeUpdateRequest): boolean {
  return Object.keys(p).length === 0
}

async function handleSave() {
  if (!form.value.name.trim()) {
    saveError.value = 'Name is required'
    return
  }
  if (!YEAR_MONTH_RE.test(form.value.first_appeared)) {
    saveError.value = 'First Appeared must be a valid month (YYYY-MM)'
    return
  }

  saving.value = true
  saveError.value = null

  try {
    if (isEdit.value) {
      const update = buildUpdatePayload()
      if (!isPayloadEmpty(update)) {
        await nodesAPI.update(props.node!.id, update)
      }
    } else {
      const create: NodeCreateRequest = {
        name: form.value.name.trim(),
        description: form.value.description,
        influence_score: form.value.influence_score ?? 5,
        first_appeared: form.value.first_appeared,
        community_id: form.value.communityId,
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

      <div class="space-y-4 py-2 max-h-[70vh] overflow-y-auto px-2">
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
          <Label for="node-name">Name</Label>
          <Input
            id="node-name"
            v-model="form.name"
            placeholder="e.g. Transformer"
          />
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
              :model-value="form.influence_score === null ? '' : String(form.influence_score)"
              @update:model-value="(v) => form.influence_score = sanitizeScore(String(v))"
              @keypress="onScoreKeypress"
              type="text"
              inputmode="decimal"
              min="0"
              max="10"
              step="0.1"
            />
          </div>
          <div class="space-y-2">
            <Label for="node-year">First Appeared</Label>
            <Input
              id="node-year"
              :model-value="form.first_appeared"
              @update:model-value="(v) => form.first_appeared = sanitizeYearMonth(String(v))"
              type="month"
              :min="'1900-01'"
              :max="'2099-12'"
              required
            />
          </div>
        </div>

        <div class="space-y-2">
          <Label>Community</Label>
          <Select
            :model-value="form.communityId === null ? '__none__' : form.communityId"
            @update:model-value="(v) => form.communityId = v === '__none__' ? null : String(v)"
          >
            <SelectTrigger class="w-full">
              <SelectValue :placeholder="communitiesLoading ? 'Loading…' : 'No community'" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__none__">No community</SelectItem>
              <SelectItem
                v-for="opt in communityOptions"
                :key="opt.id"
                :value="opt.id"
              >
                <span class="inline-flex items-center gap-2">
                  <span
                    class="inline-block h-2.5 w-2.5 rounded-sm border"
                    :style="{ backgroundColor: opt.color }"
                  />
                  <span :style="{ paddingLeft: `${opt.depth * 12}px` }">
                    {{ '— '.repeat(opt.depth) }}{{ opt.name }}
                  </span>
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
