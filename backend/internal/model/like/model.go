package like

import "time"

type VideoLike struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"user_id"`
	VideoID   uint      `gorm:"uniqueIndex:uk_user_video;not null" json:"video_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (VideoLike) TableName() string {
	return "video_likes"
}

type LikeResp struct {
	Liked bool `json:"liked"`
	Count uint `json:"count"`
}
