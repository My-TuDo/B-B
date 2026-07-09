package transcode

import (
	"context"
	"fmt"

	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetStatus returns the transcode task status for a given video.
// If no task exists, returns a completed status (for videos uploaded before transcode was introduced).
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
func (s *Service) UpdateProgress(ctx context.Context, videoID uint, status int8, progress uint8, errorMsg string) error {
	if err := s.repo.UpdateStatus(ctx, videoID, status, progress, errorMsg); err != nil {
		return fmt.Errorf("transcode.service.UpdateProgress: %w", err)
	}
	return nil
}
