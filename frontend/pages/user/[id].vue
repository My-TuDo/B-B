<template>
  <div>
    <LoadingSpinner v-if="loadingUser" />
    <ErrorMessage v-else-if="error" :message="error" :on-retry="fetchAll" />
    <div v-else-if="user">
      <!-- Banner -->
      <div class="relative rounded-2xl overflow-hidden mb-6 bg-gradient-to-br from-[var(--color-primary)]/10 via-[var(--color-surface)] to-[var(--color-primary)]/5 border border-[var(--color-border)]/30">
        <div class="px-6 pt-8 pb-6">
          <div class="flex flex-col sm:flex-row items-center sm:items-end gap-5">
            <!-- Avatar with hover upload -->
            <div
              class="w-22 h-22 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-2xl font-bold flex-shrink-0 overflow-hidden ring-4 ring-[var(--color-bg)] relative group"
              :class="isOwner ? 'cursor-pointer' : ''"
              @click="isOwner ? triggerAvatarUpload() : undefined"
            >
              <div v-if="avatarUploading" class="flex items-center justify-center w-full h-full bg-black/40">
                <svg class="animate-spin w-6 h-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
                </svg>
              </div>
              <img v-else-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" />
              <span v-else>{{ (user.nickname || user.username).charAt(0) }}</span>
              <!-- Hover overlay: always visible for owner -->
              <div v-if="isOwner && !avatarUploading" class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-full">
                <span class="text-white text-xs font-medium leading-tight text-center">更换<br>头像</span>
              </div>
            </div>
            <input ref="fileInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="handleAvatarFile" />

            <!-- Info -->
            <div class="flex-1 text-center sm:text-left">
              <h1 class="text-2xl font-bold text-[var(--color-text)]">{{ user.nickname || user.username }}</h1>
              <p class="text-sm text-[var(--color-text-secondary)]">@{{ user.username }} · 注册于 {{ formatTime(user.created_at) }}</p>
            </div>

            <!-- Buttons -->
            <div class="flex items-center gap-2 flex-shrink-0">
              <button v-if="!isOwner && userStore.isLoggedIn"
                class="px-5 py-2 text-sm font-medium rounded-full transition-all active:scale-[0.98]"
                :class="isFollowing ? 'bg-[var(--color-surface-hover)] text-[var(--color-text)]' : 'bg-[var(--color-primary)] text-white shadow-sm shadow-[var(--color-primary)]/25'"
                @click="toggleFollow"
              >{{ isFollowing ? '已关注' : '关注' }}</button>
              <button v-if="isOwner && !editing"
                class="px-5 py-2 text-sm font-medium rounded-full border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-all active:scale-[0.98]"
                @click="startEditing"
              >⚙️ 编辑资料</button>
            </div>
          </div>
        </div>

        <!-- Stats row -->
        <div class="flex justify-center gap-8 px-6 pb-4 border-t border-[var(--color-border)]/30 pt-3">
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.videos }}</p><p class="text-xs text-[var(--color-text-secondary)]">视频</p></div>
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.followers }}</p><p class="text-xs text-[var(--color-text-secondary)]">粉丝</p></div>
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.following }}</p><p class="text-xs text-[var(--color-text-secondary)]">关注</p></div>
        </div>
      </div>

      <!-- Edit form modal -->
      <AppModal :visible="editing" title="编辑资料" @close="editing = false">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">昵称</label>
            <input v-model="editForm.nickname" type="text" maxlength="50"
              class="w-full h-10 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all" />
          </div>
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">签名</label>
            <textarea v-model="editForm.bio" maxlength="500" rows="3"
              class="w-full px-3.5 py-2.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all resize-none" />
          </div>
          <div class="flex gap-2">
            <button class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm font-medium rounded-full hover:bg-[var(--color-primary-hover)] transition-all active:scale-[0.98]" @click="saveProfile">保存</button>
            <button class="px-6 py-2 bg-[var(--color-surface-hover)] text-[var(--color-text)] text-sm font-medium rounded-full hover:bg-[var(--color-border)] transition-all active:scale-[0.98]" @click="editing = false">取消</button>
          </div>
        </div>
      </AppModal>

      <!-- Tab bar -->
      <div class="flex gap-6 border-b border-[var(--color-border)] mb-6">
        <button v-for="tab in tabs" :key="tab.key"
          class="pb-3 text-sm font-medium transition-colors border-b-2"
          :class="activeTab === tab.key ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
          @click="activeTab = tab.key">
          {{ tab.label }}
        </button>
      </div>

      <!-- 主页 Tab: 双栏布局 -->
      <div v-if="activeTab === 'home'" class="flex flex-col lg:flex-row gap-6">
        <!-- Main content 75% -->
        <div class="flex-1 min-w-0">
          <div v-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <VideoCard v-for="v in videos" :key="v.id" :video="v" />
          </div>
          <EmptyState v-else message="还没有发布视频" />
        </div>
        <!-- Sidebar 25% -->
        <div class="w-full lg:w-72 flex-shrink-0 space-y-4">
          <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
            <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-2">个人简介</p>
            <p class="text-sm text-[var(--color-text)] leading-relaxed">{{ user.bio || '这个人很懒，什么都没写' }}</p>
          </div>
          <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
            <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-2">社交</p>
            <p class="text-xs text-[var(--color-text-secondary)]">粉丝 {{ stats.followers }} · 关注 {{ stats.following }}</p>
          </div>
        </div>
      </div>

      <!-- 投稿 Tab -->
      <div v-if="activeTab === 'videos'">
        <LoadingSpinner v-if="loadingVideos" />
        <div v-else-if="videos.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          <VideoCard v-for="v in videos" :key="v.id" :video="v" />
        </div>
        <EmptyState v-else message="还没有发布视频" />
      </div>

      <!-- 收藏 Tab -->
      <div v-if="activeTab === 'favorites'">
        <LoadingSpinner v-if="loadingFavorites" />
        <div v-else-if="publicFavorites.length > 0" class="space-y-3">
          <div v-for="fav in publicFavorites" :key="fav.id"
            class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer"
            @click="router.push(`/favorites/${fav.id}`)">
            <div class="flex items-center justify-between">
              <div><p class="text-sm font-medium text-[var(--color-text)]">{{ fav.name }}</p><p class="text-xs text-[var(--color-text-secondary)] mt-1">{{ fav.item_count }} 个视频</p></div>
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text-secondary)]"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
          </div>
        </div>
        <EmptyState v-else message="暂无公开收藏夹" />
      </div>

      <!-- 创作中心 Tab -->
      <div v-if="activeTab === 'creator'">
        <CreatorPanel :user-id="userId" />
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
import AppModal from '~/components/common/AppModal.vue'

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
const activeTab = ref<string>('home')
const isFollowing = ref(false)
const publicFavorites = ref<FavoriteInfo[]>([])
const avatarUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const stats = ref({ videos: 0, followers: 0, following: 0 })
const editForm = ref({ nickname: '', bio: '' })
const userId = computed(() => Number(route.params.id))
const isOwner = computed(() => userStore.userInfo?.id === userId.value)

const tabs = computed(() => {
  const list: { key: string; label: string }[] = [
    { key: 'home', label: '🏠 主页' },
    { key: 'videos', label: `🎬 投稿 (${videos.value.length})` },
    { key: 'favorites', label: '⭐ 收藏' },
  ]
  if (isOwner.value) list.push({ key: 'creator', label: '📊 创作中心' })
  return list
})

async function fetchAll() { await fetchUser(); await Promise.all([fetchVideos(), fetchProfile()]); }
async function fetchUser() {
  loadingUser.value = true; error.value = ''
  try { user.value = await get<UserInfo>(`/api/v1/users/${userId.value}`) }
  catch (err) { error.value = err instanceof Error ? err.message : '加载用户信息失败' }
  finally { loadingUser.value = false }
}
async function fetchProfile() {
  try { const data = await get<ProfileResp>(`/api/v1/users/${userId.value}/profile`); stats.value = data.stats; isFollowing.value = data.stats.following > 0 }
  catch { /* defaults */ }
}
async function fetchVideos() {
  loadingVideos.value = true
  try { const data = await get<PaginatedData<VideoInfo>>(`/api/v1/videos/users/${userId.value}/videos`, { status: 1 }); videos.value = data.items || [] }
  catch { /* non-critical */ }
  finally { loadingVideos.value = false }
}
async function fetchFavorites() {
  loadingFavorites.value = true
  try { publicFavorites.value = await get<FavoriteInfo[]>(`/api/v1/users/${userId.value}/favorites`) }
  catch { publicFavorites.value = [] }
  finally { loadingFavorites.value = false }
}
async function toggleFollow() {
  if (!userStore.isLoggedIn) { showToast('请先登录', 'error'); return }
  try { const data = await post<FollowResp>(`/api/v1/users/${userId.value}/follow`); isFollowing.value = data.following; stats.value.followers += data.following ? 1 : -1 }
  catch (err) { showToast(err instanceof Error ? err.message : '操作失败', 'error') }
}
function startEditing() { if (!user.value) return; editForm.value.nickname = user.value.nickname || ''; editForm.value.bio = user.value.bio || ''; editing.value = true }
async function saveProfile() {
  try { const updated = await put<UserInfo>(`/api/v1/users/${userId.value}`, { nickname: editForm.value.nickname || undefined, bio: editForm.value.bio || undefined }); user.value = updated; editing.value = false; showToast('资料已更新', 'success') }
  catch (err) { showToast(err instanceof Error ? err.message : '更新失败', 'error') }
}
function triggerAvatarUpload() { fileInput.value?.click() }
async function handleAvatarFile(event: Event) {
  const target = event.target as HTMLInputElement; const file = target.files?.[0]; if (!file) return
  if (file.size > 2 * 1024 * 1024) { showToast('文件大小不能超过 2MB', 'error'); target.value = ''; return }
  const allowed = ['image/jpeg', 'image/png', 'image/webp']
  if (!allowed.includes(file.type)) { showToast('仅支持 JPEG、PNG、WebP', 'error'); target.value = ''; return }
  avatarUploading.value = true
  try {
    const fd = new FormData(); fd.append('avatar', file)
    const updated = await post<UserInfo>(`/api/v1/users/${userId.value}/avatar`, fd)
    user.value = updated
    if (userStore.userInfo && updated.avatar) userStore.userInfo.avatar = updated.avatar
    showToast('头像已更新', 'success')
  } catch (err) { showToast(err instanceof Error ? err.message : '头像上传失败', 'error') }
  finally { avatarUploading.value = false; target.value = '' }
}
function formatTime(d: string) { return new Date(d).toLocaleDateString('zh-CN') }

watch(activeTab, (tab) => { if (tab === 'favorites') fetchFavorites() })
onMounted(() => { fetchAll() })
</script>
