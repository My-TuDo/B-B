package database

import (
	"context"
	"log"

	"github.com/My-TuDo/B-B/backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化 Redis 客户端并验证连接。
// 连接失败会直接 Fatal 退出。
func InitRedis(cfg *config.Config) *redis.Client {
	// 使用配置中的地址和密码创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPass,
		DB:       0, // 使用默认 0 号数据库
	})

	// Ping 验证连接可用性
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	return rdb
}
