package category

import (
	"context"
	"fmt"

	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll(ctx context.Context) ([]categorymodel.Category, error) {
	var categories []categorymodel.Category
	result := r.db.WithContext(ctx).Order("id ASC").Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("category.repository.FindAll: %w", result.Error)
	}
	return categories, nil
}
