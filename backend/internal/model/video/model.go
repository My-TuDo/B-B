package video

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// ==================== Entity ====================

type Video struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CoverURL    string         `gorm:"type:varchar(500);default:''" json:"cover_url"`
	VideoURL    string         `gorm:"type:varchar(500);not null" json:"video_url"`
	Duration    uint           `gorm:"default:0" json:"duration"`
	FileSize    uint64         `gorm:"default:0" json:"file_size"`
	CategoryID  uint           `gorm:"index;default:0" json:"category_id"`
	Status      int8           `gorm:"type:tinyint;default:0;index" json:"status"`
	Views       uint64         `gorm:"default:0" json:"views"`
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`

	User     usermodel.User     `gorm:"foreignKey:UserID" json:"-"`
	Category Category           `gorm:"foreignKey:CategoryID" json:"-"`
}

func (Video) TableName() string {
	return "videos"
}

// Category is a local alias for the category entity used in Preload.
type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50)" json:"name"`
	Slug string `gorm:"type:varchar(50)" json:"slug"`
}

func (Category) TableName() string {
	return "categories"
}

// ==================== DTOs ====================

type CreateVideoReq struct {
	Title       string `json:"title" validate:"required,max=100"`
	Description string `json:"description" validate:"max=2000"`
	CategoryID  uint   `json:"category_id"`
}

type VideoResp struct {
	ID          uint                `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	CoverURL    string              `json:"cover_url"`
	VideoURL    string              `json:"video_url"`
	Duration    uint                `json:"duration"`
	FileSize    uint64              `json:"file_size"`
	CategoryID  uint                `json:"category_id"`
	Status      int8                `json:"status"`
	Views       uint64              `json:"views"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	User        *usermodel.UserBrief `json:"user,omitempty"`
}

type VideoListResp struct {
	Items    []VideoResp `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type UpdateVideoReq struct {
	Title       *string `json:"title" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	CategoryID  *uint   `json:"category_id"`
	Status      *int8   `json:"status"`
}

type UploadSSEMessage struct {
	Uploaded int64  `json:"uploaded"`
	Total    int64  `json:"total"`
	Error    string `json:"error,omitempty"`
}
