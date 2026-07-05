import { useToast } from './useToast'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
}

const API_BASE = ''

async function request<T>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const config: RequestInit = {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  }

  // Don't set Content-Type for FormData
  if (options.body instanceof FormData) {
    const headers = { ...options.headers } as Record<string, string>
    delete headers['Content-Type']
    config.headers = headers
  }

  try {
    const response = await fetch(`${API_BASE}${url}`, config)

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

    const result: ApiResponse<T> = await response.json()

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
