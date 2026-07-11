<template>
  <div class="max-w-4xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-6">观看历史</h1>

    <LoadingSpinner v-if="loading" />
    <ErrorMessage v-else-if="error" :message="error" :on-retry="fetchHistory" />

    <!-- Timeline -->
    <div v-else-if="historyItems.length > 0" class="relative">
      <!-- Timeline line -->
      <div class="absolute left-4 top-0 bottom-0 w-px bg-[var(--color-border)]" />

      <div v-for="(group, gIdx) in groupedHistory" :key="group.date" class="relative mb-8">
        <!-- Date header -->
        <div class="flex items-center gap-3 mb-3 relative z-10">
          <div class="w-8 h-8 rounded-full bg-[var(--color-primary)]/15 flex items-center justify-center flex-shrink-0 border-2 border-[var(--color-primary)]/30">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-primary)" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
          <span class="text-sm font-semibold text-[var(--color-text)]">{{ group.dateLabel }}</span>
          <span class="text-xs text-[var(--color-text-secondary)]">{{ group.items.length }} 个视频</span>
        </div>

        <!-- Items -->
        <div class="ml-12 space-y-3">
          <div v-for="item in group.items" :key="item.video.id"
            class="flex items-center gap-4 p-3 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer"
            @click="navigateTo(`/video/${item.video.id}?t=${item.progress}`)">
            <!-- Thumbnail -->
            <NuxtLink :to="`/video/${item.video.id}?t=${item.progress}`" class="w-32 h-18 rounded-lg overflow-hidden flex-shrink-0">
              <img v-if="item.video.cover_url" :src="item.video.cover_url" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full bg-[var(--color-surface-hover)] flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="currentColor" class="text-[var(--color-text-secondary)]/30"><polygon points="8,5 19,12 8,19"/></svg>
              </div>
            </NuxtLink>
            <!-- Info -->
            <div class="flex-1 min-w-0">
              <h3 class="text-sm font-medium text-[var(--color-text)] truncate">{{ item.video.title }}</h3>
              <p class="text-xs text-[var(--color-text-secondary)] mt-0.5">{{ item.video.user?.nickname || item.video.user?.username || '未知' }}</p>
              <!-- Progress bar -->
              <div class="flex items-center gap-2 mt-2">
                <div class="flex-1 h-1 bg-[var(--color-border)] rounded-full overflow-hidden">
                  <div class="h-full bg-[var(--color-primary)] rounded-full transition-all" :style="{ width: getProgressPercent(item) + '%' }" />
                </div>
                <span class="text-xs text-[var(--color-text-secondary)] flex-shrink-0">{{ formatProgress(item) }}</span>
              </div>
            </div>
            <!-- Watched time -->
            <div class="text-xs text-[var(--color-text-secondary)] flex-shrink-0 text-right hidden sm:block">
              {{ formatTimeOfDay(item.watched_at) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="flex justify-center mt-6 ml-12">
        <button
          class="px-8 py-2.5 bg-[var(--color-surface)] text-[var(--color-text)] text-sm font-medium rounded-full border border-[var(--color-border)] hover:bg-[var(--color-surface-hover)] transition-all active:scale-[0.98] disabled:opacity-50"
          :disabled="loadingMore" @click="loadMore">{{ loadingMore ? '加载中...' : '加载更多' }}</button>
      </div>
    </div>

    <EmptyState v-else message="暂无观看记录，快去发现有趣的内容吧" />
  </div>
</template>

<script setup lang="ts">
import type { HistoryItem, HistoryListResp } from '~/types'
import { useApi } from '~/composables/useApi'

const { get } = useApi()
const router = useRouter()

const historyItems = ref<HistoryItem[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const currentPage = ref(1)
const totalItems = ref(0)
const pageSize = 20

const hasMore = computed(() => historyItems.value.length < totalItems.value)

interface GroupedItems { date: string; dateLabel: string; items: HistoryItem[] }
const groupedHistory = computed<GroupedItems[]>(() => {
  const map = new Map<string, HistoryItem[]>()
  for (const item of historyItems.value) {
    const d = new Date(item.watched_at)
    const key = d.toLocaleDateString('zh-CN')
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(item)
  }
  return Array.from(map.entries()).map(([date, items]) => {
    const d = new Date(items[0].watched_at)
    const today = new Date()
    const yesterday = new Date(today); yesterday.setDate(yesterday.getDate() - 1)
    let label: string
    if (date === today.toLocaleDateString('zh-CN')) label = '今天'
    else if (date === yesterday.toLocaleDateString('zh-CN')) label = '昨天'
    else label = date
    return { date, dateLabel: label, items }
  })
})

function getProgressPercent(item: HistoryItem): number {
  if (!item.video.duration || item.video.duration <= 0) return 0
  return Math.min(100, Math.round((item.progress / item.video.duration) * 100))
}
function formatProgress(item: HistoryItem): string {
  const pct = getProgressPercent(item)
  if (pct >= 95) return '已看完'
  return `${pct}%`
}
function formatTimeOfDay(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}
function navigateTo(url: string) { router.push(url) }

async function fetchHistory() {
  loading.value = true; error.value = ''; currentPage.value = 1
  try {
    const data = await get<HistoryListResp>('/api/v1/history/', { page: 1, page_size: pageSize })
    historyItems.value = data.items || []; totalItems.value = data.total || 0
  } catch (err) { error.value = err instanceof Error ? err.message : '加载失败' }
  finally { loading.value = false }
}
async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true; currentPage.value++
  try {
    const data = await get<HistoryListResp>('/api/v1/history/', { page: currentPage.value, page_size: pageSize })
    historyItems.value.push(...(data.items || [])); totalItems.value = data.total || 0
  } catch { currentPage.value-- }
  finally { loadingMore.value = false }
}

onMounted(() => fetchHistory())
</script>
