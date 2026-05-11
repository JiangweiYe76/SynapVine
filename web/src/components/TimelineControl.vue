<script setup lang="ts">
import { inject, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TimelineState, TimelineActions } from '../composables/useTimeline'
import { historicalEvents } from '../mock/data'

const { t } = useI18n()

const timeline = inject('timeline') as TimelineState & TimelineActions

const speeds = [0.5, 1, 2, 5]

const decadeTicks = computed(() => {
  const ticks: number[] = []
  const start = Math.ceil(timeline.range.value.minYear / 20) * 20
  const end = Math.floor(timeline.range.value.maxYear / 20) * 20
  for (let y = start; y <= end; y += 20) {
    ticks.push(y)
  }
  return ticks
})

const eventYears = computed(() => {
  const set = new Set(historicalEvents.map(e => e.year))
  return [...set].sort((a, b) => a - b)
})

function yearToPercent(year: number): number {
  const range = timeline.range.value.maxYear - timeline.range.value.minYear
  return ((year - timeline.range.value.minYear) / range) * 100
}

const progressPercent = computed(() => yearToPercent(timeline.currentYear.value))

function getCategoryClass(category?: string) {
  switch (category) {
    case 'breakthrough': return 'event-breakthrough'
    case 'release': return 'event-release'
    case 'milestone': return 'event-milestone'
    default: return ''
  }
}

function getCategoryLabel(category?: string) {
  switch (category) {
    case 'breakthrough': return t('timeline.breakthrough')
    case 'release': return t('timeline.release')
    case 'milestone': return t('timeline.milestone')
    default: return ''
  }
}

function getCategoryColor(category?: string) {
  switch (category) {
    case 'breakthrough': return 'var(--color-accent-red)'
    case 'release': return 'var(--color-accent-green)'
    case 'milestone': return 'var(--color-accent-purple)'
    default: return 'var(--color-timeline-accent)'
  }
}
</script>

<template>
  <div class="timeline-container" :class="{ 'is-playing': timeline.isPlaying.value }">
    <div class="timeline-controls-row">
      <div class="playback-btns">
        <button class="ctrl-btn" @click="timeline.seek(timeline.range.value.minYear)" :title="t('timeline.earliest', { year: timeline.range.value.minYear })">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 6h2v12H6zm3.5 6l8.5 6V6z"/></svg>
        </button>
        <button class="ctrl-btn" @click="timeline.prev()" :title="t('timeline.prevYear')">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 6h2v12H6zm9.5 6l-8.5 6V6z"/></svg>
        </button>
        <button
          class="ctrl-btn play-btn"
          @click="timeline.isPlaying.value ? timeline.pause() : timeline.play()"
        >
          <svg v-if="timeline.isPlaying.value" viewBox="0 0 24 24" fill="currentColor"><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
        </button>
        <button class="ctrl-btn" @click="timeline.next()" :title="t('timeline.nextYear')">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
        </button>
        <button class="ctrl-btn" @click="timeline.seek(timeline.range.value.maxYear)" :title="t('timeline.latest', { year: timeline.range.value.maxYear })">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M16 18h2V6h-2zm-11-7l8.5 6V6z"/></svg>
        </button>
      </div>

      <div class="year-display">
        <span class="year-num">{{ timeline.currentYear.value }}</span>
      </div>

      <div class="speed-group">
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
    </div>

    <Transition name="event-fade">
      <div
        v-if="timeline.currentEvent.value"
        class="event-info-bar"
        :class="getCategoryClass(timeline.currentEvent.value.category)"
      >
        <span
          class="info-badge"
          :style="{ background: getCategoryColor(timeline.currentEvent.value.category) }"
        >{{ getCategoryLabel(timeline.currentEvent.value.category) }}</span>
        <span class="info-title">{{ timeline.currentEvent.value.title }}</span>
      </div>
    </Transition>

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
          :min="timeline.range.value.minYear"
          :max="timeline.range.value.maxYear"
          :value="timeline.currentYear.value"
          :style="{ '--progress': `${progressPercent}%` }"
          @input="(e) => timeline.seek(Number((e.target as HTMLInputElement).value))"
        />
        <div class="event-dots">
          <div
            v-for="year in eventYears"
            :key="year"
            class="event-dot"
            :class="getCategoryClass(historicalEvents.find(e => e.year === year)?.category)"
            :style="{ left: `${yearToPercent(year)}%` }"
            :title="historicalEvents.find(e => e.year === year)?.title"
          />
        </div>
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
                : 'translateX(-50%)'
          }"
        >
          {{ decade }}
        </span>
      </div>
    </div>
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
  z-index: 100;
  backdrop-filter: blur(12px);
  transition: border-color 0.3s ease;
}

.timeline-container.is-playing {
  border-color: var(--color-timeline-accent);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3), 0 0 30px var(--color-timeline-accent-dim);
}

.timeline-controls-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.playback-btns {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
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

.ctrl-btn svg {
  width: 15px;
  height: 15px;
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

.year-display {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.year-num {
  font-family: 'JetBrains Mono', monospace;
  font-size: 32px;
  font-weight: 700;
  color: var(--color-timeline-accent);
  letter-spacing: 3px;
  line-height: 1;
}

.speed-group {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
  background: var(--color-bg-tertiary);
  border-radius: 8px;
  padding: 2px;
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

.slider-tick.tick-first {
  display: none;
}

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

.event-dots {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.event-dot {
  position: absolute;
  bottom: 6px;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  transform: translate(-50%, 0);
  background: var(--color-timeline-accent);
  opacity: 0.6;
  transition: opacity 0.2s, transform 0.2s;
}

.event-dot:hover {
  opacity: 1;
}

.event-dot.event-breakthrough {
  background: var(--color-accent-red);
  width: 7px;
  height: 7px;
  transform: translate(-50%, 1px);
}

.event-dot.event-release {
  background: var(--color-accent-green);
}

.event-dot.event-milestone {
  background: var(--color-accent-purple);
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

.event-info-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.info-badge {
  padding: 2px 10px;
  border-radius: 10px;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.info-title {
  font-size: 13px;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.event-fade-enter-active {
  transition: all 0.2s ease;
}

.event-fade-leave-active {
  transition: all 0.15s ease;
}

.event-fade-enter-from,
.event-fade-leave-to {
  opacity: 0;
  margin-bottom: -20px;
}
</style>
