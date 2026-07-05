<template>
  <div>
    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchVideos"
    />

    <!-- Video grid -->
    <div v-else-if="videos.length > 0">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        <VideoCard
          v-for="video in videos"
          :key="video.id"
          :video="video"
        />
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="flex justify-center mt-8">
        <button
          class="px-8 py-2.5 bg-[var(--color-surface)] text-[var(--color-text)] text-sm rounded-[var(--radius-full)] border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)] transition-colors disabled:opacity-50"
          :disabled="loadingMore"
          @click="loadMore"
        >
          {{ loadingMore ? '加载中...' : '加载更多' }}
        </button>
      </div>
    </div>

    <!-- Empty -->
    <EmptyState v-else message="还没有视频，快去上传吧" />
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const route = useRoute()

const videos = ref<VideoInfo[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const totalVideos = ref(0)
const pageSize = 12

const hasMore = computed(() => videos.value.length < totalVideos.value)

// Read category_id from URL query (set by sidebar), react to route change
const categoryId = computed(() => {
  const raw = route.query.category_id
  return raw ? Number(raw) : 0
})

async function fetchVideos() {
  loading.value = true
  error.value = ''
  currentPage.value = 1

  try {
    const params: Record<string, string | number> = { page: 1, page_size: pageSize }
    if (categoryId.value > 0) {
      params.category_id = categoryId.value
    }

    const data = await get<PaginatedData<VideoInfo>>('/api/v1/videos/', params)
    videos.value = data.items || []
    totalVideos.value = data.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载视频列表失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const params: Record<string, string | number> = {
      page: currentPage.value,
      page_size: pageSize,
    }
    if (categoryId.value > 0) {
      params.category_id = categoryId.value
    }

    const data = await get<PaginatedData<VideoInfo>>('/api/v1/videos/', params)
    videos.value.push(...(data.items || []))
    totalVideos.value = data.total || 0
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}

// Watch route query for sidebar category click
watch(categoryId, () => {
  fetchVideos()
})

onMounted(() => {
  fetchVideos().then(() => {
    loading.value = false
  })
})
</script>
