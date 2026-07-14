// Package video 定义视频数据模型，包含视频实体、分类别名及请求/响应结构。
package video

import (
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
)

// ==================== Entity ====================

// Video 视频实体，映射 videos 表。
type Video struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`            // 视频主键
	UserID      uint           `gorm:"index;not null" json:"user_id"`                 // 发布者用户 ID
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`      // 视频标题
	Description string         `gorm:"type:text" json:"description"`                  // 视频描述
	CoverURL    string         `gorm:"type:varchar(500);default:''" json:"cover_url"` // 封面图 URL
	VideoURL    string         `gorm:"type:varchar(500);not null" json:"video_url"`  // 视频文件 URL
	Duration    uint           `gorm:"default:0" json:"duration"`                     // 视频时长（秒）
	FileSize    uint64         `gorm:"default:0" json:"file_size"`                    // 文件大小（字节）
	CategoryID  uint           `gorm:"index;default:0" json:"category_id"`            // 所属分类 ID
	Status      int8           `gorm:"type:tinyint;default:0;index" json:"status"`   // 视频状态（0 待审 / 1 正常 / 3 拒绝）
	Views       uint64         `gorm:"default:0" json:"views"`                        // 播放量
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`                    // 发布时间
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`                    // 更新时间

	User     usermodel.User     `gorm:"foreignKey:UserID" json:"-"`                   // 关联发布者（不序列化）
	Category Category           `gorm:"foreignKey:CategoryID" json:"-"`               // 关联分类（不序列化）
}

// TableName 返回 videos 表名。
func (Video) TableName() string {
	return "videos"
}

// Category 分类实体别名，用于 Preload 关联查询。
type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`              // 分类主键
	Name string `gorm:"type:varchar(50)" json:"name"`      // 分类名称
	Slug string `gorm:"type:varchar(50)" json:"slug"`      // 分类英文标识
}

// TableName 返回 categories 表名。
func (Category) TableName() string {
	return "categories"
}

// ==================== DTOs ====================

// CreateVideoReq 创建/上传视频请求体。
type CreateVideoReq struct {
	Title       string `json:"title" validate:"required,max=100"`       // 视频标题
	Description string `json:"description" validate:"max=2000"`         // 视频描述
	CategoryID  uint   `json:"category_id"`                             // 分类 ID
}

// VideoResp 视频响应体，含发布者简要信息。
type VideoResp struct {
	ID          uint                `json:"id"`                    // 视频主键
	Title       string              `json:"title"`                 // 视频标题
	Description string              `json:"description"`           // 视频描述
	CoverURL    string              `json:"cover_url"`            // 封面图 URL
	VideoURL    string              `json:"video_url"`            // 视频文件 URL
	Duration    uint                `json:"duration"`              // 时长（秒）
	FileSize    uint64              `json:"file_size"`             // 文件大小（字节）
	CategoryID  uint                `json:"category_id"`           // 分类 ID
	Status      int8                `json:"status"`                // 视频状态
	Views       uint64              `json:"views"`                 // 播放量
	CreatedAt   time.Time           `json:"created_at"`            // 发布时间
	UpdatedAt   time.Time           `json:"updated_at"`            // 更新时间
	User        *usermodel.UserBrief `json:"user,omitempty"`       // 发布者简要信息
}

// VideoListResp 视频列表响应体。
type VideoListResp struct {
	Items    []VideoResp `json:"items"`     // 视频列表
	Total    int64       `json:"total"`     // 视频总数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
}

// UpdateVideoReq 更新视频信息请求体。
type UpdateVideoReq struct {
	Title       *string `json:"title" validate:"omitempty,max=100"`       // 视频标题
	Description *string `json:"description" validate:"omitempty,max=2000"` // 视频描述
	CategoryID  *uint   `json:"category_id"`                              // 分类 ID
	Status      *int8   `json:"status"`                                   // 视频状态
}

// UploadSSEMessage 上传进度 SSE 消息体。
type UploadSSEMessage struct {
	Uploaded int64  `json:"uploaded"`         // 已上传字节数
	Total    int64  `json:"total"`            // 文件总字节数
	Error    string `json:"error,omitempty"`  // 错误信息（如有）
}
