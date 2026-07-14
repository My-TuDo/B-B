// Package message 定义站内消息/通知数据模型，包含消息实体与响应结构。
package message

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// Message 消息实体，映射 messages 表，支持系统通知与用户间消息。
type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`           // 消息主键
	UserID     uint      `gorm:"index;not null" json:"user_id"`                // 接收用户 ID
	FromUserID uint      `gorm:"not null" json:"from_user_id"`                 // 发送用户 ID
	Type       int8      `gorm:"type:tinyint;not null" json:"type"`           // 消息类型（1 赞 / 2 评论 / 3 关注 / 4 系统通知等）
	TargetID   uint      `gorm:"default:0" json:"target_id"`                  // 关联目标 ID（视频、评论等）
	Content    string    `gorm:"type:varchar(500)" json:"content"`            // 消息内容
	IsRead     int8      `gorm:"type:tinyint;default:0" json:"is_read"`      // 是否已读（1 已读 / 0 未读）
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`                  // 消息创建时间

	FromUser usermodel.User `gorm:"foreignKey:FromUserID" json:"-"`           // 关联发送者（不序列化）
}

// TableName 返回 messages 表名。
func (Message) TableName() string {
	return "messages"
}

// MessageResp 消息响应体，含发送者简要信息。
type MessageResp struct {
	ID        uint                `json:"id"`                 // 消息主键
	Type      int8                `json:"type"`               // 消息类型
	Content   string              `json:"content"`            // 消息内容
	TargetID  uint                `json:"target_id"`          // 关联目标 ID
	IsRead    int8                `json:"is_read"`            // 是否已读
	CreatedAt time.Time           `json:"created_at"`         // 创建时间
	FromUser  *usermodel.UserBrief `json:"from_user,omitempty"` // 发送者简要信息
}

// NotificationListResp 通知列表响应体。
type NotificationListResp struct {
	Items    []MessageResp `json:"items"`     // 通知列表
	Total    int64         `json:"total"`     // 通知总数
	Page     int           `json:"page"`      // 当前页码
	PageSize int           `json:"page_size"` // 每页数量
	Unread   int64         `json:"unread"`    // 未读通知数
}
