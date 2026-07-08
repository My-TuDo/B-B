package message

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	FromUserID uint      `gorm:"not null" json:"from_user_id"`
	Type       int8      `gorm:"type:tinyint;not null" json:"type"`
	TargetID   uint      `gorm:"default:0" json:"target_id"`
	Content    string    `gorm:"type:varchar(500)" json:"content"`
	IsRead     int8      `gorm:"type:tinyint;default:0" json:"is_read"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`

	FromUser usermodel.User `gorm:"foreignKey:FromUserID" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageResp struct {
	ID        uint                `json:"id"`
	Type      int8                `json:"type"`
	Content   string              `json:"content"`
	TargetID  uint                `json:"target_id"`
	IsRead    int8                `json:"is_read"`
	CreatedAt time.Time           `json:"created_at"`
	FromUser  *usermodel.UserBrief `json:"from_user,omitempty"`
}

type NotificationListResp struct {
	Items    []MessageResp `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Unread   int64         `json:"unread"`
}
