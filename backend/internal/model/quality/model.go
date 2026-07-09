package quality

import "time"

// VideoQuality stores a specific resolution version of a video.
type VideoQuality struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID    uint      `gorm:"index;not null" json:"video_id"`
	Quality    string    `gorm:"type:varchar(10);not null" json:"quality"` // 360p / 480p / 720p / 1080p
	ObjectName string    `gorm:"type:varchar(500);not null" json:"object_name"`
	FileSize   uint64    `gorm:"default:0" json:"file_size"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
}

func (VideoQuality) TableName() string {
	return "video_qualities"
}
