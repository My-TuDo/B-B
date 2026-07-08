<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-6">排行榜</h1>

    <!-- Period tabs -->
    <div class="flex gap-2 mb-6">
      <button
        v-for="p in periods"
        :key="p.key"
        @click="activePeriod = p.key"
        :class="[
          'px-5 py-2 text-sm font-medium rounded-[var(--radius-full)] transition-all duration-[var(--transition-normal)]',
          activePeriod === p.key
            ? 'bg-[var(--color-primary)] text-white shadow-sm shadow-[var(--color-primary)]/25'
            : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[var(--color-surface-hover)]'
        ]"
      >
        {{ p.label }}
      </button>
    </div>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchRanking"
    />

    <!-- Ranking list -->
    <div v-else-if="videos.length > 0" class="space-y-3">
      <div
        v-for="(video, index) in videos"
        :key="video.id"
        class="flex items-center gap-4 p-3 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-lg)] hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer"
        @click="navigateTo(`/video/${video.id}`)"
      >
        <!-- Rank number -->
        <div
          class="w-10 h-10 flex items-center justify-center text-lg font-bold rounded-[var(--radius-md)] flex-shrink-0"
          :class="rankStyle(index + 1)"
        >
          {{ index + 1 }}
        </div>

        <!-- Video info -->
        <div class="flex-1 min-w-0">
          <h3 class="text-sm font-medium text-[var(--color-text)] truncate">{{ video.title }}</h3>
          <p class="text-xs text-[var(--color-text-secondary)] mt-0.5">
            {{ video.user?.nickname || video.user?.username || '未知' }}
          </p>
        </div>

        <!-- Views -->
        <div class="text-xs text-[var(--color-text-secondary)] flex-shrink-0">
          {{ formatViews(video.views) }} 播放
        </div>
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="flex justify-center mt-6">
        <button
          class="px-8 py-2.5 bg-[var(--color-surface)] text-[var(--color-text)] text-sm font-medium rounded-[var(--radius-full)] border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)] hover:border-[var(--color-text-secondary)]/30 transition-all duration-[var(--transition-normal)] disabled:opacity-50 active:scale-[0.98]"
          :disabled="loadingMore"
          @click="loadMore"
        >
          {{ loadingMore ? '加载中...' : '加载更多' }}
        </button>
      </div>
    </div>

    <!-- Empty -->
    <EmptyState v-else message="暂无排行榜数据" />
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { RankingData } from '~/types'

const { get } = useApi()

const periods = [
  { key: 'day', label: '日榜' },
  { key: 'week', label: '周榜' },
  { key: 'total', label: '总榜' },
]
const activePeriod = ref('day')
const videos = ref<RankingData['items']>([])
const totalVideos = ref(0)
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const pageSize = 12

const hasMore = computed(() => videos.value.length < totalVideos.value)

function rankStyle(rank: number) {
  if (rank === 1) return 'bg-yellow-500/20 text-yellow-500'
  if (rank === 2) return 'bg-gray-400/20 text-gray-400'
  if (rank === 3) return 'bg-amber-700/20 text-amber-700'
  return 'text-[var(--color-text-secondary)]'
}

function formatViews(views: number): string {
  if (views >= 10000) return `${(views / 10000).toFixed(1)}万`
  return String(views)
}

async function fetchRanking() {
  loading.value = true
  error.value = ''
  currentPage.value = 1

  try {
    const data = await get<RankingData>('/api/v1/videos/ranking', {
      period: activePeriod.value,
      page: 1,
      page_size: pageSize,
    })
    videos.value = data.items || []
    totalVideos.value = data.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载排行榜失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const data = await get<RankingData>('/api/v1/videos/ranking', {
      period: activePeriod.value,
      page: currentPage.value,
      page_size: pageSize,
    })
    videos.value.push(...(data.items || []))
    totalVideos.value = data.total || 0
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}

watch(activePeriod, fetchRanking)
onMounted(fetchRanking)
</script>
