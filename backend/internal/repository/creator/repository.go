// Package creator 提供创作者数据访问层，封装创作者中心相关的数据库操作。
// 支持查询创作者的视频列表和统计数据（总播放量、视频数、今日播放量）。
package creator

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

// Repository 创作者数据仓库，封装创作者相关的数据库查询操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建创作者数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListByUser 查询指定用户的视频列表，按创建时间倒序分页返回。
// 用于创作者管理后台展示自己的作品列表。
func (r *Repository) ListByUser(ctx context.Context, userID uint, offset, limit int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	// 筛选属于该用户的视频
	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id = ?", userID)

	// 统计视频总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("creator.repository.ListByUser.Count: %w", err)
	}

	// 预加载用户信息，按时间倒序分页查询
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("creator.repository.ListByUser.Find: %w", err)
	}

	return videos, total, nil
}

// GetStats 获取创作者的统计数据：总播放量、视频总数、今日新增播放量。
// 排除状态为 3（已删除）的视频。
// 返回值依次为：totalViews（总播放量）、totalVideos（视频总数）、todayViews（今日播放量）、err。
func (r *Repository) GetStats(ctx context.Context, userID uint) (totalViews uint64, totalVideos int64, todayViews uint64, err error) {
	// 统计用户视频总数和总播放量（排除已删除视频）
	row := r.db.WithContext(ctx).Table("videos").
		Where("user_id = ? AND status != ?", userID, 3).
		Select("COUNT(*) as total_videos, COALESCE(SUM(views), 0) as total_views").
		Row()
	if err := row.Scan(&totalVideos, &totalViews); err != nil {
		return 0, 0, 0, fmt.Errorf("creator.repository.GetStats: %w", err)
	}

	// 统计今日创建的视频的播放量总和
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COALESCE(SUM(views), 0) FROM videos WHERE user_id = ? AND status != 3 AND DATE(created_at) = CURDATE()",
		userID,
	).Scan(&todayViews).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("creator.repository.GetStats.TodayViews: %w", err)
	}

	return totalViews, totalVideos, todayViews, nil
}
