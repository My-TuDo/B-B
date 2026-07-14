// Package category 定义视频分类数据模型，包含分类实体与对外响应结构。
package category

import "time"

// ==================== Entity ====================

// Category 视频分类实体，映射 categories 表。
type Category struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`         // 分类主键
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`      // 分类名称（中文）
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"` // 分类英文标识
	CreatedAt time.Time `gorm:"not null" json:"created_at"`                 // 创建时间
}

// TableName 返回 categories 表名。
func (Category) TableName() string {
	return "categories"
}

// ==================== DTOs ====================

// CategoryResp 分类响应体，返回给客户端。
type CategoryResp struct {
	ID   uint   `json:"id"`   // 分类主键
	Name string `json:"name"` // 分类名称
	Slug string `json:"slug"` // 分类英文标识
}
