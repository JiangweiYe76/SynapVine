<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { NodeDetail as NodeDetailType, HierarchicalCommunity } from '../types/graph'
import { PALETTE } from '../types/graph'
import { X } from 'lucide-vue-next'

const { t } = useI18n()

const props = defineProps<{
  detail: NodeDetailType | null
  communities: HierarchicalCommunity[]
}>()

const emit = defineEmits<{
  close: []
  neighborClick: [neighborId: string]
}>()

const communityLookup = computed(() => {
  const map = new Map<number, HierarchicalCommunity>()
  function walk(list: HierarchicalCommunity[]) {
    for (const c of list) {
      map.set(c.id, c)
      if (c.children) walk(c.children)
    }
  }
  walk(props.communities)
  return map
})

function getCommunityPath(communityId: number) {
  if (communityId === 0) return ''
  const parts: string[] = []
  let current: HierarchicalCommunity | undefined = communityLookup.value.get(communityId)
  while (current && current.id !== 0) {
    parts.unshift(current.name)
    current = current.parent_id != null ? communityLookup.value.get(current.parent_id) : undefined
  }
  return parts.join(' > ')
}

function getCommunityColor(communityId: number) {
  const c = communityLookup.value.get(communityId)
  return c?.color || PALETTE[communityId % PALETTE.length]
}
</script>

<template>
  <Transition
    enter-active-class="transition-transform duration-300 ease-out"
    enter-from-class="translate-x-full"
    enter-to-class="translate-x-0"
    leave-active-class="transition-transform duration-300 ease-in"
    leave-from-class="translate-x-0"
    leave-to-class="translate-x-full"
  >
    <div
      v-if="detail"
      class="absolute right-12 top-12 bottom-12 w-96 bg-(--color-bg-secondary) border-(--color-border-default) flex flex-col z-40 shadow-2xl rounded-2xl transition-colors duration-300"
    >
      <div class="flex items-center justify-between px-6 py-5 border-b border-(--color-border-default) shrink-0">
        <h2 class="text-lg font-semibold text-(--color-text-primary)">{{ t('nodeDetail.title') }}</h2>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="emit('close')"
        >
          <X class="w-6 h-6" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-6 py-6 space-y-6">
        <div>
          <h3 class="text-2xl font-bold text-(--color-text-primary)">{{ detail.node.name }}</h3>
          <div class="flex items-center gap-3 mt-3">
            <div
              class="w-4 h-4 rounded-full"
              :style="{ backgroundColor: getCommunityColor(detail.node.community_id) }"
            />
            <span class="text-sm text-(--color-text-secondary)">{{ getCommunityPath(detail.node.community_id) }}</span>
            <template v-if="detail.node.first_appeared">
              <span class="text-sm text-(--color-text-muted)">·</span>
              <span class="text-sm text-(--color-text-muted)">{{ detail.node.first_appeared }}</span>
            </template>
          </div>
        </div>

        <div class="bg-(--color-bg-primary) rounded-xl p-5 transition-colors duration-300">
          <div class="text-sm text-(--color-text-secondary) mb-3">{{ t('nodeDetail.influenceScore') }}</div>
          <div class="flex items-center gap-4">
            <div class="flex-1 h-3 bg-(--color-border-default) rounded-full overflow-hidden transition-colors duration-300">
              <div
                class="h-full bg-gradient-to-r from-[#58a6ff] to-[#79c0ff] rounded-full transition-all duration-500"
                :style="{ width: `${(detail.node.influence_score / 10) * 100}%` }"
              />
            </div>
            <span class="text-lg font-bold text-(--color-text-primary) w-10 text-right">{{ detail.node.influence_score }}</span>
          </div>
        </div>

        <div>
          <div class="text-sm text-(--color-text-secondary) mb-3">{{ t('nodeDetail.description') }}</div>
          <p class="text-base text-(--color-text-primary) leading-relaxed opacity-90">{{ detail.node.description }}</p>
        </div>

        <div>
          <div class="text-sm text-(--color-text-secondary) mb-4">{{ t('nodeDetail.relatedNodes') }} ({{ detail.neighbors.length }})</div>
          <div class="space-y-3 max-h-80 overflow-y-auto pr-2">
            <div
              v-for="neighbor in detail.neighbors"
              :key="neighbor.id"
              class="flex items-center justify-between px-4 py-3.5 bg-(--color-bg-primary) rounded-xl hover:bg-(--color-bg-tertiary) cursor-pointer transition-colors group"
              @click="emit('neighborClick', neighbor.id)"
            >
              <div class="flex items-center gap-4">
                <div
                  class="w-3.5 h-3.5 rounded-full shrink-0"
                  :style="{ backgroundColor: getCommunityColor(neighbor.community_id) }"
                />
                <span class="text-base text-(--color-text-primary) group-hover:opacity-100 opacity-90">{{ neighbor.name }}</span>
              </div>
              <div class="flex items-center gap-4">
                <span class="text-sm text-(--color-text-muted)">{{ neighbor.relation }}</span>
                <span class="text-sm font-semibold text-[#58a6ff]">{{ (neighbor.weight * 100).toFixed(0) }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>
