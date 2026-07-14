// Package interaction 提供互动状态数据访问层，封装用户与视频间各种互动状态的查询操作。
// 支持判断用户是否已点赞、已收藏、已关注，以及统计投币数量。
// 该包为前端展示用户的互动状态（如按钮高亮）提供数据支撑。
package interaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository 互动数据仓库，封装跨表的互动状态查询操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建互动数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// IsLiked 判断用户是否已点赞指定视频。
// 直接查询 video_likes 表，存在记录即为已点赞。
func (r *Repository) IsLiked(ctx context.Context, userID, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("video_likes").Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("interaction.repository.IsLiked: %w", err)
	}
	return count > 0, nil
}

// CountCoins 统计用户对指定视频的总投币数（SUM 聚合 count 字段）。
// 用户可能多次投币，每次投币数量可不同，因此用 SUM 而非 COUNT。
func (r *Repository) CountCoins(ctx context.Context, userID, videoID uint) (int64, error) {
	var sum int64
	if err := r.db.WithContext(ctx).Table("video_coins").Where("user_id = ? AND video_id = ?", userID, videoID).Select("COALESCE(SUM(count), 0)").Scan(&sum).Error; err != nil {
		return 0, fmt.Errorf("interaction.repository.CountCoins: %w", err)
	}
	return sum, nil
}

// IsFavorited 判断用户是否已将指定视频收藏到任意收藏夹。
// 通过 JOIN favorites 和 favorite_items 表来关联查询。
func (r *Repository) IsFavorited(ctx context.Context, userID, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("favorite_items fi").
		Joins("JOIN favorites f ON f.id = fi.favorite_id").
		Where("f.user_id = ? AND fi.video_id = ?", userID, videoID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("interaction.repository.IsFavorited: %w", err)
	}
	return count > 0, nil
}

// IsFollowing 判断 followerID 是否关注了 followingID。
// 直接查询 follows 表，存在记录即为已关注。
func (r *Repository) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("follows").Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("interaction.repository.IsFollowing: %w", err)
	}
	return count > 0, nil
}
