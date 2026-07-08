package favorite

import (
	"context"
	"fmt"

	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateFavorite(ctx context.Context, f *favoritemodel.Favorite) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return fmt.Errorf("favorite.repository.CreateFavorite: %w", err)
	}
	return nil
}

func (r *Repository) FindFavoriteByID(ctx context.Context, id uint) (*favoritemodel.Favorite, error) {
	var f favoritemodel.Favorite
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("favorite.repository.FindFavoriteByID: %w", err)
	}
	return &f, nil
}

func (r *Repository) FindFavoritesByUserID(ctx context.Context, userID uint) ([]favoritemodel.Favorite, error) {
	var list []favoritemodel.Favorite
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("favorite.repository.FindFavoritesByUserID: %w", err)
	}
	return list, nil
}

func (r *Repository) CountItems(ctx context.Context, favoriteID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).Where("favorite_id = ?", favoriteID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("favorite.repository.CountItems: %w", err)
	}
	return count, nil
}

func (r *Repository) FindItem(ctx context.Context, favoriteID, videoID uint) (*favoritemodel.FavoriteItem, error) {
	var item favoritemodel.FavoriteItem
	if err := r.db.WithContext(ctx).Where("favorite_id = ? AND video_id = ?", favoriteID, videoID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("favorite.repository.FindItem: %w", err)
	}
	return &item, nil
}

func (r *Repository) CreateItem(ctx context.Context, item *favoritemodel.FavoriteItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("favorite.repository.CreateItem: %w", err)
	}
	return nil
}

func (r *Repository) DeleteItem(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&favoritemodel.FavoriteItem{}, id).Error; err != nil {
		return fmt.Errorf("favorite.repository.DeleteItem: %w", err)
	}
	return nil
}

func (r *Repository) FindItemsByFavoriteID(ctx context.Context, favoriteID uint, offset, limit int) ([]videomodel.Video, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).Where("favorite_id = ?", favoriteID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Count: %w", err)
	}

	var itemIDs []uint
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).
		Where("favorite_id = ?", favoriteID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Pluck("video_id", &itemIDs).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Pluck: %w", err)
	}

	if len(itemIDs) == 0 {
		return nil, total, nil
	}

	var videos []videomodel.Video
	if err := r.db.WithContext(ctx).Preload("User").Where("id IN ?", itemIDs).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Videos: %w", err)
	}

	// Preserve order from itemIDs
	ordered := make([]videomodel.Video, 0, len(itemIDs))
	vidMap := make(map[uint]videomodel.Video, len(videos))
	for _, v := range videos {
		vidMap[v.ID] = v
	}
	for _, id := range itemIDs {
		if v, ok := vidMap[id]; ok {
			ordered = append(ordered, v)
		}
	}

	return ordered, total, nil
}

func (r *Repository) FindDefaultFavorite(ctx context.Context, userID uint) (*favoritemodel.Favorite, error) {
	var f favoritemodel.Favorite
	if err := r.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, "默认收藏夹").First(&f).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("favorite.repository.FindDefaultFavorite: %w", err)
	}
	return &f, nil
}

func (r *Repository) FindFirstCover(ctx context.Context, favoriteID uint) (string, error) {
	var item favoritemodel.FavoriteItem
	if err := r.db.WithContext(ctx).Where("favorite_id = ?", favoriteID).Order("created_at ASC").First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", fmt.Errorf("favorite.repository.FindFirstCover: %w", err)
	}

	var video videomodel.Video
	if err := r.db.WithContext(ctx).Select("cover_url").First(&video, item.VideoID).Error; err != nil {
		return "", nil
	}
	return video.CoverURL, nil
}
