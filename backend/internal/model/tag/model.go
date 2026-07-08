package tag

import "time"

// ==================== Entity ====================

type Tag struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(30);uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type VideoTag struct {
	ID      uint `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID uint `gorm:"uniqueIndex:uk_video_tag;not null" json:"video_id"`
	TagID   uint `gorm:"uniqueIndex:uk_video_tag;not null" json:"tag_id"`
}

func (VideoTag) TableName() string {
	return "video_tags"
}

// ==================== DTOs ====================

type TagResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type SetVideoTagsReq struct {
	TagIDs []uint `json:"tag_ids" validate:"required"`
}

type CreateTagReq struct {
	Name string `json:"name" validate:"required,max=30"`
}
