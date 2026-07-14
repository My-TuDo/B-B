// Package coin 提供投币相关的业务逻辑服务，
// 包括投币、每日限额检查等功能。
package coin

import (
	"context"
	"fmt"
	"time"

	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	coinrepo "github.com/My-TuDo/B-B/backend/internal/repository/coin"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/redis/go-redis/v9"
)

// Service 投币服务，封装投币相关的业务逻辑与每日限额控制。
type Service struct {
	repo *coinrepo.Repository
	rdb  *redis.Client
}

// NewService 创建投币服务实例。
func NewService(repo *coinrepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// AddCoin 为指定视频投币。每个用户对同一视频只能投一次，
// 每日投币总数上限为5。使用Redis Lua脚本保证原子性。
func (s *Service) AddCoin(ctx context.Context, userID, videoID uint, count uint8) (*coinmodel.CoinResp, error) {
	// Check if user already coined this video (one-time per video)
	// 检查用户是否已对该视频投过币（每个视频仅可投一次）
	existing, err := s.repo.FindByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		return nil, fmt.Errorf("coin.service.AddCoin: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("coin.service.AddCoin: %w", newError(errcode.CoinAlreadyCoined))
	}

	// Check daily limit — 通过Redis Lua脚本原子性地检查并递增每日限额
	dailyKey := fmt.Sprintf("coin:limit:%d:%s", userID, time.Now().Format("20060102"))
	if s.rdb != nil {
		// Atomic check-and-increment via Lua script to prevent race conditions
		script := redis.NewScript(`
			local current = redis.call('GET', KEYS[1])
			current = tonumber(current) or 0
			local count = tonumber(ARGV[1])
			if current + count > 5 then
				return -1
			end
			local newval = redis.call('INCRBY', KEYS[1], count)
			redis.call('EXPIREAT', KEYS[1], ARGV[2])
			return newval
		`)
		result, err := script.Run(ctx, s.rdb, []string{dailyKey}, int64(count), endOfDay().Unix()).Result()
		if err != nil {
			return nil, fmt.Errorf("coin.service.AddCoin: %w", err)
		}
		newTotal, ok := result.(int64)
		if !ok || newTotal == -1 {
			return nil, fmt.Errorf("coin.service.AddCoin: %w", newError(errcode.CoinLimitExceeded))
		}

		// Save to MySQL — 限额检查通过后写入MySQL
		c := &coinmodel.VideoCoin{
			UserID:  userID,
			VideoID: videoID,
			Count:   count,
		}
		if err := s.repo.Create(ctx, c); err != nil {
			return nil, fmt.Errorf("coin.service.AddCoin: %w", err)
		}

		return &coinmodel.CoinResp{CoinsToday: uint(newTotal)}, nil
	}

	// No Redis → save to MySQL directly (no limit check possible)
	// 无Redis时直接写入MySQL，无法进行限额检查
	c := &coinmodel.VideoCoin{
		UserID:  userID,
		VideoID: videoID,
		Count:   count,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("coin.service.AddCoin: %w", err)
	}

	return &coinmodel.CoinResp{CoinsToday: uint(count)}, nil
}

// endOfDay 返回当天的最后一秒（23:59:59），用于设置Redis key的过期时间。
func endOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
}

// Error 服务层错误类型，携带错误码以支持在HTTP层映射为合适的响应。
type Error struct {
	Code int
	Msg  string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	return e.Msg
}

// newError 根据错误码创建带本地化消息的服务错误。
func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}
