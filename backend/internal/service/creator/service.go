package creator

import (
	"context"
	"fmt"

	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	creatorrepo "github.com/My-TuDo/B-B/backend/internal/repository/creator"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

type Service struct {
	repo *creatorrepo.Repository
}

func NewService(repo *creatorrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListVideos(ctx context.Context, userID uint, page, pageSize int) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	offset := (page - 1) * pageSize
	videos, total, err := s.repo.ListByUser(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("creator.service.ListVideos: %w", err)
	}

	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		resp := videomodel.VideoResp{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			CoverURL:    v.CoverURL,
			VideoURL:    v.VideoURL,
			Duration:    v.Duration,
			FileSize:    v.FileSize,
			CategoryID:  v.CategoryID,
			Status:      v.Status,
			Views:       v.Views,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
		if v.User.ID != 0 {
			avatar := ""
			if v.User.Avatar != "" {
				avatar = storage.GetObjectURL(v.User.Avatar)
			}
			resp.User = &usermodel.UserBrief{
				ID:       v.User.ID,
				Username: v.User.Username,
				Nickname: v.User.Nickname,
				Avatar:   avatar,
			}
		}
		// Presign cover URL — use direct public URL since bucket is public
		if resp.CoverURL != "" {
			resp.CoverURL = storage.GetObjectURL(resp.CoverURL)
		}
		items[i] = resp
	}

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) GetStats(ctx context.Context, userID uint) (*historymodel.CreatorStatsResp, error) {
	totalViews, totalVideos, todayViews, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("creator.service.GetStats: %w", err)
	}

	return &historymodel.CreatorStatsResp{
		TotalViews:   totalViews,
		TotalVideos:  totalVideos,
		TodayViews:   todayViews,
		TodayNewFans: 0,
	}, nil
}
