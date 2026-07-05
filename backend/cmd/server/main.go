package main

import (
	"fmt"
	"log"

	authhandler "github.com/My-TuDo/B-B/backend/internal/handler/auth"
	categoryhandler "github.com/My-TuDo/B-B/backend/internal/handler/category"
	userhandler "github.com/My-TuDo/B-B/backend/internal/handler/user"
	videohandler "github.com/My-TuDo/B-B/backend/internal/handler/video"
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	"github.com/My-TuDo/B-B/backend/pkg/config"
	"github.com/My-TuDo/B-B/backend/pkg/database"
	"github.com/My-TuDo/B-B/backend/pkg/jwt"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/My-TuDo/B-B/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.LogLevel)
	jwt.Init(cfg.JWTSecret)
	validator.Init()

	db := database.InitMySQL(cfg)
	rdb := database.InitRedis(cfg)
	minioClient := storage.Init(cfg)
	_ = minioClient // Used internally by storage package

	middleware.InitAuth(rdb)

	// Auto migrate
	if err := db.AutoMigrate(&usermodel.User{}, &categorymodel.Category{}, &videomodel.Video{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Seed categories if empty
	seedCategories(db)

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(rdb))

	api := r.Group("/api/v1")

	// Register routes
	authhandler.RegisterRoutes(api, db, rdb)
	userhandler.RegisterRoutes(api, db, rdb)
	videohandler.RegisterRoutes(api, db, rdb)
	categoryhandler.RegisterRoutes(api, db)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logger.Info("server starting")
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&categorymodel.Category{}).Count(&count)
	if count > 0 {
		return
	}

	categories := []categorymodel.Category{
		{Name: "动画", Slug: "anime"},
		{Name: "音乐", Slug: "music"},
		{Name: "游戏", Slug: "game"},
		{Name: "知识", Slug: "knowledge"},
		{Name: "生活", Slug: "life"},
		{Name: "影视", Slug: "movie"},
		{Name: "科技", Slug: "tech"},
	}

	for _, cat := range categories {
		db.Create(&cat)
	}
}
