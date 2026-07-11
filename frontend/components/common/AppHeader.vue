<template>
  <header class="fixed top-0 left-0 right-0 z-50 bg-[var(--color-surface)]/80 backdrop-blur-xl border-b border-[var(--color-border)]/50" style="height: var(--header-height)">
    <div class="h-full max-w-[1920px] mx-auto px-4 flex items-center justify-between">
      <!-- Left: Logo -->
      <NuxtLink to="/" class="inline-block min-w-[4rem] text-left text-2xl font-bold text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] transition-colors transition-opacity duration-200" :style="{ opacity: visible ? 1 : 0 }">
        {{ displayText }}
      </NuxtLink>

      <!-- Center: Search -->
      <div class="hidden md:flex flex-1 max-w-md mx-4 relative">
        <form class="w-full" @submit.prevent="doSearch">
          <div class="w-full h-9 bg-[var(--color-surface-hover)] rounded-[var(--radius-full)] border border-[var(--color-border)] flex items-center px-4 transition-colors duration-[var(--transition-normal)] focus-within:border-[var(--color-primary)]/40 focus-within:ring-2 focus-within:ring-[var(--color-focus-ring)]">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索视频..."
              class="flex-1 bg-transparent text-sm text-[var(--color-text)] placeholder-[var(--color-text-secondary)] focus:outline-none"
              @input="onSearchInput"
              @focus="onSearchFocus"
              @blur="onSearchBlur"
            />
            <button type="submit" class="ml-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            </button>
          </div>
        </form>
        <!-- Suggestions dropdown -->
        <div
          v-if="showSuggestions && suggestions.length > 0"
          class="absolute top-full left-0 right-0 mt-1 bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-md)] shadow-lg z-10"
        >
          <div
            v-for="item in suggestions"
            :key="item.keyword"
            class="px-4 py-2 text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-hover)] cursor-pointer"
            @mousedown.prevent="selectSuggestion(item.keyword)"
          >
            {{ item.keyword }}
          </div>
          <div class="px-4 py-1.5 text-xs text-[var(--color-text-secondary)] border-t border-[var(--color-border)]">
            搜索建议
          </div>
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
          <NuxtLink
            class="relative w-9 h-9 flex items-center justify-center rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)] transition-colors"
            to="/notifications"
            title="消息通知"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-[var(--color-text)]">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            <span
              v-if="userStore.unreadCount > 0"
              class="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 flex items-center justify-center bg-[var(--color-danger)] text-white text-[10px] font-bold rounded-full leading-none"
            >
              {{ userStore.unreadCount > 99 ? '99+' : userStore.unreadCount }}
            </span>
          </NuxtLink>

          <!-- User dropdown -->
          <div class="relative" @mouseenter="handleDropdownEnter" @mouseleave="handleDropdownLeave">
            <button class="flex items-center gap-2 h-9 px-2 rounded-[var(--radius-full)] hover:bg-[var(--color-surface-hover)] transition-colors">
              <div class="w-7 h-7 rounded-full bg-white flex items-center justify-center text-[var(--color-primary)] text-xs font-bold overflow-hidden">
                <img v-if="userStore.userInfo.avatar" :src="userStore.userInfo.avatar" class="w-full h-full object-cover" />
                <span v-else>{{ userStore.userInfo.nickname?.charAt(0) || userStore.userInfo.username?.charAt(0) || 'U' }}</span>
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
            class="h-9 px-5 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-all duration-[var(--transition-normal)] flex items-center justify-center shadow-sm shadow-[var(--color-primary)]/25 active:scale-[0.98]"
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

interface Suggestion {
  keyword: string
  count: number
}

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()
const showDropdown = ref(false)
const searchQuery = ref('')
const suggestions = ref<Suggestion[]>([])
const showSuggestions = ref(false)
const { showToast } = useToast()
const { currentTheme, toggleTheme } = useTheme()

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let blurTimer: ReturnType<typeof setTimeout> | null = null

// ── Logo animation ──
const variants = ['B-B', 'BvB', 'BoB', 'B_B', 'B^B']
const currentIndex = ref(0)
const visible = ref(true)
const displayText = computed(() => variants[currentIndex.value])

let logoTimer: ReturnType<typeof setTimeout> | null = null

function cycleLogo() {
  // fade out
  visible.value = false
  logoTimer = setTimeout(() => {
    currentIndex.value = (currentIndex.value + 1) % variants.length
    visible.value = true // fade in
    logoTimer = setTimeout(() => {
      cycleLogo() // stay then next cycle
    }, 1700) // 1.5s visible + 200ms buffer
  }, 200) // fade-out duration
}

function doSearch() {
  const q = searchQuery.value.trim()
  searchQuery.value = ''
  if (!q) {
    router.push('/search')
    return
  }
  if (route.path === '/search') {
    router.replace({ query: { q } })
  } else {
    router.push(`/search?q=${encodeURIComponent(q)}`)
  }
}

async function fetchSuggestions(q: string) {
  if (!q) {
    suggestions.value = []
    showSuggestions.value = false
    return
  }
  try {
    const data = await get<Suggestion[]>('/api/v1/search/suggestions', { q, limit: 8 })
    suggestions.value = data || []
    showSuggestions.value = suggestions.value.length > 0
  } catch {
    suggestions.value = []
    showSuggestions.value = false
  }
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchSuggestions(searchQuery.value.trim())
  }, 300)
}

function onSearchFocus() {
  if (suggestions.value.length > 0) {
    showSuggestions.value = true
  }
}

function onSearchBlur() {
  blurTimer = setTimeout(() => {
    showSuggestions.value = false
  }, 150)
}

function selectSuggestion(keyword: string) {
  if (blurTimer) clearTimeout(blurTimer)
  searchQuery.value = ''
  suggestions.value = []
  showSuggestions.value = false
  if (route.path === '/search') {
    router.replace({ query: { q: keyword } })
  } else {
    router.push(`/search?q=${encodeURIComponent(keyword)}`)
  }
}

const isDark = computed(() => currentTheme.value === 'dark')

// ── Notification polling (30s interval, lightweight: only fetches unread count) ──
let pollTimer: ReturnType<typeof setInterval> | null = null

function startPolling() {
  stopPolling()
  userStore.fetchUnread()
  pollTimer = setInterval(() => {
    userStore.fetchUnread()
  }, 30_000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

// Debounce to avoid spawning multiple pollers when isLoggedIn flickers
let pollDebounce: ReturnType<typeof setTimeout> | null = null
watch(() => userStore.isLoggedIn, (loggedIn) => {
  if (pollDebounce) clearTimeout(pollDebounce)
  pollDebounce = setTimeout(() => {
    if (loggedIn) {
      startPolling()
    } else {
      stopPolling()
      userStore.clearUnread()
    }
  }, 200)
})

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

onMounted(() => {
  // Start logo cycle after 1.5s initial display
  logoTimer = setTimeout(cycleLogo, 1500)
  // Start notification polling if already logged in
  if (userStore.isLoggedIn) {
    startPolling()
  }
})

onUnmounted(() => {
  if (closeTimer) clearTimeout(closeTimer)
  if (logoTimer) clearTimeout(logoTimer)
  if (debounceTimer) clearTimeout(debounceTimer)
  if (blurTimer) clearTimeout(blurTimer)
  stopPolling()
})
</script>
