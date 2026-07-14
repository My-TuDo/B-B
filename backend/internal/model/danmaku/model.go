// Package danmaku 定义弹幕数据模型，包含弹幕实体与请求/响应结构。
package danmaku

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// Danmaku 弹幕实体，映射 danmaku 表。
type Danmaku struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`             // 弹幕主键
	VideoID   uint      `gorm:"index;not null" json:"video_id"`                 // 所属视频 ID
	UserID    uint      `gorm:"not null" json:"user_id"`                        // 发送用户 ID
	Content   string    `gorm:"type:varchar(200);not null" json:"content"`     // 弹幕文本内容
	Color     string    `gorm:"type:varchar(7);default:#ffffff" json:"color"`  // 弹幕颜色（十六进制）
	Position  int8      `gorm:"type:tinyint;default:0" json:"position"`        // 弹幕位置（0 滚动 / 1 顶部 / 2 底部）
	Size      int8      `gorm:"type:tinyint;default:1" json:"size"`            // 弹幕字号（0 小 / 1 中 / 2 大）
	PlayTime  uint      `gorm:"default:0" json:"play_time"`                    // 弹幕出现的播放时间（秒）
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                    // 发送时间

	User usermodel.User `gorm:"foreignKey:UserID" json:"-"`                    // 关联用户（不序列化）
}

// TableName 返回 danmaku 表名。
func (Danmaku) TableName() string {
	return "danmaku"
}

// DanmakuReq 发送弹幕请求体。
type DanmakuReq struct {
	Content  string `json:"content" validate:"required,max=200"`              // 弹幕内容
	Color    string `json:"color" validate:"omitempty,max=7"`                 // 弹幕颜色
	Position *int8  `json:"position" validate:"omitempty,oneof=0 1 2"`       // 弹幕位置
	Size     *int8  `json:"size" validate:"omitempty,oneof=0 1 2"`           // 弹幕字号
	PlayTime uint   `json:"play_time" validate:"omitempty"`                   // 出现时间（秒）
}

// DanmakuResp 弹幕响应体，含发送者简要信息。
type DanmakuResp struct {
	ID       uint                `json:"id"`                 // 弹幕主键
	Content  string              `json:"content"`            // 弹幕内容
	Color    string              `json:"color"`              // 弹幕颜色
	Position int8                `json:"position"`           // 弹幕位置
	Size     int8                `json:"size"`               // 弹幕字号
	PlayTime uint                `json:"play_time"`          // 出现时间（秒）
	User     *usermodel.UserBrief `json:"user,omitempty"`    // 发送者简要信息
}
