<template>
  <div class="max-w-4xl mx-auto">
    <!-- Navigation bar -->
    <div class="mb-4">
      <!-- Back button -->
      <button
        class="inline-flex items-center gap-1.5 text-sm text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-3"
        @click="goBack"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
        返回个人中心
      </button>

      <!-- Favorite tabs -->
      <div v-if="allFavorites.length > 1" class="flex flex-wrap gap-2">
        <button
          v-for="fav in allFavorites"
          :key="fav.id"
          class="px-4 py-2 text-sm rounded-full whitespace-nowrap transition-colors border"
          :class="fav.id === favoriteId
            ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
            : 'bg-[var(--color-surface)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-surface-hover)]'"
          @click="router.push('/favorites/' + fav.id)"
        >
          {{ fav.name }}
          <span class="ml-1 text-xs opacity-70">({{ fav.item_count }})</span>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchDetail"
    />

    <!-- Content -->
    <div v-else-if="favorite">
      <!-- Header -->
      <div class="bg-[var(--color-surface)] rounded-[var(--radius-lg)] p-6 border border-[var(--color-border)] mb-6">
        <div class="flex items-center gap-3">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" class="text-[var(--color-primary)]">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
          </svg>
          <div>
            <h1 class="text-xl font-bold text-[var(--color-text)]">{{ favorite.name }}</h1>
            <p class="text-sm text-[var(--color-text-secondary)] mt-1">
              {{ favorite.is_public ? '公开' : '私有' }}收藏夹
              <span v-if="total > 0"> &middot; {{ total }} 个视频</span>
            </p>
          </div>
        </div>
      </div>

      <!-- Videos grid -->
      <LoadingSpinner v-if="loadingVideos" />

      <div v-else-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
        <VideoCard
          v-for="v in videos"
          :key="v.id"
          :video="v"
        />
      </div>

      <EmptyState v-else message="收藏夹是空的" icon="&#11088;" />

      <!-- Pagination -->
      <div v-if="total > pageSize" class="flex justify-center mt-6 gap-2">
        <button
          v-for="p in Math.ceil(total / pageSize)"
          :key="p"
          class="w-8 h-8 text-sm rounded-full transition-colors"
          :class="p === page ? 'bg-[var(--color-primary)] text-white' : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)]'"
          @click="page = p; fetchVideos()"
        >
          {{ p }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FavoriteInfo, VideoInfo, FavoriteDetailResp } from '~/types'
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import VideoCard from '~/components/video/VideoCard.vue'

const { get } = useApi()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const favorite = ref<FavoriteInfo | null>(null)
const videos = ref<VideoInfo[]>([])
const allFavorites = ref<FavoriteInfo[]>([])
const loading = ref(true)
const loadingVideos = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)

const favoriteId = computed(() => Number(route.params.id))

function goBack() {
  if (userStore.userInfo?.id) {
    router.push('/user/' + userStore.userInfo.id + '?tab=favorites')
  } else {
    router.back()
  }
}

async function fetchFavorites() {
  try {
    const data = await get<FavoriteInfo[]>('/api/v1/favorites/')
    allFavorites.value = data || []
  } catch {
    allFavorites.value = []
  }
}

async function fetchVideos() {
  loadingVideos.value = true
  try {
    const data = await get<FavoriteDetailResp>(`/api/v1/favorites/${favoriteId.value}`, {
      page: page.value,
      page_size: pageSize.value,
    })
    videos.value = data.items || []
    total.value = data.total || 0
  } catch {
    videos.value = []
  } finally {
    loadingVideos.value = false
  }
}

async function fetchDetail() {
  loading.value = true
  error.value = ''

  try {
    const data = await get<FavoriteDetailResp>(`/api/v1/favorites/${favoriteId.value}`, {
      page: page.value,
      page_size: pageSize.value,
    })
    favorite.value = data.favorite
    videos.value = data.items || []
    total.value = data.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载收藏夹失败'
  } finally {
    loading.value = false
  }
}

// Re-fetch when navigating between different favorites
watch(favoriteId, (newId, oldId) => {
  if (newId !== oldId) {
    page.value = 1
    fetchDetail()
  }
})

onMounted(() => {
  fetchDetail()
  fetchFavorites()
})
</script>
