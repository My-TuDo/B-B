<template>
  <div :class="['flex gap-3', { 'mt-3': level > 0 }]">
    <!-- Avatar (root comment) -->
    <div class="w-9 h-9 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-xs font-bold flex-shrink-0 overflow-hidden">
      <img v-if="comment.user?.avatar" :src="comment.user.avatar" class="w-full h-full object-cover" />
      <span v-else>{{ (comment.user?.nickname || comment.user?.username || '?').charAt(0) }}</span>
    </div>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium text-[var(--color-text)]">{{ comment.user?.nickname || comment.user?.username || '匿名' }}</span>
        <span class="text-xs text-[var(--color-text-secondary)]">{{ formatTime(comment.created_at) }}</span>
      </div>
      <p class="text-sm text-[var(--color-text)] mt-1 leading-relaxed break-words">{{ comment.content }}</p>

      <!-- Actions -->
      <div class="flex items-center gap-4 mt-2">
        <button
          class="flex items-center gap-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors"
          @click="$emit('like', comment.id)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>
          </svg>
          <span v-if="comment.likes > 0">{{ comment.likes }}</span>
        </button>
        <button
          class="text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors"
          @click="showReplyBox = !showReplyBox"
        >
          回复
        </button>
        <button
          v-if="canDelete"
          class="text-xs text-[var(--color-text-secondary)] hover:text-red-500 transition-colors"
          @click="$emit('delete', comment.id)"
        >
          删除
        </button>
      </div>

      <!-- Inline reply box (for replying to this root comment) -->
      <div v-if="showReplyBox" class="mt-3 flex gap-2">
        <textarea
          v-model="replyContent"
          class="flex-1 min-h-[48px] p-2 text-xs bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-[var(--radius-md)] resize-none focus:outline-none focus:border-[var(--color-primary)]/40"
          :placeholder="`回复 @${comment.user?.nickname || comment.user?.username || '匿名'}`"
          rows="2"
        />
        <div class="flex flex-col gap-1 flex-shrink-0">
          <button
            class="px-3 py-1 text-xs font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] disabled:opacity-40 transition-colors"
            :disabled="!replyContent.trim() || replySubmitting"
            @click="submitReply"
          >
            {{ replySubmitting ? '...' : '发送' }}
          </button>
          <button
            class="px-3 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors"
            @click="showReplyBox = false; replyContent = ''"
          >
            取消
          </button>
        </div>
      </div>

      <!-- Flat children replies (B站 style) -->
      <div
        v-if="props.comment.replies && props.comment.replies.length > 0"
        class="mt-2 bg-[var(--color-primary-soft)] rounded-[var(--radius-md)] px-3 py-1"
      >
        <!-- Rendered children -->
        <div
          v-for="(child, idx) in visibleChildren"
          :key="child.id"
          class="flex gap-2 py-2"
          :class="{ 'border-b border-[var(--color-border)]/50': idx < visibleChildren.length - 1 }"
        >
          <!-- Small avatar 24px -->
          <div class="w-6 h-6 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-[10px] font-bold flex-shrink-0 overflow-hidden">
            <img v-if="child.user?.avatar" :src="child.user.avatar" class="w-full h-full object-cover" />
            <span v-else>{{ (child.user?.nickname || child.user?.username || '?').charAt(0) }}</span>
          </div>

          <div class="flex-1 min-w-0">
            <p class="text-xs text-[var(--color-text)] leading-relaxed break-words">
              <span class="font-medium text-[var(--color-primary)] mr-1">{{ child.user?.nickname || child.user?.username || '匿名' }}</span>
              <span>{{ child.content }}</span>
            </p>

            <!-- Child reply actions -->
            <div class="flex items-center gap-3 mt-1">
              <span class="text-[11px] text-[var(--color-text-secondary)]">{{ formatTime(child.created_at) }}</span>
              <button
                class="flex items-center gap-0.5 text-[11px] text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors"
                @click="$emit('like', child.id)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>
                </svg>
                <span v-if="child.likes > 0">{{ child.likes }}</span>
              </button>
              <button
                class="text-[11px] text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors"
                @click="toggleChildReply(child.id)"
              >
                回复
              </button>
              <button
                v-if="canDeleteChild(child)"
                class="text-[11px] text-[var(--color-text-secondary)] hover:text-red-500 transition-colors"
                @click="$emit('delete', child.id)"
              >
                删除
              </button>
            </div>

            <!-- Child inline reply box -->
            <div v-if="activeChildReplyId === child.id" class="mt-2 flex gap-2">
              <textarea
                v-model="childReplyContents[child.id]"
                class="flex-1 min-h-[42px] p-2 text-xs bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-[var(--radius-md)] resize-none focus:outline-none focus:border-[var(--color-primary)]/40"
                :placeholder="`回复 @${child.user?.nickname || child.user?.username || '匿名'}`"
                rows="2"
              />
              <div class="flex flex-col gap-1 flex-shrink-0">
                <button
                  class="px-3 py-1 text-xs font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] disabled:opacity-40 transition-colors"
                  :disabled="!childReplyContents[child.id]?.trim() || childReplySubmitting[child.id]"
                  @click="submitChildReply(child)"
                >
                  {{ childReplySubmitting[child.id] ? '...' : '发送' }}
                </button>
                <button
                  class="px-3 py-1 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors"
                  @click="closeChildReply(child.id)"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Expand/collapse toggle -->
        <button
          v-if="props.comment.replies && props.comment.replies.length > 3"
          class="w-full py-1.5 text-xs text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] hover:underline transition-colors text-center"
          @click="showAllChildren = !showAllChildren"
        >
          {{ showAllChildren ? '收起回复' : `展开更多回复(${props.comment.replies.length - 3})` }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CommentItem } from '~/types'
import { useUserStore } from '~/stores/userStore'

const props = defineProps<{
  comment: CommentItem
  level: number
}>()

const emit = defineEmits<{
  reply: [comment: CommentItem]
  like: [commentId: number]
  delete: [commentId: number]
  'submit-reply': [parentComment: CommentItem, content: string]
}>()

const userStore = useUserStore()

// --- Root comment inline reply ---
const showReplyBox = ref(false)
const replyContent = ref('')
const replySubmitting = ref(false)

const canDelete = computed(() => {
  if (!userStore.userInfo) return false
  return userStore.userInfo.id === props.comment.user_id
})

async function submitReply() {
  const content = replyContent.value.trim()
  if (!content || replySubmitting.value) return
  replySubmitting.value = true
  const contentToSend = content
  const parentComment = props.comment
  // Clear reply UI immediately — parent handles the API call asynchronously
  replyContent.value = ''
  showReplyBox.value = false
  replySubmitting.value = false
  emit('submit-reply', parentComment, contentToSend)
}

// --- Flat children (B站 style) ---
const showAllChildren = ref(false)

const visibleChildren = computed(() => {
  const children = props.comment.replies
  if (!children || children.length === 0) return []
  if (showAllChildren.value || children.length <= 3) {
    return children
  }
  return children.slice(0, 3)
})

// --- Child reply inline ---
const activeChildReplyId = ref<number | null>(null)
const childReplyContents = reactive<Record<number, string>>({})
const childReplySubmitting = reactive<Record<number, boolean>>({})

function canDeleteChild(child: CommentItem): boolean {
  if (!userStore.userInfo) return false
  return userStore.userInfo.id === child.user_id
}

function toggleChildReply(childId: number) {
  if (activeChildReplyId.value === childId) {
    activeChildReplyId.value = null
  } else {
    activeChildReplyId.value = childId
    if (!(childId in childReplyContents)) {
      childReplyContents[childId] = ''
    }
  }
}

function closeChildReply(childId: number) {
  activeChildReplyId.value = null
  delete childReplyContents[childId]
  delete childReplySubmitting[childId]
}

function submitChildReply(child: CommentItem) {
  const content = childReplyContents[child.id]?.trim()
  if (!content || childReplySubmitting[child.id]) return

  childReplySubmitting[child.id] = true
  const contentToSend = content
  const parentComment = child

  // Clear reply UI immediately
  closeChildReply(child.id)
  emit('submit-reply', parentComment, contentToSend)
}

// --- Utility ---
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
  if (days < 30) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}
</script>
