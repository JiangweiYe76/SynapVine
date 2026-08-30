<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { injectTimeline } from '../composables/useGraph'
import {
  SkipBack, ChevronLeft, Play, Pause, ChevronRight, SkipForward,
  ChevronDown, ChevronUp, Clock,
} from 'lucide-vue-next'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

const { t } = useI18n()

const timeline = injectTimeline()

const props = defineProps<{
  modelValue?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const speeds = [0.5, 1, 2, 5]

const decadeTicks = computed(() => {
  const ticks: number[] = []
  const start = Math.ceil(timeline.range.value.min_year / 20) * 20
  const end = Math.floor(timeline.range.value.max_year / 20) * 20
  for (let y = start; y <= end; y += 20) {
    ticks.push(y)
  }
  return ticks
})

function yearToPercent(year: number): number {
  const range = timeline.range.value.max_year - timeline.range.value.min_year
  return ((year - timeline.range.value.min_year) / range) * 100
}

const progressPercent = computed(() => yearToPercent(timeline.currentYear.value))
</script>

<template>
  <div>
    <Transition name="timeline-slide">
      <div
        v-show="modelValue !== false"
        class="timeline-container"
        :class="{ 'is-playing': timeline.isPlaying.value }"
      >
        <div class="flex items-center gap-4">
          <!-- Playback -->
          <div class="flex items-center gap-0.5 shrink-0">
            <Tooltip>
              <TooltipTrigger as-child>
                <button
                  class="ctrl-btn"
                  @click="timeline.seek(timeline.range.value.min_year)"
                >
                  <SkipBack :size="15" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ t('timeline.earliest', { year: timeline.range.value.min_year }) }}</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
                <button
                  class="ctrl-btn"
                  @click="timeline.prev()"
                >
                  <ChevronLeft :size="15" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ t('timeline.prevYear') }}</p>
              </TooltipContent>
            </Tooltip>
            <button
              class="ctrl-btn play-btn"
              @click="timeline.isPlaying.value ? timeline.pause() : timeline.play()"
            >
              <Pause v-if="timeline.isPlaying.value" :size="18" />
              <Play v-else :size="18" />
            </button>
            <Tooltip>
              <TooltipTrigger as-child>
                <button
                  class="ctrl-btn"
                  @click="timeline.next()"
                >
                  <ChevronRight :size="15" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ t('timeline.nextYear') }}</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
                <button
                  class="ctrl-btn"
                  @click="timeline.seek(timeline.range.value.max_year)"
                >
                  <SkipForward :size="15" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ t('timeline.latest', { year: timeline.range.value.max_year }) }}</p>
              </TooltipContent>
            </Tooltip>
          </div>

          <!-- Year -->
          <div class="flex-1 flex items-center justify-center">
            <span class="year-num">{{ timeline.currentYear.value }}</span>
          </div>

          <!-- Speed -->
          <div class="flex items-center gap-0.5 shrink-0 bg-(--color-bg-tertiary) rounded-lg p-0.5">
            <button
              v-for="s in speeds"
              :key="s"
              class="speed-btn"
              :class="{ active: timeline.playbackSpeed.value === s }"
              @click="timeline.setSpeed(s)"
            >
              {{ s }}×
            </button>
          </div>

          <!-- Collapse -->
          <Tooltip>
            <TooltipTrigger as-child>
              <button
                class="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-(--color-bg-tertiary) text-(--color-text-muted) hover:text-(--color-text-primary) transition-colors shrink-0"
                @click="emit('update:modelValue', false)"
              >
                <ChevronDown :size="16" />
              </button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{{ t('timeline.hide') }}</p>
            </TooltipContent>
          </Tooltip>
        </div>

        <!-- Track -->
        <div class="timeline-track">
          <div class="slider-wrapper">
            <div class="slider-bg">
              <div class="slider-fill" :style="{ width: `${progressPercent}%` }" />
              <div class="slider-ticks">
                <div
                  v-for="(decade, index) in decadeTicks"
                  :key="decade"
                  class="slider-tick"
                  :style="{ left: `${yearToPercent(decade)}%` }"
                  :class="{
                    'tick-first': index === 0,
                    'tick-last': index === decadeTicks.length - 1,
                    'tick-passed': decade <= timeline.currentYear.value,
                  }"
                />
              </div>
            </div>
            <input
              type="range"
              class="timeline-slider"
              :min="timeline.range.value.min_year"
              :max="timeline.range.value.max_year"
              :value="timeline.currentYear.value"
              :style="{ '--progress': `${progressPercent}%` }"
              @input="(e) => timeline.seek(Number((e.target as HTMLInputElement).value))"
            />
          </div>
          <div class="tick-labels">
            <span
              v-for="(decade, index) in decadeTicks"
              :key="decade"
              class="tick-label"
              :class="{ 'tick-passed': decade <= timeline.currentYear.value }"
              :style="{
                left: `${yearToPercent(decade)}%`,
                transform: index === 0
                  ? 'translateX(0)'
                  : index === decadeTicks.length - 1
                    ? 'translateX(-100%)'
                    : 'translateX(-50%)',
              }"
            >
              {{ decade }}
            </span>
          </div>
        </div>
      </div>
    </Transition>

    <Transition name="trigger-fade">
      <Tooltip>
        <TooltipTrigger as-child>
          <div
            v-show="modelValue === false"
            class="timeline-trigger"
            @click="emit('update:modelValue', true)"
          >
        <Clock :size="14" />
        <span class="text-xs font-medium whitespace-nowrap">{{ t('timeline.show') }}</span>
        <ChevronUp :size="14" />
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{{ t('timeline.show') }}</p>
        </TooltipContent>
      </Tooltip>
    </Transition>
  </div>
</template>

<style scoped>
.timeline-container {
  position: fixed;
  bottom: 60px;
  left: 50%;
  transform: translateX(-50%);
  width: 92%;
  max-width: 900px;
  padding: 16px 24px 18px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border-default);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  z-index: 40;
  backdrop-filter: blur(12px);
  transition: border-color 0.3s ease;
}

.timeline-container.is-playing {
  border-color: var(--color-timeline-accent);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3), 0 0 30px var(--color-timeline-accent-dim);
}

.ctrl-btn {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}

.ctrl-btn:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.play-btn {
  width: 36px;
  height: 36px;
  background: var(--color-timeline-accent);
  border-radius: 50%;
  color: #0a0e17;
}

.play-btn:hover {
  background: var(--color-timeline-accent);
  filter: brightness(1.1);
  transform: scale(1.08);
}

.year-num {
  font-family: 'JetBrains Mono', monospace;
  font-size: 32px;
  font-weight: 700;
  color: var(--color-timeline-accent);
  letter-spacing: 3px;
  line-height: 1;
}

.speed-btn {
  padding: 4px 9px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  cursor: pointer;
  transition: all 0.15s;
}

.speed-btn:hover {
  color: var(--color-text-primary);
}

.speed-btn.active {
  background: var(--color-bg-primary);
  color: var(--color-timeline-accent);
  font-weight: 600;
}

.timeline-track {
  position: relative;
  margin-top: 12px;
  padding-bottom: 22px;
}

.slider-wrapper {
  position: relative;
  height: 28px;
}

.slider-bg {
  position: absolute;
  top: 10px;
  left: 0;
  right: 0;
  height: 6px;
  background: var(--color-timeline-track);
  border-radius: 3px;
  overflow: hidden;
}

.slider-fill {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: var(--color-timeline-accent);
  border-radius: 3px;
  transition: width 0.15s linear;
  opacity: 0.6;
}

.slider-ticks {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 100%;
}

.slider-tick {
  position: absolute;
  top: -4px;
  width: 2px;
  height: 14px;
  background: var(--color-timeline-track);
  transform: translateX(-50%);
  transition: background 0.2s;
}

.slider-tick.tick-first,
.slider-tick.tick-last {
  display: none;
}

.slider-tick.tick-passed {
  background: var(--color-timeline-accent);
}

.timeline-slider {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
  outline: none;
  cursor: pointer;
  z-index: 2;
  margin: 0;
}

.timeline-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 22px;
  height: 22px;
  background: var(--color-timeline-thumb);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 0 12px var(--color-timeline-accent-dim);
  border: 2px solid var(--color-bg-secondary);
  margin-top: -8px;
}

.timeline-slider::-moz-range-thumb {
  width: 22px;
  height: 22px;
  background: var(--color-timeline-thumb);
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid var(--color-bg-secondary);
  box-shadow: 0 0 12px var(--color-timeline-accent-dim);
}

.tick-labels {
  position: relative;
  height: 18px;
  margin-top: 6px;
}

.tick-label {
  position: absolute;
  font-size: 11px;
  color: var(--color-text-muted);
  font-family: 'JetBrains Mono', monospace;
  transition: color 0.2s;
}

.tick-label.tick-passed {
  color: var(--color-text-primary);
  font-weight: 600;
}

.timeline-trigger {
  position: fixed;
  bottom: 60px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 18px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border-default);
  border-radius: 20px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(12px);
  cursor: pointer;
  z-index: 40;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.timeline-trigger:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
  transform: translateX(-50%) translateY(-2px);
  box-shadow: 0 6px 28px rgba(0, 0, 0, 0.35);
}

/* Transitions */
.timeline-slide-enter-active,
.timeline-slide-leave-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.timeline-slide-enter-from,
.timeline-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(100%);
}

.trigger-fade-enter-active,
.trigger-fade-leave-active {
  transition: all 0.25s ease;
}

.trigger-fade-enter-from,
.trigger-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(8px);
}
</style>
