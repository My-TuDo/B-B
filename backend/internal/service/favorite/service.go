// Package favorite 提供收藏夹相关的业务逻辑服务，
// 包括收藏夹的创建、查询、详情获取以及收藏/取消收藏视频等功能。
package favorite

import (
	"context"
	"fmt"
	"time"

	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	favoriterepo "github.com/My-TuDo/B-B/backend/internal/repository/favorite"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

// Service 收藏夹服务，封装收藏夹相关的业务逻辑。
type Service struct {
	repo *favoriterepo.Repository
}

// NewService 创建收藏夹服务实例。
func NewService(repo *favoriterepo.Repository) *Service {
	return &Service{repo: repo}
}

// CreateFavorite 创建新的收藏夹。
func (s *Service) CreateFavorite(ctx context.Context, userID uint, req *favoritemodel.FavoriteReq) (*favoritemodel.FavoriteResp, error) {
	// 处理可选字段默认值
	isPublic := int8(1)
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	// 构建收藏夹实体
	f := &favoritemodel.Favorite{
		UserID:   userID,
		Name:     req.Name,
		IsPublic: isPublic,
	}

	// 写入数据库
	if err := s.repo.CreateFavorite(ctx, f); err != nil {
		return nil, fmt.Errorf("favorite.service.CreateFavorite: %w", err)
	}

	return &favoritemodel.FavoriteResp{
		ID:       f.ID,
		Name:     f.Name,
		IsPublic: f.IsPublic,
	}, nil
}

// GetUserPublicFavorites 获取指定用户的公开收藏夹列表，包含收藏数和封面。
func (s *Service) GetUserPublicFavorites(ctx context.Context, userID uint) ([]favoritemodel.FavoriteResp, error) {
	list, err := s.repo.FindFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetUserPublicFavorites: %w", err)
	}

	resps := make([]favoritemodel.FavoriteResp, 0)
	for _, f := range list {
		// 仅返回公开收藏夹
		if f.IsPublic == 0 {
			continue
		}
		// 统计收藏数量
		count, _ := s.repo.CountItems(ctx, f.ID)
		// 获取封面（第一个收藏视频的封面）
		coverURL := ""
		if url, err := s.repo.FindFirstCover(ctx, f.ID); err == nil && url != "" {
			if presigned, err := storage.GetPresignedURL(ctx, url, time.Hour); err == nil {
				coverURL = presigned
			}
		}
		resps = append(resps, favoritemodel.FavoriteResp{
			ID:        f.ID,
			Name:      f.Name,
			IsPublic:  f.IsPublic,
			ItemCount: count,
			CoverURL:  coverURL,
		})
	}

	return resps, nil
}

// GetFavorites 获取当前用户的所有收藏夹列表。如果用户没有任何收藏夹，
// 则自动创建默认收藏夹后返回。
func (s *Service) GetFavorites(ctx context.Context, userID uint) ([]favoritemodel.FavoriteResp, error) {
	list, err := s.repo.FindFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
	}

	// Auto-create default favorite if user has none
	// 用户无收藏夹时自动创建默认收藏夹
	if len(list) == 0 {
		if err := s.CreateDefaultFavorite(ctx, userID); err != nil {
			return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
		}
		// Re-fetch after creating default — 重新查询
		list, err = s.repo.FindFavoritesByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
		}
	}

	// 组装响应
	resps := make([]favoritemodel.FavoriteResp, 0, len(list))
	for _, f := range list {
		count, _ := s.repo.CountItems(ctx, f.ID)
		coverURL := ""
		if url, err := s.repo.FindFirstCover(ctx, f.ID); err == nil && url != "" {
			if presigned, err := storage.GetPresignedURL(ctx, url, time.Hour); err == nil {
				coverURL = presigned
			}
		}
		resps = append(resps, favoritemodel.FavoriteResp{
			ID:        f.ID,
			Name:      f.Name,
			IsPublic:  f.IsPublic,
			ItemCount: count,
			CoverURL:  coverURL,
		})
	}

	return resps, nil
}

// GetFavoriteDetail 获取收藏夹详情，包括其中包含的视频列表（分页）。
// 私有收藏夹仅允许所有者查看。
func (s *Service) GetFavoriteDetail(ctx context.Context, userID uint, favoriteID uint, page, pageSize int) (*favoritemodel.FavoriteDetailResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询收藏夹信息
	f, err := s.repo.FindFavoriteByID(ctx, favoriteID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", err)
	}
	if f == nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", newError(errcode.FavoriteNotFound))
	}
	// 私有收藏夹权限校验
	if f.IsPublic == 0 && f.UserID != userID {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", newError(errcode.Forbidden))
	}

	// 查询收藏夹内的视频
	offset := (page - 1) * pageSize
	videos, total, err := s.repo.FindItemsByFavoriteID(ctx, favoriteID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", err)
	}

	// 组装视频响应
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
		// Presign cover — 生成封面公网URL
		if v.CoverURL != "" {
			items[i].CoverURL = storage.GetObjectURL(v.CoverURL)
		}
	}

	return &favoritemodel.FavoriteDetailResp{
		Favorite: favoritemodel.FavoriteResp{
			ID:       f.ID,
			Name:     f.Name,
			IsPublic: f.IsPublic,
		},
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ToggleFavoriteItem 切换视频的收藏状态：已收藏则取消，未收藏则添加。
// 仅收藏夹所有者可以操作。
func (s *Service) ToggleFavoriteItem(ctx context.Context, userID uint, favoriteID uint, videoID uint) (*favoritemodel.FavoriteToggleResp, error) {
	// 校验收藏夹存在且属于当前用户
	f, err := s.repo.FindFavoriteByID(ctx, favoriteID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
	}
	if f == nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", newError(errcode.FavoriteNotFound))
	}
	if f.UserID != userID {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", newError(errcode.Forbidden))
	}

	// 检查当前收藏状态
	existing, err := s.repo.FindItem(ctx, favoriteID, videoID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
	}

	if existing != nil {
		// Remove — 已收藏则取消
		if err := s.repo.DeleteItem(ctx, existing.ID); err != nil {
			return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
		}
		return &favoritemodel.FavoriteToggleResp{Favorited: false}, nil
	}

	// Add — 未收藏则添加
	item := &favoritemodel.FavoriteItem{
		FavoriteID: favoriteID,
		VideoID:    videoID,
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
	}
	return &favoritemodel.FavoriteToggleResp{Favorited: true}, nil
}

// CreateDefaultFavorite 为用户创建名为"默认收藏夹"的公开收藏夹。
// 若已存在则跳过。
func (s *Service) CreateDefaultFavorite(ctx context.Context, userID uint) error {
	// 检查是否已存在默认收藏夹
	existing, err := s.repo.FindDefaultFavorite(ctx, userID)
	if err != nil {
		return fmt.Errorf("favorite.service.CreateDefaultFavorite: %w", err)
	}
	if existing != nil {
		return nil
	}

	// 创建默认收藏夹
	f := &favoritemodel.Favorite{
		UserID:   userID,
		Name:     "默认收藏夹",
		IsPublic: 1,
	}
	if err := s.repo.CreateFavorite(ctx, f); err != nil {
		return fmt.Errorf("favorite.service.CreateDefaultFavorite: %w", err)
	}
	return nil
}

// Error 服务层错误类型，携带错误码以支持在HTTP层映射为合适的响应。
type Error struct {
	Code int
	Msg  string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	return e.Msg
}

// newError 根据错误码创建带本地化消息的服务错误。
func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}
