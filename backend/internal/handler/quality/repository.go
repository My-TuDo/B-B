// Package quality 提供视频画质相关的数据访问层。
package quality

import (
	"context"
	"fmt"

	qmodel "github.com/My-TuDo/B-B/backend/internal/model/quality"
	"gorm.io/gorm"
)

// Repository 画质数据仓库，封装视频画质记录的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建画质数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByVideoID 根据视频 ID 查询所有画质记录，按清晰度升序排列。
func (r *Repository) FindByVideoID(ctx context.Context, videoID uint) ([]qmodel.VideoQuality, error) {
	var qualities []qmodel.VideoQuality
	result := r.db.WithContext(ctx).Where("video_id = ?", videoID).Order("quality ASC").Find(&qualities)
	if result.Error != nil {
		return nil, fmt.Errorf("quality.repository.FindByVideoID: %w", result.Error)
	}
	return qualities, nil
}

// Create 创建一条新的画质记录。
func (r *Repository) Create(ctx context.Context, q *qmodel.VideoQuality) error {
	result := r.db.WithContext(ctx).Create(q)
	if result.Error != nil {
		return fmt.Errorf("quality.repository.Create: %w", result.Error)
	}
	return nil
}

// CreateOrIgnore 当相同 video_id 和 quality 的记录不存在时才创建，避免重复。
func (r *Repository) CreateOrIgnore(ctx context.Context, q *qmodel.VideoQuality) error {
	var count int64
	r.db.WithContext(ctx).Model(&qmodel.VideoQuality{}).
		Where("video_id = ? AND quality = ?", q.VideoID, q.Quality).
		Count(&count)
	if count > 0 {
		return nil
	}
	return r.Create(ctx, q)
}
