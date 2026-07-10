export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  created_at: string
  role?: number
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

export interface Tag {
  id: number
  name: string
}

export interface HistoryItem {
  video: VideoInfo
  progress: number
  watched_at: string
}

export interface HistoryListResp {
  items: HistoryItem[]
  total: number
  page: number
  page_size: number
}

export interface SearchSuggestion {
  keyword: string
  count: number
}

export interface CreatorStats {
  total_views: number
  total_videos: number
  today_views: number
  today_new_fans: number
}

export interface AdminVideoItem {
  id: number
  title: string
  status: number
  created_at: string
  user?: UserBrief
}

export interface RankingData {
  items: VideoInfo[]
  total: number
  page: number
  page_size: number
  period: string
}

export interface DanmakuItem {
  id: number
  content: string
  color: string
  position: number
  size: number
  play_time: number
  user?: UserBrief
}

export interface CommentItem {
  id: number
  video_id: number
  user_id: number
  parent_id: number
  root_id: number
  content: string
  likes: number
  created_at: string
  user?: UserBrief
  replies?: CommentItem[]
}

export interface CommentListResp {
  items: CommentItem[]
  total: number
  page: number
  page_size: number
}

export interface CommentLikeResp {
  liked: boolean
  likes: number
}

export interface LikeResp {
  liked: boolean
  count: number
}

export interface CoinResp {
  coins_today: number
}

export interface FavoriteInfo {
  id: number
  name: string
  is_public: number
  item_count: number
  cover_url?: string
}

export interface FavoriteDetailResp {
  favorite: FavoriteInfo
  items: VideoInfo[]
  total: number
  page: number
  page_size: number
}

export interface FavoriteToggleResp {
  favorited: boolean
}

export interface FollowResp {
  following: boolean
}

export interface ProfileStats {
  videos: number
  followers: number
  following: number
}

export interface ProfileResp {
  user: UserInfo
  stats: ProfileStats
}

export interface NotificationItem {
  id: number
  type: number
  content: string
  target_id: number
  is_read: number
  created_at: string
  from_user?: UserBrief
}

export interface NotificationListResp {
  items: NotificationItem[]
  total: number
  page: number
  page_size: number
  unread: number
}

export interface InteractionStatus {
  liked: boolean
  coins: number
  favorited: boolean
  following?: boolean
}

// Phase 4: Media Processing
export interface AdminStats {
  total_users: number
  total_videos: number
  total_views: number
  total_comments: number
  total_danmaku: number
  today_new_users: number
  today_new_videos: number
}

export interface AdminUserItem {
  id: number
  username: string
  nickname: string
  avatar: string
  role: number
  created_at: string
}

export interface AdminUsersListResp {
  items: AdminUserItem[]
  total: number
  page: number
  page_size: number
}

export interface SystemInfo {
  go_version: string
  uptime: string
  db_connected: boolean
}

export interface TranscodeStatus {
  status: number
  progress: number
}

export interface VideoQuality {
  quality: string
  play_url: string
  file_size: number
}
