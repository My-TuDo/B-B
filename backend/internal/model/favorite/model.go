package favorite

import (
	"time"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
)

type Favorite struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	IsPublic  int8      `gorm:"type:tinyint;default:1" json:"is_public"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}

type FavoriteItem struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FavoriteID uint      `gorm:"uniqueIndex:uk_fav_video;not null" json:"favorite_id"`
	VideoID    uint      `gorm:"uniqueIndex:uk_fav_video;not null" json:"video_id"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
}

func (FavoriteItem) TableName() string {
	return "favorite_items"
}

type FavoriteReq struct {
	Name     string `json:"name" validate:"required,max=50"`
	IsPublic *int8  `json:"is_public" validate:"omitempty,oneof=0 1"`
}

type FavoriteResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IsPublic  int8   `json:"is_public"`
	ItemCount int64  `json:"item_count"`
	CoverURL  string `json:"cover_url,omitempty"`
}

type FavoriteDetailResp struct {
	Favorite FavoriteResp         `json:"favorite"`
	Items    []videomodel.VideoResp `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type FavoriteItemReq struct {
	VideoID uint `json:"video_id" validate:"required"`
}

type FavoriteToggleResp struct {
	Favorited bool `json:"favorited"`
}
