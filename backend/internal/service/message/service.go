package message

import (
	"context"
	"fmt"

	messagemodel "github.com/My-TuDo/B-B/backend/internal/model/message"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	messagerepo "github.com/My-TuDo/B-B/backend/internal/repository/message"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo *messagerepo.Repository
	rdb  *redis.Client
}

func NewService(repo *messagerepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

func (s *Service) GetNotifications(ctx context.Context, userID uint, page, pageSize int) (*messagemodel.NotificationListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	list, total, err := s.repo.FindByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("message.service.GetNotifications: %w", err)
	}

	items := make([]messagemodel.MessageResp, len(list))
	for i, m := range list {
		items[i] = messagemodel.MessageResp{
			ID:        m.ID,
			Type:      m.Type,
			Content:   m.Content,
			TargetID:  m.TargetID,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt,
		}
		if m.FromUser.ID != 0 {
			avatar := ""
			if m.FromUser.Avatar != "" {
				avatar = storage.GetObjectURL(m.FromUser.Avatar)
			}
			items[i].FromUser = &usermodel.UserBrief{
				ID:       m.FromUser.ID,
				Username: m.FromUser.Username,
				Nickname: m.FromUser.Nickname,
				Avatar:   avatar,
			}
		}
	}

	unread, _ := s.repo.CountUnread(ctx, userID)
	if s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		if val, err := s.rdb.Get(ctx, key).Int64(); err == nil && val > unread {
			unread = val
		}
	}

	return &messagemodel.NotificationListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Unread:   unread,
	}, nil
}

func (s *Service) ReadAll(ctx context.Context, userID uint) error {
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("message.service.ReadAll: %w", err)
	}
	if s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Set(ctx, key, 0, 0)
	}
	return nil
}

func (s *Service) MarkSingleRead(ctx context.Context, userID uint, messageID uint) error {
	updated, err := s.repo.MarkSingleRead(ctx, messageID, userID)
	if err != nil {
		return fmt.Errorf("message.service.MarkSingleRead: %w", err)
	}
	if updated && s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Decr(ctx, key)
	}
	return nil
}

func (s *Service) SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error {
	m := &messagemodel.Message{
		UserID:     userID,
		FromUserID: fromUserID,
		Type:       msgType,
		TargetID:   targetID,
		Content:    content,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return fmt.Errorf("message.service.SendNotification: %w", err)
	}

	if s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Incr(ctx, key)
	}
	return nil
}
