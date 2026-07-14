// Package transcode 提供视频转码相关的业务逻辑层。
package transcode

import (
	"context"
	"fmt"

	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
)

// Service 转码业务服务，封装转码任务的状态查询、创建和进度更新逻辑。
type Service struct {
	repo *Repository
}

// NewService 创建转码业务服务实例。
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetStatus returns the transcode task status for a given video.
// If no task exists, returns a completed status (for videos uploaded before transcode was introduced).
// 无转码记录视为已完成（兼容旧数据），返回 status=done、progress=100。
func (s *Service) GetStatus(ctx context.Context, videoID uint) (*tmodel.TranscodeTask, error) {
	task, err := s.repo.FindByVideoID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("transcode.service.GetStatus: %w", err)
	}
	if task == nil {
		// No transcode task means either pre-transcode video or already transcoded.
		// Return a synthetic completed status so the UI shows "ready".
		return &tmodel.TranscodeTask{
			VideoID:  videoID,
			Status:   tmodel.StatusDone,
			Progress: 100,
		}, nil
	}
	return task, nil
}

// CreateTask creates a new transcode task for a video.
// 使用 CreateOrGet 确保幂等：同一视频不会重复创建任务。
func (s *Service) CreateTask(ctx context.Context, videoID uint) (*tmodel.TranscodeTask, error) {
	task := &tmodel.TranscodeTask{
		VideoID: videoID,
		Status:  tmodel.StatusPending,
	}
	created, err := s.repo.CreateOrGet(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("transcode.service.CreateTask: %w", err)
	}
	return created, nil
}

// UpdateProgress updates the task status and progress.
// 转码工作进程调用此方法更新数据库中的进度。
func (s *Service) UpdateProgress(ctx context.Context, videoID uint, status int8, progress uint8, errorMsg string) error {
	if err := s.repo.UpdateStatus(ctx, videoID, status, progress, errorMsg); err != nil {
		return fmt.Errorf("transcode.service.UpdateProgress: %w", err)
	}
	return nil
}
