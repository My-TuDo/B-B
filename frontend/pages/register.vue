<template>
  <div>
    <h2 class="text-2xl font-bold text-[var(--color-text)] mb-6">注册</h2>

    <form @submit.prevent="handleRegister" class="space-y-5">
      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">用户名</label>
        <input
          v-model="username"
          type="text"
          placeholder="3-20位字母、数字或下划线"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
        <p v-if="usernameError" class="mt-1.5 text-xs text-[var(--color-danger)]">{{ usernameError }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">邮箱</label>
        <input
          v-model="email"
          type="email"
          placeholder="请输入邮箱地址"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
        <p v-if="emailError" class="mt-1.5 text-xs text-[var(--color-danger)]">{{ emailError }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">密码</label>
        <input
          v-model="password"
          type="password"
          placeholder="至少8位，包含字母和数字"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
        <p v-if="passwordError" class="mt-1.5 text-xs text-[var(--color-danger)]">{{ passwordError }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-[var(--color-text)] mb-1.5">确认密码</label>
        <input
          v-model="confirmPassword"
          type="password"
          placeholder="请再次输入密码"
          class="w-full h-11 px-3.5 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-[var(--radius-md)] text-[var(--color-text)] text-sm placeholder-[var(--color-text-secondary)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-focus-ring)] transition-all duration-[var(--transition-normal)]"
          required
        />
        <p v-if="confirmError" class="mt-1.5 text-xs text-[var(--color-danger)]">{{ confirmError }}</p>
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full h-11 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-md)] hover:bg-[var(--color-primary-hover)] transition-all duration-[var(--transition-normal)] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm shadow-[var(--color-primary)]/25 active:scale-[0.98]"
      >
        {{ loading ? '注册中...' : '注册' }}
      </button>
    </form>

    <p class="mt-6 text-center text-sm text-[var(--color-text-secondary)]">
      已有账号？<NuxtLink to="/login" class="text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] font-medium">去登录</NuxtLink>
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

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)

const usernameError = ref('')
const emailError = ref('')
const passwordError = ref('')
const confirmError = ref('')

function validate(): boolean {
  usernameError.value = ''
  emailError.value = ''
  passwordError.value = ''
  confirmError.value = ''

  const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/
  if (!usernameRegex.test(username.value)) {
    usernameError.value = '用户名需为3-20位字母、数字或下划线'
    return false
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email.value)) {
    emailError.value = '请输入有效的邮箱地址'
    return false
  }

  if (password.value.length < 8) {
    passwordError.value = '密码至少8位'
    return false
  }
  const hasLetter = /[a-zA-Z]/.test(password.value)
  const hasDigit = /[0-9]/.test(password.value)
  if (!hasLetter || !hasDigit) {
    passwordError.value = '密码需包含字母和数字'
    return false
  }

  if (password.value !== confirmPassword.value) {
    confirmError.value = '两次输入的密码不一致'
    return false
  }

  return true
}

async function handleRegister() {
  if (!validate()) return

  loading.value = true
  try {
    await userStore.register(username.value, email.value, password.value)
    showToast('注册成功', 'success')
    router.push('/')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '注册失败'
    showToast(msg, 'error')
  } finally {
    loading.value = false
  }
}
</script>
