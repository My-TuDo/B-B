<template>
  <div class="comment-section">
    <h3 class="text-base font-semibold text-[var(--color-text)] mb-4">
      评论 ({{ total }})
    </h3>

    <!-- Sort tabs -->
    <div class="flex gap-4 mb-4 border-b border-[var(--color-border)]">
      <button
        class="pb-2 text-sm font-medium transition-colors border-b-2"
        :class="sort === 'new' ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
        @click="sort = 'new'; fetchComments()"
      >
        最新
      </button>
      <button
        class="pb-2 text-sm font-medium transition-colors border-b-2"
        :class="sort === 'hot' ? 'text-[var(--color-primary)] border-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text)]'"
        @click="sort = 'hot'; fetchComments()"
      >
        最热
      </button>
    </div>

    <!-- Comment input -->
    <div v-if="isLoggedIn" class="flex gap-3 mb-6">
      <div class="w-9 h-9 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-xs font-bold flex-shrink-0 overflow-hidden">
        <img v-if="currentUser?.avatar" :src="currentUser.avatar" class="w-full h-full object-cover" />
        <span v-else>{{ currentUser?.nickname?.charAt(0) || 'U' }}</span>
      </div>
      <div class="flex-1">
        <textarea
          v-model="commentContent"
          class="w-full min-h-[60px] p-3 text-sm bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-[var(--radius-md)] resize-none focus:outline-none focus:border-[var(--color-primary)]/40 focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-colors"
          :placeholder="replyTarget ? `回复 @${replyTarget.nickname}` : '发一条友善的评论'"
          rows="2"
        />
        <div class="flex justify-between items-center mt-2">
          <button
            v-if="replyTarget"
            class="text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors"
            @click="cancelReply"
          >
            取消回复
          </button>
          <span v-else />
          <button
            class="px-4 py-1.5 text-sm font-medium bg-[var(--color-primary)] text-white rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            :disabled="!commentContent.trim() || submitting"
            @click="submitComment"
          >
            {{ submitting ? '发送中...' : '发表评论' }}
          </button>
        </div>
      </div>
    </div>
    <div v-else class="mb-6 p-4 bg-[var(--color-surface-hover)] rounded-[var(--radius-md)] text-center">
      <NuxtLink to="/login" class="text-sm text-[var(--color-primary)] hover:underline">登录</NuxtLink>
      <span class="text-sm text-[var(--color-text-secondary)]">后发表评论</span>
    </div>

    <!-- Comment list -->
    <div v-if="loading" class="py-8 text-center">
      <LoadingSpinner size="sm" />
    </div>

    <div v-else-if="comments.length === 0" class="py-8 text-center">
      <p class="text-sm text-[var(--color-text-secondary)]">暂无评论，来发第一条吧</p>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="comment in comments"
        :key="comment.id"
        class="comment-item"
      >
        <CommentNode
          :comment="comment"
          :level="0"
          @reply="handleReply"
          @like="handleLike"
          @delete="handleDelete"
          @submit-reply="handleSubmitReply"
          @load-replies="loadReplies"
        />
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="total > pageSize" class="flex justify-center mt-6 gap-2">
      <button
        v-for="p in Math.ceil(total / pageSize)"
        :key="p"
        class="w-8 h-8 text-sm rounded-full transition-colors"
        :class="p === page ? 'bg-[var(--color-primary)] text-white' : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-hover)]'"
        @click="page = p; fetchComments()"
      >
        {{ p }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CommentItem, CommentListResp, CommentLikeResp } from '~/types'
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'

const props = defineProps<{
  videoId: number
  videoAuthorId?: number
}>()

const { get, post, del } = useApi()
const userStore = useUserStore()

const comments = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const sort = ref<'new' | 'hot'>('new')
const loading = ref(false)
const submitting = ref(false)
const commentContent = ref('')
const replyTarget = ref<{ id: number; nickname: string; rootId: number } | null>(null)

const isLoggedIn = computed(() => userStore.isLoggedIn)
const currentUser = computed(() => userStore.userInfo)

async function fetchComments(silent = false) {
  if (!silent) loading.value = true
  try {
    const data = await get<CommentListResp>(`/api/v1/videos/${props.videoId}/comments`, {
      page: page.value,
      page_size: pageSize.value,
      sort: sort.value,
    })
    comments.value = data.items || []
    total.value = data.total || 0
  } catch {
    comments.value = []
    total.value = 0
  } finally {
    if (!silent) loading.value = false
  }
}

async function submitComment() {
  if (!commentContent.value.trim() || submitting.value) return
  submitting.value = true

  const content = replyTarget.value
    ? `@${replyTarget.value.nickname}: ${commentContent.value}`
    : commentContent.value
  const body: Record<string, unknown> = { content }
  if (replyTarget.value) {
    body.parent_id = replyTarget.value.id
    body.root_id = replyTarget.value.rootId
  }

  commentContent.value = ''
  const replyTo = replyTarget.value
  replyTarget.value = null

  try {
    await post(`/api/v1/videos/${props.videoId}/comments`, body)
    // Refresh to get actual server data (with correct ID, user, timestamps)
    fetchComments(true)
  } catch (err) {
    // Restore content on error so user doesn't lose their comment
    commentContent.value = content
    replyTarget.value = replyTo
  } finally {
    submitting.value = false
  }
}

function handleReply(comment: CommentItem) {
  replyTarget.value = {
    id: comment.id,
    nickname: comment.user?.nickname || comment.user?.username || '匿名',
    rootId: comment.root_id > 0 ? comment.root_id : comment.id,
  }
  // Scroll to textarea
  const textarea = document.querySelector('.comment-section textarea') as HTMLTextAreaElement
  if (textarea) textarea.focus()
}

function cancelReply() {
  replyTarget.value = null
}

async function handleSubmitReply(parentComment: CommentItem, content: string) {
  if (!content.trim()) return

  // Build the request: root_id = the root comment, parent_id = the comment being replied to
  const rootId = parentComment.root_id > 0 ? parentComment.root_id : parentComment.id
  const body: Record<string, unknown> = {
    content: `@${parentComment.user?.nickname || parentComment.user?.username || '匿名'}: ${content}`,
    parent_id: parentComment.id,
    root_id: rootId,
  }

  try {
    await post(`/api/v1/videos/${props.videoId}/comments`, body)
    // Refresh to get updated comment list including new nested reply
    fetchComments(true)
  } catch {
    // CommentNode already cleared its own UI — nothing extra needed on error
  }
}

async function handleLike(commentId: number) {
  // Optimistic update: find comment in local list and update count
  const updateLikes = (list: CommentItem[]) => {
    for (const c of list) {
      if (c.id === commentId) {
        c.likes += 1
        return true
      }
      if (c.replies) {
        if (updateLikes(c.replies)) return true
      }
    }
    return false
  }
  updateLikes(comments.value)

  try {
    const data = await post<CommentLikeResp>(`/api/v1/comments/${commentId}/like`)
    // Sync with server count
    const syncLikes = (list: CommentItem[]) => {
      for (const c of list) {
        if (c.id === commentId) {
          c.likes = data.likes
          return true
        }
        if (c.replies) {
          if (syncLikes(c.replies)) return true
        }
      }
      return false
    }
    syncLikes(comments.value)
  } catch {
    // Silent — revert optimistic update by decreasing
    const revertLikes = (list: CommentItem[]) => {
      for (const c of list) {
        if (c.id === commentId) {
          c.likes -= 1
          return true
        }
        if (c.replies) {
          if (revertLikes(c.replies)) return true
        }
      }
      return false
    }
    revertLikes(comments.value)
  }
}

async function handleDelete(commentId: number) {
  if (!confirm('确定要删除这条评论吗？')) return
  try {
    await del(`/api/v1/videos/${props.videoId}/comments/${commentId}`)
    fetchComments(true)
  } catch {
    // silent
  }
}

async function loadReplies(commentId: number) {
  // This is handled by the CommentNode component expanding children
  // Force refresh to show all replies
  fetchComments()
}

onMounted(fetchComments)
</script>
