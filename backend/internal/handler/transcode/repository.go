// Package transcode 提供视频转码任务的数据访问层。
package transcode

import (
	"context"
	"fmt"

	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	"gorm.io/gorm"
)

// Repository 转码数据仓库，封装转码任务记录的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建转码数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByVideoID 根据视频 ID 查询转码任务记录。
// 若未找到记录返回 nil（而非错误），表示该视频无转码任务。
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

// Create 创建一条新的转码任务记录。
func (r *Repository) Create(ctx context.Context, task *tmodel.TranscodeTask) error {
	result := r.db.WithContext(ctx).Create(task)
	if result.Error != nil {
		return fmt.Errorf("transcode.repository.Create: %w", result.Error)
	}
	return nil
}

// UpdateStatus 更新指定视频的转码状态、进度和错误信息。
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

// CreateOrGet 先查询再创建：若已存在则返回现有记录，否则创建新记录。
// 用于保证同一视频只存在一个转码任务。
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
