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

type Service struct {
	repo *favoriterepo.Repository
}

func NewService(repo *favoriterepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateFavorite(ctx context.Context, userID uint, req *favoritemodel.FavoriteReq) (*favoritemodel.FavoriteResp, error) {
	isPublic := int8(1)
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	f := &favoritemodel.Favorite{
		UserID:   userID,
		Name:     req.Name,
		IsPublic: isPublic,
	}

	if err := s.repo.CreateFavorite(ctx, f); err != nil {
		return nil, fmt.Errorf("favorite.service.CreateFavorite: %w", err)
	}

	return &favoritemodel.FavoriteResp{
		ID:       f.ID,
		Name:     f.Name,
		IsPublic: f.IsPublic,
	}, nil
}

func (s *Service) GetUserPublicFavorites(ctx context.Context, userID uint) ([]favoritemodel.FavoriteResp, error) {
	list, err := s.repo.FindFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetUserPublicFavorites: %w", err)
	}

	resps := make([]favoritemodel.FavoriteResp, 0)
	for _, f := range list {
		if f.IsPublic == 0 {
			continue
		}
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

func (s *Service) GetFavorites(ctx context.Context, userID uint) ([]favoritemodel.FavoriteResp, error) {
	list, err := s.repo.FindFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
	}

	// Auto-create default favorite if user has none
	if len(list) == 0 {
		if err := s.CreateDefaultFavorite(ctx, userID); err != nil {
			return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
		}
		// Re-fetch after creating default
		list, err = s.repo.FindFavoritesByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("favorite.service.GetFavorites: %w", err)
		}
	}

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

func (s *Service) GetFavoriteDetail(ctx context.Context, userID uint, favoriteID uint, page, pageSize int) (*favoritemodel.FavoriteDetailResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	f, err := s.repo.FindFavoriteByID(ctx, favoriteID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", err)
	}
	if f == nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", newError(errcode.FavoriteNotFound))
	}
	if f.IsPublic == 0 && f.UserID != userID {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", newError(errcode.Forbidden))
	}

	offset := (page - 1) * pageSize
	videos, total, err := s.repo.FindItemsByFavoriteID(ctx, favoriteID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.GetFavoriteDetail: %w", err)
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
		// Presign cover
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

func (s *Service) ToggleFavoriteItem(ctx context.Context, userID uint, favoriteID uint, videoID uint) (*favoritemodel.FavoriteToggleResp, error) {
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

	existing, err := s.repo.FindItem(ctx, favoriteID, videoID)
	if err != nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
	}

	if existing != nil {
		// Remove
		if err := s.repo.DeleteItem(ctx, existing.ID); err != nil {
			return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
		}
		return &favoritemodel.FavoriteToggleResp{Favorited: false}, nil
	}

	// Add
	item := &favoritemodel.FavoriteItem{
		FavoriteID: favoriteID,
		VideoID:    videoID,
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("favorite.service.ToggleFavoriteItem: %w", err)
	}
	return &favoritemodel.FavoriteToggleResp{Favorited: true}, nil
}

func (s *Service) CreateDefaultFavorite(ctx context.Context, userID uint) error {
	existing, err := s.repo.FindDefaultFavorite(ctx, userID)
	if err != nil {
		return fmt.Errorf("favorite.service.CreateDefaultFavorite: %w", err)
	}
	if existing != nil {
		return nil
	}

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

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}
