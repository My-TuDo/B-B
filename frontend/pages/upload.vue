<template>
  <div class="max-w-2xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-8">上传视频</h1>

    <form @submit.prevent="handleUpload" class="space-y-6">
      <!-- File picker -->
      <div>
        <label class="block text-sm text-[var(--color-text-secondary)] mb-2">视频文件</label>
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

      <!-- Title -->
      <div>
        <label class="block text-sm text-[var(--color-text-secondary)] mb-2">标题</label>
        <input
          v-model="title"
          type="text"
          placeholder="视频标题（最多100字）"
          maxlength="100"
          class="w-full h-10 px-3 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] transition-colors"
          required
        />
      </div>

      <!-- Description -->
      <div>
        <label class="block text-sm text-[var(--color-text-secondary)] mb-2">简介</label>
        <textarea
          v-model="description"
          placeholder="视频简介（选填）"
          maxlength="2000"
          rows="4"
          class="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] transition-colors resize-none"
        ></textarea>
      </div>

      <!-- Category -->
      <div>
        <label class="block text-sm text-[var(--color-text-secondary)] mb-2">分类</label>
        <select
          v-model="categoryId"
          class="w-full h-10 px-3 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm focus:outline-none focus:border-[var(--color-primary)] transition-colors"
        >
          <option :value="0">请选择分类</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>
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
        class="w-full h-11 bg-[var(--color-primary)] text-white text-sm rounded-[var(--radius-md)] hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ uploading ? '上传中...' : '上传视频' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import type { Category } from '~/types'

definePageMeta({
  middleware: 'auth',
})

const { get, post } = useApi()
const router = useRouter()
const { showToast } = useToast()

const fileInput = ref<HTMLInputElement>()
const categories = ref<Category[]>([])
const dragOver = ref(false)

const selectedFile = ref<File | null>(null)
const title = ref('')
const description = ref('')
const categoryId = ref<number>(0)

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

function formatFileSize(bytes: number): string {
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${bytes} B`
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

  try {
    // Use fetch directly for SSE progress
    const response = await fetch('/api/v1/videos/', {
      method: 'POST',
      credentials: 'include',
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

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6))
            if (data.error) {
              throw new Error(data.error)
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
    const data = await get<Category[]>('/api/v1/categories/')
    categories.value = data || []
  } catch {
    // Non-critical
  }
})
</script>
