package coin

import (
	"context"
	"fmt"

	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *coinmodel.VideoCoin) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("coin.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) CountByVideoID(ctx context.Context, videoID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&coinmodel.VideoCoin{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("coin.repository.CountByVideoID: %w", err)
	}
	return count, nil
}

func (r *Repository) FindByUserAndVideo(ctx context.Context, userID, videoID uint) (*coinmodel.VideoCoin, error) {
	var c coinmodel.VideoCoin
	if err := r.db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("coin.repository.FindByUserAndVideo: %w", err)
	}
	return &c, nil
}
