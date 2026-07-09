package meta

import "time"

// VideoMeta holds ffprobe-extracted metadata about a video.
type VideoMeta struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID   uint      `gorm:"uniqueIndex;not null" json:"video_id"`
	Duration  float64   `gorm:"not null" json:"duration"`
	Width     uint      `gorm:"default:0" json:"width"`
	Height    uint      `gorm:"default:0" json:"height"`
	Codec     string    `gorm:"type:varchar(50);default:''" json:"codec"`
	Bitrate   uint      `gorm:"default:0" json:"bitrate"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (VideoMeta) TableName() string {
	return "video_metas"
}
