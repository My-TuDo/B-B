<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-[var(--color-text)]">稿件管理</h1>
      <NuxtLink
        to="/upload"
        class="inline-flex items-center gap-1.5 px-4 py-2 bg-[var(--color-primary)] text-white text-sm rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
        上传视频
      </NuxtLink>
    </div>

    <!-- Status tabs -->
    <div class="flex gap-2 mb-6">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="activeTab = tab.key"
        :class="[
          'px-4 py-2 text-sm rounded-[var(--radius-full)] transition-colors',
          activeTab === tab.key
            ? 'bg-[var(--color-primary)] text-white'
            : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)]'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" />

    <!-- Error -->
    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchVideos"
    />

    <!-- Video list -->
    <div v-else-if="drafts.length > 0" class="space-y-4">
      <div
        v-for="video in drafts"
        :key="video.id"
        class="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-4"
      >
        <div class="flex flex-col sm:flex-row sm:items-center gap-4">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="text-base font-medium text-[var(--color-text)] truncate">{{ video.title }}</h3>
              <span
                :class="[
                  'inline-block px-2 py-0.5 text-xs rounded-[var(--radius-full)]',
                  video.status === 0
                    ? 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)]'
                    : 'bg-green-900/30 text-green-400'
                ]"
              >
                {{ video.status === 0 ? '草稿' : '已发布' }}
              </span>
            </div>
            <p class="text-xs text-[var(--color-text-secondary)] mt-1">
              上传于 {{ formatTime(video.created_at) }} · {{ formatFileSize(video.file_size) }}
            </p>
          </div>

          <div class="flex gap-2 flex-shrink-0">
            <template v-if="video.status === 1">
              <NuxtLink
                :to="`/video/${video.id}`"
                class="px-4 py-1.5 text-sm bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors inline-block"
              >
                查看
              </NuxtLink>
            </template>
            <template v-else>
              <button
                class="px-4 py-1.5 text-sm bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
                @click="publishDraft(video)"
              >
                发布
              </button>
              <button
                class="px-4 py-1.5 text-sm bg-[var(--color-surface-hover)] text-[var(--color-text)] rounded-[var(--radius-full)] hover:bg-[var(--color-border)] transition-colors"
                @click="openEdit(video)"
              >
                编辑
              </button>
              <button
                class="px-4 py-1.5 text-sm bg-red-900/30 text-[var(--color-danger)] rounded-[var(--radius-full)] hover:bg-red-900/50 transition-colors"
                @click="deleteVideo(video)"
              >
                删除
              </button>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <EmptyState v-else :message="emptyMessage" icon="📝" />

    <!-- Edit modal -->
    <div
      v-if="editingDraft"
      class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 p-4"
      @click.self="editingDraft = null"
    >
      <div class="w-full max-w-lg bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-6">
        <h2 class="text-lg font-semibold text-[var(--color-text)] mb-4">编辑视频信息</h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm text-[var(--color-text-secondary)] mb-1">标题</label>
            <input
              v-model="editForm.title"
              type="text"
              maxlength="100"
              class="w-full h-10 px-3 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)]"
            />
          </div>
          <div>
            <label class="block text-sm text-[var(--color-text-secondary)] mb-1">简介</label>
            <textarea
              v-model="editForm.description"
              maxlength="2000"
              rows="3"
              class="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] resize-none"
            ></textarea>
          </div>
        </div>

        <div class="flex justify-end gap-2 mt-6">
          <button
            class="px-4 py-2 text-sm bg-[var(--color-surface-hover)] text-[var(--color-text)] rounded-[var(--radius-full)] hover:bg-[var(--color-border)] transition-colors"
            @click="editingDraft = null"
          >
            取消
          </button>
          <button
            class="px-4 py-2 text-sm bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
            @click="saveEdit"
          >
            保存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'
import type { VideoInfo, PaginatedData } from '~/types'

definePageMeta({
  middleware: 'auth',
})

const { get, put, del } = useApi()
const userStore = useUserStore()
const { showToast } = useToast()

const tabs = [
  { key: 'all', label: '全部' },
  { key: 'draft', label: '草稿' },
  { key: 'published', label: '已发布' },
]
const activeTab = ref('all')
const drafts = ref<VideoInfo[]>([])
const loading = ref(true)
const error = ref('')
const editingDraft = ref<VideoInfo | null>(null)
const editForm = ref({ title: '', description: '' })

const emptyMessage = computed(() => {
  switch (activeTab.value) {
    case 'draft': return '还没有草稿'
    case 'published': return '还没有已发布的视频'
    default: return '还没有视频，快去上传吧'
  }
})

watch(activeTab, () => {
  fetchVideos()
})

async function fetchVideos() {
  if (!userStore.userInfo) {
    error.value = '请先登录'
    loading.value = false
    return
  }

  loading.value = true
  error.value = ''

  try {
    const params: Record<string, number> = {}
    if (activeTab.value === 'draft') {
      params.status = 0
    } else if (activeTab.value === 'published') {
      params.status = 1
    }

    const data = await get<PaginatedData<VideoInfo>>(`/api/v1/videos/users/${userStore.userInfo.id}/videos`, params)
    drafts.value = data.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载视频列表失败'
  } finally {
    loading.value = false
  }
}

async function publishDraft(draft: VideoInfo) {
  try {
    await put(`/api/v1/videos/${draft.id}`, { status: 1 })
    showToast('视频已发布', 'success')
    await fetchVideos()
  } catch (err) {
    const msg = err instanceof Error ? err.message : '发布失败'
    showToast(msg, 'error')
  }
}

function openEdit(draft: VideoInfo) {
  editingDraft.value = draft
  editForm.value.title = draft.title
  editForm.value.description = draft.description
}

async function saveEdit() {
  if (!editingDraft.value) return

  try {
    await put(`/api/v1/videos/${editingDraft.value.id}`, {
      title: editForm.value.title,
      description: editForm.value.description,
    })
    showToast('修改已保存', 'success')
    editingDraft.value = null
    await fetchVideos()
  } catch (err) {
    const msg = err instanceof Error ? err.message : '保存失败'
    showToast(msg, 'error')
  }
}

async function deleteVideo(draft: VideoInfo) {
  if (!confirm('确定要删除这个视频吗？')) return

  try {
    await del(`/api/v1/videos/${draft.id}`)
    showToast('视频已删除', 'success')
    drafts.value = drafts.value.filter(v => v.id !== draft.id)
    await fetchVideos()
  } catch (err) {
    const msg = err instanceof Error ? err.message : '删除失败'
    showToast(msg, 'error')
  }
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

function formatFileSize(bytes: number): string {
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${bytes} B`
}

onMounted(fetchVideos)
</script>
