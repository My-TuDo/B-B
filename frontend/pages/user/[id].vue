<template>
  <div class="max-w-4xl mx-auto">
    <!-- Loading -->
    <LoadingSpinner v-if="loadingUser" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchAll"
    />

    <!-- User profile -->
    <div v-else-if="user">
      <div class="bg-[var(--color-surface)] rounded-[var(--radius-lg)] p-6 border border-[var(--color-border)]">
        <div class="flex flex-col sm:flex-row items-center gap-4">
          <!-- Avatar -->
          <div
            class="w-20 h-20 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-2xl font-bold flex-shrink-0 overflow-hidden relative group"
            :class="editing ? 'cursor-pointer' : ''"
            :title="editing ? '点击更换头像' : ''"
            @click="editing ? triggerAvatarUpload() : undefined"
          >
            <div v-if="avatarUploading" class="flex items-center justify-center w-full h-full bg-black/40">
              <svg class="animate-spin w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
            </div>
            <img v-else-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" />
            <span v-else>{{ (user.nickname || user.username).charAt(0) }}</span>
            <!-- Hover overlay for edit mode -->
            <div v-if="editing && !avatarUploading" class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-full">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M23 19a2 2 0 01-2 2H3a2 2 0 01-2-2V8a2 2 0 012-2h4l2-3h6l2 3h4a2 2 0 012 2z"/>
                <circle cx="12" cy="13" r="4"/>
              </svg>
            </div>
          </div>
          <!-- Hidden file input -->
          <input
            ref="fileInput"
            type="file"
            accept="image/jpeg,image/png,image/webp"
            class="hidden"
            @change="handleAvatarFile"
          />

          <!-- Info -->
          <div class="flex-1 text-center sm:text-left">
            <h1 class="text-xl font-bold text-[var(--color-text)]">{{ user.nickname || user.username }}</h1>
            <p class="text-sm text-[var(--color-text-secondary)] mt-1">@{{ user.username }}</p>
            <p v-if="user.bio" class="text-sm text-[var(--color-text)] mt-2">{{ user.bio }}</p>
            <p class="text-xs text-[var(--color-text-secondary)] mt-2">注册于 {{ formatTime(user.created_at) }}</p>
          </div>

          <!-- Buttons -->
          <div class="flex items-center gap-2">
            <!-- Follow button (not shown for own profile) -->
            <button
              v-if="!isOwner && userStore.isLoggedIn"
              class="px-5 py-2 text-sm font-medium rounded-[var(--radius-full)] transition-all active:scale-[0.98]"
              :class="isFollowing ? 'bg-[var(--color-surface-hover)] text-[var(--color-text)] hover:bg-[var(--color-border)]' : 'bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] shadow-sm shadow-[var(--color-primary)]/25'"
              @click="toggleFollow"
            >
              {{ isFollowing ? '已关注' : '关注' }}
            </button>

            <!-- Edit button (own profile only) -->
            <button
              v-if="isOwner && !editing"
              class="px-5 py-2 text-sm font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-all active:scale-[0.98] shadow-sm shadow-[var(--color-primary)]/25"
              @click="startEditing"
            >
              编辑资料
            </button>
          </div>
        </div>

        <!-- Stats -->
        <div class="flex justify-center sm:justify-start gap-6 mt-4 pt-4 border-t border-[var(--color-border)]">
          <div class="text-center">
            <p class="text-lg font-bold text-[var(--color-text)]">{{ stats.videos }}</p>
            <p class="text-xs text-[var(--color-text-secondary)]">视频</p>
          </div>
          <div class="text-center">
            <p class="text-lg font-bold text-[var(--color-text)]">{{ stats.followers }}</p>
            <p class="text-xs text-[var(--color-text-secondary)]">粉丝</p>
          </div>
          <div class="text-center">
            <p class="text-lg font-bold text-[var(--color-text)]">{{ stats.following }}</p>
            <p class="text-xs text-[var(--color-text-secondary)]">关注</p>
          </div>
        </div>

        <!-- Edit form -->
        <div v-if="editing" class="mt-6 border-t border-[var(--color-border)] pt-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">昵称</label>
            <input
              v-model="editForm.nickname"
              type="text"
              maxlength="50"
              class="w-full h-10 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">签名</label>
            <textarea
              v-model="editForm.bio"
              maxlength="500"
              rows="3"
              class="w-full px-3.5 py-2.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all resize-none"
            ></textarea>
          </div>
          <div class="flex gap-2">
            <button
              class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-all active:scale-[0.98] shadow-sm shadow-[var(--color-primary)]/25"
              @click="saveProfile"
            >
              保存
            </button>
            <button
              class="px-6 py-2 bg-[var(--color-surface-hover)] text-[var(--color-text)] text-sm font-medium rounded-[var(--radius-full)] hover:bg-[var(--color-border)] transition-all active:scale-[0.98]"
              @click="editing = false"
            >
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="mt-8">
        <div class="flex gap-6 border-b border-[var(--color-border)] mb-4">
          <button
            class="pb-2 text-sm font-medium transition-colors border-b-2"
            :class="activeTab === 'videos' ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
            @click="activeTab = 'videos'"
          >
            投稿
          </button>
          <button
            class="pb-2 text-sm font-medium transition-colors border-b-2"
            :class="activeTab === 'favorites' ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
            @click="activeTab = 'favorites'"
          >
            收藏
          </button>
        </div>

        <!-- Videos tab -->
        <div v-if="activeTab === 'videos'">
          <LoadingSpinner v-if="loadingVideos" />

          <div v-else-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            <VideoCard
              v-for="v in videos"
              :key="v.id"
              :video="v"
            />
          </div>

          <EmptyState v-else message="还没有发布视频" icon="📹" />
        </div>

        <!-- Favorites tab -->
        <div v-if="activeTab === 'favorites'">
          <LoadingSpinner v-if="loadingFavorites" />

          <div v-else-if="publicFavorites.length > 0" class="space-y-3">
            <div
              v-for="fav in publicFavorites"
              :key="fav.id"
              class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-md)] hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer"
              @click="router.push(`/favorites/${fav.id}`)"
            >
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-[var(--color-text)]">{{ fav.name }}</p>
                  <p class="text-xs text-[var(--color-text-secondary)] mt-1">{{ fav.item_count }} 个视频</p>
                </div>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text-secondary)]">
                  <polyline points="9 18 15 12 9 6"/>
                </svg>
              </div>
            </div>
          </div>

          <EmptyState v-else message="暂无公开收藏夹" icon="⭐" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { UserInfo, VideoInfo, PaginatedData, ProfileResp, FollowResp, FavoriteInfo } from '~/types'
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'
import VideoCard from '~/components/video/VideoCard.vue'

const { get, put, post } = useApi()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { showToast } = useToast()

const user = ref<UserInfo | null>(null)
const videos = ref<VideoInfo[]>([])
const loadingUser = ref(true)
const loadingVideos = ref(true)
const loadingFavorites = ref(false)
const error = ref('')
const editing = ref(false)
const activeTab = ref<'videos' | 'favorites'>('videos')
const isFollowing = ref(false)
const publicFavorites = ref<FavoriteInfo[]>([])
const avatarUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const stats = ref({
  videos: 0,
  followers: 0,
  following: 0,
})

const editForm = ref({
  nickname: '',
  bio: '',
})

const userId = computed(() => Number(route.params.id))
const isOwner = computed(() => userStore.userInfo?.id === userId.value)

async function fetchAll() {
  await fetchUser()
  await Promise.all([fetchVideos(), fetchProfile()])
}

async function fetchUser() {
  loadingUser.value = true
  error.value = ''

  try {
    user.value = await get<UserInfo>(`/api/v1/users/${userId.value}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载用户信息失败'
  } finally {
    loadingUser.value = false
  }
}

async function fetchProfile() {
  try {
    const data = await get<ProfileResp>(`/api/v1/users/${userId.value}/profile`)
    stats.value = data.stats
    isFollowing.value = data.stats.following > 0
  } catch {
    // Use defaults
  }
}

async function fetchVideos() {
  loadingVideos.value = true

  try {
    const data = await get<PaginatedData<VideoInfo>>(`/api/v1/videos/users/${userId.value}/videos`, {
      status: 1,
    })
    videos.value = data.items || []
  } catch {
    // Non-critical
  } finally {
    loadingVideos.value = false
  }
}

async function fetchFavorites() {
  if (activeTab.value !== 'favorites') return
  loadingFavorites.value = true
  try {
    publicFavorites.value = await get<FavoriteInfo[]>(`/api/v1/users/${userId.value}/favorites`)
  } catch {
    publicFavorites.value = []
  } finally {
    loadingFavorites.value = false
  }
}

async function toggleFollow() {
  if (!userStore.isLoggedIn) {
    showToast('请先登录', 'error')
    return
  }
  try {
    const data = await post<FollowResp>(`/api/v1/users/${userId.value}/follow`)
    isFollowing.value = data.following
    stats.value.followers += data.following ? 1 : -1
  } catch (err) {
    const msg = err instanceof Error ? err.message : '操作失败'
    showToast(msg, 'error')
  }
}

function startEditing() {
  if (!user.value) return
  editForm.value.nickname = user.value.nickname || ''
  editForm.value.bio = user.value.bio || ''
  editing.value = true
}

async function saveProfile() {
  try {
    const updated = await put<UserInfo>(`/api/v1/users/${userId.value}`, {
      nickname: editForm.value.nickname || undefined,
      bio: editForm.value.bio || undefined,
    })
    user.value = updated
    editing.value = false
    showToast('资料已更新', 'success')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '更新失败'
    showToast(msg, 'error')
  }
}

function triggerAvatarUpload() {
  fileInput.value?.click()
}

async function handleAvatarFile(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Validate size on client side
  if (file.size > 2 * 1024 * 1024) {
    showToast('文件大小不能超过 2MB', 'error')
    target.value = ''
    return
  }

  // Validate type on client side
  const allowedTypes = ['image/jpeg', 'image/png', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    showToast('仅支持 JPEG、PNG、WebP 格式', 'error')
    target.value = ''
    return
  }

  avatarUploading.value = true

  try {
    const formData = new FormData()
    formData.append('avatar', file)

    const updated = await post<UserInfo>(`/api/v1/users/${userId.value}/avatar`, formData)
    user.value = updated
    // Sync to global store so Header avatar updates immediately
    if (userStore.userInfo && updated.avatar) {
      userStore.userInfo.avatar = updated.avatar
    }
    showToast('头像已更新', 'success')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '头像上传失败'
    showToast(msg, 'error')
  } finally {
    avatarUploading.value = false
    target.value = ''
  }
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

// Watch tab changes
watch(activeTab, (tab) => {
  if (tab === 'favorites') fetchFavorites()
})

onMounted(() => {
  const tab = route.query.tab
  if (tab === 'favorites') {
    activeTab.value = 'favorites'
  } else if (tab === 'posts') {
    activeTab.value = 'videos'
  }
  fetchAll()
})
</script>
