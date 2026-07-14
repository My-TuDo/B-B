// Package favorite 提供收藏夹数据访问层，封装收藏夹和收藏项相关的数据库操作。
// 支持创建收藏夹、管理收藏项（添加/删除视频）、查询收藏夹内容和默认收藏夹等。
package favorite

import (
	"context"
	"fmt"

	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

// Repository 收藏夹数据仓库，封装收藏夹相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建收藏夹数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateFavorite 创建一个新收藏夹。
func (r *Repository) CreateFavorite(ctx context.Context, f *favoritemodel.Favorite) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return fmt.Errorf("favorite.repository.CreateFavorite: %w", err)
	}
	return nil
}

// FindFavoriteByID 根据收藏夹 ID 查找收藏夹。
// 若未找到则返回 (nil, nil)。
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

// FindFavoritesByUserID 查询用户的所有收藏夹，按创建时间倒序排列。
func (r *Repository) FindFavoritesByUserID(ctx context.Context, userID uint) ([]favoritemodel.Favorite, error) {
	var list []favoritemodel.Favorite
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("favorite.repository.FindFavoritesByUserID: %w", err)
	}
	return list, nil
}

// CountItems 统计指定收藏夹中的视频数量。
func (r *Repository) CountItems(ctx context.Context, favoriteID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).Where("favorite_id = ?", favoriteID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("favorite.repository.CountItems: %w", err)
	}
	return count, nil
}

// FindItem 查询收藏夹中是否已存在指定的视频项。
// 用于判断视频是否已被收藏到该收藏夹，避免重复收藏。
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

// CreateItem 向收藏夹中添加一个视频项。
func (r *Repository) CreateItem(ctx context.Context, item *favoritemodel.FavoriteItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("favorite.repository.CreateItem: %w", err)
	}
	return nil
}

// DeleteItem 从收藏夹中移除一个视频项（硬删除）。
func (r *Repository) DeleteItem(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&favoritemodel.FavoriteItem{}, id).Error; err != nil {
		return fmt.Errorf("favorite.repository.DeleteItem: %w", err)
	}
	return nil
}

// FindItemsByFavoriteID 分页查询指定收藏夹中的视频列表。
// 先通过收藏项表获取视频 ID 列表（保持添加顺序），再批量查询视频详情。
// 返回的视频列表保持与收藏项一致的顺序。
func (r *Repository) FindItemsByFavoriteID(ctx context.Context, favoriteID uint, offset, limit int) ([]videomodel.Video, int64, error) {
	var total int64
	// 统计收藏夹内视频总数
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).Where("favorite_id = ?", favoriteID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Count: %w", err)
	}

	// 按创建时间倒序获取分页后的视频 ID 列表
	var itemIDs []uint
	if err := r.db.WithContext(ctx).Model(&favoritemodel.FavoriteItem{}).
		Where("favorite_id = ?", favoriteID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Pluck("video_id", &itemIDs).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Pluck: %w", err)
	}

	// 无收藏项时提前返回
	if len(itemIDs) == 0 {
		return nil, total, nil
	}

	// 批量查询视频详情（含用户信息）
	var videos []videomodel.Video
	if err := r.db.WithContext(ctx).Preload("User").Where("id IN ?", itemIDs).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("favorite.repository.FindItems.Videos: %w", err)
	}

	// 保持视频顺序与 itemIDs 一致（收藏项添加顺序）
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

// FindDefaultFavorite 查找用户的默认收藏夹（名称为"默认收藏夹"）。
// 新用户注册后自动创建的默认收藏夹，用于快速收藏。
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

// FindFirstCover 获取收藏夹中第一个视频的封面 URL，作为收藏夹的展示封面。
// 按添加时间升序取第一个收藏项，再查对应视频的封面地址。
func (r *Repository) FindFirstCover(ctx context.Context, favoriteID uint) (string, error) {
	var item favoritemodel.FavoriteItem
	if err := r.db.WithContext(ctx).Where("favorite_id = ?", favoriteID).Order("created_at ASC").First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", fmt.Errorf("favorite.repository.FindFirstCover: %w", err)
	}

	// 查询对应视频的封面 URL
	var video videomodel.Video
	if err := r.db.WithContext(ctx).Select("cover_url").First(&video, item.VideoID).Error; err != nil {
		return "", nil
	}
	return video.CoverURL, nil
}
