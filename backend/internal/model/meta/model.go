// Package meta 定义视频元数据模型，存储通过 ffprobe 提取的媒体信息。
package meta

import "time"

// VideoMeta 视频元数据实体，映射 video_metas 表。
// 存储通过 ffprobe 提取的视频技术参数。
type VideoMeta struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`           // 记录主键
	VideoID   uint      `gorm:"uniqueIndex;not null" json:"video_id"`         // 关联视频 ID
	Duration  float64   `gorm:"not null" json:"duration"`                     // 视频时长（秒）
	Width     uint      `gorm:"default:0" json:"width"`                       // 视频宽度（像素）
	Height    uint      `gorm:"default:0" json:"height"`                      // 视频高度（像素）
	Codec     string    `gorm:"type:varchar(50);default:''" json:"codec"`    // 编码格式（如 h264、hevc）
	Bitrate   uint      `gorm:"default:0" json:"bitrate"`                     // 码率（bps）
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                   // 创建时间
}

// TableName 返回 video_metas 表名。
func (VideoMeta) TableName() string {
	return "video_metas"
}
