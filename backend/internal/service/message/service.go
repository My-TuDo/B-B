// Package message 提供消息通知相关的业务逻辑服务，
// 包括通知列表查询、全部已读、单条已读以及发送通知等功能。
// 该服务同时实现了其他模块所需的 Notifier 接口。
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

// Service 消息通知服务，封装消息相关的业务逻辑。
type Service struct {
	repo *messagerepo.Repository
	rdb  *redis.Client
}

// NewService 创建消息通知服务实例。
func NewService(repo *messagerepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// GetNotifications 分页获取当前用户的通知列表，同时返回未读数量。
// 未读数优先使用Redis缓存值，当Redis值更大时覆盖MySQL统计值。
func (s *Service) GetNotifications(ctx context.Context, userID uint, page, pageSize int) (*messagemodel.NotificationListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 分页查询通知列表
	offset := (page - 1) * pageSize
	list, total, err := s.repo.FindByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("message.service.GetNotifications: %w", err)
	}

	// 组装响应，含发送者信息
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
		// 填充发送者简要信息
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

	// 未读数：取MySQL统计与Redis缓存中的较大值
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

// ReadAll 将当前用户所有通知标记为已读，同时将Redis未读计数归零。
func (s *Service) ReadAll(ctx context.Context, userID uint) error {
	// 标记MySQL中所有通知为已读
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("message.service.ReadAll: %w", err)
	}
	// 同步清零Redis未读计数
	if s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Set(ctx, key, 0, 0)
	}
	return nil
}

// MarkSingleRead 将单条通知标记为已读，同时递减Redis未读计数。
func (s *Service) MarkSingleRead(ctx context.Context, userID uint, messageID uint) error {
	// 标记MySQL中单条通知为已读
	updated, err := s.repo.MarkSingleRead(ctx, messageID, userID)
	if err != nil {
		return fmt.Errorf("message.service.MarkSingleRead: %w", err)
	}
	// 若确实更新了一条记录，递减Redis计数
	if updated && s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Decr(ctx, key)
	}
	return nil
}

// SendNotification 向指定用户发送一条通知消息，同时递增Redis未读计数。
// 该方法实现了其他模块所需的 Notifier 接口。
func (s *Service) SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error {
	// 构建消息实体
	m := &messagemodel.Message{
		UserID:     userID,
		FromUserID: fromUserID,
		Type:       msgType,
		TargetID:   targetID,
		Content:    content,
	}
	// 写入MySQL
	if err := s.repo.Create(ctx, m); err != nil {
		return fmt.Errorf("message.service.SendNotification: %w", err)
	}

	// 递增Redis未读计数
	if s.rdb != nil {
		key := fmt.Sprintf("unread:%d", userID)
		s.rdb.Incr(ctx, key)
	}
	return nil
}
