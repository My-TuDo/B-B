package video

import (
	"context"
	"fmt"

	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, video *videomodel.Video) error {
	result := r.db.WithContext(ctx).Create(video)
	if result.Error != nil {
		return fmt.Errorf("video.repository.Create: %w", result.Error)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*videomodel.Video, error) {
	var video videomodel.Video
	result := r.db.WithContext(ctx).Preload("User").First(&video, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("video.repository.FindByID: %w", result.Error)
	}
	return &video, nil
}

func (r *Repository) Update(ctx context.Context, video *videomodel.Video) error {
	result := r.db.WithContext(ctx).Save(video)
	if result.Error != nil {
		return fmt.Errorf("video.repository.Update: %w", result.Error)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, page, pageSize int, categoryID uint) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("status = ?", 1)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.List.Count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.List.Find: %w", err)
	}

	return videos, total, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uint, status *int8, page, pageSize int) ([]videomodel.Video, int64, error) {
	var videos []videomodel.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	} else {
		query = query.Where("status != ?", 3)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByUser.Count: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&videos).Error; err != nil {
		return nil, 0, fmt.Errorf("video.repository.ListByUser.Find: %w", err)
	}

	return videos, total, nil
}

func (r *Repository) IncrementViews(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&videomodel.Video{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1"))
	if result.Error != nil {
		return fmt.Errorf("video.repository.IncrementViews: %w", result.Error)
	}
	return nil
}
