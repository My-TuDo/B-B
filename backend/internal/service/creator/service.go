// Package creator 提供创作者中心相关的业务逻辑服务，
// 包括创作者视频列表查询和创作数据统计等功能。
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

// Service 创作者服务，封装创作者相关的业务逻辑。
type Service struct {
	repo *creatorrepo.Repository
}

// NewService 创建创作者服务实例。
func NewService(repo *creatorrepo.Repository) *Service {
	return &Service{repo: repo}
}

// ListVideos 分页查询指定用户的视频列表。
func (s *Service) ListVideos(ctx context.Context, userID uint, page, pageSize int) (*videomodel.VideoListResp, error) {
	// 参数校验与修正
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询该用户的视频
	offset := (page - 1) * pageSize
	videos, total, err := s.repo.ListByUser(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("creator.service.ListVideos: %w", err)
	}

	// 组装响应
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
		// 填充作者信息
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
		// Presign cover URL — 为封面生成公网访问URL
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

// GetStats 获取创作者的统计数据，包括总播放量、总视频数和今日播放量。
func (s *Service) GetStats(ctx context.Context, userID uint) (*historymodel.CreatorStatsResp, error) {
	// 查询统计数据
	totalViews, totalVideos, todayViews, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("creator.service.GetStats: %w", err)
	}

	return &historymodel.CreatorStatsResp{
		TotalViews:   totalViews,
		TotalVideos:  totalVideos,
		TodayViews:   todayViews,
		TodayNewFans: 0, // TODO: 实现今日新增粉丝统计
	}, nil
}
