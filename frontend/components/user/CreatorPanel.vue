<template>
  <div class="space-y-6">
    <LoadingSpinner v-if="loading" />
    <div v-else>
      <!-- Stats cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
          <p class="text-xs text-[var(--color-text-secondary)] uppercase tracking-wide">总播放</p>
          <p class="mt-1 text-xl font-bold text-[var(--color-text)]">{{ stats.total_views }}</p>
        </div>
        <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
          <p class="text-xs text-[var(--color-text-secondary)] uppercase tracking-wide">总视频</p>
          <p class="mt-1 text-xl font-bold text-[var(--color-text)]">{{ stats.total_videos }}</p>
        </div>
        <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
          <p class="text-xs text-[var(--color-text-secondary)] uppercase tracking-wide">今日播放</p>
          <p class="mt-1 text-xl font-bold text-[var(--color-text)]">{{ stats.today_views }}</p>
        </div>
        <div class="p-4 bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl">
          <p class="text-xs text-[var(--color-text-secondary)] uppercase tracking-wide">新粉丝</p>
          <p class="mt-1 text-xl font-bold text-[var(--color-text)]">{{ stats.today_new_fans }}</p>
        </div>
      </div>

      <!-- Video management table -->
      <div class="bg-[var(--color-surface)] border border-[var(--color-border)]/30 rounded-xl overflow-hidden">
        <div class="px-5 py-3 border-b border-[var(--color-border)] flex items-center justify-between">
          <h3 class="text-sm font-semibold text-[var(--color-text)]">稿件管理</h3>
          <NuxtLink to="/upload" class="px-4 py-1.5 text-xs font-medium bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-colors">上传视频</NuxtLink>
        </div>
        <table class="w-full">
          <thead>
            <tr class="border-b border-[var(--color-border)]">
              <th class="px-5 py-2.5 text-left text-xs font-medium text-[var(--color-text-secondary)]">视频</th>
              <th class="px-5 py-2.5 text-left text-xs font-medium text-[var(--color-text-secondary)] hidden sm:table-cell">状态</th>
              <th class="px-5 py-2.5 text-left text-xs font-medium text-[var(--color-text-secondary)] hidden sm:table-cell">播放</th>
              <th class="px-5 py-2.5 text-right text-xs font-medium text-[var(--color-text-secondary)]">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--color-border)]">
            <tr v-for="v in videos" :key="v.id" class="hover:bg-[var(--color-surface-hover)]/50 transition-colors">
              <td class="px-5 py-3">
                <div class="flex items-center gap-3 min-w-0">
                  <NuxtLink :to="`/video/${v.id}`" class="w-28 h-16 rounded bg-[var(--color-surface-hover)] overflow-hidden flex-shrink-0">
                    <img v-if="v.cover_url" :src="v.cover_url" class="w-full h-full object-cover" />
                    <div v-else class="w-full h-full flex items-center justify-center text-[var(--color-text-secondary)] text-xs">无封面</div>
                  </NuxtLink>
                  <div class="min-w-0">
                    <NuxtLink :to="`/video/${v.id}`" class="text-sm font-medium text-[var(--color-text)] hover:text-[var(--color-primary)] transition-colors line-clamp-2">{{ v.title }}</NuxtLink>
                  </div>
                </div>
              </td>
              <td class="px-5 py-3 hidden sm:table-cell">
                <span class="text-xs px-2 py-0.5 rounded-full" :class="statusBadge(v.status)">{{ statusText(v.status) }}</span>
              </td>
              <td class="px-5 py-3 text-sm text-[var(--color-text-secondary)] hidden sm:table-cell">{{ formatNum(v.views) }}</td>
              <td class="px-5 py-3 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <button class="px-3 py-1 text-xs rounded-full border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors" @click="startEdit(v)">编辑</button>
                  <button class="px-3 py-1 text-xs rounded-full border border-red-500/20 text-red-400 hover:bg-red-500/10 transition-colors" @click="confirmDelete(v)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="videos.length === 0" class="px-5 py-10 text-center text-sm text-[var(--color-text-secondary)]">暂无稿件，<NuxtLink to="/upload" class="text-[var(--color-primary)] hover:underline">上传第一个视频</NuxtLink></div>
      </div>

      <!-- Edit modal -->
      <AppModal :visible="editVisible" title="编辑视频" @close="editVisible = false">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">标题</label>
            <input v-model="editForm.title" type="text" maxlength="100"
              class="w-full h-10 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] transition-all" />
          </div>
          <div>
            <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">简介</label>
            <textarea v-model="editForm.description" maxlength="500" rows="3"
              class="w-full px-3.5 py-2.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] transition-all resize-none" />
          </div>
          <button class="w-full py-2 bg-[var(--color-primary)] text-white text-sm font-medium rounded-full hover:bg-[var(--color-primary-hover)] transition-all active:scale-[0.98]" @click="saveEdit">保存</button>
        </div>
      </AppModal>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CreatorStats, VideoInfo, PaginatedData } from '~/types'
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'

const props = defineProps<{ userId: number }>()
const { get, put, del } = useApi()
const { showToast } = useToast()
const loading = ref(true)
const stats = ref<CreatorStats>({ total_views: 0, total_videos: 0, today_views: 0, today_new_fans: 0 })
const videos = ref<VideoInfo[]>([])
const editVisible = ref(false)
const editForm = ref({ id: 0, title: '', description: '' })

async function load() {
  loading.value = true
  try { stats.value = await get<CreatorStats>('/api/v1/creator/stats') }
  catch { /* defaults */ }
  try {
    const data = await get<PaginatedData<VideoInfo>>(`/api/v1/videos/users/${props.userId}/videos`, { page_size: 50 })
    videos.value = data.items || []
  } catch { videos.value = [] }
  loading.value = false
}

function statusText(s: number): string {
  switch (s) { case 0: return '草稿'; case 1: return '已发布'; case 2: return '审核中'; case 3: return '已删除'; default: return '未知' }
}
function statusBadge(s: number): string {
  switch (s) { case 0: return 'bg-yellow-500/10 text-yellow-500'; case 1: return 'bg-green-500/10 text-green-500'; case 2: return 'bg-blue-500/10 text-blue-500'; default: return 'bg-red-500/10 text-red-500' }
}
function formatNum(n: number) { return n >= 10000 ? `${(n / 10000).toFixed(1)}万` : String(n) }

function startEdit(v: VideoInfo) {
  editForm.value = { id: v.id, title: v.title, description: v.description || '' }
  editVisible.value = true
}
async function saveEdit() {
  try {
    await put(`/api/v1/videos/${editForm.value.id}`, { title: editForm.value.title, description: editForm.value.description })
    const idx = videos.value.findIndex(v => v.id === editForm.value.id)
    if (idx >= 0) { videos.value[idx].title = editForm.value.title; videos.value[idx].description = editForm.value.description }
    editVisible.value = false; showToast('已更新', 'success')
  } catch (err) { showToast(err instanceof Error ? err.message : '更新失败', 'error') }
}
async function confirmDelete(v: VideoInfo) {
  if (!confirm(`确定删除「${v.title}」？`)) return
  try {
    await del(`/api/v1/videos/${v.id}`)
    videos.value = videos.value.filter(x => x.id !== v.id)
    showToast('已删除', 'success')
  } catch (err) { showToast(err instanceof Error ? err.message : '删除失败', 'error') }
}

onMounted(() => load())
</script>
