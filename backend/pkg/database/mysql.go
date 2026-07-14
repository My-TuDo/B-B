// Package database 提供 MySQL 和 Redis 的初始化与连接管理。
// MySQL 使用 GORM 作为 ORM，Redis 使用 go-redis/v9 客户端。
package database

import (
	"log"
	"time"

	"github.com/My-TuDo/B-B/backend/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitMySQL 初始化 MySQL 连接，配置连接池参数，返回 *gorm.DB。
// 连接失败会直接 Fatal 退出。
func InitMySQL(cfg *config.Config) *gorm.DB {
	// 从配置生成 DSN 连接字符串
	dsn := cfg.DSN()

	// 打开数据库连接，设置日志级别为 Warn（只记录慢查询和错误）
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}

	// 获取底层 sql.DB 以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(100)     // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)      // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间

	return db
}
