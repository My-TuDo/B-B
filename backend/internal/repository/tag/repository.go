// Package tag 提供标签数据访问层，封装视频标签相关的数据库操作。
// 支持标签的创建、查询、按名称查找，以及视频与标签的关联管理（批量替换、查询关联标签）。
package tag

import (
	"context"
	"fmt"

	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 标签数据仓库，封装标签相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建标签数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一个新标签。
func (r *Repository) Create(ctx context.Context, tag *tagmodel.Tag) error {
	result := r.db.WithContext(ctx).Create(tag)
	if result.Error != nil {
		return fmt.Errorf("tag.repository.Create: %w", result.Error)
	}
	return nil
}

// FindAll 查询所有标签，按 ID 升序排列。
// 用于管理后台或上传页面的标签选择器。
func (r *Repository) FindAll(ctx context.Context) ([]tagmodel.Tag, error) {
	var tags []tagmodel.Tag
	result := r.db.WithContext(ctx).Order("id ASC").Find(&tags)
	if result.Error != nil {
		return nil, fmt.Errorf("tag.repository.FindAll: %w", result.Error)
	}
	return tags, nil
}

// FindByName 根据标签名称精确查找标签。
// 若未找到则返回 (nil, nil)，用于判断标签是否已存在，避免重复创建。
func (r *Repository) FindByName(ctx context.Context, name string) (*tagmodel.Tag, error) {
	var tag tagmodel.Tag
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&tag)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("tag.repository.FindByName: %w", result.Error)
	}
	return &tag, nil
}

// ReplaceVideoTags 替换视频的标签关联（先删后插策略）。
// 先删除该视频的所有旧标签关联，再逐条创建新的关联。
// 使用 OnConflict{DoNothing: true} 防止重复关联报错。
func (r *Repository) ReplaceVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	tx := r.db.WithContext(ctx)
	// 第一步：删除该视频现有的所有标签关联
	if err := tx.Where("video_id = ?", videoID).Delete(&tagmodel.VideoTag{}).Error; err != nil {
		return fmt.Errorf("tag.repository.ReplaceVideoTags.delete: %w", err)
	}
	// 第二步：逐条插入新的标签关联
	for _, tagID := range tagIDs {
		vt := tagmodel.VideoTag{VideoID: videoID, TagID: tagID}
		// OnConflict 保证即使存在重复也不会报错
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&vt).Error; err != nil {
			return fmt.Errorf("tag.repository.ReplaceVideoTags.create: %w", err)
		}
	}
	return nil
}

// GetVideoTags 查询指定视频关联的所有标签。
// 通过 JOIN video_tags 中间表获取标签列表，按标签 ID 升序排列。
func (r *Repository) GetVideoTags(ctx context.Context, videoID uint) ([]tagmodel.Tag, error) {
	var tags []tagmodel.Tag
	result := r.db.WithContext(ctx).
		Joins("JOIN video_tags ON video_tags.tag_id = tags.id").
		Where("video_tags.video_id = ?", videoID).
		Order("tags.id ASC").
		Find(&tags)
	if result.Error != nil {
		return nil, fmt.Errorf("tag.repository.GetVideoTags: %w", result.Error)
	}
	return tags, nil
}
