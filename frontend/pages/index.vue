<template>
  <div>
    <!-- Tabs: 推荐 | 最新 | 关注 -->
    <div class="flex gap-1 p-1 bg-[var(--color-surface)]/60 backdrop-blur-sm rounded-[var(--radius-full)] border border-[var(--color-border)]/30 w-fit mb-6">
      <button
        v-for="tab in displayTabs"
        :key="tab.key"
        @click="activeTab = tab.key"
        :class="[
          'px-5 py-2 text-sm font-medium rounded-[var(--radius-full)] transition-all duration-[var(--transition-normal)]',
          activeTab === tab.key
            ? 'bg-[var(--color-primary)] text-white shadow-sm shadow-[var(--color-primary)]/25'
            : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:bg-[var(--color-surface-hover)]'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Carousel (hot tab only) -->
    <VideoCarousel v-if="activeTab === 'hot' && carouselVideos.length > 0" :videos="carouselVideos" />

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
    <EmptyState v-else message="还没有视频，快去上传吧" />
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import type { VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const route = useRoute()
const userStore = useUserStore()

const videos = ref<VideoInfo[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const totalVideos = ref(0)
const pageSize = 12

const tabs = [
  { key: 'hot', label: '推荐' },
  { key: 'latest', label: '最新' },
]
const activeTab = ref(route.query.category_id ? 'latest' : 'hot')

const displayTabs = computed(() => {
  if (userStore.isLoggedIn) {
    return [...tabs, { key: 'following', label: '关注' }]
  }
  return tabs
})

const hasMore = computed(() => videos.value.length < totalVideos.value)

const categoryId = computed(() => {
  const raw = route.query.category_id
  return raw ? Number(raw) : 0
})

async function fetchVideos() {
  loading.value = true
  error.value = ''
  currentPage.value = 1

  try {
    if (activeTab.value === 'following') {
      const data = await get<PaginatedData<VideoInfo>>('/api/v1/feed', { page: 1, page_size: pageSize })
      videos.value = data.items || []
      totalVideos.value = data.total || 0
    } else if (activeTab.value === 'hot') {
      const params: Record<string, string | number> = { page: 1, page_size: pageSize }
      const data = await get<PaginatedData<VideoInfo>>('/api/v1/videos/hot', params)
      videos.value = data.items || []
      totalVideos.value = data.total || 0
    } else {
      const params: Record<string, string | number> = { page: 1, page_size: pageSize }
      if (categoryId.value > 0) {
        params.category_id = categoryId.value
      }
      const data = await get<PaginatedData<VideoInfo>>('/api/v1/videos/', params)
      videos.value = data.items || []
      totalVideos.value = data.total || 0
    }
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
    if (activeTab.value === 'following') {
      const data = await get<PaginatedData<VideoInfo>>('/api/v1/feed', { page: currentPage.value, page_size: pageSize })
      videos.value.push(...(data.items || []))
      totalVideos.value = data.total || 0
    } else if (activeTab.value === 'hot') {
      const params = { page: currentPage.value, page_size: pageSize }
      const data = await get<PaginatedData<VideoInfo>>('/api/v1/videos/hot', params)
      videos.value.push(...(data.items || []))
      totalVideos.value = data.total || 0
    } else {
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
    }
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}

watch(activeTab, () => {
  fetchVideos()
})

// If user logs out while on "following" tab, reset to "hot"
watch(() => userStore.isLoggedIn, (loggedIn) => {
  if (!loggedIn && activeTab.value === 'following') {
    activeTab.value = 'hot'
  }
})

// When category_id changes, auto-switch to 'latest' tab (hot tab doesn't support category)
watch(categoryId, (newVal) => {
  if (newVal > 0 && activeTab.value !== 'latest') {
    activeTab.value = 'latest'
  } else if (activeTab.value === 'latest') {
    fetchVideos()
  }
})

onMounted(() => {
  fetchVideos()
})
</script>
