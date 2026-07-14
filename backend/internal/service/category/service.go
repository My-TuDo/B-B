// Package category 提供视频分类相关的业务逻辑服务，
// 包括分类列表查询等功能。
package category

import (
	"context"
	"fmt"

	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	categoryrepo "github.com/My-TuDo/B-B/backend/internal/repository/category"
)

// Service 分类服务，封装视频分类相关的业务逻辑。
type Service struct {
	repo *categoryrepo.Repository
}

// NewService 创建分类服务实例。
func NewService(repo *categoryrepo.Repository) *Service {
	return &Service{repo: repo}
}

// List 获取全部分类列表，按存储顺序返回。
func (s *Service) List(ctx context.Context) ([]categorymodel.CategoryResp, error) {
	// 查询所有分类
	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("category.service.List: %w", err)
	}

	// 转换为响应结构
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
