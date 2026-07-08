<template>
  <div ref="containerRef" class="danmaku-layer" />
</template>

<script setup lang="ts">
import type { DanmakuItem } from '~/types'
import { useApi } from '~/composables/useApi'

const props = defineProps<{
  videoId: number
  enabled: boolean
  videoEl?: HTMLVideoElement | null
}>()

const { get } = useApi()
const containerRef = ref<HTMLDivElement | null>(null)
let ws: WebSocket | null = null

// ---- Constants ----
const TRACK_COUNT = 5
const ANIMATION_DURATION_MS = 8000
const TIME_WINDOW_BEFORE = 1 // show danmaku 1s before its play_time
const TIME_WINDOW_AFTER = 5  // show danmaku up to 5s after its play_time

// ---- State ----
const danmakuPool = ref<DanmakuItem[]>([])
const playedHistoryIds = new Set<number>()
const trackBusyUntil: number[] = new Array(TRACK_COUNT).fill(0)
const recentContentKeys = new Map<string, number>() // contentKey -> timestamp, for dedup
let isPaused = false
let checkInterval: ReturnType<typeof setInterval> | null = null
let videoEl: HTMLVideoElement | null = null
let onVideoPlay: (() => void) | null = null
let onVideoPause: (() => void) | null = null
let onVideoSeeked: (() => void) | null = null

// ---- Helpers ----
function makeContentKey(item: DanmakuItem): string {
  return `${item.content}|${item.user?.id || 0}`
}

function getTrackTop(trackIndex: number, containerHeight: number): number {
  const startPercent = 0.05 // start at 5% from top
  const spacing = (0.9 * containerHeight) / TRACK_COUNT // use 90% of height
  return startPercent * containerHeight + trackIndex * spacing
}

function assignTrack(): { track: number; top: number } {
  const container = containerRef.value
  const containerHeight = container?.clientHeight || 400
  const now = Date.now()

  // Find the track that becomes free earliest
  let bestTrack = 0
  let earliestFree = trackBusyUntil[0]

  for (let i = 0; i < TRACK_COUNT; i++) {
    if (now >= trackBusyUntil[i]) {
      // Track is free now
      bestTrack = i
      break
    }
    if (trackBusyUntil[i] < earliestFree) {
      earliestFree = trackBusyUntil[i]
      bestTrack = i
    }
  }

  // Mark track busy for animation duration
  trackBusyUntil[bestTrack] = Math.max(now, trackBusyUntil[bestTrack]) + ANIMATION_DURATION_MS
  const top = getTrackTop(bestTrack, containerHeight)

  return { track: bestTrack, top }
}

// ---- Dedup check ----
function isDuplicate(item: DanmakuItem): boolean {
  const key = makeContentKey(item)
  const lastTime = recentContentKeys.get(key)
  if (lastTime && Date.now() - lastTime < 15000) {
    // Same content+user within 15s — it's a duplicate
    return true
  }
  recentContentKeys.set(key, Date.now())

  // Clean old entries periodically
  if (recentContentKeys.size > 50) {
    const cutoff = Date.now() - 30000
    for (const [k, t] of recentContentKeys) {
      if (t < cutoff) recentContentKeys.delete(k)
    }
  }
  return false
}

// ---- Create danmaku DOM element ----
function createDanmakuEl(item: DanmakuItem): HTMLElement | null {
  if (!containerRef.value) return null

  const el = document.createElement('div')
  el.className = 'danmaku-item'
  el.dataset.danmakuId = String(item.id)

  const color = item.color || '#ffffff'
  const position = item.position || 0

  el.style.color = color
  el.style.position = 'absolute'
  el.style.whiteSpace = 'nowrap'
  el.style.pointerEvents = 'none'
  el.style.textShadow = '1px 1px 2px rgba(0,0,0,0.8)'
  el.style.fontWeight = 'bold'
  el.style.zIndex = '10'
  el.style.willChange = 'transform'

  // Size
  if (item.size === 0) el.style.fontSize = '14px'
  else if (item.size === 2) el.style.fontSize = '20px'
  else el.style.fontSize = '16px'

  el.textContent = item.content

  const { track, top } = assignTrack()
  el.dataset.track = String(track)

  if (position === 1) {
    // Top danmaku
    el.style.top = `${top}px`
    el.style.left = '50%'
    el.style.transform = 'translateX(-50%)'
    el.style.animation = 'danmaku-fade 3s ease-out forwards'
  } else if (position === 2) {
    // Bottom danmaku
    el.style.top = `${containerRef.value.clientHeight * 0.7 + (track * 30)}px`
    el.style.left = '50%'
    el.style.transform = 'translateX(-50%)'
    el.style.animation = 'danmaku-fade 3s ease-out forwards'
  } else {
    // Scroll danmaku
    const containerWidth = containerRef.value.clientWidth || window.innerWidth
    el.style.setProperty('--danmaku-start-x', `${containerWidth}px`)
    el.style.top = `${top}px`
    el.style.left = '0'
    el.style.animation = `danmaku-scroll ${ANIMATION_DURATION_MS / 1000}s linear forwards`
  }

  if (isPaused) {
    el.style.animationPlayState = 'paused'
  }

  containerRef.value.appendChild(el)

  el.addEventListener('animationend', () => {
    if (el.parentNode) el.parentNode.removeChild(el)
  }, { once: true })

  return el
}

// ---- Public: add danmaku (called from parent for optimistic render) ----
function addDanmaku(item: DanmakuItem, isRealTime = true) {
  if (!props.enabled) return
 
  if (isRealTime && isDuplicate(item)) {
    return
  }

  if (isRealTime) {
    createDanmakuEl(item)
    danmakuPool.value.push(item)
    playedHistoryIds.add(item.id)
  } else {
    // History danmaku: add to pool for time-based rendering
    danmakuPool.value.push(item)
    // Check if should show now
    const currentTime = videoEl?.currentTime || 0
    if (isInTimeWindow(item.play_time, currentTime) && !playedHistoryIds.has(item.id)) {
      playedHistoryIds.add(item.id)
      createDanmakuEl(item)
    }
  }
}

// ---- Time checking ----
function isInTimeWindow(playTime: number, currentTime: number): boolean {
  return currentTime >= playTime - TIME_WINDOW_BEFORE && currentTime <= playTime + TIME_WINDOW_AFTER
}

function checkDanmaku() {
  if (!props.enabled || isPaused) return

  const currentTime = videoEl?.currentTime || 0

  for (const item of danmakuPool.value) {
    // Skip if already played
    if (playedHistoryIds.has(item.id)) continue

    // Check time window
    if (isInTimeWindow(item.play_time, currentTime)) {
      playedHistoryIds.add(item.id)
      createDanmakuEl(item)
    }
  }
}

function resetPlayedHistory() {
  // When seeking, clear played history so danmaku can reappear
  playedHistoryIds.clear()
}

function checkDanmakuOnSeek() {
  if (!props.enabled) return
  const currentTime = videoEl?.currentTime || 0

  // Clear existing DOM elements (they're from old time position)
  if (containerRef.value) {
    const existing = containerRef.value.querySelectorAll('.danmaku-item')
    existing.forEach((el) => el.remove())
  }

  // Reset played history and re-check
  resetPlayedHistory()

  for (const item of danmakuPool.value) {
    if (isInTimeWindow(item.play_time, currentTime)) {
      playedHistoryIds.add(item.id)
      createDanmakuEl(item)
    }
  }
}

// ---- Pause / Resume ----
function pauseAll() {
  isPaused = true
  if (!containerRef.value) return
  const els = containerRef.value.querySelectorAll<HTMLElement>('.danmaku-item')
  els.forEach((el) => {
    el.style.animationPlayState = 'paused'
  })
}

function resumeAll() {
  isPaused = false
  if (!containerRef.value) return
  const els = containerRef.value.querySelectorAll<HTMLElement>('.danmaku-item')
  els.forEach((el) => {
    el.style.animationPlayState = 'running'
  })
}

// ---- WebSocket ----
function connectWebSocket() {
  if (!props.enabled) return

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const isDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
  const wsHost = isDev ? window.location.hostname + ':8080' : window.location.host
  ws = new WebSocket(`${protocol}://${wsHost}/api/v1/ws/danmaku/${props.videoId}`)

  ws.onmessage = (event) => {
    try {
      const data: DanmakuItem = JSON.parse(event.data)
      // Direct call — no debounce, no delay
      addDanmaku(data, true)
    } catch {
      // ignore parse errors
    }
  }

  ws.onclose = () => {
    setTimeout(() => connectWebSocket(), 3000)
  }

  ws.onerror = () => {
    ws?.close()
  }
}

// ---- History loading ----
async function loadHistory() {
  try {
    const data = await get<DanmakuItem[]>(`/api/v1/videos/${props.videoId}/danmaku`)
    if (data && data.length > 0) {
      // Mark all as history (non-real-time)
      data.forEach((item) => {
        danmakuPool.value.push(item)
      })
      // Check immediately for current video time
      checkDanmaku()
    }
  } catch {
    // silent fail
  }
}

// ---- Video element binding ----
function bindVideoEvents(el: HTMLVideoElement) {
  videoEl = el

  onVideoPlay = () => resumeAll()
  onVideoPause = () => pauseAll()
  onVideoSeeked = () => checkDanmakuOnSeek()

  el.addEventListener('play', onVideoPlay)
  el.addEventListener('pause', onVideoPause)
  el.addEventListener('seeked', onVideoSeeked)
}

function unbindVideoEvents() {
  if (!videoEl) return
  if (onVideoPlay) videoEl.removeEventListener('play', onVideoPlay)
  if (onVideoPause) videoEl.removeEventListener('pause', onVideoPause)
  if (onVideoSeeked) videoEl.removeEventListener('seeked', onVideoSeeked)
  videoEl = null
}

// ---- Time check interval ----
function startTimeCheck() {
  checkInterval = setInterval(() => {
    if (props.enabled && !isPaused && videoEl && !videoEl.paused) {
      checkDanmaku()
    }
  }, 250)
}

function stopTimeCheck() {
  if (checkInterval) {
    clearInterval(checkInterval)
    checkInterval = null
  }
}

// ---- Watch videoEl prop ----
watch(
  () => props.videoEl,
  (el) => {
    unbindVideoEvents()
    if (el) {
      bindVideoEvents(el)
      // Sync pause state
      if (el.paused) {
        isPaused = true
      }
    }
  },
)

// ---- Lifecycle ----
onMounted(() => {
  loadHistory()
  connectWebSocket()
  startTimeCheck()

  // If videoEl prop is already set, bind it
  if (props.videoEl) {
    bindVideoEvents(props.videoEl)
  }
})

onUnmounted(() => {
  stopTimeCheck()
  unbindVideoEvents()
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
})

defineExpose({ addDanmaku, pauseAll, resumeAll })
</script>

<style scoped>
.danmaku-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  pointer-events: none;
  z-index: 5;
}
</style>

<style>
@keyframes danmaku-scroll {
  from {
    transform: translateX(var(--danmaku-start-x, 100vw));
  }
  to {
    transform: translateX(-100%);
  }
}

@keyframes danmaku-fade {
  0% {
    opacity: 1;
  }
  80% {
    opacity: 1;
  }
  100% {
    opacity: 0;
  }
}
</style>
