package history

import (
	"time"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
)

// ==================== Entity ====================

type VideoHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:uk_user_video;index:idx_user_watched,priority:1;not null" json:"user_id"`
	VideoID   uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"video_id"`
	Progress  uint      `gorm:"default:0" json:"progress"`
	WatchedAt time.Time `gorm:"index:idx_user_watched,priority:2;not null" json:"watched_at"`

	Video videomodel.Video `gorm:"foreignKey:VideoID" json:"-"`
}

func (VideoHistory) TableName() string {
	return "video_histories"
}

// ==================== DTOs ====================

type CreateHistoryReq struct {
	VideoID  uint    `json:"video_id" validate:"required"`
	Progress uint    `json:"progress"`
	Duration float64 `json:"duration,omitempty"`
}

type HistoryItemResp struct {
	Video     videomodel.VideoResp `json:"video"`
	Progress  uint                  `json:"progress"`
	WatchedAt time.Time             `json:"watched_at"`
}

type HistoryListResp struct {
	Items    []HistoryItemResp `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type SearchSuggestionResp struct {
	Keyword string `json:"keyword"`
	Count   int64  `json:"count"`
}

type SearchListResp struct {
	Items    []videomodel.VideoResp `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type CreatorStatsResp struct {
	TotalViews   uint64 `json:"total_views"`
	TotalVideos  int64  `json:"total_videos"`
	TodayViews   uint64 `json:"today_views"`
	TodayNewFans int64  `json:"today_new_fans"`
}

type RankingListResp struct {
	Items    []videomodel.VideoResp `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Period   string                 `json:"period"`
}

type AdminReviewReq struct {
	Status int8 `json:"status" validate:"required,oneof=1 3"`
}
