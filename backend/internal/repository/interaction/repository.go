package interaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IsLiked(ctx context.Context, userID, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("video_likes").Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("interaction.repository.IsLiked: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) CountCoins(ctx context.Context, userID, videoID uint) (int64, error) {
	var sum int64
	if err := r.db.WithContext(ctx).Table("video_coins").Where("user_id = ? AND video_id = ?", userID, videoID).Select("COALESCE(SUM(count), 0)").Scan(&sum).Error; err != nil {
		return 0, fmt.Errorf("interaction.repository.CountCoins: %w", err)
	}
	return sum, nil
}

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

func (r *Repository) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("follows").Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("interaction.repository.IsFollowing: %w", err)
	}
	return count > 0, nil
}
