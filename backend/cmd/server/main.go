// B-B 视频平台后端服务入口。
//
// 启动流程：
//  1. 加载配置（环境变量 > YAML > 默认值）
//  2. 初始化基础设施：日志、JWT、校验器
//  3. 初始化外部依赖：MySQL、Redis、MinIO
//  4. 初始化中间件和 WebSocket Hub
//  5. 自动迁移数据库表结构
//  6. 创建全文索引、种子数据
//  7. 启动 RabbitMQ 消费者（转码任务）
//  8. 注册路由并启动 HTTP 服务
package main

import (
	"encoding/json"
	"fmt"
	"log"

	// Handler 层 — HTTP 请求处理
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
	qualityhandler "github.com/My-TuDo/B-B/backend/internal/handler/quality"
	searchhandler "github.com/My-TuDo/B-B/backend/internal/handler/search"
	taghandler "github.com/My-TuDo/B-B/backend/internal/handler/tag"
	transcodehdlr "github.com/My-TuDo/B-B/backend/internal/handler/transcode"
	userhandler "github.com/My-TuDo/B-B/backend/internal/handler/user"
	videohandler "github.com/My-TuDo/B-B/backend/internal/handler/video"

	// Middleware 层
	"github.com/My-TuDo/B-B/backend/internal/middleware"

	// Model 层 — 数据库表结构定义
	categorymodel "github.com/My-TuDo/B-B/backend/internal/model/category"
	coinmodel "github.com/My-TuDo/B-B/backend/internal/model/coin"
	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	danmakumodel "github.com/My-TuDo/B-B/backend/internal/model/danmaku"
	favoritemodel "github.com/My-TuDo/B-B/backend/internal/model/favorite"
	followmodel "github.com/My-TuDo/B-B/backend/internal/model/follow"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	likemodel "github.com/My-TuDo/B-B/backend/internal/model/like"
	messagemodel "github.com/My-TuDo/B-B/backend/internal/model/message"
	metamodel "github.com/My-TuDo/B-B/backend/internal/model/meta"
	qualitymodel "github.com/My-TuDo/B-B/backend/internal/model/quality"
	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	transcodemodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"

	// Repository 层 — 数据访问
	messagerepo "github.com/My-TuDo/B-B/backend/internal/repository/message"
	adminrepo "github.com/My-TuDo/B-B/backend/internal/repository/admin"
	creatorrepo "github.com/My-TuDo/B-B/backend/internal/repository/creator"

	// Service 层 — 业务逻辑
	messageservice "github.com/My-TuDo/B-B/backend/internal/service/message"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	creatorservice "github.com/My-TuDo/B-B/backend/internal/service/creator"

	// Worker 和 WebSocket
	"github.com/My-TuDo/B-B/backend/internal/worker"
	"github.com/My-TuDo/B-B/backend/internal/ws"

	// 基础设施包
	"github.com/My-TuDo/B-B/backend/pkg/config"
	"github.com/My-TuDo/B-B/backend/pkg/database"
	"github.com/My-TuDo/B-B/backend/pkg/jwt"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/rabbitmq"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/My-TuDo/B-B/backend/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// main 是服务主入口，按顺序完成初始化、依赖注入和 HTTP 启动。
func main() {
	// ==================== 第 1 步：加载配置 ====================
	cfg := config.Load()

	// ==================== 第 2 步：初始化基础设施 ====================
	logger.Init(cfg.LogLevel) // 结构化日志
	jwt.Init(cfg.JWTSecret)   // JWT 签名密钥
	validator.Init()          // 请求参数校验器

	// ==================== 第 3 步：初始化外部依赖 ====================
	db := database.InitMySQL(cfg)     // MySQL 数据库连接
	rdb := database.InitRedis(cfg)    // Redis 缓存连接
	minioClient := storage.Init(cfg)  // MinIO 对象存储客户端
	_ = minioClient                   // minioClient 由 storage 包内部持有

	// ==================== 第 4 步：初始化中间件和 WebSocket ====================
	middleware.InitAuth(rdb) // 认证中间件（依赖 Redis 做 Token 白名单校验）
	ws.InitHub()             // WebSocket 连接中心

	// ==================== 第 5 步：数据库表自动迁移 ====================
	// GORM AutoMigrate 会根据 Model 定义自动创建/更新表结构
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
		&transcodemodel.TranscodeTask{},
		&qualitymodel.VideoQuality{},
		&metamodel.VideoMeta{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// ==================== 第 6 步：数据库修补 ====================
	// 删除旧版模型遗留的外键约束（parent_id 自引用）
	// MySQL 8.0 不支持 DROP FOREIGN KEY IF EXISTS，忽略错误
	_ = db.Exec("ALTER TABLE comments DROP FOREIGN KEY fk_comments_children").Error

	// 为视频表创建全文索引（用于搜索功能）
	var idxCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='videos' AND index_name='ft_videos_title_desc'").Scan(&idxCount)
	if idxCount == 0 {
		db.Exec("ALTER TABLE videos ADD FULLTEXT INDEX ft_videos_title_desc (title, description)")
	}

	// 初始化分类种子数据
	seedCategories(db)

	// ==================== 第 7 步：初始化消息队列和转码 Worker ====================
	rmqCfg := &rabbitmq.Config{
		Host:     cfg.RabbitMQHost,
		Port:     cfg.RabbitMQPort,
		User:     cfg.RabbitMQUser,
		Password: cfg.RabbitMQPass,
	}
	rmqClient, rmqErr := rabbitmq.Init(rmqCfg)
	if rmqErr != nil {
		logger.Warn("rabbitmq init failed, transcode will run without queue")
	} else {
		logger.Info("rabbitmq connected")
		defer rmqClient.Close()
	}

	// 启动 RabbitMQ 消费者协程，持续监听转码任务
	if rmqClient != nil {
		go func() {
			deliveries, err := rmqClient.ConsumeTranscodeTask()
			if err != nil {
				logger.Error("rabbitmq consume failed", zap.Error(err))
				return
			}
			for d := range deliveries {
				var msg rabbitmq.TranscodeMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					logger.Error("rabbitmq unmarshal failed", zap.Error(err))
					d.Nack(false, false) // 解析失败，不重新入队
					continue
				}
				// 处理转码任务
				worker.ProcessVideo(msg.VideoID, db)
				d.Ack(false) // 手动确认消息
			}
		}()
	}

	// ==================== 第 8 步：构建 HTTP 服务并注册路由 ====================
	r := gin.New()

	// --- 全局中间件 ---
	r.Use(middleware.Recovery())    // Panic 恢复
	r.Use(middleware.RequestID())   // 请求 ID 注入
	r.Use(middleware.Logger())      // 请求日志
	r.Use(middleware.CORS())        // 跨域
	r.Use(middleware.RateLimit(rdb)) // 限流

	api := r.Group("/api/v1")

	// CSRF 防护中间件（仅对写操作生效）
	api.Use(middleware.CSRF())

	// --- 注册各模块路由 ---
	// 认证和用户模块（需要 Redis 做 Token 管理）
	authhandler.RegisterRoutes(api, db, rdb)
	userhandler.RegisterRoutes(api, db, rdb)

	// 视频模块（需要转码发布函数将任务投递到队列）
	var transcodePublisher func(uint) error
	if rmqClient != nil {
		transcodePublisher = rmqClient.PublishTranscodeTask
	}
	videohandler.RegisterRoutes(api, db, rdb, transcodePublisher)

	// 基础内容模块
	categoryhandler.RegisterRoutes(api, db)
	taghandler.RegisterRoutes(api, db)
	historyhandler.RegisterRoutes(api, db)
	searchhandler.RegisterRoutes(api, db)

	// 创作者中心（独立的 Repository → Service → Handler 链）
	creatorRepo := creatorrepo.NewRepository(db)
	creatorSvc := creatorservice.NewService(creatorRepo)
	creatorhandler.RegisterRoutes(api, creatorSvc)

	// 管理后台
	adminRepo := adminrepo.NewRepository(db)
	adminSvc := adminservice.NewService(adminRepo)
	adminhandler.RegisterRoutes(api, db, adminSvc)

	// 弹幕模块
	danmakuhandler.RegisterRoutes(api, db, rdb)

	// 社交互动模块（共享 messageSvc 用于发送通知）
	messageRepo := messagerepo.NewRepository(db)
	messageSvc := messageservice.NewService(messageRepo, rdb)

	commenthandler.RegisterRoutes(api, db, rdb, messageSvc)
	likehandler.RegisterRoutes(api, db, rdb, messageSvc)
	coinhandler.RegisterRoutes(api, db, rdb)
	favoritehandler.RegisterRoutes(api, db)
	followhandler.RegisterRoutes(api, db, messageSvc)
	notificationhandler.RegisterRoutes(api, messageSvc)
	interactionhandler.RegisterRoutes(api, db, rdb)

	// 转码和画质模块
	transcodehdlr.RegisterRoutes(api, db)
	qualityhandler.RegisterRoutes(api, db)

	// --- 启动 HTTP 服务 ---
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logger.Info("server starting")
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// seedCategories 当分类表为空时，插入预设的视频分类数据。
// db 为 GORM 数据库实例，用于查询和写入分类表。
func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&categorymodel.Category{}).Count(&count)
	if count > 0 {
		return // 已有数据，跳过
	}

	// 预设分类列表
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
