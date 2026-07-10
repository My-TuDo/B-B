package search

import (
	"context"
	"fmt"

	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	searchrepo "github.com/My-TuDo/B-B/backend/internal/repository/search"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

type Service struct {
	repo *searchrepo.Repository
}

func NewService(repo *searchrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, q string, page, pageSize int) (*historymodel.SearchListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	videos, total, err := s.repo.Search(ctx, q, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("search.service.Search: %w", err)
	}

	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = videomodel.VideoResp{
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
			items[i].User = &usermodel.UserBrief{
				ID:       v.User.ID,
				Username: v.User.Username,
				Nickname: v.User.Nickname,
				Avatar:   avatar,
			}
		}
	}

	// Generate presigned URLs for covers (failures are silently ignored)
	for i := range items {
		if items[i].CoverURL != "" {
			items[i].CoverURL = storage.GetObjectURL(items[i].CoverURL)
		}
	}

	return &historymodel.SearchListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) Suggestions(ctx context.Context, q string, limit int) ([]historymodel.SearchSuggestionResp, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}

	rows, err := s.repo.SearchSuggestions(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("search.service.Suggestions: %w", err)
	}

	resp := make([]historymodel.SearchSuggestionResp, len(rows))
	for i, r := range rows {
		resp[i] = historymodel.SearchSuggestionResp{Keyword: r.Keyword, Count: r.Count}
	}

	return resp, nil
}
