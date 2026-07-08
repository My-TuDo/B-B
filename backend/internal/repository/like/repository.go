package like

import (
	"context"
	"fmt"

	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, l *likemodel.VideoLike) error {
	if err := r.db.WithContext(ctx).Create(l).Error; err != nil {
		return fmt.Errorf("like.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, videoID uint) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&likemodel.VideoLike{}).Error; err != nil {
		return fmt.Errorf("like.repository.Delete: %w", err)
	}
	return nil
}

func (r *Repository) Exists(ctx context.Context, userID, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&likemodel.VideoLike{}).Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("like.repository.Exists: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) CountByVideoID(ctx context.Context, videoID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&likemodel.VideoLike{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("like.repository.CountByVideoID: %w", err)
	}
	return count, nil
}

func (r *Repository) Upsert(ctx context.Context, l *likemodel.VideoLike) error {
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(l).Error; err != nil {
		return fmt.Errorf("like.repository.Upsert: %w", err)
	}
	return nil
}

func (r *Repository) FindVideoAuthor(ctx context.Context, videoID uint) (uint, error) {
	var authorID uint
	if err := r.db.WithContext(ctx).Table("videos").Select("user_id").Where("id = ?", videoID).Scan(&authorID).Error; err != nil {
		return 0, fmt.Errorf("like.repository.FindVideoAuthor: %w", err)
	}
	return authorID, nil
}

func (r *Repository) FindUserName(ctx context.Context, userID uint) (string, error) {
	var name string
	if err := r.db.WithContext(ctx).Table("users").Select("COALESCE(NULLIF(nickname,''), username)").Where("id = ?", userID).Scan(&name).Error; err != nil {
		return "", fmt.Errorf("like.repository.FindUserName: %w", err)
	}
	return name, nil
}
