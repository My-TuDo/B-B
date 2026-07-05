<template>
  <header class="fixed top-0 left-0 right-0 z-50 bg-[var(--color-surface)] border-b border-[var(--color-border)]" style="height: var(--header-height)">
    <div class="h-full max-w-[1920px] mx-auto px-4 flex items-center justify-between">
      <!-- Left: Logo -->
      <NuxtLink to="/" class="text-2xl font-bold text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] transition-colors">
        B-B
      </NuxtLink>

      <!-- Center: Search placeholder -->
      <div class="hidden md:flex flex-1 max-w-md mx-4">
        <div class="w-full h-9 bg-[var(--color-bg)] rounded-[var(--radius-full)] border border-[var(--color-border)] flex items-center px-4">
          <span class="text-sm text-[var(--color-text-secondary)]">搜索功能即将上线</span>
        </div>
      </div>

      <!-- Right: actions -->
      <div class="flex items-center gap-1">
        <!-- Theme toggle -->
        <button
          class="w-9 h-9 flex items-center justify-center rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)] transition-colors"
          :title="isDark ? '切换亮色模式' : '切换暗色模式'"
          @click="toggleTheme"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text)]">
            <circle v-if="isDark" cx="12" cy="12" r="5" />
            <path v-if="isDark" d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
            <path v-else d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
        </button>

        <template v-if="userStore.isLoggedIn && userStore.userInfo">
          <!-- Notifications -->
          <button
            class="relative w-9 h-9 flex items-center justify-center rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)] transition-colors opacity-50 cursor-not-allowed"
            title="通知（即将上线）"
            disabled
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text)]">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
          </button>

          <!-- User dropdown -->
          <div class="relative" @mouseenter="handleDropdownEnter" @mouseleave="handleDropdownLeave">
            <button class="flex items-center gap-2 h-9 px-2 rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)] transition-colors">
              <div class="w-7 h-7 rounded-full bg-[var(--color-primary)] flex items-center justify-center text-white text-xs font-bold">
                {{ userStore.userInfo.nickname?.charAt(0) || userStore.userInfo.username?.charAt(0) || 'U' }}
              </div>
            </button>

            <div
              v-if="showDropdown"
              class="absolute right-0 top-full mt-1 w-44 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-md)] shadow-lg py-1"
            >
              <div class="px-4 py-2 border-b border-[var(--color-border)]">
                <p class="text-sm text-[var(--color-text)] font-medium truncate">{{ userStore.userInfo.nickname || userStore.userInfo.username }}</p>
                <p class="text-xs text-[var(--color-text-secondary)] truncate">@{{ userStore.userInfo.username }}</p>
              </div>
              <button
                class="w-full text-left px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors opacity-50 cursor-not-allowed"
                disabled
                title="即将上线"
              >
                设置
              </button>
              <button
                class="w-full text-left px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] transition-colors opacity-50 cursor-not-allowed"
                disabled
                title="即将上线"
              >
                切换账号
              </button>
              <div class="border-t border-[var(--color-border)] my-1"></div>
              <button
                class="w-full text-left px-4 py-2 text-sm text-[var(--color-danger)] hover:bg-[var(--color-surface-hover)] transition-colors"
                @click="handleLogout"
              >
                退出登录
              </button>
            </div>
          </div>
        </template>

        <template v-else>
          <NuxtLink
            to="/login"
            class="h-9 px-4 bg-[var(--color-primary)] text-white text-sm rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors flex items-center justify-center"
          >
            登录
          </NuxtLink>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'
import { useTheme } from '~/composables/useTheme'

const userStore = useUserStore()
const router = useRouter()
const showDropdown = ref(false)
const { showToast } = useToast()
const { currentTheme, toggleTheme } = useTheme()

const isDark = computed(() => currentTheme.value === 'dark')

let closeTimer: ReturnType<typeof setTimeout> | null = null

function handleDropdownEnter() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
  showDropdown.value = true
}

function handleDropdownLeave() {
  closeTimer = setTimeout(() => {
    showDropdown.value = false
  }, 150)
}

async function handleLogout() {
  await userStore.logout()
  showDropdown.value = false
  showToast('已退出登录', 'info')
  router.push('/')
}

onUnmounted(() => {
  if (closeTimer) clearTimeout(closeTimer)
})
</script>
