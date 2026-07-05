package auth

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

func (r *Repository) Create(ctx context.Context, user *usermodel.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("auth.repository.Create: %w", result.Error)
	}
	return nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("auth.repository.FindByUsername: %w", result.Error)
	}
	return &user, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("auth.repository.FindByEmail: %w", result.Error)
	}
	return &user, nil
}

func (r *Repository) FindByUsernameOrEmail(ctx context.Context, account string) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("username = ? OR email = ?", account, account).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("auth.repository.FindByUsernameOrEmail: %w", result.Error)
	}
	return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("auth.repository.FindByID: %w", result.Error)
	}
	return &user, nil
}
