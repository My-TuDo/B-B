<template>
  <div class="max-w-3xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-bold text-[var(--color-text)]">消息通知</h1>
      <button
        v-if="unread > 0"
        class="px-4 py-1.5 text-sm font-medium text-[var(--color-primary)] bg-[var(--color-primary)]/10 rounded-[var(--radius-full)] hover:bg-[var(--color-primary)]/20 transition-colors"
        @click="readAll"
      >
        全部已读
      </button>
    </div>

    <div v-if="loading">
      <LoadingSpinner />
    </div>

    <ErrorMessage
      v-else-if="error"
      :message="error"
      :on-retry="fetchNotifications"
    />

    <div v-else-if="notifications.length === 0">
      <EmptyState message="暂无消息" icon="&#128276;" />
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="item in notifications"
        :key="item.id"
        class="flex items-start gap-4 p-4 rounded-[var(--radius-md)] transition-colors cursor-pointer"
        :class="item.is_read === 0 ? 'bg-[var(--color-primary)]/5 hover:bg-[var(--color-primary)]/10' : 'bg-[var(--color-surface)] hover:bg-[var(--color-surface-hover)]'"
        @click="handleClick(item)"
      >
        <!-- Type icon -->
        <div class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" :class="typeStyle(item.type).bg">
          <span class="text-lg">{{ typeStyle(item.type).icon }}</span>
        </div>

        <!-- Content -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-[var(--color-text)]">
              {{ item.from_user?.nickname || item.from_user?.username || '系统' }}
            </span>
            <span class="text-xs text-[var(--color-text-secondary)]">
              {{ typeStyle(item.type).label }}
            </span>
            <div v-if="item.is_read === 0" class="w-2 h-2 rounded-full bg-[var(--color-primary)] flex-shrink-0" />
          </div>
          <p class="text-sm text-[var(--color-text-secondary)] mt-1 truncate">{{ item.content }}</p>
          <p class="text-xs text-[var(--color-text-secondary)]/60 mt-1">{{ formatTime(item.created_at) }}</p>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="total > pageSize" class="flex justify-center mt-6 gap-2">
      <button
        v-for="p in Math.ceil(total / pageSize)"
        :key="p"
        class="w-8 h-8 text-sm rounded-full transition-colors"
        :class="p === page ? 'bg-[var(--color-primary)] text-white' : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)]'"
        @click="page = p; fetchNotifications()"
      >
        {{ p }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NotificationItem, NotificationListResp } from '~/types'
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'

definePageMeta({ middleware: 'auth' })

const { get, post } = useApi()
const userStore = useUserStore()
const router = useRouter()

const notifications = ref<NotificationItem[]>([])
const total = ref(0)
const unread = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const error = ref('')

function typeStyle(type: number): { icon: string; label: string; bg: string } {
  switch (type) {
    case 1: return { icon: '💬', label: '评论了你', bg: 'bg-blue-500/10' }
    case 2: return { icon: '👍', label: '赞了你', bg: 'bg-pink-500/10' }
    case 3: return { icon: '👤', label: '关注了你', bg: 'bg-green-500/10' }
    case 4: return { icon: '📋', label: '审核通知', bg: 'bg-yellow-500/10' }
    default: return { icon: '📌', label: '通知', bg: 'bg-gray-500/10' }
  }
}

async function fetchNotifications() {
  loading.value = true
  error.value = ''
  try {
    const data = await get<NotificationListResp>('/api/v1/notifications/', {
      page: page.value,
      page_size: pageSize.value,
    })
    notifications.value = data.items || []
    total.value = data.total || 0
    unread.value = data.unread || 0
    userStore.unreadCount = data.unread || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function readAll() {
  try {
    await post('/api/v1/notifications/read-all')
    userStore.clearUnread()
    fetchNotifications()
  } catch {
    // silent
  }
}

async function handleClick(item: NotificationItem) {
  // Mark as read if unread
  if (item.is_read === 0) {
    item.is_read = 1
    unread.value = Math.max(0, unread.value - 1)
    userStore.decrementUnread()
    post('/api/v1/notifications/' + item.id + '/read').catch(() => {})
  }

  if (item.type === 1 && item.target_id > 0) {
    router.push(`/video/${item.target_id}`)
  } else if (item.type === 3 && item.from_user) {
    router.push(`/user/${item.from_user.id}`)
  }
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}

onMounted(fetchNotifications)
</script>
