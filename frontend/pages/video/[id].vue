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
        <!-- Video player -->
        <div class="w-full bg-black rounded-[var(--radius-lg)] overflow-hidden">
          <video
            v-if="playUrl"
            :key="playUrl"
            controls
            class="w-full"
            style="aspect-ratio: 16/9"
          >
            <source :src="playUrl" type="video/mp4" />
            您的浏览器不支持视频播放
          </video>
          <div v-else class="flex items-center justify-center" style="aspect-ratio: 16/9">
            <LoadingSpinner />
          </div>
        </div>

        <!-- Video info -->
        <div class="mt-4">
          <h1 class="text-xl font-bold text-[var(--color-text)]">{{ video.title }}</h1>

          <div class="flex items-center gap-3 mt-3">
            <NuxtLink :to="`/user/${video.user?.id}`" class="flex items-center gap-2 hover:opacity-80 transition-opacity">
              <div class="w-10 h-10 rounded-full bg-[var(--color-primary)] flex items-center justify-center text-white text-sm font-bold overflow-hidden">
                <img v-if="video.user?.avatar" :src="video.user.avatar" class="w-full h-full object-cover" />
                <span v-else>{{ (video.user?.nickname || video.user?.username || 'U').charAt(0) }}</span>
              </div>
              <div>
                <p class="text-sm font-medium text-[var(--color-text)]">{{ video.user?.nickname || video.user?.username }}</p>
                <p class="text-xs text-[var(--color-text-secondary)]">{{ formatViews(video.views) }} 播放 · {{ formatTime(video.created_at) }}</p>
              </div>
            </NuxtLink>
          </div>

          <!-- Description -->
          <div class="mt-4 p-4 bg-[var(--color-surface)] rounded-[var(--radius-md)]">
            <p
              class="text-sm text-[var(--color-text)] whitespace-pre-wrap"
              :class="{ 'line-clamp-3': !showFullDesc && descriptionLines > 3 }"
            >{{ video.description || '暂无简介' }}</p>
            <button
              v-if="descriptionLines > 3"
              class="text-sm text-[var(--color-primary)] mt-1 hover:underline"
              @click="showFullDesc = !showFullDesc"
            >
              {{ showFullDesc ? '收起' : '展开' }}
            </button>
          </div>

          <!-- Comments placeholder -->
          <div class="mt-6 p-8 bg-[var(--color-surface)] rounded-[var(--radius-md)] text-center">
            <p class="text-sm text-[var(--color-text-secondary)]">弹幕和评论功能将在后续版本上线</p>
          </div>
        </div>
      </div>

      <!-- Sidebar: related videos -->
      <div class="w-full lg:w-80 flex-shrink-0">
        <h3 class="text-base font-semibold text-[var(--color-text)] mb-3">推荐视频</h3>
        <div v-if="relatedVideos.length > 0" class="space-y-3">
          <VideoCard
            v-for="v in relatedVideos"
            :key="v.id"
            :video="v"
          />
        </div>
        <EmptyState v-else message="暂无推荐" icon="📺" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import type { VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const route = useRoute()

const video = ref<VideoInfo | null>(null)
const relatedVideos = ref<VideoInfo[]>([])
const playUrl = ref('')
const loading = ref(true)
const error = ref('')
const showFullDesc = ref(false)

const descriptionLines = computed(() => {
  if (!video.value?.description) return 0
  return video.value.description.split('\n').length
})

async function fetchVideo() {
  loading.value = true
  error.value = ''

  try {
    const id = route.params.id as string
    video.value = await get<VideoInfo>(`/api/v1/videos/${id}`)

    // Fetch play URL
    const playData = await get<{ play_url: string }>(`/api/v1/videos/${id}/play-url`)
    playUrl.value = playData.play_url

    // Fetch related videos
    if (video.value.category_id > 0) {
      const related = await get<PaginatedData<VideoInfo>>('/api/v1/videos/', {
        category_id: video.value.category_id,
        page_size: 10,
      })
      relatedVideos.value = (related.items || []).filter((v) => v.id !== video.value?.id).slice(0, 8)
    }
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

onMounted(fetchVideo)
</script>
