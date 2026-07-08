package tag

import (
	"context"
	"fmt"

	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, tag *tagmodel.Tag) error {
	result := r.db.WithContext(ctx).Create(tag)
	if result.Error != nil {
		return fmt.Errorf("tag.repository.Create: %w", result.Error)
	}
	return nil
}

func (r *Repository) FindAll(ctx context.Context) ([]tagmodel.Tag, error) {
	var tags []tagmodel.Tag
	result := r.db.WithContext(ctx).Order("id ASC").Find(&tags)
	if result.Error != nil {
		return nil, fmt.Errorf("tag.repository.FindAll: %w", result.Error)
	}
	return tags, nil
}

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

func (r *Repository) ReplaceVideoTags(ctx context.Context, videoID uint, tagIDs []uint) error {
	tx := r.db.WithContext(ctx)
	// Delete existing
	if err := tx.Where("video_id = ?", videoID).Delete(&tagmodel.VideoTag{}).Error; err != nil {
		return fmt.Errorf("tag.repository.ReplaceVideoTags.delete: %w", err)
	}
	// Insert new
	for _, tagID := range tagIDs {
		vt := tagmodel.VideoTag{VideoID: videoID, TagID: tagID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&vt).Error; err != nil {
			return fmt.Errorf("tag.repository.ReplaceVideoTags.create: %w", err)
		}
	}
	return nil
}

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
