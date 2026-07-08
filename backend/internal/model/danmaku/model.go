package danmaku

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

type Danmaku struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID   uint      `gorm:"index;not null" json:"video_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Content   string    `gorm:"type:varchar(200);not null" json:"content"`
	Color     string    `gorm:"type:varchar(7);default:#ffffff" json:"color"`
	Position  int8      `gorm:"type:tinyint;default:0" json:"position"`
	Size      int8      `gorm:"type:tinyint;default:1" json:"size"`
	PlayTime  uint      `gorm:"default:0" json:"play_time"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`

	User usermodel.User `gorm:"foreignKey:UserID" json:"-"`
}

func (Danmaku) TableName() string {
	return "danmaku"
}

type DanmakuReq struct {
	Content  string `json:"content" validate:"required,max=200"`
	Color    string `json:"color" validate:"omitempty,max=7"`
	Position *int8  `json:"position" validate:"omitempty,oneof=0 1 2"`
	Size     *int8  `json:"size" validate:"omitempty,oneof=0 1 2"`
	PlayTime uint   `json:"play_time" validate:"omitempty"`
}

type DanmakuResp struct {
	ID       uint                `json:"id"`
	Content  string              `json:"content"`
	Color    string              `json:"color"`
	Position int8                `json:"position"`
	Size     int8                `json:"size"`
	PlayTime uint                `json:"play_time"`
	User     *usermodel.UserBrief `json:"user,omitempty"`
}
