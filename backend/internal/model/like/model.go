// Package like 定义视频点赞数据模型，包含点赞实体与响应结构。
package like

import "time"

// VideoLike 视频点赞实体，映射 video_likes 表。
type VideoLike struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`                      // 记录主键
	UserID    uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"user_id"`       // 点赞用户 ID
	VideoID   uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"video_id"`      // 被点赞视频 ID
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                              // 点赞时间
}

// TableName 返回 video_likes 表名。
func (VideoLike) TableName() string {
	return "video_likes"
}

// LikeResp 点赞操作响应体。
type LikeResp struct {
	Liked bool `json:"liked"` // 当前用户是否已点赞
	Count uint `json:"count"` // 视频最新点赞数
}
