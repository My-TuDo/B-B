<template>
  <div class="max-w-2xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-8">上传视频</h1>

    <form @submit.prevent="handleUpload" class="space-y-6">
      <!-- File picker -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">视频文件</label>
        <div
          class="border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 text-center hover:border-[var(--color-primary)] transition-colors cursor-pointer"
          :class="{ 'border-[var(--color-primary)]': dragOver }"
          @click="triggerFileInput"
          @dragover.prevent="dragOver = true"
          @dragleave.prevent="dragOver = false"
          @drop.prevent="handleDrop"
        >
          <input
            ref="fileInput"
            type="file"
            accept="video/*"
            class="hidden"
            @change="handleFileChange"
          />
          <template v-if="!selectedFile">
            <div class="text-4xl mb-3 opacity-50">&#9654;</div>
            <p class="text-[var(--color-text-secondary)] text-sm">点击或拖拽视频文件到此处</p>
            <p class="text-[var(--color-text-secondary)] text-xs mt-1">支持 MP4, WebM, OGG, MOV, AVI, MKV，最大 500MB</p>
          </template>
          <template v-else>
            <div class="text-[var(--color-text)]">
              <p class="font-medium">{{ selectedFile.name }}</p>
              <p class="text-sm text-[var(--color-text-secondary)] mt-1">{{ formatFileSize(selectedFile.size) }}</p>
            </div>
            <button type="button" class="mt-3 text-sm text-[var(--color-danger)] hover:underline" @click.stop="clearFile">
              移除
            </button>
          </template>
        </div>
      </div>

      <!-- Cover image picker (optional) -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">封面图片（可选）</label>
        <div
          class="border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-lg)] p-6 text-center hover:border-[var(--color-primary)] transition-colors cursor-pointer"
          :class="{ 'border-[var(--color-primary)]': coverDragOver }"
          @click="triggerCoverInput"
          @dragover.prevent="coverDragOver = true"
          @dragleave.prevent="coverDragOver = false"
          @drop.prevent="handleCoverDrop"
        >
          <input
            ref="coverInput"
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif"
            class="hidden"
            @change="handleCoverChange"
          />
          <template v-if="!selectedCover">
            <div class="text-3xl mb-2 opacity-50">&#128444;</div>
            <p class="text-[var(--color-text-secondary)] text-sm">点击或拖拽封面图片（选填）</p>
            <p class="text-[var(--color-text-secondary)] text-xs mt-1">支持 JPEG, PNG, WebP, GIF，最大 5MB</p>
          </template>
          <template v-else>
            <div class="flex items-center justify-center gap-4">
              <img
                :src="coverPreviewUrl"
                class="h-20 rounded-[var(--radius-md)] object-cover"
                alt="封面预览"
              />
              <div class="text-left">
                <p class="text-[var(--color-text)] text-sm font-medium">{{ selectedCover.name }}</p>
                <p class="text-xs text-[var(--color-text-secondary)]">{{ formatFileSize(selectedCover.size) }}</p>
              </div>
            </div>
            <button
              type="button"
              class="mt-3 text-sm text-[var(--color-danger)] hover:underline"
              @click.stop="clearCover"
            >
              移除
            </button>
          </template>
        </div>
      </div>

      <!-- Title -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">标题</label>
        <input
          v-model="title"
          type="text"
          placeholder="视频标题（最多100字）"
          maxlength="100"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
      </div>

      <!-- Description -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">简介</label>
        <textarea
          v-model="description"
          placeholder="视频简介（选填）"
          maxlength="2000"
          rows="4"
          class="w-full px-3.5 py-2.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)] resize-none"
        ></textarea>
      </div>

      <!-- Category -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">分类</label>
        <select
          v-model="categoryId"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
        >
          <option :value="0">请选择分类</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>
      </div>

      <!-- Tags -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-2">标签</label>
        <div class="flex flex-wrap gap-2 mb-2">
          <span
            v-for="tag in selectedTags"
            :key="tag.id"
            class="inline-flex items-center gap-1 px-2 py-1 text-xs bg-[var(--color-primary)]/20 text-[var(--color-primary)] rounded-[var(--radius-full)]"
          >
            {{ tag.name }}
            <button type="button" class="hover:text-[var(--color-danger)] transition-colors" @click="removeTag(tag)">&times;</button>
          </span>
        </div>
        <div class="relative">
          <input
            v-model="tagInput"
            type="text"
            placeholder="输入标签名后按回车添加"
            class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
            @keydown.enter.prevent="addTag"
            @input="onTagInput"
          />
          <!-- Tag suggestions -->
          <div
            v-if="tagSuggestions.length > 0 && showTagSuggestions"
            class="absolute left-0 right-0 top-full mt-1 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-md)] shadow-lg z-50 max-h-40 overflow-y-auto"
          >
            <button
              v-for="t in tagSuggestions"
              :key="t.id"
              type="button"
              class="w-full text-left px-3 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors"
              @click="selectTagSuggestion(t)"
            >
              {{ t.name }}
            </button>
          </div>
        </div>
      </div>

      <!-- Progress -->
      <div v-if="uploading" class="space-y-2">
        <div class="w-full bg-[var(--color-bg)] rounded-[var(--radius-full)] h-2 overflow-hidden">
          <div
            class="h-full bg-[var(--color-primary)] rounded-[var(--radius-full)] transition-all duration-300"
            :style="{ width: uploadPercent + '%' }"
          ></div>
        </div>
        <p class="text-xs text-[var(--color-text-secondary)] text-center">
          {{ formatFileSize(uploadedBytes) }} / {{ formatFileSize(totalBytes) }} ({{ uploadPercent }}%)
        </p>
      </div>

      <!-- Submit -->
      <button
        type="submit"
        :disabled="!selectedFile || !title || uploading"
        class="w-full h-11 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-md)] hover:bg-[var(--color-primary-hover)] transition-all duration-[var(--transition-normal)] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm shadow-[var(--color-primary)]/25 active:scale-[0.98]"
      >
        {{ uploading ? '上传中...' : '上传视频' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useApi, getCSRFToken } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import type { Category, Tag } from '~/types'

definePageMeta({
  middleware: 'auth',
})

const { get, post } = useApi()
const router = useRouter()
const { showToast } = useToast()

const fileInput = ref<HTMLInputElement>()
const coverInput = ref<HTMLInputElement>()
const categories = ref<Category[]>([])
const dragOver = ref(false)
const coverDragOver = ref(false)

const selectedFile = ref<File | null>(null)
const selectedCover = ref<File | null>(null)
const coverPreviewUrl = ref<string>('')
const title = ref('')
const description = ref('')
const categoryId = ref<number>(0)

// Tags
const allTags = ref<Tag[]>([])
const selectedTags = ref<Tag[]>([])
const tagInput = ref('')
const tagSuggestions = ref<Tag[]>([])
const showTagSuggestions = ref(false)

const uploading = ref(false)
const uploadedBytes = ref(0)
const totalBytes = ref(0)

const uploadPercent = computed(() => {
  if (totalBytes.value === 0) return 0
  return Math.round((uploadedBytes.value / totalBytes.value) * 100)
})

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) {
    validateAndSetFile(input.files[0])
  }
}

function handleDrop(e: DragEvent) {
  dragOver.value = false
  if (e.dataTransfer?.files.length) {
    validateAndSetFile(e.dataTransfer.files[0])
  }
}

function validateAndSetFile(file: File) {
  const maxSize = 500 * 1024 * 1024 // 500MB
  if (file.size > maxSize) {
    showToast('文件大小不能超过 500MB', 'error')
    return
  }
  const validTypes = ['video/mp4', 'video/webm', 'video/ogg', 'video/quicktime', 'video/x-msvideo', 'video/x-matroska']
  if (!validTypes.includes(file.type) && file.type !== '') {
    showToast('不支持的视频格式', 'error')
    return
  }
  selectedFile.value = file
}

function clearFile() {
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function triggerCoverInput() {
  coverInput.value?.click()
}

function handleCoverChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) {
    validateAndSetCover(input.files[0])
  }
}

function handleCoverDrop(e: DragEvent) {
  coverDragOver.value = false
  if (e.dataTransfer?.files.length) {
    validateAndSetCover(e.dataTransfer.files[0])
  }
}

function validateAndSetCover(file: File) {
  const maxSize = 5 * 1024 * 1024 // 5MB
  if (file.size > maxSize) {
    showToast('封面图片大小不能超过 5MB', 'error')
    return
  }
  const validTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
  if (!validTypes.includes(file.type) && file.type !== '') {
    showToast('封面图片格式不支持，仅支持 JPEG/PNG/WebP/GIF', 'error')
    return
  }
  // Revoke old preview URL
  if (coverPreviewUrl.value) {
    URL.revokeObjectURL(coverPreviewUrl.value)
  }
  selectedCover.value = file
  coverPreviewUrl.value = URL.createObjectURL(file)
}

function clearCover() {
  if (coverPreviewUrl.value) {
    URL.revokeObjectURL(coverPreviewUrl.value)
    coverPreviewUrl.value = ''
  }
  selectedCover.value = null
  if (coverInput.value) {
    coverInput.value.value = ''
  }
}

function formatFileSize(bytes: number): string {
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${bytes} B`
}

function addTag() {
  const name = tagInput.value.trim()
  if (!name) return
  const existing = allTags.value.find(t => t.name === name)
  if (existing) {
    if (!selectedTags.value.find(t => t.id === existing.id)) {
      selectedTags.value.push(existing)
    }
  } else {
    const newTag: Tag = { id: 0, name }
    selectedTags.value.push(newTag)
  }
  tagInput.value = ''
  tagSuggestions.value = []
  showTagSuggestions.value = false
}

function removeTag(tag: Tag) {
  selectedTags.value = selectedTags.value.filter(t => t.id !== tag.id && t.name !== tag.name)
}

function onTagInput() {
  const q = tagInput.value.trim().toLowerCase()
  if (!q) {
    tagSuggestions.value = []
    showTagSuggestions.value = false
    return
  }
  tagSuggestions.value = allTags.value.filter(
    t => t.name.toLowerCase().includes(q) && !selectedTags.value.find(st => st.id === t.id),
  ).slice(0, 5)
  showTagSuggestions.value = tagSuggestions.value.length > 0
}

function selectTagSuggestion(tag: Tag) {
  if (!selectedTags.value.find(t => t.id === tag.id)) {
    selectedTags.value.push(tag)
  }
  tagInput.value = ''
  tagSuggestions.value = []
  showTagSuggestions.value = false
}

async function handleUpload() {
  if (!selectedFile.value || !title.value) return

  uploading.value = true
  uploadedBytes.value = 0
  totalBytes.value = selectedFile.value.size

  const formData = new FormData()
  formData.append('file', selectedFile.value)
  formData.append('title', title.value)
  formData.append('description', description.value)
  formData.append('category_id', String(categoryId.value))

  // Append cover if selected
  if (selectedCover.value) {
    formData.append('cover', selectedCover.value)
  }

  try {
    // Use fetch directly for SSE progress
    const csrfToken = getCSRFToken()
    const headers: Record<string, string> = {}
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }

    const response = await fetch('/api/v1/videos/', {
      method: 'POST',
      credentials: 'include',
      headers,
      body: formData,
    })

    if (!response.ok) {
      if (response.status === 401) {
        showToast('请先登录', 'error')
        router.push('/login')
        return
      }
      const errData = await response.json()
      throw new Error(errData.message || '上传失败')
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('无法读取响应')
    }

    const decoder = new TextDecoder()
    let buffer = ''
    let uploadedVideoId = 0
    let expectCompleteEvent = false

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event: complete')) {
          expectCompleteEvent = true
          continue
        }
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6))
            if (data.error) {
              throw new Error(data.error)
            }
            if (expectCompleteEvent) {
              if (data.id) uploadedVideoId = data.id
              expectCompleteEvent = false
              break
            }
            if (data.uploaded !== undefined && data.total !== undefined) {
              uploadedBytes.value = data.uploaded
              totalBytes.value = data.total
            }
          } catch (err) {
            if (err instanceof Error && err.message !== 'Unexpected token' && !err.message.startsWith('data:')) {
              throw err
            }
          }
        }
      }
    }

    // After upload, create/lookup tags by name and associate with video
    if (uploadedVideoId && selectedTags.value.length > 0) {
      try {
        const tagIds: number[] = []
        for (const tag of selectedTags.value) {
          const resp = await post<Tag>(`/api/v1/tags/`, { name: tag.name })
          if (resp && resp.id) {
            tagIds.push(resp.id)
          }
        }
        if (tagIds.length > 0) {
          try {
            await post(`/api/v1/videos/${uploadedVideoId}/tags`, { tag_ids: tagIds })
          } catch { /* ignore tag association errors */ }
        }
      } catch { /* ignore tag creation errors on upload */ }
    }

    showToast('视频已保存为草稿', 'success')
    router.push('/drafts')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '上传失败'
    showToast(msg, 'error')
  } finally {
    uploading.value = false
  }
}

onMounted(async () => {
  try {
    const [cats, tags] = await Promise.all([
      get<Category[]>('/api/v1/categories/'),
      get<Tag[]>('/api/v1/tags/'),
    ])
    categories.value = cats || []
    allTags.value = tags || []
  } catch {
    // Non-critical
  }
})
</script>
