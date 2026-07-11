<template>
  <div>
    <LoadingSpinner v-if="loadingUser" />
    <ErrorMessage v-else-if="error" :message="error" :on-retry="fetchAll" />
    <div v-else-if="user">
      <!-- Banner -->
      <div class="relative rounded-2xl overflow-hidden mb-6 bg-gradient-to-br from-[var(--color-primary)]/10 via-[var(--color-surface)] to-[var(--color-primary)]/5 border border-[var(--color-border)]/30">
        <div class="px-6 pt-6 pb-5">
          <div class="flex flex-col sm:flex-row items-center sm:items-end gap-4">
            <!-- Avatar -->
            <div class="w-20 h-20 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-xl font-bold flex-shrink-0 overflow-hidden ring-4 ring-[var(--color-bg)] relative group"
              :class="isOwner ? 'cursor-pointer' : ''" @click="isOwner ? triggerAvatarUpload() : undefined">
              <div v-if="avatarUploading" class="w-full h-full bg-black/40 flex items-center justify-center">
                <svg class="animate-spin w-6 h-6 text-white" viewBox="0 0 24 24" fill="none"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
              </div>
              <img v-else-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" />
              <span v-else>{{ (user.nickname || user.username).charAt(0) }}</span>
              <div v-if="isOwner && !avatarUploading" class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-full">
                <span class="text-white text-[10px] font-medium text-center leading-tight">更换<br>头像</span>
              </div>
            </div>
            <input ref="fileInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="handleAvatarFile" />

            <div class="flex-1 text-center sm:text-left">
              <h1 class="text-xl font-bold text-[var(--color-text)]">{{ user.nickname || user.username }}</h1>
              <p class="text-sm text-[var(--color-text-secondary)]">@{{ user.username }} · 注册于 {{ formatTime(user.created_at) }}</p>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <button v-if="!isOwner && userStore.isLoggedIn"
                class="px-5 py-2 text-sm font-medium rounded-full transition-all active:scale-[0.98]"
                :class="isFollowing ? 'bg-[var(--color-surface-hover)] text-[var(--color-text)]' : 'bg-[var(--color-primary)] text-white shadow-sm shadow-[var(--color-primary)]/25'"
                @click="toggleFollow">{{ isFollowing ? '已关注' : '关注' }}</button>
              <button v-if="isOwner && !editing"
                class="px-4 py-2 text-sm font-medium rounded-full border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-all active:scale-[0.98] flex items-center gap-1.5"
                @click="startEditing">
                <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑资料
              </button>
            </div>
          </div>
        </div>
        <div class="flex justify-center gap-8 px-6 pb-4 border-t border-[var(--color-border)]/30 pt-3">
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.videos }}</p><p class="text-xs text-[var(--color-text-secondary)]">视频</p></div>
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.followers }}</p><p class="text-xs text-[var(--color-text-secondary)]">粉丝</p></div>
          <div class="text-center"><p class="text-lg font-bold text-[var(--color-text)]">{{ stats.following }}</p><p class="text-xs text-[var(--color-text-secondary)]">关注</p></div>
        </div>
      </div>

      <!-- Edit modal -->
      <AppModal :visible="editing" title="编辑资料" @close="editing = false">
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">昵称</label><input v-model="editForm.nickname" type="text" maxlength="50" class="w-full h-10 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all" /></div>
          <div><label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">签名</label><textarea v-model="editForm.bio" maxlength="500" rows="3" class="w-full px-3.5 py-2.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all resize-none" /></div>
          <div class="flex gap-2">
            <button class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm font-medium rounded-full hover:bg-[var(--color-primary-hover)] active:scale-[0.98]" @click="saveProfile">保存</button>
            <button class="px-6 py-2 bg-[var(--color-surface-hover)] text-[var(--color-text)] text-sm font-medium rounded-full hover:bg-[var(--color-border)] active:scale-[0.98]" @click="editing = false">取消</button>
          </div>
        </div>
      </AppModal>

      <!-- Tab bar with stats -->
      <div class="flex items-center gap-6 border-b border-[var(--color-border)] mb-6">
        <div class="flex gap-6 flex-1">
          <button v-for="tab in tabs" :key="tab.key"
            class="pb-3 text-sm font-medium transition-colors border-b-2 flex items-center gap-1.5"
            :class="activeTab === tab.key ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
            @click="activeTab = tab.key">
            <span v-html="tab.icon" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
        <div class="hidden sm:flex items-center gap-4 text-xs text-[var(--color-text-secondary)] pb-3">
          <span>关注 {{ stats.following }}</span>
          <span class="text-[var(--color-border)]">|</span>
          <span>粉丝 {{ stats.followers }}</span>
        </div>
      </div>

      <!-- 主页 Tab: horz scroll + right sidebar -->
      <div v-if="activeTab === 'home'" class="flex flex-col lg:flex-row gap-6">
        <div class="flex-1 min-w-0 space-y-8">
          <div>
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-base font-semibold text-[var(--color-text)]">投稿 · {{ videos.length }}</h3>
              <NuxtLink to="?tab=videos" class="text-xs text-[var(--color-primary)] hover:underline">查看更多 →</NuxtLink>
            </div>
            <div v-if="videos.length > 0" class="flex gap-4 overflow-x-auto pb-2">
              <div v-for="v in videos.slice(0, 5)" :key="v.id" class="w-52 flex-shrink-0"><VideoCard :video="v" /></div>
            </div>
            <EmptyState v-else message="还没有发布视频" />
          </div>
          <div v-if="favSummary.length > 0">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-base font-semibold text-[var(--color-text)]">收藏夹</h3>
              <NuxtLink to="?tab=favorites" class="text-xs text-[var(--color-primary)] hover:underline">查看更多 →</NuxtLink>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div v-for="fav in favSummary" :key="fav.id" class="p-3 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer" @click="router.push(`/favorites/${fav.id}`)"><p class="text-sm font-medium text-[var(--color-text)]">{{ fav.name }}</p><p class="text-xs text-[var(--color-text-secondary)] mt-1">{{ fav.item_count }} 个视频</p></div>
            </div>
          </div>
        </div>
        <div class="w-full lg:w-64 flex-shrink-0 space-y-3">
          <div v-if="isOwner" class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
            <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-2">创作中心</p>
            <NuxtLink to="/creator" class="block text-center py-2 text-xs font-medium bg-[var(--color-primary)]/10 text-[var(--color-primary)] rounded-lg hover:bg-[var(--color-primary)]/20 transition-colors">视频投稿</NuxtLink>
          </div>
          <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
            <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-2">个人简介</p>
            <p class="text-sm text-[var(--color-text)] leading-relaxed">{{ user.bio || '这个人很懒，什么都没写' }}</p>
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
          <div v-for="fav in publicFavorites" :key="fav.id" class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-surface-hover)] transition-colors cursor-pointer" @click="router.push(`/favorites/${fav.id}`)">
            <div class="flex items-center justify-between"><div><p class="text-sm font-medium text-[var(--color-text)]">{{ fav.name }}</p><p class="text-xs text-[var(--color-text-secondary)] mt-1">{{ fav.item_count }} 个视频</p></div><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text-secondary)]"><polyline points="9 18 15 12 9 6"/></svg></div>
          </div>
        </div>
        <EmptyState v-else message="暂无公开收藏夹" />
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
const favSummary = computed(() => publicFavorites.value.slice(0, 3))

const homeIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`
const videoIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>`
const starIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>`

const tabs = computed(() => [
  { key: 'home', label: '主页', icon: homeIcon },
  { key: 'videos', label: `投稿 (${videos.value.length})`, icon: videoIcon },
  { key: 'favorites', label: '收藏', icon: starIcon },
])

async function fetchAll() { await fetchUser(); await Promise.all([fetchVideos(), fetchProfile()]); fetchFavorites() }
async function fetchUser() {
  loadingUser.value = true; error.value = ''
  try { user.value = await get<UserInfo>(`/api/v1/users/${userId.value}`) }
  catch (err) { error.value = err instanceof Error ? err.message : '加载用户信息失败' }
  finally { loadingUser.value = false }
}
async function fetchProfile() {
  try { const data = await get<ProfileResp>(`/api/v1/users/${userId.value}/profile`); stats.value = data.stats; isFollowing.value = data.stats.following > 0 }
  catch { /* */ }
}
async function fetchVideos() {
  loadingVideos.value = true
  try { const data = await get<PaginatedData<VideoInfo>>(`/api/v1/videos/users/${userId.value}/videos`, { status: 1 }); videos.value = data.items || [] }
  catch { }
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
  if (!['image/jpeg','image/png','image/webp'].includes(file.type)) { showToast('仅支持 JPEG、PNG、WebP', 'error'); target.value = ''; return }
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
onMounted(() => { fetchAll() })
</script>
