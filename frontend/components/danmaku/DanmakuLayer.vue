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
const SCROLL_DURATION = 7   // seconds of video time to cross the screen (Bilibili style)
const FADE_DURATION = 3     // seconds of video time for fixed (top/bottom) danmaku to stay

// ---- State ----
interface PoolEntry {
  key: string
  item: DanmakuItem
}

const danmakuPool = ref<PoolEntry[]>([])
const recentContentKeys = new Map<string, number>()
let rafId: number | null = null
let videoEl: HTMLVideoElement | null = null
let onVideoSeeked: (() => void) | null = null
let optCounter = 0

// ---- Helpers ----
function makeContentKey(item: DanmakuItem): string {
  return `${item.content}|${item.user?.id || 0}`
}

function getTrackTop(trackIndex: number, containerHeight: number): number {
  const startPercent = 0.05
  const spacing = (0.9 * containerHeight) / TRACK_COUNT
  return startPercent * containerHeight + trackIndex * spacing
}

/** Generate a stable unique key for a pool entry. */
function makePoolKey(item: DanmakuItem): string {
  if (item.id > 0) return String(item.id)
  return `o:${++optCounter}`
}

// ---- Dedup ----
function isDuplicate(item: DanmakuItem): boolean {
  const key = makeContentKey(item)
  const lastTime = recentContentKeys.get(key)
  if (lastTime && Date.now() - lastTime < 15000) return true
  recentContentKeys.set(key, Date.now())
  if (recentContentKeys.size > 50) {
    const cutoff = Date.now() - 30000
    for (const [k, t] of recentContentKeys) {
      if (t < cutoff) recentContentKeys.delete(k)
    }
  }
  return false
}

// ---- RAF loop: drive all danmaku positions by video currentTime ----
function startRafLoop() {
  stopRafLoop()
  const tick = () => {
    updateDanmakus()
    rafId = requestAnimationFrame(tick)
  }
  rafId = requestAnimationFrame(tick)
}

function stopRafLoop() {
  if (rafId !== null) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
}

function updateDanmakus() {
  if (!containerRef.value || !videoEl || !props.enabled) return

  const currentTime = videoEl.currentTime
  const container = containerRef.value
  const cw = container.clientWidth
  const ch = container.clientHeight

  // Track which tracks are occupied by current visible danmakus
  const trackOccupied = new Array(TRACK_COUNT).fill(false)
  const activeKeys = new Set<string>()

  // ── Step 1: update positions of existing elements ──
  const existing = container.querySelectorAll<HTMLElement>('.danmaku-item')
  for (const el of existing) {
    const playTime = parseFloat(el.dataset.playTime || '0')
    const posType = parseInt(el.dataset.position || '0')
    const track = parseInt(el.dataset.track || '0') || 0
    const elapsed = currentTime - playTime

    if (posType === 0) {
      // Scrolling danmaku
      if (elapsed < 0 || elapsed >= SCROLL_DURATION) {
        el.remove()
        continue
      }
      const progress = elapsed / SCROLL_DURATION
      const startX = cw + 20
      const x = startX - (startX + el.offsetWidth) * progress
      el.style.transform = `translateX(${x}px)`
    } else {
      // Fixed (top / bottom) danmaku
      if (elapsed < 0 || elapsed >= FADE_DURATION) {
        el.remove()
        continue
      }
      // Fade out during the last 20 % of duration
      const fadeStart = FADE_DURATION * 0.8
      if (elapsed >= fadeStart) {
        el.style.opacity = String(1 - (elapsed - fadeStart) / (FADE_DURATION - fadeStart))
      }
    }

    trackOccupied[track] = true
    const key = el.dataset.danmakuId || ''
    if (key) activeKeys.add(key)
  }

  // ── Step 2: activate new danmakus from pool ──
  for (const entry of danmakuPool.value) {
    if (activeKeys.has(entry.key)) continue

    const { item } = entry
    const elapsed = currentTime - item.play_time
    if (elapsed < 0) continue

    const maxDur = (item.position || 0) === 0 ? SCROLL_DURATION : FADE_DURATION
    if (elapsed >= maxDur) continue

    // Find a free track
    let freeTrack = -1
    for (let i = 0; i < TRACK_COUNT; i++) {
      if (!trackOccupied[i]) { freeTrack = i; break }
    }
    if (freeTrack === -1) continue

    // Create DOM element
    const el = document.createElement('div')
    el.className = 'danmaku-item'
    el.dataset.danmakuId = entry.key
    el.dataset.playTime = String(item.play_time)
    el.dataset.position = String(item.position || 0)
    el.dataset.track = String(freeTrack)

    const color = item.color || '#ffffff'
    const posType = item.position || 0

    el.style.color = color
    el.style.position = 'absolute'
    el.style.whiteSpace = 'nowrap'
    el.style.pointerEvents = 'none'
    el.style.textShadow = '1px 1px 2px rgba(0,0,0,0.8)'
    el.style.fontWeight = 'bold'
    el.style.zIndex = '10'
    el.style.willChange = 'transform'
    el.style.fontSize = item.size === 0 ? '22px' : item.size === 2 ? '30px' : '26px'
    el.textContent = item.content

    const top = getTrackTop(freeTrack, ch)
    el.style.top = `${top}px`

    if (posType === 1) {
      // Top fixed
      el.style.left = '50%'
      el.style.transform = 'translateX(-50%)'
      el.style.opacity = '1'
    } else if (posType === 2) {
      // Bottom fixed
      el.style.top = `${ch * 0.7 + freeTrack * 30}px`
      el.style.left = '50%'
      el.style.transform = 'translateX(-50%)'
      el.style.opacity = '1'
    } else {
      // Scrolling — calculate position from elapsed video time
      el.style.left = '0'
      // Append first, then read offsetWidth for correct calculation
      container.appendChild(el)
      const progress = elapsed / SCROLL_DURATION
      const startX = cw + 20
      const x = startX - (startX + el.offsetWidth) * progress
      el.style.transform = `translateX(${x}px)`
    }

    // For non-scrolling types that weren't appended above
    if (!el.parentNode) {
      container.appendChild(el)
    }

    trackOccupied[freeTrack] = true
  }
}

// ---- Public: add danmaku (called from parent for optimistic render) ----
function addDanmaku(item: DanmakuItem) {
  if (!props.enabled) return
  if (isDuplicate(item)) return

  const key = makePoolKey(item)
  danmakuPool.value.push({ key, item })
  // RAF loop will pick it up on next frame (~16ms)
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
      addDanmaku(data)
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
      data.forEach((item) => {
        const key = makePoolKey(item)
        danmakuPool.value.push({ key, item })
      })
    }
  } catch {
    // silent fail
  }
}

// ---- Video element binding ----
function bindVideoEvents(el: HTMLVideoElement) {
  videoEl = el

  // On seek: RAF loop handles positions on next frame automatically
  // No special handling needed — currentTime changes, next RAF frame recalculates.

  // Listen only for seeked to add a safety re-check (some browsers pause on seek)
  onVideoSeeked = () => {
    // ensure any danmaku that should now be visible is picked up
    updateDanmakus()
  }

  el.addEventListener('seeked', onVideoSeeked)
}

function unbindVideoEvents() {
  if (!videoEl) return
  if (onVideoSeeked) videoEl.removeEventListener('seeked', onVideoSeeked)
  videoEl = null
}

// ---- Watch videoEl prop ----
watch(
  () => props.videoEl,
  (el) => {
    unbindVideoEvents()
    if (el) {
      bindVideoEvents(el)
    }
  },
)

// ---- Lifecycle ----
onMounted(() => {
  loadHistory()
  connectWebSocket()
  startRafLoop()

  // If videoEl prop is already set, bind it
  if (props.videoEl) {
    bindVideoEvents(props.videoEl)
  }
})

onUnmounted(() => {
  stopRafLoop()
  unbindVideoEvents()
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
})

defineExpose({ addDanmaku })
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
