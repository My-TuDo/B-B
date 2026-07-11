<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-6">排行榜</h1>

    <!-- Period tabs -->
    <div class="flex gap-1 p-1 bg-[var(--color-surface)]/60 backdrop-blur-sm rounded-full border border-[var(--color-border)]/30 w-fit mb-6">
      <button v-for="p in periods" :key="p.key" @click="activePeriod = p.key"
        :class="['px-5 py-2 text-sm font-medium rounded-full transition-all duration-200', activePeriod === p.key ? 'bg-[var(--color-primary)] text-white shadow-sm shadow-[var(--color-primary)]/25' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]']">
        {{ p.label }}
      </button>
    </div>

    <LoadingSpinner v-if="loading" />
    <ErrorMessage v-else-if="error" :message="error" :on-retry="fetchRanking" />
    <div v-else-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
      <div v-for="(video, index) in videos" :key="video.id" class="relative group">
        <div class="absolute top-2 left-2 z-10 w-7 h-7 rounded-lg flex items-center justify-center text-xs font-bold"
          :class="index < 3 ? rankTopBg(index) : 'bg-black/50 text-white/70'">{{ index + 1 }}</div>
        <VideoCard :video="video" />
      </div>
    </div>
    <EmptyState v-else message="暂无排行榜数据" />

    <div v-if="hasMore" class="flex justify-center mt-10">
      <button class="px-8 py-2.5 bg-[var(--color-surface)] text-[var(--color-text)] text-sm font-medium rounded-full border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)] transition-all active:scale-[0.98] disabled:opacity-50" :disabled="loadingMore" @click="loadMore">{{ loadingMore ? '加载中...' : '加载更多' }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { RankingData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const periods = [{ key: 'day', label: '日榜' },{ key: 'week', label: '周榜' },{ key: 'total', label: '总榜' }]
const activePeriod = ref('day')
const videos = ref<RankingData['items']>([])
const totalVideos = ref(0)
const loading = ref(true); const loadingMore = ref(false); const error = ref(''); const currentPage = ref(1); const pageSize = 12
const hasMore = computed(() => videos.value.length < totalVideos.value)

function rankTopBg(i: number) {
  if (i === 0) return 'bg-yellow-500 text-white'
  if (i === 1) return 'bg-gray-300 text-gray-800'
  return 'bg-amber-700 text-white'
}
async function fetchRanking() {
  loading.value = true; error.value = ''; currentPage.value = 1
  try { const data = await get<RankingData>('/api/v1/videos/ranking', { period: activePeriod.value, page: 1, page_size: pageSize }); videos.value = data.items || []; totalVideos.value = data.total || 0 }
  catch (err) { error.value = err instanceof Error ? err.message : '加载排行榜失败' }
  finally { loading.value = false }
}
async function loadMore() {
  if (loadingMore.value || !hasMore.value) return; loadingMore.value = true; currentPage.value++
  try { const data = await get<RankingData>('/api/v1/videos/ranking', { period: activePeriod.value, page: currentPage.value, page_size: pageSize }); videos.value.push(...(data.items || [])); totalVideos.value = data.total || 0 }
  catch { currentPage.value-- }
  finally { loadingMore.value = false }
}
watch(activePeriod, fetchRanking)
onMounted(fetchRanking)
</script>
