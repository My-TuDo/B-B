package like

import (
	"context"
	"fmt"

	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	likerepo "github.com/My-TuDo/B-B/backend/internal/repository/like"
	"github.com/redis/go-redis/v9"
)

type Notifier interface {
	SendNotification(ctx context.Context, userID uint, fromUserID uint, msgType int8, targetID uint, content string) error
}

type Service struct {
	repo     *likerepo.Repository
	rdb      *redis.Client
	notifier Notifier
}

func NewService(repo *likerepo.Repository, rdb *redis.Client, notifier Notifier) *Service {
	return &Service{repo: repo, rdb: rdb, notifier: notifier}
}

func (s *Service) ToggleLike(ctx context.Context, userID, videoID uint) (*likemodel.LikeResp, error) {
	key := fmt.Sprintf("video:like:%d", videoID)

	liked := false
	if s.rdb != nil {
		added, err := s.rdb.SAdd(ctx, key, userID).Result()
		if err != nil {
			return nil, fmt.Errorf("like.service.ToggleLike: %w", err)
		}
		if added == 0 {
			s.rdb.SRem(ctx, key, userID)
		} else {
			liked = true
		}
	}

	// Send notification when a like is added (not self-like)
	if liked && s.notifier != nil {
		authorID, err := s.repo.FindVideoAuthor(ctx, videoID)
		if err == nil && authorID != 0 && authorID != userID {
			// Look up the liking user's name
			likeUsername := "有人"
			if name, nameErr := s.repo.FindUserName(ctx, userID); nameErr == nil && name != "" {
				likeUsername = name
			}
			content := fmt.Sprintf("%s 赞了你的视频", likeUsername)
			_ = s.notifier.SendNotification(ctx, authorID, userID, 2, videoID, content)
		}
	}

	likeCount, _ := s.getLikeCount(ctx, videoID)
	return &likemodel.LikeResp{Liked: liked, Count: likeCount}, nil
}

func (s *Service) getLikeCount(ctx context.Context, videoID uint) (uint, error) {
	key := fmt.Sprintf("video:like:%d", videoID)
	if s.rdb != nil {
		count, err := s.rdb.SCard(ctx, key).Result()
		if err == nil {
			return uint(count), nil
		}
	}
	dbCount, err := s.repo.CountByVideoID(ctx, videoID)
	if err != nil {
		return 0, fmt.Errorf("like.service.getLikeCount: %w", err)
	}
	return uint(dbCount), nil
}

func (s *Service) SyncLikes(ctx context.Context) {
	// Called periodically to sync Redis Set to MySQL
}
