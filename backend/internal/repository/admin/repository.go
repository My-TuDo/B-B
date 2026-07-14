// Package admin 提供管理员数据访问层，封装与管理后台相关的数据库操作。
// 包括审核视频列表查询等功能，是管理员业务逻辑与数据库之间的桥梁。
package admin

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

// Repository 管理员数据仓库，封装与管理员功能相关的数据库操作。
// 通过持有的 gorm.DB 实例与数据库交互，支持上下文传递和事务管理。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建管理员数据仓库实例。
// 接收一个已初始化的 gorm.DB 指针，返回可用的 Repository 指针。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListVideos 根据视频审核状态分页查询视频列表。
// status 为视频审核状态码（0:待审 1:通过 2:拒绝 3:删除），
// offset 和 limit 控制分页偏移量和每页数量。
// 返回视频切片、总数统计以及可能的错误。
func (r *Repository) ListVideos(ctx context.Context, status int8, offset, limit int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	// 构建基础查询：按状态筛选视频
	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("status = ?", status)

	// 先统计符合条件的视频总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("admin.repository.ListVideos.Count: %w", err)
	}

	// 预加载关联用户信息，按创建时间倒序，分页查询视频列表
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("admin.repository.ListVideos.Find: %w", err)
	}

	return videos, total, nil
}
