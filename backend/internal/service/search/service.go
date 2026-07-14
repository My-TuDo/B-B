// Package search 提供搜索相关的业务逻辑服务，
// 包括视频搜索和搜索建议等功能。
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

// Service 搜索服务，封装搜索相关的业务逻辑。
type Service struct {
	repo *searchrepo.Repository
}

// NewService 创建搜索服务实例。
func NewService(repo *searchrepo.Repository) *Service {
	return &Service{repo: repo}
}

// Search 根据关键词分页搜索视频，返回视频列表及总数。
func (s *Service) Search(ctx context.Context, q string, page, pageSize int) (*historymodel.SearchListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 执行搜索
	videos, total, err := s.repo.Search(ctx, q, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("search.service.Search: %w", err)
	}

	// 组装响应
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
		// 填充作者信息
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
	// 为封面生成公网访问URL（失败时静默忽略）
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

// Suggestions 根据输入前缀返回搜索建议列表，最多返回 limit 条。
func (s *Service) Suggestions(ctx context.Context, q string, limit int) ([]historymodel.SearchSuggestionResp, error) {
	// 参数校验
	if limit < 1 || limit > 20 {
		limit = 10
	}

	// 查询搜索建议
	rows, err := s.repo.SearchSuggestions(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("search.service.Suggestions: %w", err)
	}

	// 转换为响应格式
	resp := make([]historymodel.SearchSuggestionResp, len(rows))
	for i, r := range rows {
		resp[i] = historymodel.SearchSuggestionResp{Keyword: r.Keyword, Count: r.Count}
	}

	return resp, nil
}
