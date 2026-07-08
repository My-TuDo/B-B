package history

import (
	"context"
	"fmt"
	"time"

	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateOrUpdate(ctx context.Context, h *historymodel.VideoHistory) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "video_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"progress", "watched_at"}),
	}).Create(h)
	if result.Error != nil {
		return fmt.Errorf("history.repository.CreateOrUpdate: %w", result.Error)
	}
	return nil
}

func (r *Repository) FindByUserAndVideo(ctx context.Context, userID, videoID uint) (*historymodel.VideoHistory, error) {
	var h historymodel.VideoHistory
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&h)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("history.repository.FindByUserAndVideo: %w", result.Error)
	}
	return &h, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]historymodel.VideoHistory, int64, error) {
	var histories []historymodel.VideoHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&historymodel.VideoHistory{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("history.repository.ListByUser.Count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Video").Preload("Video.User").
		Order("watched_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&histories).Error; err != nil {
		return nil, 0, fmt.Errorf("history.repository.ListByUser.Find: %w", err)
	}

	// Set watched_at to now if not set
	now := time.Now()
	for i := range histories {
		if histories[i].WatchedAt.IsZero() {
			histories[i].WatchedAt = now
		}
	}

	return histories, total, nil
}
