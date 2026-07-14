// Package follow 提供关注数据访问层，封装用户关注/粉丝相关的数据库操作。
// 支持创建和删除关注关系、查询粉丝和关注列表、统计关注数和粉丝数等。
package follow

import (
	"context"
	"fmt"

	followmodel "github.com/My-TuDo/B-B/backend/internal/model/follow"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"gorm.io/gorm"
)

// Repository 关注数据仓库，封装关注相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建关注数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条关注记录。
// followerID 关注 followingID，建立单向关注关系。
func (r *Repository) Create(ctx context.Context, f *followmodel.Follow) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return fmt.Errorf("follow.repository.Create: %w", err)
	}
	return nil
}

// Delete 删除关注关系（取关操作）。
// 根据关注者和被关注者 ID 删除对应的关注记录。
func (r *Repository) Delete(ctx context.Context, followerID, followingID uint) error {
	if err := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&followmodel.Follow{}).Error; err != nil {
		return fmt.Errorf("follow.repository.Delete: %w", err)
	}
	return nil
}

// Exists 判断两个用户之间是否存在关注关系。
func (r *Repository) Exists(ctx context.Context, followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("follow.repository.Exists: %w", err)
	}
	return count > 0, nil
}

// CountFollowers 统计指定用户的粉丝数量（有多少人关注了该用户）。
func (r *Repository) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("following_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountFollowers: %w", err)
	}
	return count, nil
}

// CountFollowing 统计指定用户关注了多少人。
func (r *Repository) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("follower_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountFollowing: %w", err)
	}
	return count, nil
}

// CountVideosByUser 统计用户发布的公开视频数量（status = 1 表示已审核通过）。
func (r *Repository) CountVideosByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&struct{ ID uint }{}).Table("videos").Where("user_id = ? AND status = 1", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountVideosByUser: %w", err)
	}
	return count, nil
}

// FindFollowers 分页查询用户的粉丝列表。
// 返回粉丝用户列表、粉丝总数和可能的错误。
func (r *Repository) FindFollowers(ctx context.Context, userID uint, offset, limit int) ([]usermodel.User, int64, error) {
	var total int64
	var ids []uint

	// 构建子查询：按关注时间倒序获取粉丝 ID 列表
	subQuery := r.db.WithContext(ctx).Model(&followmodel.Follow{}).
		Where("following_id = ?", userID).
		Order("created_at DESC")

	// 统计粉丝总数
	if err := subQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Count: %w", err)
	}

	// 分页获取粉丝用户 ID
	if err := subQuery.Offset(offset).Limit(limit).Pluck("follower_id", &ids).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Pluck: %w", err)
	}

	// 无粉丝时提前返回
	if len(ids) == 0 {
		return nil, total, nil
	}

	// 根据 ID 列表批量查询用户信息
	var users []usermodel.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Users: %w", err)
	}
	return users, total, nil
}

// FindFollowing 分页查询用户关注的人列表。
// 返回关注用户列表、关注总数和可能的错误。
func (r *Repository) FindFollowing(ctx context.Context, userID uint, offset, limit int) ([]usermodel.User, int64, error) {
	var total int64
	var ids []uint

	// 构建子查询：按关注时间倒序获取关注者 ID 列表
	subQuery := r.db.WithContext(ctx).Model(&followmodel.Follow{}).
		Where("follower_id = ?", userID).
		Order("created_at DESC")

	if err := subQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowing.Count: %w", err)
	}

	if err := subQuery.Offset(offset).Limit(limit).Pluck("following_id", &ids).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowing.Pluck: %w", err)
	}

	if len(ids) == 0 {
		return nil, total, nil
	}

	var users []usermodel.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowing.Users: %w", err)
	}
	return users, total, nil
}

// FindFeed 根据用户关注列表获取关注用户的视频动态（待实现）。
// 当前为占位方法，返回空结果。
func (r *Repository) FindFeed(ctx context.Context, userID uint, offset, limit int) ([]struct {
	ID         uint
	UserID     uint
	Title      string
	CoverURL   string
	VideoURL   string
	Duration   uint
	FileSize   uint64
	CategoryID uint
	Status     int8
	Views      uint64
	CreatedAt  interface{}
	UpdatedAt  interface{}
}, int64, error) {
	return nil, 0, nil
}

// FindFollowingIDs 获取用户关注的所有用户 ID 列表。
// 用于构建关注动态流等场景，一次性获取所有关注对象的 ID。
func (r *Repository) FindFollowingIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).
		Where("follower_id = ?", userID).
		Pluck("following_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("follow.repository.FindFollowingIDs: %w", err)
	}
	return ids, nil
}
