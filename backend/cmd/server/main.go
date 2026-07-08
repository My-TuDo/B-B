package main

import (
	"fmt"
	"log"

	adminhandler "github.com/My-TuDo/B-B/backend/internal/handler/admin"
	authhandler "github.com/My-TuDo/B-B/backend/internal/handler/auth"
	categoryhandler "github.com/My-TuDo/B-B/backend/internal/handler/category"
	coinhandler "github.com/My-TuDo/B-B/backend/internal/handler/coin"
	commenthandler "github.com/My-TuDo/B-B/backend/internal/handler/comment"
	creatorhandler "github.com/My-TuDo/B-B/backend/internal/handler/creator"
	danmakuhandler "github.com/My-TuDo/B-B/backend/internal/handler/danmaku"
	favoritehandler "github.com/My-TuDo/B-B/backend/internal/handler/favorite"
	followhandler "github.com/My-TuDo/B-B/backend/internal/handler/follow"
	historyhandler "github.com/My-TuDo/B-B/backend/internal/handler/history"
	interactionhandler "github.com/My-TuDo/B-B/backend/internal/handler/interaction"
	likehandler "github.com/My-TuDo/B-B/backend/internal/handler/like"
	notificationhandler "github.com/My-TuDo/B-B/backend/internal/handler/notification"
	searchhandler "github.com/My-TuDo/B-B/backend/internal/handler/search"
	taghandler "github.com/My-TuDo/B-B/backend/internal/handler/tag"
	userhandler "github.com/My-TuDo/B-B/backend/internal/handler/user"
	videohandler "github.com/My-TuDo/B-B/backend/internal/handler/video"
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	followmodel "github.com/My-TuDo/B-B/backend/internal/model/follow"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	messagemodel "github.com/My-TuDo/B-B/backend/internal/model/message"
	messagerepo "github.com/My-TuDo/B-B/backend/internal/repository/message"
	messageservice "github.com/My-TuDo/B-B/backend/internal/service/message"
	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	adminrepo "github.com/My-TuDo/B-B/backend/internal/repository/admin"
	creatorrepo "github.com/My-TuDo/B-B/backend/internal/repository/creator"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	creatorservice "github.com/My-TuDo/B-B/backend/internal/service/creator"
	"github.com/My-TuDo/B-B/backend/internal/ws"
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
	ws.InitHub()

	// Auto migrate
	if err := db.AutoMigrate(
		&usermodel.User{},
		&categorymodel.Category{},
		&videomodel.Video{},
		&tagmodel.Tag{},
		&tagmodel.VideoTag{},
		&historymodel.VideoHistory{},
		&danmakumodel.Danmaku{},
		&commentmodel.Comment{},
		&likemodel.VideoLike{},
		&coinmodel.VideoCoin{},
		&favoritemodel.Favorite{},
		&favoritemodel.FavoriteItem{},
		&followmodel.Follow{},
		&messagemodel.Message{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Drop stale FK constraint from previous model version (parent_id self-reference).
	// MySQL 8.0 does not support DROP FOREIGN KEY IF EXISTS, so ignore errors
	// (the constraint may already have been dropped).
	_ = db.Exec("ALTER TABLE comments DROP FOREIGN KEY fk_comments_children").Error

	// FULLTEXT index for video search
	var idxCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='videos' AND index_name='ft_videos_title_desc'").Scan(&idxCount)
	if idxCount == 0 {
		db.Exec("ALTER TABLE videos ADD FULLTEXT INDEX ft_videos_title_desc (title, description)")
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

	// CSRF middleware for write operations
	api.Use(middleware.CSRF())

	// Register routes
	authhandler.RegisterRoutes(api, db, rdb)
	userhandler.RegisterRoutes(api, db, rdb)
	videohandler.RegisterRoutes(api, db, rdb)
	categoryhandler.RegisterRoutes(api, db)
	taghandler.RegisterRoutes(api, db)
	historyhandler.RegisterRoutes(api, db)
	searchhandler.RegisterRoutes(api, db)
	creatorRepo := creatorrepo.NewRepository(db)
	creatorSvc := creatorservice.NewService(creatorRepo)
	creatorhandler.RegisterRoutes(api, creatorSvc)

	adminRepo := adminrepo.NewRepository(db)
	adminSvc := adminservice.NewService(adminRepo)
	adminhandler.RegisterRoutes(api, db, adminSvc)

	danmakuhandler.RegisterRoutes(api, db, rdb)

	// Create message service once for notifications
	messageRepo := messagerepo.NewRepository(db)
	messageSvc := messageservice.NewService(messageRepo, rdb)

	commenthandler.RegisterRoutes(api, db, rdb, messageSvc)
	likehandler.RegisterRoutes(api, db, rdb, messageSvc)
	coinhandler.RegisterRoutes(api, db, rdb)
	favoritehandler.RegisterRoutes(api, db)
	followhandler.RegisterRoutes(api, db, messageSvc)
	notificationhandler.RegisterRoutes(api, messageSvc)
	interactionhandler.RegisterRoutes(api, db, rdb)

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
