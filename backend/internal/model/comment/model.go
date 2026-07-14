// Package comment 定义视频评论数据模型，包含评论实体与请求/响应结构。
package comment

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// Comment 评论实体，映射 comments 表，支持嵌套回复。
type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`          // 评论主键
	VideoID   uint      `gorm:"index;not null" json:"video_id"`              // 所属视频 ID
	UserID    uint      `gorm:"not null" json:"user_id"`                     // 评论用户 ID
	ParentID  uint      `gorm:"index;default:0" json:"parent_id"`            // 父评论 ID（0 表示顶层评论）
	RootID    uint      `gorm:"index;default:0" json:"root_id"`              // 根评论 ID（0 表示顶层评论）
	Content   string    `gorm:"type:text;not null" json:"content"`           // 评论内容
	Likes     uint      `gorm:"default:0" json:"likes"`                      // 评论点赞数
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                  // 评论时间

	User usermodel.User `gorm:"foreignKey:UserID" json:"-"`                 // 关联用户（不序列化）
}

// TableName 返回 comments 表名。
func (Comment) TableName() string {
	return "comments"
}

// CommentReq 发表评论请求体。
type CommentReq struct {
	Content  string `json:"content" validate:"required"` // 评论内容
	ParentID *uint  `json:"parent_id"`                    // 父评论 ID（可选，用于回复）
	RootID   *uint  `json:"root_id"`                      // 根评论 ID（可选，用于嵌套回复）
}

// CommentResp 评论响应体，含用户信息与子回复。
type CommentResp struct {
	ID        uint                `json:"id"`                 // 评论主键
	VideoID   uint                `json:"video_id"`           // 所属视频 ID
	UserID    uint                `json:"user_id"`            // 评论用户 ID
	ParentID  uint                `json:"parent_id"`          // 父评论 ID
	RootID    uint                `json:"root_id"`            // 根评论 ID
	Content   string              `json:"content"`            // 评论内容
	Likes     uint                `json:"likes"`              // 点赞数
	CreatedAt time.Time           `json:"created_at"`         // 评论时间
	User      *usermodel.UserBrief `json:"user,omitempty"`    // 评论用户简要信息
	Replies   []CommentResp       `json:"replies,omitempty"`  // 子回复列表
}

// CommentListResp 评论列表响应体。
type CommentListResp struct {
	Items    []CommentResp `json:"items"`     // 评论列表
	Total    int64         `json:"total"`     // 评论总数
	Page     int           `json:"page"`      // 当前页码
	PageSize int           `json:"page_size"` // 每页数量
}

// CommentLikeResp 评论点赞/取消点赞响应体。
type CommentLikeResp struct {
	Liked bool `json:"liked"` // 当前用户是否已点赞
	Likes uint `json:"likes"` // 评论最新点赞数
}
