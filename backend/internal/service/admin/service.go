package admin

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	adminrepo "github.com/My-TuDo/B-B/backend/internal/repository/admin"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

type Service struct {
	repo *adminrepo.Repository
}

func NewService(repo *adminrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListVideos(ctx context.Context, status int8, page, pageSize int) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	offset := (page - 1) * pageSize
	videos, total, err := s.repo.ListVideos(ctx, status, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("admin.service.ListVideos: %w", err)
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
		items[i] = resp
	}

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
