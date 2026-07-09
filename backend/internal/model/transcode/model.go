package transcode

import "time"

// TranscodeTask represents a single video transcode job.
type TranscodeTask struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID   uint      `gorm:"uniqueIndex;not null" json:"video_id"`
	Status    int8      `gorm:"type:tinyint;default:0" json:"status"` // 0=等待, 1=处理中, 2=完成, 3=失败
	Progress  uint8     `gorm:"type:tinyint unsigned;default:0" json:"progress"`
	ErrorMsg  string    `gorm:"type:varchar(500);default:''" json:"error_msg"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (TranscodeTask) TableName() string {
	return "transcode_tasks"
}

// Task status constants
const (
	StatusPending    int8 = 0
	StatusProcessing int8 = 1
	StatusDone       int8 = 2
	StatusFailed     int8 = 3
)
