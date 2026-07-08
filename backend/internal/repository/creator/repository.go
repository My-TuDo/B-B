package creator

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

func (r *Repository) ListByUser(ctx context.Context, userID uint, offset, limit int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("creator.repository.ListByUser.Count: %w", err)
	}

	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("creator.repository.ListByUser.Find: %w", err)
	}

	return videos, total, nil
}

func (r *Repository) GetStats(ctx context.Context, userID uint) (totalViews uint64, totalVideos int64, todayViews uint64, err error) {
	row := r.db.WithContext(ctx).Table("videos").
		Where("user_id = ? AND status != ?", userID, 3).
		Select("COUNT(*) as total_videos, COALESCE(SUM(views), 0) as total_views").
		Row()
	if err := row.Scan(&totalVideos, &totalViews); err != nil {
		return 0, 0, 0, fmt.Errorf("creator.repository.GetStats: %w", err)
	}

	if err := r.db.WithContext(ctx).Raw(
		"SELECT COALESCE(SUM(views), 0) FROM videos WHERE user_id = ? AND status != 3 AND DATE(created_at) = CURDATE()",
		userID,
	).Scan(&todayViews).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("creator.repository.GetStats.TodayViews: %w", err)
	}

	return totalViews, totalVideos, todayViews, nil
}
