<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { HierarchicalCommunity } from '../types/graph'

const { t } = useI18n()

const props = defineProps<{
  communities: HierarchicalCommunity[]
  highlightedCommunity: number[]
}>()

const emit = defineEmits<{
  highlight: [communityId: number | null]
}>()

const collapsed = ref<Set<number>>(new Set())

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
</script>

<template>
  <div class="bg-(--color-bg-secondary) rounded-2xl p-5 w-72 select-none shadow-2xl transition-colors duration-300">
    <h3 class="text-(--color-text-primary) font-semibold mb-4 pl-2">{{ t('communityLegend.title') }}</h3>

    <div class="flex flex-col gap-1">
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

      <template v-for="comm in topLevelChildren" :key="comm.id">
        <div
          class="flex items-center gap-1.5 cursor-pointer rounded-lg px-2 py-2 transition-colors"
          :class="isHighlighted(comm.id)
            ? 'bg-(--color-accent-blue)/10'
            : 'hover:bg-(--color-bg-tertiary)'"
          @click="handleClick(comm)"
        >
          <button
            v-if="comm.children && comm.children.length > 0"
            class="w-5 h-5 flex items-center justify-center text-(--color-text-muted) hover:text-(--color-text-primary) shrink-0 rounded transition-colors"
            @click.stop="toggleCollapse(comm.id)"
          >
            <svg
              class="w-3.5 h-3.5 transition-transform duration-200"
              :class="collapsed.has(comm.id) ? '' : 'rotate-90'"
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <div v-else class="w-5 shrink-0" />
          <div
            class="w-3 h-3 rounded-full shrink-0"
            :style="{ backgroundColor: comm.color }"
          />
          <span
            class="flex-1 text-sm truncate"
            :class="isHighlighted(comm.id) ? 'text-(--color-accent-blue) font-medium' : 'text-(--color-text-primary)'"
          >{{ comm.name }}</span>
          <span class="text-xs text-(--color-text-muted) tabular-nums">{{ comm.node_count }}</span>
        </div>

        <div
          v-if="comm.children && comm.children.length > 0 && !collapsed.has(comm.id)"
          class="ml-4 border-l-2 border-(--color-border-muted) pl-3 mt-0.5"
        >
          <div
            v-for="child in comm.children"
            :key="child.id"
            class="flex items-center gap-3 cursor-pointer rounded-lg px-2 py-1.5 transition-colors"
            :class="isHighlighted(child.id)
              ? 'bg-(--color-accent-blue)/10'
              : 'hover:bg-(--color-bg-tertiary)'"
            @click="handleClick(child)"
          >
            <div
              class="w-2.5 h-2.5 rounded-full shrink-0"
              :style="{ backgroundColor: child.color }"
            />
            <span
              class="flex-1 text-sm truncate"
              :class="isHighlighted(child.id) ? 'text-(--color-accent-blue) font-medium' : 'text-(--color-text-primary)'"
            >{{ child.name }}</span>
            <span class="text-xs text-(--color-text-muted) tabular-nums">{{ child.node_count }}</span>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
