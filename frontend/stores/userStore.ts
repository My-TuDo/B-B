import { defineStore } from 'pinia'
import { useApi } from '~/composables/useApi'
import type { UserInfo, LoginResp } from '~/types'

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)
  const isLoggedIn = computed(() => !!userInfo.value)

  async function fetchUser(id: number) {
    const { get } = useApi()
    const data = await get<UserInfo>(`/api/v1/users/${id}`)
    userInfo.value = data
  }

  async function login(account: string, password: string) {
    const { post } = useApi()
    const data = await post<LoginResp>('/api/v1/auth/login', {
      account,
      password,
    })
    userInfo.value = {
      id: data.id,
      username: data.username,
      nickname: data.nickname,
      avatar: data.avatar,
      bio: '',
      created_at: '',
    }
  }

  async function register(username: string, email: string, password: string) {
    const { post } = useApi()
    const data = await post<LoginResp>('/api/v1/auth/register', {
      username,
      email,
      password,
    })
    userInfo.value = {
      id: data.id,
      username: data.username,
      nickname: data.nickname,
      avatar: data.avatar,
      bio: '',
      created_at: '',
    }
  }

  async function logout() {
    const { post } = useApi()
    try {
      await post('/api/v1/auth/logout')
    } catch {
      // Ignore errors on logout
    }
    userInfo.value = null
  }

  function clearUser() {
    userInfo.value = null
  }

  // Restore login state from cookie on app startup
  async function init() {
    try {
      const { get } = useApi()
      const data = await get<UserInfo>('/api/v1/auth/me')
      userInfo.value = data
    } catch {
      // Not logged in or token expired - stay logged out
      clearUser()
    }
  }

  return {
    userInfo,
    isLoggedIn,
    fetchUser,
    login,
    register,
    logout,
    clearUser,
    init,
  }
})
