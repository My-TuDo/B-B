package coin

import "time"

type VideoCoin struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	VideoID   uint      `gorm:"index;not null" json:"video_id"`
	Count     uint8     `gorm:"default:1" json:"count"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (VideoCoin) TableName() string {
	return "video_coins"
}

type CoinReq struct {
	Count uint8 `json:"count" validate:"required,oneof=1 2"`
}

type CoinResp struct {
	CoinsToday uint `json:"coins_today"`
}
