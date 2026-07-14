// Package quality 提供视频画质相关的业务逻辑层。
package quality

import (
	"context"
	"fmt"

	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

// Service 画质业务服务，封装画质查询逻辑。
type Service struct {
	repo *Repository
}

// NewService 创建画质业务服务实例。
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// QualityInfo 画质信息响应结构体，包含清晰度标签、播放地址和文件大小。
type QualityInfo struct {
	Quality  string `json:"quality"`   // 清晰度标签，如 "1080p"
	PlayURL  string `json:"play_url"`  // HLS 播放地址
	FileSize uint64 `json:"file_size"` // 文件大小（字节）
}

// GetQualities 获取指定视频的所有画质版本信息。
// 将存储中的对象名转换为可直接访问的播放 URL。
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
