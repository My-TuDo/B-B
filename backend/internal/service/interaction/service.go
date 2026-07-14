// Package interaction 提供用户与视频/用户之间的交互状态查询服务，
// 包括视频的点赞/投币/收藏状态，以及用户间的关注关系状态。
package interaction

import (
	"context"
	"fmt"

	interactionrepo "github.com/My-TuDo/B-B/backend/internal/repository/interaction"
	"github.com/redis/go-redis/v9"
)

// InteractionStatus 描述当前用户对目标视频或用户的交互状态。
type InteractionStatus struct {
	Liked     bool `json:"liked"`               // 是否已点赞
	Coins     uint `json:"coins"`               // 已投币数
	Favorited bool `json:"favorited"`           // 是否已收藏
	Following bool `json:"following,omitempty"` // 是否已关注（仅用户交互时有值）
}

// Service 交互状态服务，封装用户交互状态的聚合查询逻辑。
type Service struct {
	repo *interactionrepo.Repository
	rdb  *redis.Client
}

// NewService 创建交互状态服务实例。
func NewService(repo *interactionrepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// GetVideoInteractions 获取当前用户对指定视频的交互状态（点赞、投币、收藏）。
// 点赞状态优先从Redis读取以提高性能。
func (s *Service) GetVideoInteractions(ctx context.Context, userID, videoID uint) (*InteractionStatus, error) {
	status := &InteractionStatus{}

	// 点赞状态：优先从Redis Set查询
	if s.rdb != nil {
		key := fmt.Sprintf("video:like:%d", videoID)
		liked, _ := s.rdb.SIsMember(ctx, key, userID).Result()
		status.Liked = liked
	} else {
		// 降级到MySQL查询
		status.Liked, _ = s.repo.IsLiked(ctx, userID, videoID)
	}

	// 投币状态：从MySQL查询
	coinCount, _ := s.repo.CountCoins(ctx, userID, videoID)
	status.Coins = uint(coinCount)

	// 收藏状态：从MySQL查询
	status.Favorited, _ = s.repo.IsFavorited(ctx, userID, videoID)

	return status, nil
}

// GetUserInteractions 获取当前查看者对目标用户的交互状态（当前仅含关注关系）。
// viewerID 为 0（未登录）时直接返回空状态。
func (s *Service) GetUserInteractions(ctx context.Context, viewerID, targetUserID uint) (*InteractionStatus, error) {
	status := &InteractionStatus{}
	if viewerID == 0 {
		return status, nil
	}
	// 查询关注关系
	status.Following, _ = s.repo.IsFollowing(ctx, viewerID, targetUserID)
	return status, nil
}
