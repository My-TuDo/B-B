// Package coin 定义视频投币记录模型，包含投币实体与请求/响应结构。
package coin

import "time"

// VideoCoin 视频投币记录实体，映射 video_coins 表。
type VideoCoin struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`          // 记录主键
	UserID    uint      `gorm:"index;not null" json:"user_id"`               // 投币用户 ID
	VideoID   uint      `gorm:"index;not null" json:"video_id"`              // 被投币视频 ID
	Count     uint8     `gorm:"default:1" json:"count"`                      // 投币数量（1 或 2）
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                  // 投币时间
}

// TableName 返回 video_coins 表名。
func (VideoCoin) TableName() string {
	return "video_coins"
}

// CoinReq 投币请求体。
type CoinReq struct {
	Count uint8 `json:"count" validate:"required,oneof=1 2"` // 投币数量，只能为 1 或 2
}

// CoinResp 投币响应体。
type CoinResp struct {
	CoinsToday uint `json:"coins_today"` // 今日已投币总数
}
