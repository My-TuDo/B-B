package user

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("user.repository.FindByID: %w", result.Error)
	}
	return &user, nil
}

func (r *Repository) Update(ctx context.Context, user *usermodel.User) error {
	result := r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"bio":      user.Bio,
	})
	if result.Error != nil {
		return fmt.Errorf("user.repository.Update: %w", result.Error)
	}
	return nil
}
