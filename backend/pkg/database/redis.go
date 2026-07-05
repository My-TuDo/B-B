package database

import (
	"context"
	"log"

	"github.com/My-TuDo/B-B/backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPass,
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	return rdb
}
