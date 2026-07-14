// Package admin 提供管理员后台相关的业务逻辑服务，
// 包括视频列表管理、审核等功能。
package admin

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	adminrepo "github.com/My-TuDo/B-B/backend/internal/repository/admin"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

// Service 管理员服务，封装管理后台相关的业务逻辑。
type Service struct {
	repo *adminrepo.Repository
}

// NewService 创建管理员服务实例。
func NewService(repo *adminrepo.Repository) *Service {
	return &Service{repo: repo}
}

// ListVideos 分页查询视频列表，支持按审核状态过滤。
// status 为视频状态码（0=待审, 1=通过, 2=驳回, 3=删除），
// page 和 pageSize 控制分页，不合法时会自动修正为默认值。
func (s *Service) ListVideos(ctx context.Context, status int8, page, pageSize int) (*videomodel.VideoListResp, error) {
	// 参数校验与修正
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询数据库
	offset := (page - 1) * pageSize
	videos, total, err := s.repo.ListVideos(ctx, status, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("admin.service.ListVideos: %w", err)
	}

	// 组装响应，将模型数据转换为对外响应结构
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
		// 填充作者简要信息
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
