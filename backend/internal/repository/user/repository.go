// Package user 提供用户数据访问层，封装用户信息相关的数据库操作。
// 支持根据 ID 查找用户和更新用户个人资料（昵称、头像、简介）。
package user

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"gorm.io/gorm"
)

// Repository 用户数据仓库，封装用户信息相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建用户数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByID 根据用户主键 ID 查找用户信息。
// 若未找到则返回 (nil, nil)，用于获取用户个人主页信息。
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

// Update 更新用户的个人资料信息（昵称、头像、简介）。
// 使用 map[string]interface{} 指定仅更新这三个字段，避免意外覆盖其他字段。
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
