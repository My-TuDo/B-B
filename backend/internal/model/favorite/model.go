// Package favorite 定义视频收藏夹数据模型，包含收藏夹、收藏项实体与请求/响应结构。
package favorite

import (
	"time"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
)

// Favorite 收藏夹实体，映射 favorites 表。
type Favorite struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`           // 收藏夹主键
	UserID    uint      `gorm:"index;not null" json:"user_id"`                // 所属用户 ID
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`        // 收藏夹名称
	IsPublic  int8      `gorm:"type:tinyint;default:1" json:"is_public"`     // 是否公开（1 公开 / 0 私密）
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                   // 创建时间
}

// TableName 返回 favorites 表名。
func (Favorite) TableName() string {
	return "favorites"
}

// FavoriteItem 收藏项实体，映射 favorite_items 表，记录收藏夹中的视频。
type FavoriteItem struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`                             // 记录主键
	FavoriteID uint      `gorm:"uniqueIndex:uk_fav_video;not null" json:"favorite_id"`           // 所属收藏夹 ID
	VideoID    uint      `gorm:"uniqueIndex:uk_fav_video;not null" json:"video_id"`              // 被收藏视频 ID
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`                                     // 收藏时间
}

// TableName 返回 favorite_items 表名。
func (FavoriteItem) TableName() string {
	return "favorite_items"
}

// FavoriteReq 创建/更新收藏夹请求体。
type FavoriteReq struct {
	Name     string `json:"name" validate:"required,max=50"`                // 收藏夹名称
	IsPublic *int8  `json:"is_public" validate:"omitempty,oneof=0 1"`       // 是否公开
}

// FavoriteResp 收藏夹响应体。
type FavoriteResp struct {
	ID        uint   `json:"id"`                  // 收藏夹主键
	Name      string `json:"name"`                // 收藏夹名称
	IsPublic  int8   `json:"is_public"`           // 是否公开
	ItemCount int64  `json:"item_count"`          // 收藏视频数量
	CoverURL  string `json:"cover_url,omitempty"` // 封面图 URL
}

// FavoriteDetailResp 收藏夹详情响应体，含视频列表。
type FavoriteDetailResp struct {
	Favorite FavoriteResp         `json:"favorite"` // 收藏夹信息
	Items    []videomodel.VideoResp `json:"items"`  // 收藏视频列表
	Total    int64                `json:"total"`    // 视频总数
	Page     int                  `json:"page"`     // 当前页码
	PageSize int                  `json:"page_size"` // 每页数量
}

// FavoriteItemReq 添加/移除收藏项请求体。
type FavoriteItemReq struct {
	VideoID uint `json:"video_id" validate:"required"` // 视频 ID
}

// FavoriteToggleResp 收藏切换响应体。
type FavoriteToggleResp struct {
	Favorited bool `json:"favorited"` // 当前是否已收藏
}
