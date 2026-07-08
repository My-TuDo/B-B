<template>
  <div>
    <h2 class="text-2xl font-bold text-[var(--color-text)] mb-6">登录</h2>

    <form @submit.prevent="handleLogin" class="space-y-5">
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">账号</label>
        <input
          v-model="account"
          type="text"
          placeholder="用户名或邮箱"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">密码</label>
        <input
          v-model="password"
          type="password"
          placeholder="请输入密码"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full h-11 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-md)] hover:bg-[var(--color-primary-hover)] transition-all duration-[var(--transition-normal)] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm shadow-[var(--color-primary)]/25 active:scale-[0.98]"
      >
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>

    <p class="mt-6 text-center text-sm text-[var(--color-text-secondary)]">
      没有账号？<NuxtLink to="/register" class="text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] font-medium">去注册</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '~/stores/userStore'
import { useToast } from '~/composables/useToast'

definePageMeta({
  layout: 'auth',
  middleware: 'guest',
})

const userStore = useUserStore()
const router = useRouter()
const { showToast } = useToast()

const account = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!account.value || !password.value) return

  loading.value = true
  try {
    await userStore.login(account.value, password.value)
    showToast('登录成功', 'success')
    router.push('/')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '登录失败'
    showToast(msg, 'error')
  } finally {
    loading.value = false
  }
}
</script>
