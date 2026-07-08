import { useToast } from './useToast'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
}

const API_BASE = ''

function getCSRFToken(): string {
  if (typeof document === 'undefined') return ''
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

async function fetchCSRFToken(): Promise<void> {
  try {
    await fetch('/api/v1/csrf-token', { credentials: 'include' })
  } catch {
    // Non-critical
  }
}

async function request<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const method = (options.method || 'GET').toUpperCase()
  const headers: Record<string, string> = { ...options.headers as Record<string, string> }

  // Set Content-Type for JSON bodies
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  } else {
    delete headers['Content-Type']
  }

  // Add CSRF token for write operations
  if (['POST', 'PUT', 'DELETE'].includes(method)) {
    const csrfToken = getCSRFToken()
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
  }

  const config: RequestInit = {
    ...options,
    credentials: 'include',
    headers,
  }

  try {
    const response = await fetch(`${API_BASE}${url}`, config)

    // If CSRF check fails (403), try refreshing the token once
    let result: ApiResponse<T> | null = null
    if (response.status === 403 && ['POST', 'PUT', 'DELETE'].includes(method)) {
      const body = await response.json()
      if (body.message && body.message.includes('CSRF')) {
        await fetchCSRFToken()
        const newToken = getCSRFToken()
        if (newToken) {
          headers['X-CSRF-Token'] = newToken
          const retryConfig = { ...config, headers }
          const retryResponse = await fetch(`${API_BASE}${url}`, retryConfig)
          const retryResult: ApiResponse<T> = await retryResponse.json()
          if (retryResult.code !== 200 && retryResult.code !== 201) {
            throw new Error(retryResult.message || '请求失败')
          }
          return retryResult.data
        }
      }
      // Save body for reuse — response.json() was already consumed
      result = body as ApiResponse<T>
    }

    // /api/v1/auth/* 的 401 是业务语义（密码错误等），不是会话过期，不触发全局拦截
    const isAuthRoute = url.startsWith('/api/v1/auth/')
    if (response.status === 401 && !isAuthRoute) {
      const { showToast } = useToast()
      showToast('请先登录', 'error')
      navigateTo('/login')
      throw new Error('Unauthorized')
    }

    if (response.status >= 500) {
      const { showToast } = useToast()
      showToast('服务器错误，请稍后重试', 'error')
      throw new Error('Server error')
    }

    // Use saved body if CSRF check already consumed response.json(), otherwise read now
    if (!result) {
      result = await response.json() as ApiResponse<T>
    }

    if (result.code !== 200 && result.code !== 201) {
      throw new Error(result.message || '请求失败')
    }

    return result.data
  } catch (err) {
    if (err instanceof Error && err.message === 'Unauthorized') {
      throw err
    }
    if (err instanceof Error && err.message === 'Server error') {
      throw err
    }
    throw err
  }
}

export function useApi() {
  const get = <T>(url: string, params?: Record<string, string | number>) => {
    let queryString = ''
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          searchParams.append(key, String(value))
        }
      })
      queryString = `?${searchParams.toString()}`
    }
    return request<T>(`${url}${queryString}`)
  }

  const post = <T>(url: string, body?: unknown) => {
    return request<T>(url, {
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
    })
  }

  const put = <T>(url: string, body?: unknown) => {
    return request<T>(url, {
      method: 'PUT',
      body: JSON.stringify(body),
    })
  }

  const del = <T>(url: string) => {
    return request<T>(url, {
      method: 'DELETE',
    })
  }

  return { get, post, put, del }
}

export { fetchCSRFToken, getCSRFToken }
