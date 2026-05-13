<script setup lang="ts">
import { reactive, watch, onMounted, onUnmounted, inject, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Settings, X } from 'lucide-vue-next'

const { t } = useI18n()

const emit = defineEmits<{
  close: []
}>()

const themeComposable = inject<any>('theme')!

const currentTheme = computed(() => themeComposable.theme.value as 'dark' | 'light')

interface SettingsData {
  nodeScale: number
  edgeOpacity: number
  autoPlayTimeline: boolean
  timelineSpeed: number
}

const defaults: SettingsData = {
  nodeScale: 1.0,
  edgeOpacity: 0.6,
  autoPlayTimeline: false,
  timelineSpeed: 1,
}
const STORAGE_KEY = 'ai-graph-settings'

function loadSettings(): SettingsData {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return { ...defaults, ...JSON.parse(stored) }
    }
  } catch {}
  return { ...defaults }
}

const settings = reactive<SettingsData>(loadSettings())

watch(
  () => ({ ...settings }),
  (val) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
  },
  { deep: true }
)

function resetDefaults() {
  Object.assign(settings, defaults)
}

const speeds = [0.5, 1, 2, 5]

function handleOverlayClick(e: MouseEvent) {
  if ((e.target as HTMLElement).dataset.overlay !== undefined) {
    emit('close')
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div
    data-overlay
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]"
    @click="handleOverlayClick"
  >
    <div
      class="bg-(--color-bg-secondary) border border-(--color-border-default) rounded-2xl shadow-2xl w-[440px] max-h-[85vh] flex flex-col transition-colors duration-300"
    >
      <div class="flex items-center justify-between px-6 py-5 border-b border-(--color-border-default) shrink-0">
        <div class="flex items-center gap-3">
          <Settings class="w-6 h-6 text-(--color-text-secondary)" />
          <h2 class="text-lg font-semibold text-(--color-text-primary)">{{ t('settings.title') }}</h2>
        </div>
        <button
          class="w-10 h-10 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-secondary) hover:text-(--color-text-primary) transition-colors"
          @click="emit('close')"
        >
          <X class="w-6 h-6" />
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-6 py-6 space-y-8">
        <section>
          <h3 class="text-sm font-semibold text-(--color-text-secondary) uppercase tracking-wider mb-4">{{ t('settings.appearance') }}</h3>

          <div class="space-y-5">
            <div class="flex items-center justify-between">
              <label class="text-sm text-(--color-text-primary)">{{ t('settings.theme') }}</label>
              <div class="flex bg-(--color-bg-tertiary) rounded-lg p-1 gap-1">
                <button
                  class="px-4 py-1.5 rounded-md text-sm transition-colors"
                  :class="currentTheme === 'dark'
                    ? 'bg-(--color-bg-primary) text-(--color-text-primary) shadow-sm'
                    : 'text-(--color-text-secondary) hover:text-(--color-text-primary)'"
                  @click="themeComposable.setTheme('dark')"
                >
                  {{ t('settings.themeDark') }}
                </button>
                <button
                  class="px-4 py-1.5 rounded-md text-sm transition-colors"
                  :class="currentTheme === 'light'
                    ? 'bg-(--color-bg-primary) text-(--color-text-primary) shadow-sm'
                    : 'text-(--color-text-secondary) hover:text-(--color-text-primary)'"
                  @click="themeComposable.setTheme('light')"
                >
                  {{ t('settings.themeLight') }}
                </button>
              </div>
            </div>

            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="text-sm text-(--color-text-primary)">{{ t('settings.nodeScale') }}</label>
                <span class="text-sm font-mono text-(--color-accent-blue)">{{ settings.nodeScale.toFixed(1) }}x</span>
              </div>
              <input
                type="range"
                min="0.5"
                max="2.0"
                step="0.1"
                :value="settings.nodeScale"
                @input="settings.nodeScale = parseFloat(($event.target as HTMLInputElement).value)"
                class="w-full h-2 bg-(--color-bg-tertiary) rounded-full appearance-none cursor-pointer
                  [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4
                  [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-(--color-accent-blue)
                  [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:transition-transform
                  [&::-webkit-slider-thumb]:hover:scale-125"
              />
            </div>

            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="text-sm text-(--color-text-primary)">{{ t('settings.edgeOpacity') }}</label>
                <span class="text-sm font-mono text-(--color-accent-blue)">{{ (settings.edgeOpacity * 100).toFixed(0) }}%</span>
              </div>
              <input
                type="range"
                min="0.3"
                max="1.0"
                step="0.05"
                :value="settings.edgeOpacity"
                @input="settings.edgeOpacity = parseFloat(($event.target as HTMLInputElement).value)"
                class="w-full h-2 bg-(--color-bg-tertiary) rounded-full appearance-none cursor-pointer
                  [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4
                  [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-(--color-accent-blue)
                  [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:transition-transform
                  [&::-webkit-slider-thumb]:hover:scale-125"
              />
            </div>
          </div>
        </section>

        <section>
          <h3 class="text-sm font-semibold text-(--color-text-secondary) uppercase tracking-wider mb-4">{{ t('settings.timeline') }}</h3>

          <div class="space-y-5">
            <div class="flex items-center justify-between">
              <label class="text-sm text-(--color-text-primary)">{{ t('settings.defaultSpeed') }}</label>
              <select
                :value="settings.timelineSpeed"
                @change="settings.timelineSpeed = parseFloat(($event.target as HTMLSelectElement).value)"
                class="bg-(--color-bg-tertiary) border border-(--color-border-default) rounded-lg px-3 py-1.5 text-sm text-(--color-text-primary) outline-none focus:border-(--color-accent-blue) transition-colors cursor-pointer"
              >
                <option v-for="s in speeds" :key="s" :value="s">{{ s }}x</option>
              </select>
            </div>

            <div class="flex items-center justify-between">
              <label class="text-sm text-(--color-text-primary)">{{ t('settings.autoPlayTimeline') }}</label>
              <button
                role="switch"
                :aria-checked="settings.autoPlayTimeline"
                class="relative w-11 h-6 rounded-full transition-colors cursor-pointer"
                :class="settings.autoPlayTimeline ? 'bg-(--color-accent-blue)' : 'bg-(--color-bg-tertiary)'"
                @click="settings.autoPlayTimeline = !settings.autoPlayTimeline"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform shadow-sm"
                  :class="settings.autoPlayTimeline ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>
          </div>
        </section>


      </div>

      <div class="px-6 py-4 border-t border-(--color-border-default) flex justify-between items-center shrink-0">
        <button
          class="px-4 py-2 text-sm text-(--color-text-secondary) hover:text-(--color-accent-red) rounded-lg hover:bg-(--color-bg-tertiary) transition-colors"
          @click="resetDefaults"
        >
          {{ t('settings.resetDefaults') }}
        </button>
        <button
          class="px-6 py-2 text-sm font-medium bg-(--color-accent-blue) text-white rounded-lg hover:opacity-90 transition-opacity"
          @click="emit('close')"
        >
          {{ t('settings.done') }}
        </button>
      </div>
    </div>
  </div>
</template>
