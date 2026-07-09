package quality

import (
	"context"
	"fmt"

	qmodel "github.com/My-TuDo/B-B/backend/internal/model/quality"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByVideoID(ctx context.Context, videoID uint) ([]qmodel.VideoQuality, error) {
	var qualities []qmodel.VideoQuality
	result := r.db.WithContext(ctx).Where("video_id = ?", videoID).Order("quality ASC").Find(&qualities)
	if result.Error != nil {
		return nil, fmt.Errorf("quality.repository.FindByVideoID: %w", result.Error)
	}
	return qualities, nil
}

func (r *Repository) Create(ctx context.Context, q *qmodel.VideoQuality) error {
	result := r.db.WithContext(ctx).Create(q)
	if result.Error != nil {
		return fmt.Errorf("quality.repository.Create: %w", result.Error)
	}
	return nil
}

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
