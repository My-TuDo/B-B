// Package coin 提供投币数据访问层，封装用户对视频投币相关的数据库操作。
// 支持创建投币记录、统计视频投币总数、查询用户对特定视频的投币情况。
package coin

import (
	"context"
	"fmt"

	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	"gorm.io/gorm"
)

// Repository 投币数据仓库，封装投币相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建投币数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条投币记录。
// 记录用户对指定视频的投币信息（含投币数量）。
func (r *Repository) Create(ctx context.Context, c *coinmodel.VideoCoin) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("coin.repository.Create: %w", err)
	}
	return nil
}

// CountByVideoID 统计指定视频的总投币数（按记录条数统计）。
// 用于展示视频的投币热度。
func (r *Repository) CountByVideoID(ctx context.Context, videoID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&coinmodel.VideoCoin{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("coin.repository.CountByVideoID: %w", err)
	}
	return count, nil
}

// FindByUserAndVideo 查找用户对特定视频的投币记录。
// 若未找到则返回 (nil, nil)，用于判断用户是否已对该视频投过币。
func (r *Repository) FindByUserAndVideo(ctx context.Context, userID, videoID uint) (*coinmodel.VideoCoin, error) {
	var c coinmodel.VideoCoin
	if err := r.db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("coin.repository.FindByUserAndVideo: %w", err)
	}
	return &c, nil
}
