export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  created_at: string
}

export interface UserBrief {
  id: number
  username: string
  nickname: string
  avatar: string
}

export interface VideoInfo {
  id: number
  title: string
  description: string
  cover_url: string
  video_url: string
  duration: number
  file_size: number
  category_id: number
  status: number
  views: number
  created_at: string
  updated_at: string
  user?: UserBrief
}

export interface Category {
  id: number
  name: string
  slug: string
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
}

export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface LoginResp {
  id: number
  username: string
  nickname: string
  avatar: string
}
