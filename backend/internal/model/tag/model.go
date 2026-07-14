// Package tag 定义视频标签数据模型，包含标签、视频-标签关联实体与请求/响应结构。
package tag

import "time"

// ==================== Entity ====================

// Tag 标签实体，映射 tags 表。
type Tag struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`           // 标签主键
	Name      string    `gorm:"type:varchar(30);uniqueIndex;not null" json:"name"` // 标签名称
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                   // 创建时间
}

// TableName 返回 tags 表名。
func (Tag) TableName() string {
	return "tags"
}

// VideoTag 视频-标签关联实体，映射 video_tags 表。
type VideoTag struct {
	ID      uint `gorm:"primaryKey;autoIncrement" json:"id"`                       // 记录主键
	VideoID uint `gorm:"uniqueIndex:uk_video_tag;not null" json:"video_id"`       // 视频 ID
	TagID   uint `gorm:"uniqueIndex:uk_video_tag;not null" json:"tag_id"`         // 标签 ID
}

// TableName 返回 video_tags 表名。
func (VideoTag) TableName() string {
	return "video_tags"
}

// ==================== DTOs ====================

// TagResp 标签响应体。
type TagResp struct {
	ID   uint   `json:"id"`   // 标签主键
	Name string `json:"name"` // 标签名称
}

// SetVideoTagsReq 设置视频标签请求体。
type SetVideoTagsReq struct {
	TagIDs []uint `json:"tag_ids" validate:"required"` // 标签 ID 列表
}

// CreateTagReq 创建标签请求体。
type CreateTagReq struct {
	Name string `json:"name" validate:"required,max=30"` // 标签名称
}
