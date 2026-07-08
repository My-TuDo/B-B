package interaction

import (
	"context"
	"fmt"

	interactionrepo "github.com/My-TuDo/B-B/backend/internal/repository/interaction"
	"github.com/redis/go-redis/v9"
)

type InteractionStatus struct {
	Liked     bool `json:"liked"`
	Coins     uint `json:"coins"`
	Favorited bool `json:"favorited"`
	Following bool `json:"following,omitempty"`
}

type Service struct {
	repo *interactionrepo.Repository
	rdb  *redis.Client
}

func NewService(repo *interactionrepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

func (s *Service) GetVideoInteractions(ctx context.Context, userID, videoID uint) (*InteractionStatus, error) {
	status := &InteractionStatus{}

	if s.rdb != nil {
		key := fmt.Sprintf("video:like:%d", videoID)
		liked, _ := s.rdb.SIsMember(ctx, key, userID).Result()
		status.Liked = liked
	} else {
		status.Liked, _ = s.repo.IsLiked(ctx, userID, videoID)
	}

	coinCount, _ := s.repo.CountCoins(ctx, userID, videoID)
	status.Coins = uint(coinCount)

	status.Favorited, _ = s.repo.IsFavorited(ctx, userID, videoID)

	return status, nil
}

func (s *Service) GetUserInteractions(ctx context.Context, viewerID, targetUserID uint) (*InteractionStatus, error) {
	status := &InteractionStatus{}
	if viewerID == 0 {
		return status, nil
	}
	status.Following, _ = s.repo.IsFollowing(ctx, viewerID, targetUserID)
	return status, nil
}
