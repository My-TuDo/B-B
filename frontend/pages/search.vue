<template>
  <div class="max-w-6xl mx-auto">

    <!-- Search header -->
    <h1 v-if="query" class="text-lg font-semibold text-[var(--color-text)] mb-6">搜索结果：{{ query }}</h1>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Results -->
    <template v-else-if="searched">
      <div v-if="results.length > 0">
        <p class="text-sm text-[var(--color-text-secondary)] mb-4">找到 {{ totalResults }} 个结果</p>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
          <VideoCard
            v-for="video in results"
            :key="video.id"
            :video="video"
          />
        </div>

        <!-- Load more -->
        <div v-if="hasMore" class="flex justify-center mt-8">
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
      <EmptyState v-else message="未找到相关视频" icon="🔍" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const route = useRoute()

const query = ref('')
const results = ref<VideoInfo[]>([])
const totalResults = ref(0)
const loading = ref(false)
const loadingMore = ref(false)
const searched = ref(false)
const currentPage = ref(1)
const pageSize = 12

const hasMore = computed(() => results.value.length < totalResults.value)

// Sync query from URL and trigger search
watch(
  () => route.query.q,
  (newQ) => {
    const q = (newQ as string) || ''
    if (q !== query.value) {
      query.value = q
      if (q) {
        doSearch()
      } else {
        results.value = []
        totalResults.value = 0
        searched.value = false
      }
    }
  },
  { immediate: true },
)

async function doSearch() {
  const q = query.value.trim()
  if (!q) return

  loading.value = true
  searched.value = true
  currentPage.value = 1

  try {
    const data = await get<PaginatedData<VideoInfo>>('/api/v1/search/', { q, page: 1, page_size: pageSize })
    results.value = data.items || []
    totalResults.value = data.total || 0
  } catch {
    results.value = []
    totalResults.value = 0
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const data = await get<PaginatedData<VideoInfo>>('/api/v1/search/', {
      q: query.value.trim(),
      page: currentPage.value,
      page_size: pageSize,
    })
    results.value.push(...(data.items || []))
    totalResults.value = data.total || 0
  } catch {
    currentPage.value--
  } finally {
    loadingMore.value = false
  }
}
</script>
