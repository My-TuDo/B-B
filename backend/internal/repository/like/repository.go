// Package like 提供点赞数据访问层，封装视频点赞相关的数据库操作。
// 支持创建和删除点赞记录、判断点赞状态、统计点赞数，以及 Upsert 操作。
// 同时提供查询视频作者和用户名称的辅助方法，用于点赞通知。
package like

import (
	"context"
	"fmt"

	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 点赞数据仓库，封装点赞相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建点赞数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条点赞记录。
func (r *Repository) Create(ctx context.Context, l *likemodel.VideoLike) error {
	if err := r.db.WithContext(ctx).Create(l).Error; err != nil {
		return fmt.Errorf("like.repository.Create: %w", err)
	}
	return nil
}

// Delete 删除点赞记录（取消点赞）。
// 根据用户 ID 和视频 ID 定位并删除对应的点赞记录。
func (r *Repository) Delete(ctx context.Context, userID, videoID uint) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&likemodel.VideoLike{}).Error; err != nil {
		return fmt.Errorf("like.repository.Delete: %w", err)
	}
	return nil
}

// Exists 判断用户是否已点赞指定视频。
func (r *Repository) Exists(ctx context.Context, userID, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&likemodel.VideoLike{}).Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("like.repository.Exists: %w", err)
	}
	return count > 0, nil
}

// CountByVideoID 统计指定视频的总点赞数。
func (r *Repository) CountByVideoID(ctx context.Context, videoID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&likemodel.VideoLike{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("like.repository.CountByVideoID: %w", err)
	}
	return count, nil
}

// Upsert 插入或忽略点赞记录（幂等操作）。
// 使用 OnConflict{DoNothing: true}，若已存在则静默跳过，避免重复点赞报错。
func (r *Repository) Upsert(ctx context.Context, l *likemodel.VideoLike) error {
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(l).Error; err != nil {
		return fmt.Errorf("like.repository.Upsert: %w", err)
	}
	return nil
}

// FindVideoAuthor 查询指定视频的作者 ID。
// 用于点赞通知场景：需要知道视频作者是谁以便发送通知。
func (r *Repository) FindVideoAuthor(ctx context.Context, videoID uint) (uint, error) {
	var authorID uint
	if err := r.db.WithContext(ctx).Table("videos").Select("user_id").Where("id = ?", videoID).Scan(&authorID).Error; err != nil {
		return 0, fmt.Errorf("like.repository.FindVideoAuthor: %w", err)
	}
	return authorID, nil
}

// FindUserName 获取用户的显示名称（优先使用昵称，无昵称则使用用户名）。
// 使用 COALESCE(NULLIF(...)) 实现：nickname 为空字符串时回退到 username。
func (r *Repository) FindUserName(ctx context.Context, userID uint) (string, error) {
	var name string
	if err := r.db.WithContext(ctx).Table("users").Select("COALESCE(NULLIF(nickname,''), username)").Where("id = ?", userID).Scan(&name).Error; err != nil {
		return "", fmt.Errorf("like.repository.FindUserName: %w", err)
	}
	return name, nil
}
