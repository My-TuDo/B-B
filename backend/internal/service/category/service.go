package category

import (
	"context"
	"fmt"

	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	categoryrepo "github.com/My-TuDo/B-B/backend/internal/repository/category"
)

type Service struct {
	repo *categoryrepo.Repository
}

func NewService(repo *categoryrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]categorymodel.CategoryResp, error) {
	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("category.service.List: %w", err)
	}

	resp := make([]categorymodel.CategoryResp, len(categories))
	for i, cat := range categories {
		resp[i] = categorymodel.CategoryResp{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}

	return resp, nil
}
