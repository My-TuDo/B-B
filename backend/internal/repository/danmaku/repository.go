// Package danmaku 提供弹幕数据访问层，封装视频弹幕相关的数据库操作。
// 支持创建弹幕和按视频 ID 查询弹幕列表，弹幕按播放时间升序排列以匹配视频进度。
package danmaku

import (
	"context"
	"fmt"

	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	"gorm.io/gorm"
)

// Repository 弹幕数据仓库，封装弹幕相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建弹幕数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条弹幕记录。
// 弹幕包含文本内容、播放时间位置、颜色等信息。
func (r *Repository) Create(ctx context.Context, d *danmakumodel.Danmaku) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("danmaku.repository.Create: %w", err)
	}
	return nil
}

// FindByVideoID 根据视频 ID 查询该视频的所有弹幕。
// 弹幕按播放时间（play_time）升序排列，保证与视频播放进度同步。
// 预加载用户信息，用于前端显示发送者昵称和头像。
func (r *Repository) FindByVideoID(ctx context.Context, videoID uint) ([]danmakumodel.Danmaku, error) {
	var list []danmakumodel.Danmaku
	if err := r.db.WithContext(ctx).Preload("User").Where("video_id = ?", videoID).Order("play_time ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("danmaku.repository.FindByVideoID: %w", err)
	}
	return list, nil
}
