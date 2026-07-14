// Package like 提供视频点赞相关的业务逻辑服务，
// 包括点赞/取消点赞、点赞数查询以及点赞数据同步等功能。
package like

import (
	"context"
	"fmt"

	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	likerepo "github.com/My-TuDo/B-B/backend/internal/repository/like"
	"github.com/redis/go-redis/v9"
)

// Notifier 通知发送接口，用于在点赞时向视频作者发送通知。
type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

// Service 点赞服务，封装视频点赞相关的业务逻辑。
type Service struct {
	repo     *likerepo.Repository
	rdb      *redis.Client
	notifier Notifier
}

// NewService 创建点赞服务实例。
func NewService(repo *likerepo.Repository, rdb *redis.Client, notifier Notifier) *Service {
	return &Service{repo: repo, rdb: rdb, notifier: notifier}
}

// ToggleLike 切换点赞状态：已点赞则取消，未点赞则点赞。
// 使用Redis Set存储点赞用户ID。点赞成功时向视频作者发送通知。
func (s *Service) ToggleLike(ctx context.Context, userID, videoID uint) (*likemodel.LikeResp, error) {
	key := fmt.Sprintf("video:like:%d", videoID)

	liked := false
	if s.rdb != nil {
		// SAdd 返回1表示成功添加（点赞），0表示已存在
		added, err := s.rdb.SAdd(ctx, key, userID).Result()
		if err != nil {
			return nil, fmt.Errorf("like.service.ToggleLike: %w", err)
		}
		if added == 0 {
			// 已存在 → 取消点赞
			s.rdb.SRem(ctx, key, userID)
		} else {
			liked = true
		}
	}

	// Send notification when a like is added (not self-like)
	// 点赞成功且非自点赞时，向视频作者发送通知
	if liked && s.notifier != nil {
		authorID, err := s.repo.FindVideoAuthor(ctx, videoID)
		if err == nil && authorID != 0 && authorID != userID {
			// Look up the liking user's name — 查询点赞用户的昵称
			likeUsername := "有人"
			if name, nameErr := s.repo.FindUserName(ctx, userID); nameErr == nil && name != "" {
				likeUsername = name
			}
			content := fmt.Sprintf("%s 赞了你的视频", likeUsername)
			_ = s.notifier.SendNotification(ctx, authorID, userID, 2, videoID, content)
		}
	}

	// 获取最新点赞数
	likeCount, _ := s.getLikeCount(ctx, videoID)
	return &likemodel.LikeResp{Liked: liked, Count: likeCount}, nil
}

// getLikeCount 获取视频的点赞总数，优先从Redis读取，失败时回退MySQL。
func (s *Service) getLikeCount(ctx context.Context, videoID uint) (uint, error) {
	key := fmt.Sprintf("video:like:%d", videoID)
	if s.rdb != nil {
		count, err := s.rdb.SCard(ctx, key).Result()
		if err == nil {
			return uint(count), nil
		}
	}
	// Redis不可用时从MySQL查询
	dbCount, err := s.repo.CountByVideoID(ctx, videoID)
	if err != nil {
		return 0, fmt.Errorf("like.service.getLikeCount: %w", err)
	}
	return uint(dbCount), nil
}

// SyncLikes 定时将Redis中的点赞数据同步回MySQL。
// Called periodically to sync Redis Set to MySQL
func (s *Service) SyncLikes(ctx context.Context) {
	// TODO: 实现Redis点赞数据到MySQL的同步逻辑
}
