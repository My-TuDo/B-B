// Package history 定义视频观看历史与搜索相关数据模型。
package history

import (
	"time"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
)

// ==================== Entity ====================

// VideoHistory 观看历史实体，映射 video_histories 表。
type VideoHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`                                                    // 记录主键
	UserID    uint      `gorm:"uniqueIndex:uk_user_video;index:idx_user_watched,priority:1;not null" json:"user_id"`  // 用户 ID
	VideoID   uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"video_id"`                                   // 视频 ID
	Progress  uint      `gorm:"default:0" json:"progress"`                                                            // 观看进度（秒）
	WatchedAt time.Time `gorm:"index:idx_user_watched,priority:2;not null" json:"watched_at"`                         // 最近观看时间

	Video videomodel.Video `gorm:"foreignKey:VideoID" json:"-"`                                                       // 关联视频（不序列化）
}

// TableName 返回 video_histories 表名。
func (VideoHistory) TableName() string {
	return "video_histories"
}

// ==================== DTOs ====================

// CreateHistoryReq 创建/更新观看历史请求体。
type CreateHistoryReq struct {
	VideoID  uint    `json:"video_id" validate:"required"` // 视频 ID
	Progress uint    `json:"progress"`                      // 观看进度（秒）
	Duration float64 `json:"duration,omitempty"`            // 视频总时长（秒）
}

// HistoryItemResp 观看历史项响应体。
type HistoryItemResp struct {
	Video     videomodel.VideoResp `json:"video"`      // 视频信息
	Progress  uint                  `json:"progress"`   // 观看进度（秒）
	WatchedAt time.Time             `json:"watched_at"` // 最近观看时间
}

// HistoryListResp 观看历史列表响应体。
type HistoryListResp struct {
	Items    []HistoryItemResp `json:"items"`     // 历史记录列表
	Total    int64             `json:"total"`     // 历史记录总数
	Page     int               `json:"page"`      // 当前页码
	PageSize int               `json:"page_size"` // 每页数量
}

// SearchSuggestionResp 搜索建议响应体。
type SearchSuggestionResp struct {
	Keyword string `json:"keyword"` // 搜索关键词
	Count   int64  `json:"count"`   // 关联数量
}

// SearchListResp 搜索结果列表响应体。
type SearchListResp struct {
	Items    []videomodel.VideoResp `json:"items"`     // 搜索结果视频列表
	Total    int64                  `json:"total"`     // 搜索结果总数
	Page     int                    `json:"page"`      // 当前页码
	PageSize int                    `json:"page_size"` // 每页数量
}

// CreatorStatsResp 创作者统计数据响应体。
type CreatorStatsResp struct {
	TotalViews   uint64 `json:"total_views"`   // 累计播放量
	TotalVideos  int64  `json:"total_videos"`  // 视频总数
	TodayViews   uint64 `json:"today_views"`   // 今日播放量
	TodayNewFans int64  `json:"today_new_fans"` // 今日新增粉丝
}

// RankingListResp 排行榜响应体。
type RankingListResp struct {
	Items    []videomodel.VideoResp `json:"items"`     // 排行榜视频列表
	Total    int64                  `json:"total"`     // 总数
	Page     int                    `json:"page"`      // 当前页码
	PageSize int                    `json:"page_size"` // 每页数量
	Period   string                 `json:"period"`    // 排行周期标识
}

// AdminReviewReq 管理员审核请求体。
type AdminReviewReq struct {
	Status int8 `json:"status" validate:"required,oneof=1 3"` // 审核状态（1 通过 / 3 拒绝）
}
