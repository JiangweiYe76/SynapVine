<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { HierarchicalCommunity } from '../types/graph'
import { ChevronRight, ChevronLeft } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const { t } = useI18n()

const props = defineProps<{
  communities: HierarchicalCommunity[]
  highlightedCommunity: number[]
}>()

const emit = defineEmits<{
  highlight: [communityId: number | null]
}>()

const collapsed = ref<Set<number>>(new Set())
const panelVisible = ref(true)

function toggleCollapse(id: number) {
  const next = new Set(collapsed.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  collapsed.value = next
}

function isHighlighted(id: number): boolean {
  return props.highlightedCommunity.includes(id)
}

function handleClick(community: HierarchicalCommunity) {
  if (props.highlightedCommunity.includes(community.id)) {
    emit('highlight', null)
  } else {
    emit('highlight', community.id)
  }
}

function totalNodeCount(): number {
  return props.communities[0]?.node_count || 0
}

const topLevelChildren = computed(() => props.communities[0]?.children || [])

interface FlatNode {
  comm: HierarchicalCommunity
  depth: number
  hasChildren: boolean
}

// Flatten the community tree into a depth-annotated list so the
// template can render arbitrary nesting levels without recursion.
const flatList = computed<FlatNode[]>(() => {
  const result: FlatNode[] = []
  function walk(list: HierarchicalCommunity[], depth: number) {
    for (const c of list) {
      const hasChildren = !!(c.children && c.children.length > 0)
      result.push({ comm: c, depth, hasChildren })
      if (hasChildren && !collapsed.value.has(c.id)) {
        walk(c.children!, depth + 1)
      }
    }
  }
  walk(topLevelChildren.value, 0)
  return result
})
</script>

<template>
  <Button
    v-if="!panelVisible"
    variant="ghost"
    class="absolute left-0 top-12 w-10 h-14 z-30 bg-(--color-bg-secondary) border border-l-0 border-(--color-border-default) rounded-r-xl shadow-lg hover:bg-(--color-bg-tertiary)"
    @click="panelVisible = true"
  >
    <ChevronRight class="w-5 h-5 text-(--color-text-secondary)" />
  </Button>

  <Transition
    enter-active-class="transition-transform duration-300 ease-out"
    enter-from-class="-translate-x-full"
    enter-to-class="translate-x-0"
    leave-active-class="transition-transform duration-300 ease-in"
    leave-from-class="translate-x-0"
    leave-to-class="-translate-x-full"
  >
    <div
      v-if="panelVisible"
      class="absolute left-12 top-12 bottom-12 w-72 bg-(--color-bg-secondary) flex flex-col z-30 shadow-2xl rounded-2xl select-none transition-colors duration-300"
    >
      <div class="flex items-center justify-between px-6 py-5 border-b border-(--color-border-default) shrink-0">
        <h2 class="text-lg font-semibold text-(--color-text-primary)">{{ t('communityLegend.title') }}</h2>
        <Button
          variant="ghost"
          size="icon"
          @click="panelVisible = false"
        >
          <ChevronLeft class="w-5 h-5" />
        </Button>
      </div>

      <div class="flex-1 overflow-y-auto px-6 py-6 space-y-1">
        <div
          class="flex items-center gap-3 cursor-pointer rounded-lg px-3 py-2.5 transition-colors"
          :class="highlightedCommunity.length === 0
            ? 'bg-(--color-accent-blue)/15 text-(--color-accent-blue)'
            : 'hover:bg-(--color-bg-tertiary) text-(--color-text-primary)'"
          @click="emit('highlight', null)"
        >
          <div class="w-4 h-4 rounded-full bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 shrink-0" />
          <span class="flex-1 text-sm font-medium">{{ t('communityLegend.all') }}</span>
          <span class="text-xs text-(--color-text-muted) tabular-nums">{{ totalNodeCount() }}</span>
        </div>

        <div class="h-px bg-(--color-border-default) my-2" />

        <div
          v-for="item in flatList"
          :key="item.comm.id"
          class="flex items-center gap-1.5 cursor-pointer rounded-lg py-1.5 transition-colors"
          :class="isHighlighted(item.comm.id)
            ? 'bg-(--color-accent-blue)/10'
            : 'hover:bg-(--color-bg-tertiary)'"
          :style="{ paddingLeft: `${item.depth * 16 + 8}px` }"
          @click="handleClick(item.comm)"
        >
          <Button
            v-if="item.hasChildren"
            variant="ghost"
            size="icon"
            class="size-5 shrink-0"
            @click.stop="toggleCollapse(item.comm.id)"
          >
            <ChevronRight
              class="size-3.5 transition-transform duration-200"
              :class="{ 'rotate-90': !collapsed.has(item.comm.id) }"
            />
          </Button>
          <div v-else class="w-5 shrink-0" />
          <div
            class="rounded-full shrink-0"
            :class="item.depth === 0 ? 'w-3 h-3' : 'w-2.5 h-2.5'"
            :style="{ backgroundColor: item.comm.color }"
          />
          <span
            class="flex-1 text-sm truncate"
            :class="isHighlighted(item.comm.id) ? 'text-(--color-accent-blue) font-medium' : 'text-(--color-text-primary)'"
          >{{ item.comm.name }}</span>
          <span class="text-xs text-(--color-text-muted) tabular-nums">{{ item.comm.node_count }}</span>
        </div>
      </div>
    </div>
  </Transition>
</template>
