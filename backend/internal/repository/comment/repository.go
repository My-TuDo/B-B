package comment

import (
	"context"
	"fmt"

	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *commentmodel.Comment) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("comment.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*commentmodel.Comment, error) {
	var c commentmodel.Comment
	if err := r.db.WithContext(ctx).Preload("User").First(&c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("comment.repository.FindByID: %w", err)
	}
	return &c, nil
}

func (r *Repository) FindRootComments(ctx context.Context, videoID uint, sort string, offset, limit int) ([]commentmodel.Comment, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&commentmodel.Comment{}).Where("video_id = ? AND parent_id = 0", videoID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("comment.repository.FindRootComments.Count: %w", err)
	}

	orderBy := "created_at DESC"
	if sort == "hot" {
		orderBy = "likes DESC"
	}

	var list []commentmodel.Comment
	if err := query.Preload("User").Order(orderBy).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("comment.repository.FindRootComments.Find: %w", err)
	}
	return list, total, nil
}

func (r *Repository) FindRepliesByRootIDs(ctx context.Context, rootIDs []uint) ([]commentmodel.Comment, error) {
	if len(rootIDs) == 0 {
		return nil, nil
	}
	var list []commentmodel.Comment
	if err := r.db.WithContext(ctx).Preload("User").Where("root_id IN ?", rootIDs).Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("comment.repository.FindRepliesByRootIDs: %w", err)
	}
	return list, nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&commentmodel.Comment{}, id).Error; err != nil {
		return fmt.Errorf("comment.repository.Delete: %w", err)
	}
	return nil
}

func (r *Repository) UpdateLikes(ctx context.Context, id uint, likes uint) error {
	if err := r.db.WithContext(ctx).Model(&commentmodel.Comment{}).Where("id = ?", id).Update("likes", likes).Error; err != nil {
		return fmt.Errorf("comment.repository.UpdateLikes: %w", err)
	}
	return nil
}

func (r *Repository) FindVideoAuthor(ctx context.Context, videoID uint, authorID *uint) error {
	if err := r.db.WithContext(ctx).Table("videos").Select("user_id").Where("id = ?", videoID).Scan(authorID).Error; err != nil {
		return fmt.Errorf("comment.repository.FindVideoAuthor: %w", err)
	}
	return nil
}

func (r *Repository) ExistsVideo(ctx context.Context, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("videos").Where("id = ?", videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("comment.repository.ExistsVideo: %w", err)
	}
	return count > 0, nil
}
