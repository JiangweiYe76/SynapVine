<script setup lang="ts">
import { reactive, watch, onMounted, onUnmounted, inject, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Tabs,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const { t } = useI18n()

const props = defineProps<{
  open?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
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

function close() {
  emit('update:open', false)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    close()
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
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[440px]">
      <DialogHeader>
        <DialogTitle>{{ t('settings.title') }}</DialogTitle>
      </DialogHeader>

      <div class="space-y-8 py-2">
        <section>
          <h3 class="text-sm font-semibold text-(--color-text-secondary) uppercase tracking-wider mb-4">{{ t('settings.appearance') }}</h3>

          <div class="space-y-5">
            <div class="flex items-center justify-between">
              <label class="text-sm text-(--color-text-primary)">{{ t('settings.theme') }}</label>
              <Tabs :model-value="currentTheme" @update:model-value="themeComposable.setTheme($event as any)">
                <TabsList class="grid w-32 grid-cols-2">
                  <TabsTrigger value="dark">{{ t('settings.themeDark') }}</TabsTrigger>
                  <TabsTrigger value="light">{{ t('settings.themeLight') }}</TabsTrigger>
                </TabsList>
              </Tabs>
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
              <Select :model-value="settings.timelineSpeed.toString()" @update:model-value="(v: any) => { if (v) settings.timelineSpeed = parseFloat(v) }">
                <SelectTrigger class="w-20">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="s in speeds" :key="s" :value="s.toString()">
                    {{ s }}x
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="flex items-center justify-between">
              <label class="text-sm text-(--color-text-primary)">{{ t('settings.autoPlayTimeline') }}</label>
              <Switch
                :checked="settings.autoPlayTimeline"
                @update:checked="settings.autoPlayTimeline = $event"
              />
            </div>
          </div>
        </section>
      </div>

      <DialogFooter>
        <Button
          variant="ghost"
          class="text-(--color-text-secondary) hover:text-(--color-accent-red)"
          @click="resetDefaults"
        >
          {{ t('settings.resetDefaults') }}
        </Button>
        <Button @click="close">
          {{ t('settings.done') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
