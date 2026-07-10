<template>
  <div class="max-w-7xl mx-auto">
    <!-- Page header -->
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-[var(--color-text)]">管理后台</h1>
      <p class="mt-1 text-sm text-[var(--color-text-secondary)]">系统运营数据概览与用户管理</p>
    </div>

    <!-- Tab navigation -->
    <div class="flex gap-1 mb-6 p-1 bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="px-4 py-2 text-sm font-medium rounded-lg transition-colors"
        :class="activeTab === tab.id ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)]' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text)]'"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Stats overview tab -->
    <template v-if="activeTab === 'overview'">
      <!-- Stats cards row -->
      <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div
          v-for="card in statCards"
          :key="card.label"
          class="p-5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl"
        >
          <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide">{{ card.label }}</p>
          <p class="mt-2 text-2xl font-bold text-[var(--color-text)]">{{ card.value }}</p>
          <p v-if="card.sub" class="mt-1 text-xs text-[var(--color-text-secondary)]">{{ card.sub }}</p>
        </div>
      </div>

      <!-- Today cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-10">
        <div class="p-5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl">
          <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide">今日新用户</p>
          <p class="mt-2 text-xl font-bold text-[var(--color-text)]">{{ stats?.today_new_users ?? '-' }}</p>
        </div>
        <div class="p-5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl">
          <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide">今日新视频</p>
          <p class="mt-2 text-xl font-bold text-[var(--color-text)]">{{ stats?.today_new_videos ?? '-' }}</p>
        </div>
      </div>

      <!-- Users section -->
      <div class="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
        <div class="px-6 py-4 border-b border-[var(--color-border)] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <h2 class="text-base font-semibold text-[var(--color-text)]">用户管理</h2>
        <div class="flex items-center gap-2">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索用户名或昵称..."
            class="h-9 px-3 text-sm bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-lg focus:outline-none focus:border-[var(--color-primary)]/40 w-56"
            @keyup.enter="loadUsers(1)"
          />
          <button
            class="h-9 px-4 text-sm font-medium bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors active:scale-[0.98]"
            @click="loadUsers(1)"
          >
            搜索
          </button>
        </div>
      </div>

      <LoadingSpinner v-if="loadingUsers" />

      <EmptyState
        v-else-if="!loadingUsers && users.length === 0"
        message="暂无用户数据"
      />

      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-[var(--color-border)]">
            <th class="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">用户</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider hidden sm:table-cell">角色</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider hidden md:table-cell">注册时间</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--color-border)]">
          <tr v-for="user in users" :key="user.id" class="hover:bg-[var(--color-surface-hover)]/50 transition-colors">
            <td class="px-6 py-3">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-xs font-bold overflow-hidden flex-shrink-0">
                  <img v-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" alt="" />
                  <span v-else>{{ (user.nickname || user.username || 'U').charAt(0) }}</span>
                </div>
                <div class="min-w-0">
                  <p class="text-sm font-medium text-[var(--color-text)] truncate">{{ user.nickname || user.username }}</p>
                  <p class="text-xs text-[var(--color-text-secondary)] truncate">@{{ user.username }}</p>
                </div>
              </div>
            </td>
            <td class="px-6 py-3 hidden sm:table-cell">
              <span
                class="inline-block px-2 py-0.5 text-xs font-medium rounded"
                :class="roleBadgeClass(user.role)"
              >
                {{ roleName(user.role) }}
              </span>
            </td>
            <td class="px-6 py-3 text-sm text-[var(--color-text-secondary)] hidden md:table-cell">
              {{ formatDate(user.created_at) }}
            </td>
            <td class="px-6 py-3 text-right">
              <select
                :value="user.role"
                class="h-8 px-2 text-xs bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded focus:outline-none focus:border-[var(--color-primary)]/40 cursor-pointer"
                @change="(e: Event) => handleRoleChange(user.id, Number((e.target as HTMLSelectElement).value))"
              >
                <option :value="1">普通用户</option>
                <option :value="2">UP主</option>
                <option :value="3">管理员</option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="usersTotal > pageSize" class="px-6 py-3 border-t border-[var(--color-border)] flex items-center justify-between">
        <span class="text-xs text-[var(--color-text-secondary)]">共 {{ usersTotal }} 个用户</span>
        <div class="flex items-center gap-1">
          <button
            class="w-8 h-8 flex items-center justify-center text-xs rounded hover:bg-[var(--color-surface-hover)] transition-colors disabled:opacity-30"
            :disabled="currentPage <= 1"
            @click="loadUsers(currentPage - 1)"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <span class="px-2 text-xs text-[var(--color-text-secondary)]">{{ currentPage }} / {{ Math.ceil(usersTotal / pageSize) }}</span>
          <button
            class="w-8 h-8 flex items-center justify-center text-xs rounded hover:bg-[var(--color-surface-hover)] transition-colors disabled:opacity-30"
            :disabled="currentPage >= Math.ceil(usersTotal / pageSize)"
            @click="loadUsers(currentPage + 1)"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
          </button>
        </div>
      </div>
    </div>
    </template>

    <!-- System tab -->
    <template v-if="activeTab === 'system'">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="p-5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl">
          <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-3">运行时</p>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-[var(--color-text-secondary)]">Go 版本</span>
              <span class="text-[var(--color-text)] font-mono">{{ sysInfo?.go_version || '-' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-[var(--color-text-secondary)]">运行时间</span>
              <span class="text-[var(--color-text)] font-mono">{{ sysInfo?.uptime || '-' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-[var(--color-text-secondary)]">数据库连接</span>
              <span
                class="font-medium"
                :class="sysInfo?.db_connected ? 'text-green-500' : 'text-red-500'"
              >{{ sysInfo?.db_connected ? '正常' : '断开' }}</span>
            </div>
          </div>
        </div>
        <div class="p-5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl">
          <p class="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide mb-3">构建信息</p>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-[var(--color-text-secondary)]">项目版本</span>
              <span class="text-[var(--color-text)]">B-B v1.0</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-[var(--color-text-secondary)]">前端框架</span>
              <span class="text-[var(--color-text)]">Nuxt 4 + Vue 3</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { AdminStats, AdminUserItem, AdminUsersListResp, SystemInfo } from '~/types'
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'

definePageMeta({
  middleware: 'auth',
})

const { get, put } = useApi()
const { showToast } = useToast()

const activeTab = ref('overview')
const tabs = [
  { id: 'overview', label: '概览' },
  { id: 'system', label: '系统' },
]
const sysInfo = ref<SystemInfo | null>(null)
const stats = ref<AdminStats | null>(null)
const users = ref<AdminUserItem[]>([])
const usersTotal = ref(0)
const currentPage = ref(1)
const pageSize = 20
const searchQuery = ref('')
const loadingUsers = ref(false)

const statCards = computed(() => {
  const s = stats.value
  if (!s) {
    return [
      { label: '总用户', value: '-' },
      { label: '总视频', value: '-' },
      { label: '总播放', value: '-' },
      { label: '总评论', value: '-' },
      { label: '总弹幕', value: '-' },
    ]
  }
  return [
    { label: '总用户', value: formatNum(s.total_users), sub: undefined },
    { label: '总视频', value: formatNum(s.total_videos), sub: undefined },
    { label: '总播放', value: formatNum(s.total_views), sub: undefined },
    { label: '总评论', value: formatNum(s.total_comments), sub: undefined },
    { label: '总弹幕', value: formatNum(s.total_danmaku), sub: undefined },
  ]
})

function formatNum(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  return String(n)
}

function roleName(role: number): string {
  switch (role) {
    case 3: return '管理员'
    case 2: return 'UP主'
    default: return '普通用户'
  }
}

function roleBadgeClass(role: number): string {
  switch (role) {
    case 3: return 'bg-purple-500/15 text-purple-400'
    case 2: return 'bg-blue-500/15 text-blue-400'
    default: return 'bg-[var(--color-surface-hover)] text-[var(--color-text-secondary)]'
  }
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN')
}

async function loadStats() {
  try {
    stats.value = await get<AdminStats>('/api/v1/admin/stats')
  } catch {
    // Silently fail - stats may not be available
  }
}

async function loadSystemInfo() {
  try {
    sysInfo.value = await get<SystemInfo>('/api/v1/admin/system')
  } catch {
    sysInfo.value = null
  }
}

async function loadUsers(page?: number) {
  loadingUsers.value = true
  if (page) currentPage.value = page

  try {
    const params: Record<string, string | number> = {
      page: currentPage.value,
      page_size: pageSize,
    }
    if (searchQuery.value.trim()) {
      params.q = searchQuery.value.trim()
    }
    const data = await get<AdminUsersListResp>('/api/v1/admin/users', params)
    users.value = data.items || []
    usersTotal.value = data.total || 0
  } catch {
    users.value = []
    usersTotal.value = 0
  } finally {
    loadingUsers.value = false
  }
}

async function handleRoleChange(userId: number, newRole: number) {
  try {
    await put(`/api/v1/admin/users/${userId}/role`, { role: newRole })
    showToast('角色更新成功', 'success')
    // Update local state
    const user = users.value.find(u => u.id === userId)
    if (user) user.role = newRole
  } catch (err: any) {
    showToast(err?.message || '角色更新失败', 'error')
    // Revert: reload from server
    loadUsers()
  }
}

onMounted(() => {
  loadStats()
  loadUsers()
  loadSystemInfo()
})
</script>
