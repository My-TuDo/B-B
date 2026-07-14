// Package category 提供分类数据访问层，封装视频分类相关的数据库操作。
// 支持查询全部分类列表，为前端分类导航和筛选提供数据支持。
package category

import (
	"context"
	"fmt"

	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	"gorm.io/gorm"
)

// Repository 分类数据仓库，封装分类相关的数据库查询操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建分类数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindAll 查询所有分类，按 ID 升序排列。
// 返回完整分类列表，供前端渲染分类导航和筛选器使用。
func (r *Repository) FindAll(ctx context.Context) ([]categorymodel.Category, error) {
	var categories []categorymodel.Category
	result := r.db.WithContext(ctx).Order("id ASC").Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("category.repository.FindAll: %w", result.Error)
	}
	return categories, nil
}
