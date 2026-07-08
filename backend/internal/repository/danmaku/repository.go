package danmaku

import (
	"context"
	"fmt"

	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, d *danmakumodel.Danmaku) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("danmaku.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) FindByVideoID(ctx context.Context, videoID uint) ([]danmakumodel.Danmaku, error) {
	var list []danmakumodel.Danmaku
	if err := r.db.WithContext(ctx).Preload("User").Where("video_id = ?", videoID).Order("play_time ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("danmaku.repository.FindByVideoID: %w", err)
	}
	return list, nil
}
