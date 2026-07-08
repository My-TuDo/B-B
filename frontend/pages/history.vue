<template>
  <div class="max-w-4xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-6">观看历史</h1>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchHistory"
    />

    <!-- History list -->
    <div v-else-if="historyItems.length > 0" class="space-y-3">
      <div
        v-for="item in historyItems"
        :key="item.video.id"
        class="flex items-center gap-4 p-3 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-lg)] hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer"
        @click="navigateTo(`/video/${item.video.id}?t=${item.progress}`)"
      >
        <!-- Video info -->
        <div class="flex-1 min-w-0">
          <h3 class="text-sm font-medium text-[var(--color-text)] truncate">{{ item.video.title }}</h3>
          <p class="text-xs text-[var(--color-text-secondary)] mt-0.5">
            {{ item.video.user?.nickname || item.video.user?.username || '未知' }}
          </p>
        </div>

        <!-- Progress -->
        <div class="flex items-center gap-3 flex-shrink-0">
          <div class="flex items-center gap-2">
            <div class="flex gap-[1px]">
              <div v-for="i in totalSegments" :key="i"
                class="w-2 h-4 rounded-[2px]"
                :class="i <= getFilledSegments(item) ? 'bg-[var(--color-primary)]' : 'bg-[var(--color-border)]'"
              ></div>
            </div>
            <span class="text-xs text-[var(--color-text-secondary)] w-14 text-right">{{ formatProgressText(item) }}</span>
          </div>
          <div class="text-xs text-[var(--color-text-secondary)]">
            {{ formatTime(item.watched_at) }}
          </div>
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
    <EmptyState v-else message="暂无观看记录" icon="🕐" />
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { HistoryItem, HistoryListResp } from '~/types'

definePageMeta({
  middleware: 'auth',
})

const { get } = useApi()

const route = useRoute()

const historyItems = ref<HistoryItem[]>([])
const totalItems = ref(0)
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const pageSize = 12

const hasMore = computed(() => historyItems.value.length < totalItems.value)

const totalSegments = 20

function getFilledSegments(item: HistoryItem): number {
  const { progress, video } = item
  if (progress <= 0) return 0

  if (video.duration > 0) {
    return Math.min(totalSegments, Math.round(progress / video.duration * totalSegments))
  }

  // Fallback: logarithmic mapping when duration is unknown
  if (progress < 10) return 1
  if (progress < 30) return 2
  if (progress < 60) return 4
  if (progress < 120) return 6
  if (progress < 300) return 10
  if (progress < 600) return 14
  if (progress < 1800) return 17
  return totalSegments
}

function formatProgressText(item: HistoryItem): string {
  const { progress, video } = item
  if (video.duration > 0) {
    return Math.round(progress / video.duration * 100) + '%'
  }
  return formatDuration(progress)
}

function formatDuration(sec: number): string {
  if (sec < 60) return `${Math.floor(sec)}秒`
  if (sec < 3600) {
    const m = Math.floor(sec / 60)
    const s = Math.floor(sec % 60)
    return `${m}分${s}秒`
  }
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  return `${h}时${m}分`
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return date.toLocaleDateString('zh-CN')
}

async function fetchHistory() {
  loading.value = true
  error.value = ''
  currentPage.value = 1

  try {
    const data = await get<HistoryListResp>('/api/v1/history/', { page: 1, page_size: pageSize })
    historyItems.value = data.items || []
    totalItems.value = data.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载观看历史失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const data = await get<HistoryListResp>('/api/v1/history/', {
      page: currentPage.value,
      page_size: pageSize,
    })
    historyItems.value.push(...(data.items || []))
    totalItems.value = data.total || 0
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}

onMounted(fetchHistory)

// Re-fetch when navigating back to /history (e.g. from video page via sidebar).
// Nuxt may reuse the page component without triggering onMounted a second time.
watch(
  () => route.fullPath,
  (newPath, oldPath) => {
    // Only re-fetch when actually landing on /history, not on sub-navigation
    if (newPath.startsWith('/history') && oldPath !== newPath) {
      fetchHistory()
    }
  },
)
</script>
