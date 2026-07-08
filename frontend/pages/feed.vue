<template>
  <div>
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-xl font-bold text-[var(--color-text)]">关注动态</h1>
      <p class="mt-1 text-sm text-[var(--color-text-secondary)]">你关注的 UP 主的最新视频</p>
    </div>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchFeed"
    />

    <!-- Video grid -->
    <div v-else-if="videos.length > 0">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
        <VideoCard
          v-for="video in videos"
          :key="video.id"
          :video="video"
        />
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="flex justify-center mt-10">
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
    <EmptyState
      v-else
      message="关注一些 UP 主，这里会显示他们的最新视频"
      icon="👥"
    />
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import type { VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const userStore = useUserStore()
const router = useRouter()

const videos = ref<VideoInfo[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const totalVideos = ref(0)
const pageSize = 12

const hasMore = computed(() => videos.value.length < totalVideos.value)

async function fetchFeed() {
  loading.value = true
  error.value = ''
  currentPage.value = 1

  try {
    const data = await get<PaginatedData<VideoInfo>>('/api/v1/feed', { page: 1, page_size: pageSize })
    videos.value = data.items || []
    totalVideos.value = data.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载关注动态失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const data = await get<PaginatedData<VideoInfo>>('/api/v1/feed', { page: currentPage.value, page_size: pageSize })
    videos.value.push(...(data.items || []))
    totalVideos.value = data.total || 0
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}

// Redirect to login if not authenticated (feed requires auth)
onMounted(() => {
  if (!userStore.isLoggedIn) {
    router.replace('/login')
    return
  }
  fetchFeed()
})
</script>
