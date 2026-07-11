<template>
  <div class="max-w-[1800px] mx-auto">
    <div v-if="loading">
      <LoadingSpinner />
    </div>

    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchVideo"
    />

    <div v-else-if="video" class="flex flex-col lg:flex-row gap-6">
      <!-- Main content -->
      <div class="flex-1 min-w-0">
        <!-- Back button -->
        <button class="mb-3 px-4 py-1.5 text-sm bg-[var(--color-surface)] border border-[var(--color-border)]/50 rounded-xl text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors flex items-center gap-1" @click="router.back()">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
          返回
        </button>
        <!-- Video player with danmaku layer -->
        <div class="relative w-full bg-black rounded-[var(--radius-lg)] overflow-hidden shadow-lg shadow-black/30 ring-1 ring-white/5">
          <video
            ref="videoPlayerRef"
            controls
            class="w-full"
            style="aspect-ratio: 16/9"
            @loadedmetadata="onVideoLoaded"
            @play="startHistoryRecording"
            @pause="stopHistoryRecording"
            @ended="stopHistoryRecording"
            @seeked="onSeeked"
            @timeupdate="onTimeUpdate"
          >
            您的浏览器不支持视频播放
          </video>

          <!-- Danmaku layer overlay -->
          <DanmakuLayer
            v-if="showDanmaku"
            ref="danmakuLayerRef"
            :video-id="video.id"
            :enabled="showDanmaku"
            :video-el="videoPlayerRef"
          />
        </div>

        <!-- Danmaku controls -->
        <div class="mt-2 flex items-center gap-3 flex-wrap">
          <!-- Quality selector -->
          <div v-if="qualities.length > 0" class="relative">
            <button
              class="flex items-center gap-1 px-3 py-1.5 text-xs bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)] rounded-[var(--radius-full)] hover:text-[var(--color-primary)] transition-colors"
              @click="showQualityPicker = !showQualityPicker"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 20V10"/><path d="M18 20V4"/><path d="M6 20v-4"/>
              </svg>
              {{ currentQuality }}
            </button>
            <div
              v-if="showQualityPicker"
              class="absolute bottom-full left-0 mb-1 w-28 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg shadow-lg py-1 z-20"
            >
              <button
                v-for="q in qualities"
                :key="q.quality"
                class="w-full text-left px-3 py-1.5 text-xs hover:bg-[var(--color-surface-hover)] transition-colors"
                :class="currentQuality === q.quality ? 'text-[var(--color-primary)] font-medium' : 'text-[var(--color-text)]'"
                @click="switchQuality(q)"
              >
                {{ q.quality }}
              </button>
            </div>
          </div>

          <!-- Transcode status -->
          <div v-if="transcodeStatus && transcodeStatus.status !== 2" class="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-full)]" :class="transcodeStatusClass">
            <svg v-if="transcodeStatus.status === 0 || transcodeStatus.status === 1" class="animate-spin" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
            <span>{{ transcodeStatusText }}</span>
          </div>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-full)] transition-colors"
            :class="showDanmaku ? 'bg-[var(--color-primary)]/20 text-[var(--color-primary)]' : 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)]'"
            @click="showDanmaku = !showDanmaku"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/>
            </svg>
            弹幕
          </button>
          <div v-if="showDanmaku && userStore.isLoggedIn" class="flex items-center gap-2 flex-1">
            <input
              v-model="danmakuContent"
              type="text"
              maxlength="200"
              placeholder="发送弹幕..."
              class="flex-1 h-7 px-3 text-xs bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-[var(--radius-full)] focus:outline-none focus:border-[var(--color-primary)]/40"
              @keyup.enter="sendDanmaku"
            />
            <input
              v-model="danmakuColor"
              type="color"
              class="w-6 h-6 rounded cursor-pointer border-0 p-0"
              title="弹幕颜色"
            />
            <button
              class="px-3 py-1 text-xs font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-40"
              :disabled="!danmakuContent.trim()"
              @click="sendDanmaku"
            >
              发送
            </button>
          </div>
        </div>

        <!-- Resume prompt -->
        <div
          v-if="showResumePrompt"
          class="mt-4 p-4 bg-[var(--color-primary)]/10 border border-[var(--color-primary)]/20 rounded-[var(--radius-md)] flex items-center justify-between gap-4"
        >
          <div class="flex items-center gap-2.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-primary)] flex-shrink-0">
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </svg>
            <span class="text-sm text-[var(--color-text)]">上次观看至 <span class="font-medium text-[var(--color-primary)]">{{ formatProgress(savedProgress) }}</span>，是否继续？</span>
          </div>
          <div class="flex gap-2 flex-shrink-0">
            <button
              class="px-4 py-2 text-sm font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors active:scale-[0.98]"
              @click="resumePlayback"
            >
              继续播放
            </button>
            <button
              class="px-4 py-2 text-sm bg-[var(--color-surface-hover)] text-[var(--color-text)] rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)]/80 transition-colors active:scale-[0.98]"
              @click="showResumePrompt = false"
            >
              从头开始
            </button>
          </div>
        </div>

        <!-- Video info -->
        <div class="mt-5">
          <h1 class="text-xl font-bold text-[var(--color-text)] leading-snug font-display">{{ video.title }}</h1>

          <!-- UP主 + interaction buttons row -->
          <div class="flex flex-wrap items-center justify-between gap-3 mt-4">
            <NuxtLink :to="`/user/${video.user?.id}`" class="flex items-center gap-3 group/up hover:opacity-100 transition-opacity">
              <div class="w-10 h-10 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-sm font-bold overflow-hidden ring-2 ring-transparent group-hover/up:ring-[var(--color-primary)]/30 transition-shadow">
                <img v-if="video.user?.avatar" :src="video.user.avatar" class="w-full h-full object-cover" />
                <span v-else>{{ (video.user?.nickname || video.user?.username || 'U').charAt(0) }}</span>
              </div>
              <div>
                <p class="text-sm font-medium text-[var(--color-text)] group-hover/up:text-[var(--color-primary)] transition-colors">{{ video.user?.nickname || video.user?.username }}</p>
                <p class="text-xs text-[var(--color-text-secondary)]">{{ formatViews(video.views) }} 播放 <span class="text-[var(--color-text-secondary)]/40">-</span> {{ formatTime(video.created_at) }}</p>
              </div>
            </NuxtLink>

            <!-- Interaction buttons -->
            <div class="flex items-center gap-1 relative">
              <!-- Like button -->
              <button
                class="flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-full)] transition-all active:scale-[0.95]"
                :class="interactions.liked ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)]' : 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]'"
                @click="toggleLike"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" :fill="interactions.liked ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>
                </svg>
                <span class="text-sm font-medium">{{ likeCount }}</span>
              </button>

              <!-- Coin button -->
              <button
                class="flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-full)] transition-all active:scale-[0.95]"
                :class="interactions.coins > 0 ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)]' : 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]'"
                :disabled="coinDisabled"
                @click="!coinDisabled && (showCoinPicker = !showCoinPicker)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/><text x="12" y="16" text-anchor="middle" font-size="12" fill="currentColor" stroke="none">B</text>
                </svg>
                <span class="text-sm font-medium">{{ coinDisabled ? '已投币' : '投币' }}</span>
              </button>
              <!-- Favorite button: short press = quick favorite, long press = show picker -->
              <button
                class="flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-full)] transition-all active:scale-[0.95]"
                :class="isFavorited ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)]' : 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]'"
                @mousedown="favPressStart"
                @mouseup="favPressEnd"
                @mouseleave="favPressCancel"
                @touchstart.prevent="favPressStart"
                @touchend.prevent="favPressEnd"
                @touchcancel="favPressCancel"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" :fill="isFavorited ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
                </svg>
                <span class="text-sm font-medium">收藏</span>
              </button>
            </div>
          </div>

          <!-- Description -->
          <div class="mt-5 p-4 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-md)]">
            <div class="flex items-center gap-2 mb-2">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text-secondary)]">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
              <span class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">简介</span>
            </div>
            <p
              class="text-sm text-[var(--color-text)] leading-relaxed whitespace-pre-wrap"
              :class="{ 'line-clamp-3': !showFullDesc && descriptionLines > 3 }"
            >{{ video.description || '暂无简介' }}</p>
            <button
              v-if="descriptionLines > 3"
              class="text-sm text-[var(--color-primary)] mt-2 hover:underline font-medium"
              @click="showFullDesc = !showFullDesc"
            >
              {{ showFullDesc ? '收起' : '展开更多' }}
            </button>
          </div>

          <!-- Tags -->
          <div v-if="tags.length > 0" class="mt-3 flex flex-wrap gap-2">
            <NuxtLink
              v-for="tag in tags"
              :key="tag.id"
              :to="`/search?q=${encodeURIComponent(tag.name)}`"
              class="px-3 py-1 text-xs font-medium bg-[var(--color-primary)]/15 text-[var(--color-primary)] rounded-[var(--radius-full)] hover:bg-[var(--color-primary)]/25 transition-colors"
            >
              {{ tag.name }}
            </NuxtLink>
          </div>

          <!-- Comments section -->
          <div class="mt-8">
            <CommentList
              :video-id="video.id"
              :video-author-id="video.user?.id"
            />
          </div>
        </div>
      </div>

      <!-- Sidebar: related videos -->
      <div class="w-full lg:w-80 flex-shrink-0">
        <h3 class="text-base font-semibold text-[var(--color-text)] mb-4">推荐视频</h3>
        <div v-if="relatedVideos.length > 0" class="space-y-4">
          <VideoCard
            v-for="v in relatedVideos"
            :key="v.id"
            :video="v"
          />
        </div>
        <EmptyState v-else message="暂无推荐" icon="&#9654;" />
      </div>
    </div>
  </div>

    <!-- Coin picker modal -->
    <AppModal :visible="showCoinPicker" title="投币" @close="showCoinPicker = false">
      <div class="flex gap-3 justify-center">
        <button class="px-5 py-2 text-sm font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors active:scale-[0.98]" @click="addCoin(1); showCoinPicker = false">1 枚</button>
        <button class="px-5 py-2 text-sm font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors active:scale-[0.98]" @click="addCoin(2); showCoinPicker = false">2 枚</button>
      </div>
    </AppModal>

    <!-- Favorite picker modal -->
    <AppModal :visible="showFavoritePicker" title="选择收藏夹" @close="showFavoritePicker = false">
      <div class="max-h-64 overflow-y-auto -mx-2 px-2">
        <div v-if="favoriteLoading && favorites.length === 0" class="text-xs text-[var(--color-text-secondary)] py-2">加载中...</div>

        <label
          v-for="fav in favorites"
          :key="fav.id"
          class="flex items-center gap-2.5 px-2 py-1.5 hover:bg-[var(--color-surface-hover)] rounded cursor-pointer transition-colors"
        >
          <input
            type="checkbox"
            :checked="favoritedFolderIds.includes(fav.id)"
            class="w-4 h-4 rounded accent-[var(--color-primary)] cursor-pointer flex-shrink-0"
            @change="toggleFavorite(fav.id)"
          />
          <span class="text-sm text-[var(--color-text)] flex-1 truncate">{{ fav.name }}</span>
          <span class="text-xs text-[var(--color-text-secondary)] flex-shrink-0">{{ fav.item_count }}</span>
        </label>

        <div v-if="!favoriteLoading && favorites.length === 0 && !showNewFavorite" class="text-xs text-[var(--color-text-secondary)] py-2">暂无收藏夹</div>

        <!-- Create new favorite -->
        <div class="border-t border-[var(--color-border)] mt-2 pt-2">
          <div v-if="showNewFavorite" class="flex gap-1">
            <input
              v-model="newFavoriteName"
              type="text"
              maxlength="50"
              placeholder="收藏夹名称"
              class="flex-1 h-7 px-2 text-xs bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded focus:outline-none focus:border-[var(--color-primary)]/40"
              @keyup.enter="createFavorite"
            />
            <button
              class="px-2 py-0.5 text-xs font-medium bg-[var(--color-primary)] text-white rounded hover:bg-[var(--color-primary-hover)] disabled:opacity-40 transition-colors active:scale-[0.98]"
              :disabled="!newFavoriteName.trim()"
              @click="createFavorite"
            >
              确定
            </button>
          </div>
          <button
            v-else
            class="w-full text-left px-2 py-1.5 text-xs text-[var(--color-primary)] hover:bg-[var(--color-surface-hover)] rounded transition-colors"
            @click="showNewFavorite = true"
          >
            + 创建新收藏夹
          </button>
        </div>

        <!-- Done button -->
        <button
          class="mt-2 w-full py-1.5 text-xs font-medium bg-[var(--color-surface-hover)] text-[var(--color-text)] rounded hover:bg-[var(--color-border)] transition-colors active:scale-[0.98]"
          @click="showFavoritePicker = false"
        >
          完成
        </button>
      </div>
    </AppModal>

</template>

<script setup lang="ts">
import type { VideoInfo, PaginatedData, HistoryItem, HistoryListResp, Tag, LikeResp, CoinResp, FavoriteInfo, FavoriteToggleResp, InteractionStatus, VideoQuality as VideoQualityType, TranscodeStatus } from '~/types'
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'
import VideoCard from '~/components/video/VideoCard.vue'
import AppModal from '~/components/common/AppModal.vue'
import DanmakuLayer from '~/components/danmaku/DanmakuLayer.vue'
import CommentList from '~/components/comment/CommentList.vue'

const { get, post } = useApi()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { showToast } = useToast()

const video = ref<VideoInfo | null>(null)
const tags = ref<Tag[]>([])
const relatedVideos = ref<VideoInfo[]>([])
const playUrl = ref('')
const fallbackMp4Url = ref('')  // pre-signed MP4 URL, target for ALL HLS fallback paths
const loading = ref(true)
const error = ref('')
const showFullDesc = ref(false)
const showResumePrompt = ref(false)
const savedProgress = ref(0)
const videoPlayerRef = ref<HTMLVideoElement | null>(null)
const danmakuLayerRef = ref<InstanceType<typeof DanmakuLayer> | null>(null)
let historyInterval: ReturnType<typeof setInterval> | null = null
let seekTimer: ReturnType<typeof setTimeout> | null = null

// Quality / Transcode state
const qualities = ref<VideoQualityType[]>([])
const currentQuality = ref('原始')
const showQualityPicker = ref(false)
const transcodeStatus = ref<TranscodeStatus | null>(null)
let transcodePollTimer: ReturnType<typeof setInterval> | null = null

// Danmaku state
const showDanmaku = ref(true)
const danmakuContent = ref('')
const danmakuColor = ref('#ffffff')
const currentTime = ref(0)

// Interaction state
const interactions = ref<InteractionStatus>({ liked: false, coins: 0, favorited: false })
const favoritedFolderIds = ref<number[]>([])
const coinDisabled = ref(false)
const likeCount = ref(0)
const showCoinPicker = ref(false)
const showFavoritePicker = ref(false)
const favorites = ref<FavoriteInfo[]>([])
const favoriteLoading = ref(false)
const showNewFavorite = ref(false)
const newFavoriteName = ref('')

// Favorite button state
const isFavorited = computed(() => interactions.value.favorited || favoritedFolderIds.value.length > 0)
let favPressTimer: ReturnType<typeof setTimeout> | null = null
let isLongPress = false
let crossReferenced = false

// HLS state
let hlsModule: typeof import('hls.js') | null = null
let hlsLoadPromise: Promise<void> | null = null
const hlsInstance = ref<any>(null)
const isQualitySwitching = ref(false)
let pendingSeekTime = 0
let pendingAutoPlay = false

const transcodeStatusClass = computed(() => {
  if (!transcodeStatus.value) return ''
  if (transcodeStatus.value.status === 3) return 'bg-red-500/15 text-red-400'
  return 'bg-yellow-500/15 text-yellow-400'
})

const transcodeStatusText = computed(() => {
  if (!transcodeStatus.value) return ''
  switch (transcodeStatus.value.status) {
    case 0: return '转码等待中...'
    case 1: return `转码中 ${transcodeStatus.value.progress}%`
    case 3: return '转码失败'
    default: return ''
  }
})

const descriptionLines = computed(() => {
  if (!video.value?.description) return 0
  return video.value.description.split('\n').length
})

function onTimeUpdate() {
  if (videoPlayerRef.value) {
    currentTime.value = videoPlayerRef.value.currentTime
  }
}

function formatProgress(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

function onVideoLoaded() {
  const videoEl = videoPlayerRef.value
  if (!videoEl) return

  if (savedProgress.value > 0 && showResumePrompt.value) {
    // Wait for user action via resume button
  } else if (savedProgress.value > 0) {
    videoEl.currentTime = savedProgress.value
  }
}

function resumePlayback() {
  showResumePrompt.value = false
  const videoEl = videoPlayerRef.value
  if (videoEl && savedProgress.value > 0) {
    videoEl.currentTime = savedProgress.value
  }
}

async function recordProgress() {
  const videoEl = videoPlayerRef.value
  if (!videoEl || !video.value) return

  const progress = Math.floor(videoEl.currentTime)
  if (progress <= 0) return

  const body: Record<string, number> = {
    video_id: video.value.id,
    progress,
  }
  if (videoEl.duration && isFinite(videoEl.duration)) {
    body.duration = Math.floor(videoEl.duration)
  }

  try {
    await post('/api/v1/history/', body)
  } catch {
    // Non-critical
  }
}

function onSeeked() {
  if (seekTimer) clearTimeout(seekTimer)
  seekTimer = setTimeout(() => recordProgress(), 300)
}

function startHistoryRecording() {
  if (historyInterval) return
  historyInterval = setInterval(recordProgress, 5000)
}

async function stopHistoryRecording() {
  if (historyInterval) {
    clearInterval(historyInterval)
    historyInterval = null
  }
  await recordProgress()
}

// Danmaku
async function sendDanmaku() {
  if (!danmakuContent.value.trim() || !userStore.isLoggedIn) return

  const content = danmakuContent.value
  const color = danmakuColor.value
  const playTime = Math.floor(currentTime.value)

  // Optimistic render: show danmaku immediately without waiting for API
  danmakuLayerRef.value?.addDanmaku({
    id: 0, // temp ID — real ID will come via WebSocket, dedup by content+user
    content,
    color,
    position: 0,
    size: 1,
    play_time: playTime,
    user: {
      id: userStore.userInfo?.id || 0,
      username: userStore.userInfo?.username || '',
      nickname: userStore.userInfo?.nickname || '',
      avatar: userStore.userInfo?.avatar || '',
    },
  })

  danmakuContent.value = ''

  try {
    await post(`/api/v1/videos/${video.value?.id}/danmaku`, {
      content,
      color,
      play_time: playTime,
    })
  } catch {
    showToast('弹幕发送失败', 'error')
  }
}

// Interactions
async function fetchInteractions() {
  if (!video.value) return
  try {
    const data = await get<InteractionStatus>(`/api/v1/videos/${video.value.id}/interactions`)
    interactions.value = data
    if (data.coins > 0) coinDisabled.value = true
    // Reset folder tracking until cross-referenced
    favoritedFolderIds.value = []
    crossReferenced = false
  } catch {
    // Not logged in or error - use defaults
  }
}

async function toggleLike() {
  if (!userStore.isLoggedIn) {
    showToast('请先登录', 'error')
    return
  }
  // Optimistic update
  const wasLiked = interactions.value.liked
  interactions.value.liked = !wasLiked
  likeCount.value += wasLiked ? -1 : 1

  try {
    const data = await post<LikeResp>(`/api/v1/videos/${video.value?.id}/like`)
    // Sync with actual server state
    interactions.value.liked = data.liked
    likeCount.value = data.count
  } catch {
    // Revert on error
    interactions.value.liked = wasLiked
    likeCount.value += wasLiked ? 1 : -1
  }
}

async function addCoin(count: number) {
  if (!userStore.isLoggedIn) {
    showToast('请先登录', 'error')
    return
  }
  try {
    await post<CoinResp>(`/api/v1/videos/${video.value?.id}/coin`, { count })
    coinDisabled.value = true
    interactions.value.coins = count
    showToast('投币成功', 'success')
  } catch (err: any) {
    const msg = err?.message || '投币失败'
    showToast(msg, 'error')
  }
}

async function loadFavorites() {
  if (!userStore.isLoggedIn) return
  favoriteLoading.value = true
  try {
    favorites.value = await get<FavoriteInfo[]>('/api/v1/favorites/')
  } catch {
    favorites.value = []
  } finally {
    favoriteLoading.value = false
  }
}

async function toggleFavorite(favoriteId: number) {
  const idx = favoritedFolderIds.value.indexOf(favoriteId)
  const wasInFolder = idx >= 0

  // Optimistic update
  if (wasInFolder) {
    favoritedFolderIds.value.splice(idx, 1)
  } else {
    favoritedFolderIds.value.push(favoriteId)
  }
  interactions.value.favorited = favoritedFolderIds.value.length > 0

  try {
    await post<FavoriteToggleResp>(`/api/v1/favorites/${favoriteId}/items`, { video_id: video.value?.id })
  } catch {
    // Revert on error
    if (wasInFolder) {
      favoritedFolderIds.value.push(favoriteId)
    } else {
      const revertIdx = favoritedFolderIds.value.indexOf(favoriteId)
      if (revertIdx >= 0) favoritedFolderIds.value.splice(revertIdx, 1)
    }
    interactions.value.favorited = favoritedFolderIds.value.length > 0
  }
}

async function createFavorite() {
  const name = newFavoriteName.value.trim()
  if (!name) return

  try {
    const fav = await post<FavoriteInfo>('/api/v1/favorites/', { name, is_public: 1 })
    favorites.value.push(fav)
    newFavoriteName.value = ''
    showNewFavorite.value = false
    // Auto-add current video to the new folder
    if (video.value) {
      await post<FavoriteToggleResp>(`/api/v1/favorites/${fav.id}/items`, { video_id: video.value.id })
      favoritedFolderIds.value.push(fav.id)
      interactions.value.favorited = true
    }
    showToast('收藏夹创建成功', 'success')
  } catch {
    // error handled by useApi
  }
}

// Long-press / short-press logic for favorite button
function favPressStart() {
  isLongPress = false
  favPressTimer = setTimeout(() => {
    isLongPress = true
    showFavoritePicker.value = true
    if (favorites.value.length === 0) loadFavorites()
    if (!crossReferenced) crossReferenceFavorites()
  }, 500)
}

function favPressEnd() {
  if (favPressTimer) { clearTimeout(favPressTimer); favPressTimer = null }
  if (!isLongPress) {
    if (showFavoritePicker.value) {
      showFavoritePicker.value = false
    } else {
      quickFavorite()
    }
  }
}

function favPressCancel() {
  if (favPressTimer) { clearTimeout(favPressTimer); favPressTimer = null }
  isLongPress = false
}

async function quickFavorite() {
  if (!userStore.isLoggedIn) {
    showToast('请先登录', 'error')
    return
  }
  // Ensure favorites are loaded
  if (favorites.value.length === 0) {
    await loadFavorites()
  }
  const defaultFav = favorites.value.find(f => f.name === '默认收藏夹')
  if (!defaultFav) {
    showToast('默认收藏夹不存在', 'error')
    return
  }
  await toggleFavorite(defaultFav.id)
  if (!showFavoritePicker.value) {
    const isNowFav = favoritedFolderIds.value.includes(defaultFav.id)
    showToast(isNowFav ? '已收藏至默认收藏夹' : '已取消收藏', 'success')
  }
}

async function crossReferenceFavorites() {
  if (favorites.value.length === 0 || !video.value) return
  const videoId = video.value.id

  const results = await Promise.allSettled(
    favorites.value.map(fav =>
      get<{ favorite: FavoriteInfo; items: VideoInfo[]; total: number; page: number; page_size: number }>(`/api/v1/favorites/${fav.id}?page=1&page_size=50`)
    )
  )

  const ids: number[] = []
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      const detail = result.value
      if (detail.items?.some((item: VideoInfo) => item.id === videoId)) {
        ids.push(favorites.value[index].id)
      }
    }
  })

  favoritedFolderIds.value = ids
  interactions.value.favorited = ids.length > 0
  crossReferenced = true
}

// HLS engine
async function ensureHlsReady(): Promise<boolean> {
  if (hlsModule) return true
  if (!hlsLoadPromise) {
    hlsLoadPromise = import('hls.js').then(m => { hlsModule = m }).catch(() => { hlsLoadPromise = null })
  }
  try { await hlsLoadPromise; return hlsModule != null } catch { return false }
}

function destroyHLS() {
  if (hlsInstance.value) {
    try { hlsInstance.value.destroy() } catch { /* ignore */ }
    hlsInstance.value = null
  }
}

async function startHls(url: string) {
  const videoEl = videoPlayerRef.value
  if (!videoEl || !url) return

  await ensureHlsReady()
  if (!hlsModule) { videoEl.src = fallbackMp4Url.value; videoEl.load(); return }

  const Hls = hlsModule.default || (hlsModule as any).Hls
  if (!Hls || !Hls.isSupported()) { videoEl.src = fallbackMp4Url.value; videoEl.load(); return }

  destroyHLS()
  const hls = new Hls()
  hls.on(Hls.Events.ERROR, (_evt: any, data: any) => {
    if (data.fatal) { destroyHLS(); videoEl.src = fallbackMp4Url.value; videoEl.load() }
  })
  hls.on(Hls.Events.MANIFEST_PARSED, () => {
    const el = videoPlayerRef.value
    if (!el) return
    if (pendingSeekTime > 0) { el.currentTime = pendingSeekTime; pendingSeekTime = 0 }
    if (pendingAutoPlay) { pendingAutoPlay = false; el.play().catch(() => {}) }
    if (!isQualitySwitching.value) { el.currentTime = 0 }
    isQualitySwitching.value = false
  })
  hls.loadSource(url)
  hls.attachMedia(videoEl)
  hlsInstance.value = hls

  // Safety net: if no data after 4s, fall back to mp4
  const hlsRef = hls
  setTimeout(() => {
    const el = videoPlayerRef.value
    if (el && !el.currentSrc && hlsInstance.value === hlsRef) {
      destroyHLS()
      el.src = fallbackMp4Url.value
      el.load()
    }
  }, 4000)
}

async function initPlayer() {
  const url = playUrl.value
  if (!url) return
  await nextTick()
  const videoEl = videoPlayerRef.value
  if (!videoEl) return

  if (!url.includes('.m3u8')) {
    videoEl.src = url
    videoEl.load()
    return
  }
  await startHls(url)
}

async function fetchVideo() {
  loading.value = true
  error.value = ''

  try {
    const id = route.params.id as string
    video.value = await get<VideoInfo>(`/api/v1/videos/${id}`)

    // Fetch qualities
    try {
      qualities.value = await get<VideoQualityType[]>(`/api/v1/videos/${id}/qualities`)
    } catch {
      qualities.value = []
    }

    // Fetch transcode status
    fetchTranscodeStatus()
    startTranscodePolling()

    const playData = await get<{ play_url: string }>(`/api/v1/videos/${id}/play-url`)
    fallbackMp4Url.value = playData.play_url

    // Set initial playUrl: prefer best quality m3u8, fall back to mp4
    if (qualities.value.length > 0) {
      const best = qualities.value[qualities.value.length - 1] // highest quality last
      playUrl.value = best.play_url
      currentQuality.value = best.quality
    } else {
      playUrl.value = playData.play_url
    }

    const urlTime = parseInt(route.query.t as string)
    if (!isNaN(urlTime) && urlTime > 0) {
      savedProgress.value = urlTime
    }

    if (savedProgress.value === 0) {
      try {
        const history = await get<HistoryListResp>('/api/v1/history/', { page: 1, page_size: 50 })
        if (history?.items) {
          const currentHistory = history.items.find((h: HistoryItem) => h.video.id === video.value?.id)
          if (currentHistory && currentHistory.progress > 0 && video.value?.duration && currentHistory.progress < video.value.duration) {
            savedProgress.value = currentHistory.progress
            showResumePrompt.value = true
          }
        }
      } catch {
        // Not logged in or no history
      }
    }

    try {
      tags.value = await get<Tag[]>(`/api/v1/videos/${id}/tags`)
    } catch {
      tags.value = []
    }

    // Fetch interactions
    fetchInteractions()
    loadFavorites()

    if (video.value.category_id > 0) {
      const related = await get<PaginatedData<VideoInfo>>('/api/v1/videos/', {
        category_id: video.value.category_id,
        page_size: 10,
      })
      relatedVideos.value = (related.items || []).filter((v) => v.id !== video.value?.id).slice(0, 8)
    }
    // Ensure DOM is ready before setting video src
    loading.value = false
    await nextTick()
    // Start playback
    await initPlayer()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载视频失败'
  } finally {
    loading.value = false
  }
}

function formatViews(views: number): string {
  if (views >= 10000) return `${(views / 10000).toFixed(1)}万`
  return String(views)
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

// Quality switching
async function switchQuality(q: VideoQualityType) {
  const videoEl = videoPlayerRef.value
  if (!videoEl) return
  pendingSeekTime = videoEl.currentTime
  pendingAutoPlay = !videoEl.paused
  isQualitySwitching.value = true
  currentQuality.value = q.quality
  showQualityPicker.value = false
  playUrl.value = q.play_url  // keep watch compatibility

  if (q.play_url.includes('.m3u8')) {
    await startHls(q.play_url)
  } else {
    videoEl.src = q.play_url
    videoEl.load()
  }
}

// Transcode status
async function fetchTranscodeStatus() {
  try {
    const id = route.params.id as string
    transcodeStatus.value = await get<TranscodeStatus>(`/api/v1/videos/${id}/transcode-status`)
  } catch {
    transcodeStatus.value = null
  }
}

function startTranscodePolling() {
  stopTranscodePolling()
  transcodePollTimer = setInterval(() => {
    if (transcodeStatus.value && transcodeStatus.value.status === 2) {
      stopTranscodePolling()
      return
    }
    fetchTranscodeStatus()
  }, 10000)
}

function stopTranscodePolling() {
  if (transcodePollTimer) {
    clearInterval(transcodePollTimer)
    transcodePollTimer = null
  }
}

onMounted(() => {
  ensureHlsReady()  // preload hls.js in parallel with fetchVideo
  fetchVideo()
})
onUnmounted(() => {
  stopHistoryRecording()
  stopTranscodePolling()
  destroyHLS()
  if (seekTimer) clearTimeout(seekTimer)
})
</script>
