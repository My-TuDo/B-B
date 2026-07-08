import { fetchCSRFToken } from '~/composables/useApi'

export default defineNuxtPlugin(async () => {
  const userStore = useUserStore()

  // Race init against a timeout to prevent the app from blocking indefinitely
  // (a slow / unreachable backend would otherwise show a permanent black screen)
  await Promise.race([
    userStore.init(),
    new Promise<void>((resolve) => setTimeout(resolve, 5000)),
  ])

  // Fetch CSRF token for write operations (client-side only)
  if (typeof window !== 'undefined') {
    try {
      await fetchCSRFToken()
    } catch {
      // Non-critical
    }
  }
})
