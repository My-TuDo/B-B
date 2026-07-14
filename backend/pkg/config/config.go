// Package config 提供应用配置的加载与管理。
// 配置优先级：环境变量 > config.yaml 文件 > 默认值。
// 使用 Viper 实现配置读取，支持 YAML 文件和自动环境变量绑定。
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 聚合应用所有运行时配置项。
type Config struct {
	ServerPort string // HTTP 服务监听端口
	DBHost     string // MySQL 主机地址
	DBPort     string // MySQL 端口
	DBUser     string // MySQL 用户名
	DBPassword string // MySQL 密码
	DBName     string // MySQL 数据库名

	RedisHost string // Redis 主机地址
	RedisPort string // Redis 端口
	RedisPass string // Redis 密码

	MinioEndpoint       string // MinIO 服务端点（内网）
	MinioPublicEndpoint string // MinIO 公网访问端点
	MinioAccessKey      string // MinIO 访问密钥
	MinioSecretKey      string // MinIO 密钥
	MinioBucket         string // MinIO 存储桶名称
	MinioUseSSL         bool   // MinIO 是否启用 SSL

	JWTSecret string // JWT 签名密钥
	LogLevel  string // 日志级别：debug / info / warn / error

	RabbitMQHost string // RabbitMQ 主机地址
	RabbitMQPort string // RabbitMQ 端口
	RabbitMQUser string // RabbitMQ 用户名
	RabbitMQPass string // RabbitMQ 密码
}

// Load 加载配置并返回 *Config。
// 依次设置默认值、读取 YAML 配置文件、绑定环境变量，
// 最后用环境变量显式覆盖（兜底）。
func Load() *Config {
	v := viper.New()

	// ---------- 设置默认值 ----------
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "3307")
	v.SetDefault("DB_USER", "bb_user")
	v.SetDefault("DB_PASSWORD", "bb_password")
	v.SetDefault("DB_NAME", "bb")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	v.SetDefault("MINIO_PUBLIC_ENDPOINT", "localhost:9000")
	v.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	v.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	v.SetDefault("MINIO_BUCKET", "bb-videos")
	v.SetDefault("MINIO_USE_SSL", "false")
	v.SetDefault("JWT_SECRET", "dev-secret-change-in-production")
	v.SetDefault("LOG_LEVEL", "debug")

	v.SetDefault("RABBITMQ_HOST", "localhost")
	v.SetDefault("RABBITMQ_PORT", "5672")
	v.SetDefault("RABBITMQ_USER", "guest")
	v.SetDefault("RABBITMQ_PASS", "guest")

	// ---------- 读取 YAML 配置文件 ----------
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// 自动绑定环境变量（例如 SERVER_PORT 会覆盖 config.yaml 中的 server_port）
	v.AutomaticEnv()

	// 尝试读取配置文件，文件不存在时不报错
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: config file read error: %v\n", err)
		}
	}

	// ---------- 从 Viper 构建 Config 结构体 ----------
	cfg := &Config{
		ServerPort:         v.GetString("SERVER_PORT"),
		DBHost:             v.GetString("DB_HOST"),
		DBPort:             v.GetString("DB_PORT"),
		DBUser:             v.GetString("DB_USER"),
		DBPassword:         v.GetString("DB_PASSWORD"),
		DBName:             v.GetString("DB_NAME"),
		RedisHost:          v.GetString("REDIS_HOST"),
		RedisPort:          v.GetString("REDIS_PORT"),
		RedisPass:          v.GetString("REDIS_PASSWORD"),
		MinioEndpoint:      v.GetString("MINIO_ENDPOINT"),
		MinioPublicEndpoint: v.GetString("MINIO_PUBLIC_ENDPOINT"),
		MinioAccessKey:     v.GetString("MINIO_ACCESS_KEY"),
		MinioSecretKey:     v.GetString("MINIO_SECRET_KEY"),
		MinioBucket:        v.GetString("MINIO_BUCKET"),
		MinioUseSSL:        v.GetBool("MINIO_USE_SSL"),
		JWTSecret:          v.GetString("JWT_SECRET"),
		LogLevel:           v.GetString("LOG_LEVEL"),
		RabbitMQHost:       v.GetString("RABBITMQ_HOST"),
		RabbitMQPort:       v.GetString("RABBITMQ_PORT"),
		RabbitMQUser:       v.GetString("RABBITMQ_USER"),
		RabbitMQPass:       v.GetString("RABBITMQ_PASS"),
	}

	// ---------- 环境变量显式覆盖（Viper AutomaticEnv 的兜底） ----------
	// 直接读取 os.Getenv，确保非标准命名的环境变量也能生效
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.ServerPort = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DBPort = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.RedisHost = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		cfg.RedisPort = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.RedisPass = v
	}
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		cfg.MinioEndpoint = v
	}
	if v := os.Getenv("MINIO_PUBLIC_ENDPOINT"); v != "" {
		cfg.MinioPublicEndpoint = v
	}
	if v := os.Getenv("MINIO_ACCESS_KEY"); v != "" {
		cfg.MinioAccessKey = v
	}
	if v := os.Getenv("MINIO_SECRET_KEY"); v != "" {
		cfg.MinioSecretKey = v
	}
	if v := os.Getenv("MINIO_BUCKET"); v != "" {
		cfg.MinioBucket = v
	}
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		cfg.MinioUseSSL = v == "true"
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("RABBITMQ_HOST"); v != "" {
		cfg.RabbitMQHost = v
	}
	if v := os.Getenv("RABBITMQ_PORT"); v != "" {
		cfg.RabbitMQPort = v
	}
	if v := os.Getenv("RABBITMQ_USER"); v != "" {
		cfg.RabbitMQUser = v
	}
	if v := os.Getenv("RABBITMQ_PASS"); v != "" {
		cfg.RabbitMQPass = v
	}

	return cfg
}

// DSN 返回 MySQL 连接字符串（DataSourceName）。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// RedisAddr 返回 Redis 地址，格式为 host:port。
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
