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

type Service struct {
	repo *coinrepo.Repository
	rdb  *redis.Client
}

func NewService(repo *coinrepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

func (s *Service) AddCoin(ctx context.Context, userID, videoID uint, count uint8) (*coinmodel.CoinResp, error) {
	// Check if user already coined this video (one-time per video)
	existing, err := s.repo.FindByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		return nil, fmt.Errorf("coin.service.AddCoin: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("coin.service.AddCoin: %w", newError(errcode.CoinAlreadyCoined))
	}

	// Check daily limit
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

		// Save to MySQL
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

func endOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
}

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

func newError(code int) *Error {
	return &Error{Code: code, Msg: errcode.Message(code)}
}
