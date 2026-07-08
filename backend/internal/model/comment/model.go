package comment

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VideoID   uint      `gorm:"index;not null" json:"video_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ParentID  uint      `gorm:"index;default:0" json:"parent_id"`
	RootID    uint      `gorm:"index;default:0" json:"root_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Likes     uint      `gorm:"default:0" json:"likes"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`

	User usermodel.User `gorm:"foreignKey:UserID" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}

type CommentReq struct {
	Content  string `json:"content" validate:"required"`
	ParentID *uint  `json:"parent_id"`
	RootID   *uint  `json:"root_id"`
}

type CommentResp struct {
	ID        uint                `json:"id"`
	VideoID   uint                `json:"video_id"`
	UserID    uint                `json:"user_id"`
	ParentID  uint                `json:"parent_id"`
	RootID    uint                `json:"root_id"`
	Content   string              `json:"content"`
	Likes     uint                `json:"likes"`
	CreatedAt time.Time           `json:"created_at"`
	User      *usermodel.UserBrief `json:"user,omitempty"`
	Replies   []CommentResp       `json:"replies,omitempty"`
}

type CommentListResp struct {
	Items    []CommentResp `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type CommentLikeResp struct {
	Liked bool `json:"liked"`
	Likes uint `json:"likes"`
}
