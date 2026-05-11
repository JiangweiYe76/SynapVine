import { ref, computed, watch, type Ref } from 'vue'
import type { GraphNode, GraphEdge, HistoricalEvent, TimelineRange } from '../types/graph'
import { historicalEvents, getTimelineRange } from '../mock/data'

export interface TimelineState {
  currentYear: Ref<number>
  isPlaying: Ref<boolean>
  playbackSpeed: Ref<number>
  range: Ref<TimelineRange>
  visibleNodes: Ref<GraphNode[]>
  visibleEdges: Ref<GraphEdge[]>
  currentEvent: Ref<HistoricalEvent | null>
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
  allEdges: Ref<GraphEdge[]>
): TimelineComposable {
  const range = ref<TimelineRange>(getTimelineRange())
  const currentYear = ref(range.value.maxYear)
  const isPlaying = ref(false)
  const playbackSpeed = ref(1)
  const currentEvent = ref<HistoricalEvent | null>(null)

  let playInterval: ReturnType<typeof setInterval> | null = null

  const visibleNodes = computed(() => {
    return allNodes.value.filter(n => 
      (n.first_appeared || 3000) <= currentYear.value
    )
  })

  const visibleEdges = computed(() => {
    const visibleNodeIds = new Set(visibleNodes.value.map(n => n.id))
    return allEdges.value.filter(e => 
      visibleNodeIds.has(e.source) && visibleNodeIds.has(e.target)
    )
  })

  watch(currentYear, (year) => {
    const event = historicalEvents.find(e => e.year === year)
    currentEvent.value = event || null
  }, { immediate: true })

  function play() {
    if (isPlaying.value) return
    isPlaying.value = true
    
    playInterval = setInterval(() => {
      if (currentYear.value >= range.value.maxYear) {
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
    currentYear.value = Math.max(range.value.minYear, Math.min(range.value.maxYear, year))
  }

  function setSpeed(speed: number) {
    playbackSpeed.value = speed
    if (isPlaying.value) {
      pause()
      play()
    }
  }

  function next() {
    if (currentYear.value < range.value.maxYear) {
      currentYear.value += 1
    }
  }

  function prev() {
    if (currentYear.value > range.value.minYear) {
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
    currentEvent,
    play,
    pause,
    seek,
    setSpeed,
    next,
    prev,
  }
}
