<template>
  <div class="max-w-4xl mx-auto">
    <!-- Loading -->
    <LoadingSpinner v-if="loadingUser" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchUser"
    />

    <!-- User profile -->
    <div v-else-if="user">
      <div class="bg-[var(--color-surface)] rounded-[var(--radius-lg)] p-6 border border-[var(--color-border)]">
        <div class="flex flex-col sm:flex-row items-center gap-4">
          <!-- Avatar -->
          <div class="w-20 h-20 rounded-full bg-[var(--color-primary)] flex items-center justify-center text-white text-2xl font-bold flex-shrink-0 overflow-hidden">
            <img v-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" />
            <span v-else>{{ (user.nickname || user.username).charAt(0) }}</span>
          </div>

          <!-- Info -->
          <div class="flex-1 text-center sm:text-left">
            <h1 class="text-xl font-bold text-[var(--color-text)]">{{ user.nickname || user.username }}</h1>
            <p class="text-sm text-[var(--color-text-secondary)] mt-1">@{{ user.username }}</p>
            <p v-if="user.bio" class="text-sm text-[var(--color-text)] mt-2">{{ user.bio }}</p>
            <p class="text-xs text-[var(--color-text-secondary)] mt-2">注册于 {{ formatTime(user.created_at) }}</p>
          </div>

          <!-- Edit button (only for own profile) -->
          <button
            v-if="isOwner && !editing"
            class="px-4 py-2 text-sm bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
            @click="startEditing"
          >
            编辑资料
          </button>
        </div>

        <!-- Edit form -->
        <div v-if="editing" class="mt-6 border-t border-[var(--color-border)] pt-6 space-y-4">
          <div>
            <label class="block text-sm text-[var(--color-text-secondary)] mb-1">昵称</label>
            <input
              v-model="editForm.nickname"
              type="text"
              maxlength="50"
              class="w-full h-10 px-3 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)]"
            />
          </div>
          <div>
            <label class="block text-sm text-[var(--color-text-secondary)] mb-1">签名</label>
            <textarea
              v-model="editForm.bio"
              maxlength="500"
              rows="3"
              class="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] resize-none"
            ></textarea>
          </div>
          <div class="flex gap-2">
            <button
              class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
              @click="saveProfile"
            >
              保存
            </button>
            <button
              class="px-6 py-2 bg-[var(--color-surface-hover)] text-[var(--color-text)] text-sm rounded-[var(--radius-full)] hover:bg-[var(--color-border)] transition-colors"
              @click="editing = false"
            >
              取消
            </button>
          </div>
        </div>
      </div>

      <!-- User videos -->
      <div class="mt-8">
        <h2 class="text-lg font-semibold text-[var(--color-text)] mb-4">发布的视频</h2>

        <LoadingSpinner v-if="loadingVideos" />

        <div v-else-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <VideoCard
            v-for="v in videos"
            :key="v.id"
            :video="v"
          />
        </div>

        <EmptyState v-else message="还没有发布视频" icon="📹" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'
import type { UserInfo, VideoInfo, PaginatedData } from '~/types'
import VideoCard from '~/components/video/VideoCard.vue'

const { get, put } = useApi()
const route = useRoute()
const userStore = useUserStore()
const { showToast } = useToast()

const user = ref<UserInfo | null>(null)
const videos = ref<VideoInfo[]>([])
const loadingUser = ref(true)
const loadingVideos = ref(true)
const error = ref('')
const editing = ref(false)

const editForm = ref({
  nickname: '',
  bio: '',
})

const userId = computed(() => Number(route.params.id))
const isOwner = computed(() => userStore.userInfo?.id === userId.value)

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

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchUser()
  fetchVideos()
})
</script>
