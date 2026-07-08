package message

import (
	"context"
	"fmt"

	messagemodel "github.com/My-TuDo/B-B/backend/internal/model/message"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, m *messagemodel.Message) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("message.repository.Create: %w", err)
	}
	return nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID uint, offset, limit int) ([]messagemodel.Message, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("message.repository.FindByUserID.Count: %w", err)
	}

	var list []messagemodel.Message
	if err := query.Preload("FromUser").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("message.repository.FindByUserID.Find: %w", err)
	}
	return list, total, nil
}

func (r *Repository) CountUnread(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ? AND is_read = 0", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("message.repository.CountUnread: %w", err)
	}
	return count, nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).Model(&messagemodel.Message{}).Where("user_id = ? AND is_read = 0", userID).Update("is_read", 1).Error; err != nil {
		return fmt.Errorf("message.repository.MarkAllRead: %w", err)
	}
	return nil
}

func (r *Repository) MarkSingleRead(ctx context.Context, messageID uint, userID uint) (bool, error) {
	result := r.db.WithContext(ctx).Model(&messagemodel.Message{}).
		Where("id = ? AND user_id = ? AND is_read = 0", messageID, userID).
		Update("is_read", 1)
	if result.Error != nil {
		return false, fmt.Errorf("message.repository.MarkSingleRead: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}
