package transcode

import (
	"context"
	"fmt"

	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByVideoID(ctx context.Context, videoID uint) (*tmodel.TranscodeTask, error) {
	var task tmodel.TranscodeTask
	result := r.db.WithContext(ctx).Where("video_id = ?", videoID).First(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("transcode.repository.FindByVideoID: %w", result.Error)
	}
	return &task, nil
}

func (r *Repository) Create(ctx context.Context, task *tmodel.TranscodeTask) error {
	result := r.db.WithContext(ctx).Create(task)
	if result.Error != nil {
		return fmt.Errorf("transcode.repository.Create: %w", result.Error)
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, videoID uint, status int8, progress uint8, errorMsg string) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	result := r.db.WithContext(ctx).Model(&tmodel.TranscodeTask{}).
		Where("video_id = ?", videoID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("transcode.repository.UpdateStatus: %w", result.Error)
	}
	return nil
}

func (r *Repository) CreateOrGet(ctx context.Context, task *tmodel.TranscodeTask) (*tmodel.TranscodeTask, error) {
	existing, err := r.FindByVideoID(ctx, task.VideoID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	if err := r.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}
