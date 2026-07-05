package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  string
	RedisPass  string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	JWTSecret  string
	LogLevel   string
}

func Load() *Config {
	v := viper.New()

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
	v.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	v.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	v.SetDefault("MINIO_BUCKET", "bb-videos")
	v.SetDefault("MINIO_USE_SSL", "false")
	v.SetDefault("JWT_SECRET", "dev-secret-change-in-production")
	v.SetDefault("LOG_LEVEL", "debug")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: config file read error: %v\n", err)
		}
	}

	cfg := &Config{
		ServerPort:     v.GetString("SERVER_PORT"),
		DBHost:         v.GetString("DB_HOST"),
		DBPort:         v.GetString("DB_PORT"),
		DBUser:         v.GetString("DB_USER"),
		DBPassword:     v.GetString("DB_PASSWORD"),
		DBName:         v.GetString("DB_NAME"),
		RedisHost:      v.GetString("REDIS_HOST"),
		RedisPort:      v.GetString("REDIS_PORT"),
		RedisPass:      v.GetString("REDIS_PASSWORD"),
		MinioEndpoint:  v.GetString("MINIO_ENDPOINT"),
		MinioAccessKey: v.GetString("MINIO_ACCESS_KEY"),
		MinioSecretKey: v.GetString("MINIO_SECRET_KEY"),
		MinioBucket:    v.GetString("MINIO_BUCKET"),
		MinioUseSSL:    v.GetBool("MINIO_USE_SSL"),
		JWTSecret:      v.GetString("JWT_SECRET"),
		LogLevel:       v.GetString("LOG_LEVEL"),
	}

	// Override from env directly (Viper AutomaticEnv sometimes needs this for non-standard names)
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

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
