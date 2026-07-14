// Package quality 定义视频清晰度数据模型，存储同一视频的不同画质版本。
package quality

import "time"

// VideoQuality 视频清晰度实体，映射 video_qualities 表。
// 存储同一视频的多种分辨率版本。
type VideoQuality struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`               // 记录主键
	VideoID    uint      `gorm:"index;not null" json:"video_id"`                   // 关联视频 ID
	Quality    string    `gorm:"type:varchar(10);not null" json:"quality"`         // 清晰度标识（360p / 480p / 720p / 1080p）
	ObjectName string    `gorm:"type:varchar(500);not null" json:"object_name"`   // OSS 对象名（存储路径）
	FileSize   uint64    `gorm:"default:0" json:"file_size"`                       // 文件大小（字节）
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`                       // 创建时间
}

// TableName 返回 video_qualities 表名。
func (VideoQuality) TableName() string {
	return "video_qualities"
}
