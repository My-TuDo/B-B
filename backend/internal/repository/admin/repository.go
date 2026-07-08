package admin

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListVideos(ctx context.Context, status int8, offset, limit int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("admin.repository.ListVideos.Count: %w", err)
	}

	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("admin.repository.ListVideos.Find: %w", err)
	}

	return videos, total, nil
}
