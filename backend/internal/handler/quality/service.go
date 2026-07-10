package quality

import (
	"context"
	"fmt"

	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type QualityInfo struct {
	Quality  string `json:"quality"`
	PlayURL  string `json:"play_url"`
	FileSize uint64 `json:"file_size"`
}

func (s *Service) GetQualities(ctx context.Context, videoID uint) ([]QualityInfo, error) {
	qualities, err := s.repo.FindByVideoID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("quality.service.GetQualities: %w", err)
	}

	result := make([]QualityInfo, 0, len(qualities))
	for _, q := range qualities {
		// HLS playback needs a stable URL without presigned query params so
		// Video.js can resolve relative .ts segment paths (e.g. seg_000.ts).
		// The bucket is public-download so unauthenticated GETs succeed.
		playURL := storage.GetObjectURL(q.ObjectName)
		result = append(result, QualityInfo{
			Quality:  q.Quality,
			PlayURL:  playURL,
			FileSize: q.FileSize,
		})
	}

	return result, nil
}
