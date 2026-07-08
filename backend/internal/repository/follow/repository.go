package follow

import (
	"context"
	"fmt"

	followmodel "github.com/My-TuDo/B-B/backend/internal/model/follow"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, f *followmodel.Follow) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return fmt.Errorf("follow.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, followerID, followingID uint) error {
	if err := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&followmodel.Follow{}).Error; err != nil {
		return fmt.Errorf("follow.repository.Delete: %w", err)
	}
	return nil
}

func (r *Repository) Exists(ctx context.Context, followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("follow.repository.Exists: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("following_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountFollowers: %w", err)
	}
	return count, nil
}

func (r *Repository) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).Where("follower_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountFollowing: %w", err)
	}
	return count, nil
}

func (r *Repository) CountVideosByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&struct{ ID uint }{}).Table("videos").Where("user_id = ? AND status = 1", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("follow.repository.CountVideosByUser: %w", err)
	}
	return count, nil
}

func (r *Repository) FindFollowers(ctx context.Context, userID uint, offset, limit int) ([]usermodel.User, int64, error) {
	var total int64
	var ids []uint

	subQuery := r.db.WithContext(ctx).Model(&followmodel.Follow{}).
		Where("following_id = ?", userID).
		Order("created_at DESC")

	if err := subQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Count: %w", err)
	}

	if err := subQuery.Offset(offset).Limit(limit).Pluck("follower_id", &ids).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Pluck: %w", err)
	}

	if len(ids) == 0 {
		return nil, total, nil
	}

	var users []usermodel.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("follow.repository.FindFollowers.Users: %w", err)
	}
	return users, total, nil
}

func (r *Repository) FindFollowing(ctx context.Context, userID uint, offset, limit int) ([]usermodel.User, int64, error) {
	var total int64
	var ids []uint

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

func (r *Repository) FindFeed(ctx context.Context, userID uint, offset, limit int) ([]struct {
	ID        uint
	UserID    uint
	Title     string
	CoverURL  string
	VideoURL  string
	Duration  uint
	FileSize  uint64
	CategoryID uint
	Status    int8
	Views     uint64
	CreatedAt interface{}
	UpdatedAt interface{}
}, int64, error) {
	return nil, 0, nil
}

func (r *Repository) FindFollowingIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&followmodel.Follow{}).
		Where("follower_id = ?", userID).
		Pluck("following_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("follow.repository.FindFollowingIDs: %w", err)
	}
	return ids, nil
}
