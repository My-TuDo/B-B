// Package transcode 定义视频转码任务数据模型，追踪转码进度与状态。
package transcode

import "time"

// TranscodeTask 转码任务实体，映射 transcode_tasks 表。
// 记录单个视频的转码进度与状态。
type TranscodeTask struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`             // 任务主键
	VideoID   uint      `gorm:"uniqueIndex;not null" json:"video_id"`           // 关联视频 ID
	Status    int8      `gorm:"type:tinyint;default:0" json:"status"`           // 转码状态（0=等待, 1=处理中, 2=完成, 3=失败）
	Progress  uint8     `gorm:"type:tinyint unsigned;default:0" json:"progress"` // 转码进度（0-100）
	ErrorMsg  string    `gorm:"type:varchar(500);default:''" json:"error_msg"` // 失败时的错误信息
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                     // 创建时间
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`                     // 更新时间
}

// TableName 返回 transcode_tasks 表名。
func (TranscodeTask) TableName() string {
	return "transcode_tasks"
}

// 转码任务状态常量
const (
	StatusPending    int8 = 0 // 等待中
	StatusProcessing int8 = 1 // 处理中
	StatusDone       int8 = 2 // 已完成
	StatusFailed     int8 = 3 // 失败
)
