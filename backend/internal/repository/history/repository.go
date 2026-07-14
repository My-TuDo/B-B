// Package history 提供观看历史数据访问层，封装用户视频观看历史相关的数据库操作。
// 支持创建/更新观看记录（Upsert 模式）、查询单个记录和分页列表。
package history

import (
	"context"
	"fmt"
	"time"

	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 观看历史数据仓库，封装观看历史相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建观看历史数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateOrUpdate 创建或更新观看历史记录（Upsert 模式）。
// 若用户对同一视频已有观看记录，则更新播放进度和观看时间；
// 否则创建新记录。使用 GORM 的 OnConflict 子句实现冲突时更新。
func (r *Repository) CreateOrUpdate(ctx context.Context, h *historymodel.VideoHistory) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "video_id"}}, // 冲突键：用户+视频唯一
		DoUpdates: clause.AssignmentColumns([]string{"progress", "watched_at"}), // 冲突时更新进度和时间
	}).Create(h)
	if result.Error != nil {
		return fmt.Errorf("history.repository.CreateOrUpdate: %w", result.Error)
	}
	return nil
}

// FindByUserAndVideo 查询用户对特定视频的观看历史记录。
// 若未找到则返回 (nil, nil)，用于获取上次观看进度。
func (r *Repository) FindByUserAndVideo(ctx context.Context, userID, videoID uint) (*historymodel.VideoHistory, error) {
	var h historymodel.VideoHistory
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		First(&h)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("history.repository.FindByUserAndVideo: %w", result.Error)
	}
	return &h, nil
}

// ListByUser 分页查询用户的观看历史列表，按观看时间倒序排列。
// 预加载视频信息和视频作者信息，用于前端渲染历史列表。
func (r *Repository) ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]historymodel.VideoHistory, int64, error) {
	var histories []historymodel.VideoHistory
	var total int64

	// 构建用户观看历史查询
	query := r.db.WithContext(ctx).Model(&historymodel.VideoHistory{}).Where("user_id = ?", userID)

	// 统计总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("history.repository.ListByUser.Count: %w", err)
	}

	// 计算分页偏移量，预加载视频及视频作者，按观看时间倒序
	offset := (page - 1) * pageSize
	if err := query.Preload("Video").Preload("Video.User").
		Order("watched_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&histories).Error; err != nil {
		return nil, 0, fmt.Errorf("history.repository.ListByUser.Find: %w", err)
	}

	// 若 watched_at 为零值（刚创建），设置为当前时间以保证前端显示正常
	now := time.Now()
	for i := range histories {
		if histories[i].WatchedAt.IsZero() {
			histories[i].WatchedAt = now
		}
	}

	return histories, total, nil
}
