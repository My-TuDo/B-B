// Package auth 提供认证数据访问层，封装用户注册、登录等认证相关的数据库操作。
// 支持通过用户名、邮箱、ID 等多种方式查找用户，是认证系统与数据库之间的桥梁。
package auth

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"gorm.io/gorm"
)

// Repository 认证数据仓库，封装用户认证相关的数据库操作。
// 通过持有的 gorm.DB 实例进行用户表的 CRUD 操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建认证数据仓库实例。
// 接收一个已初始化的 gorm.DB 指针，返回可用的 Repository 指针。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 在数据库中创建新用户记录。
// 使用 GORM 的 Create 方法插入用户数据，若用户名或邮箱重复会返回数据库错误。
func (r *Repository) Create(ctx context.Context, user *usermodel.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("auth.repository.Create: %w", result.Error)
	}
	return nil
}

// FindByUsername 根据用户名精确查找用户。
// 若未找到匹配记录则返回 (nil, nil)，其他数据库错误会包装后返回。
func (r *Repository) FindByUsername(ctx context.Context, username string) (*usermodel.User, error) {
	var user usermodel.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		// 记录不存在视为正常情况，返回 nil
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("auth.repository.FindByUsername: %w", result.Error)
	}
	return &user, nil
}

// FindByEmail 根据邮箱地址精确查找用户。
// 若未找到匹配记录则返回 (nil, nil)，其他数据库错误会包装后返回。
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

// FindByUsernameOrEmail 根据用户名或邮箱查找用户（二者任一匹配即返回）。
// 支持用户使用用户名或邮箱作为登录凭据进行统一查找。
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

// FindByID 根据用户主键 ID 查找用户。
// 用于通过已解析的用户 ID（如 JWT 中的 subject）快速获取用户信息。
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
