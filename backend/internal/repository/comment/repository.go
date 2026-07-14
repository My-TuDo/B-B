// Package comment 提供评论数据访问层，封装视频评论相关的数据库操作。
// 支持根评论和回复的分层查询、评论点赞数更新、删除评论以及验证视频和作者信息。
package comment

import (
	"context"
	"fmt"

	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	"gorm.io/gorm"
)

// Repository 评论数据仓库，封装评论相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建评论数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条新评论（可以是根评论或回复）。
// 评论数据通过指针传入，创建成功后 model 中的 ID 等字段会被回填。
func (r *Repository) Create(ctx context.Context, c *commentmodel.Comment) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("comment.repository.Create: %w", err)
	}
	return nil
}

// FindByID 根据评论 ID 查找单条评论，同时预加载关联的用户信息。
// 若未找到则返回 (nil, nil)。
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

// FindRootComments 查询指定视频的顶级评论（parent_id = 0），支持按时间或热度排序。
// sort 参数："hot" 按点赞数降序，其他值按创建时间降序。
// 返回根评论列表、总数以及可能的错误。
func (r *Repository) FindRootComments(ctx context.Context, videoID uint, sort string, offset, limit int) ([]commentmodel.Comment, int64, error) {
	var total int64
	// 筛选视频 ID 且父评论 ID 为 0（即顶级评论）
	query := r.db.WithContext(ctx).Model(&commentmodel.Comment{}).Where("video_id = ? AND parent_id = 0", videoID)

	// 统计符合条件的根评论总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("comment.repository.FindRootComments.Count: %w", err)
	}

	// 根据排序参数选择排序字段
	orderBy := "created_at DESC"
	if sort == "hot" {
		orderBy = "likes DESC"
	}

	// 预加载用户信息，按指定排序和分页查询
	var list []commentmodel.Comment
	if err := query.Preload("User").Order(orderBy).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("comment.repository.FindRootComments.Find: %w", err)
	}
	return list, total, nil
}

// FindRepliesByRootIDs 根据一批根评论 ID 批量查询其下的子回复。
// rootIDs 为空时直接返回 nil，避免执行无效查询。
// 回复按创建时间升序排列，保证对话的时间顺序。
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

// Delete 根据 ID 删除一条评论（硬删除）。
func (r *Repository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&commentmodel.Comment{}, id).Error; err != nil {
		return fmt.Errorf("comment.repository.Delete: %w", err)
	}
	return nil
}

// UpdateLikes 更新指定评论的点赞数量。
// 通常在点赞/取消点赞操作后同步更新评论表的计数字段。
func (r *Repository) UpdateLikes(ctx context.Context, id uint, likes uint) error {
	if err := r.db.WithContext(ctx).Model(&commentmodel.Comment{}).Where("id = ?", id).Update("likes", likes).Error; err != nil {
		return fmt.Errorf("comment.repository.UpdateLikes: %w", err)
	}
	return nil
}

// FindVideoAuthor 查询指定视频的作者 ID，结果写入 authorID 指针。
// 用于评论通知等场景，需要知道视频作者是谁。
func (r *Repository) FindVideoAuthor(ctx context.Context, videoID uint, authorID *uint) error {
	if err := r.db.WithContext(ctx).Table("videos").Select("user_id").Where("id = ?", videoID).Scan(authorID).Error; err != nil {
		return fmt.Errorf("comment.repository.FindVideoAuthor: %w", err)
	}
	return nil
}

// ExistsVideo 判断指定视频是否存在（用于评论前验证视频有效性）。
func (r *Repository) ExistsVideo(ctx context.Context, videoID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("videos").Where("id = ?", videoID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("comment.repository.ExistsVideo: %w", err)
	}
	return count > 0, nil
}
