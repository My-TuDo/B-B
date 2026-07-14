// Package message 提供消息数据访问层，封装用户消息通知相关的数据库操作。
// 支持创建消息、分页查询消息列表、统计未读数量和批量/单条标记已读。
package message

import (
	"context"
	"fmt"

	messagemodel "github.com/My-TuDo/B-B/backend/internal/model/message"
	"gorm.io/gorm"
)

// Repository 消息数据仓库，封装消息相关的数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建消息数据仓库实例。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建一条消息通知。
// 消息可包含发送者信息（FromUser），用于展示通知来源。
func (r *Repository) Create(ctx context.Context, m *messagemodel.Message) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("message.repository.Create: %w", err)
	}
	return nil
}

// FindByUserID 分页查询用户的消息列表，按创建时间倒序排列。
// 预加载发送者信息（FromUser），用于前端渲染消息详情。
func (r *Repository) FindByUserID(ctx context.Context, userID uint, offset, limit int) ([]messagemodel.Message, int64, error) {
	var total int64
	// 构建用户消息查询
	query := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ?", userID)

	// 统计消息总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("message.repository.FindByUserID.Count: %w", err)
	}

	// 预加载发送者，按时间倒序分页查询
	var list []messagemodel.Message
	if err := query.Preload("FromUser").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("message.repository.FindByUserID.Find: %w", err)
	}
	return list, total, nil
}

// CountUnread 统计用户未读消息数量（is_read = 0）。
// 用于前端显示未读消息角标。
func (r *Repository) CountUnread(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ? AND is_read = 0", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("message.repository.CountUnread: %w", err)
	}
	return count, nil
}

// MarkAllRead 将用户所有未读消息标记为已读。
// 批量更新：将所有 is_read = 0 的记录改为 is_read = 1。
func (r *Repository) MarkAllRead(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ? AND is_read = 0", userID).Update("is_read", 1).Error; err != nil {
		return fmt.Errorf("message.repository.MarkAllRead: %w", err)
	}
	return nil
}

// MarkSingleRead 将单条消息标记为已读。
// 返回是否确实有记录被更新（若消息不存在或已读则返回 false），用于幂等处理。
func (r *Repository) MarkSingleRead(ctx context.Context, messageID uint, userID uint) (bool, error) {
	result := r.db.WithContext(ctx).Model(&messagemodel.Message{}).
		Where("id = ? AND user_id = ? AND is_read = 0", messageID, userID).
		Update("is_read", 1)
	if result.Error != nil {
		return false, fmt.Errorf("message.repository.MarkSingleRead: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}
