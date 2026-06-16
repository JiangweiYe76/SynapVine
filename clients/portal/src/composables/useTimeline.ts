import { ref, computed, watch, type Ref } from 'vue'
import type { GraphNode, GraphEdge, TimelineRange } from '../types/graph'

export interface TimelineState {
  currentYear: Ref<number>
  isPlaying: Ref<boolean>
  playbackSpeed: Ref<number>
  range: Ref<TimelineRange>
  visibleNodes: Ref<GraphNode[]>
  visibleEdges: Ref<GraphEdge[]>
}

export interface TimelineActions {
  play: () => void
  pause: () => void
  seek: (year: number) => void
  setSpeed: (speed: number) => void
  next: () => void
  prev: () => void
}

export type TimelineComposable = TimelineState & TimelineActions

export function useTimeline(
  allNodes: Ref<GraphNode[]>,
  allEdges: Ref<GraphEdge[]>,
  timelineRange: Ref<TimelineRange>,
): TimelineComposable {
  // Use the server-computed range as the source of truth. It reflects
  // the full graph extent, not the currently loaded window, so the
  // timeline slider always spans the actual dataset.
  const range = computed<TimelineRange>(() => timelineRange.value)

  const fallbackYear = new Date().getFullYear()
  const currentYear = ref(fallbackYear)
  const isPlaying = ref(false)
  const playbackSpeed = ref(1)

  // Clamp currentYear into the (possibly newly arrived) range so the
  // slider position and visibleNodes stay in sync with the data.
  watch(range, (r) => {
    if (currentYear.value > r.max_year) currentYear.value = r.max_year
    if (currentYear.value < r.min_year) currentYear.value = r.min_year
  })

  let playInterval: ReturnType<typeof setInterval> | null = null

  const visibleNodes = computed(() => {
    return allNodes.value.filter(n => {
      const year = n.first_appeared ? parseInt(n.first_appeared.split('-')[0], 10) : 3000
      return year <= currentYear.value
    })
  })

  const visibleEdges = computed(() => {
    const visibleNodeIds = new Set(visibleNodes.value.map(n => n.id))
    return allEdges.value.filter(e => 
      visibleNodeIds.has(e.source) && visibleNodeIds.has(e.target)
    )
  })

  function play() {
    if (isPlaying.value) return
    isPlaying.value = true
    
    playInterval = setInterval(() => {
      if (currentYear.value >= range.value.max_year) {
        pause()
        return
      }
      currentYear.value += 1
    }, 1000 / playbackSpeed.value)
  }

  function pause() {
    isPlaying.value = false
    if (playInterval) {
      clearInterval(playInterval)
      playInterval = null
    }
  }

  function seek(year: number) {
    currentYear.value = Math.max(range.value.min_year, Math.min(range.value.max_year, year))
  }

  function setSpeed(speed: number) {
    playbackSpeed.value = speed
    if (isPlaying.value) {
      pause()
      play()
    }
  }

  function next() {
    if (currentYear.value < range.value.max_year) {
      currentYear.value += 1
    }
  }

  function prev() {
    if (currentYear.value > range.value.min_year) {
      currentYear.value -= 1
    }
  }

  return {
    currentYear,
    isPlaying,
    playbackSpeed,
    range,
    visibleNodes,
    visibleEdges,
    play,
    pause,
    seek,
    setSpeed,
    next,
    prev,
  }
}
