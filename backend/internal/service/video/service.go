// Package video 提供视频相关的核心业务逻辑服务，
// 包括视频上传（含封面）、视频查询、播放链接生成、视频编辑/删除、
// 视频列表、热门视频排行以及转码任务触发等功能。
package video

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strconv"
	"time"

	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	"github.com/My-TuDo/B-B/backend/internal/worker"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 视频服务，封装视频相关的核心业务逻辑。
type Service struct {
	repo       *videorepo.Repository
	rdb        *redis.Client
	rmqPublish func(videoID uint) error // 转码任务发布函数
	db         *gorm.DB
}

// NewService 创建视频服务实例（不含Redis）。
func NewService(repo *videorepo.Repository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithRedis 创建视频服务实例（含Redis）。
func NewServiceWithRedis(repo *videorepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// SetTranscodePublisher 设置上传后发布转码任务所使用的回调函数。
// SetTranscodePublisher sets the function used to publish transcode tasks after upload.
func (s *Service) SetTranscodePublisher(fn func(videoID uint) error) {
	s.rmqPublish = fn
}

// SetDB 设置数据库连接，用于转码时的直接回退处理。
// SetDB sets the database connection for fallback direct processing.
func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

// UploadVideo 上传视频文件到MinIO存储，同时可选上传封面图，
// 创建数据库记录（状态为草稿），并触发转码任务。
// progressFn 用于回调上传进度。
func (s *Service) UploadVideo(ctx context.Context, userID uint, file io.Reader, fileName string, fileSize int64, contentType string, title, description string, categoryID uint, progressFn func(uploaded, total int64), coverFile multipart.File, coverHeader *multipart.FileHeader) (*videomodel.VideoResp, error) {
	// Validate file size (500MB) — 校验文件大小上限为500MB
	maxSize := int64(500 * 1024 * 1024)
	if fileSize > maxSize {
		return nil, fmt.Errorf("video.service.UploadVideo: %w", newError(errcode.FileTooLarge))
	}

	// Generate object name — 生成唯一对象名：{userID}/{uuid}.{ext}
	ext := getFileExt(fileName)
	objectName := fmt.Sprintf("%d/%s%s", userID, uuid.New().String(), ext)

	// Create progress reader — 包装上传进度回调
	reader := &progressReader{
		reader: file,
		total:  fileSize,
		fn:     progressFn,
	}

	// Upload to MinIO — 上传视频到对象存储
	if err := storage.UploadVideo(ctx, objectName, reader, fileSize, contentType); err != nil {
		return nil, fmt.Errorf("video.service.UploadVideo: %w", err)
	}

	// Upload cover if present (failure does not block video)
	// 若提供了封面图则上传，封面失败不阻塞视频上传
	var coverObjectName string
	var coverErr error
	if coverFile != nil && coverHeader != nil {
		coverObjectName, coverErr = uploadCoverToStorage(ctx, userID, coverFile, coverHeader)
	}

	// Create DB record — 创建视频数据库记录，初始状态为草稿
	video := &videomodel.Video{
		UserID:      userID,
		Title:       title,
		Description: description,
		CoverURL:    coverObjectName,
		VideoURL:    objectName,
		FileSize:    uint64(fileSize),
		CategoryID:  categoryID,
		Status:      0, // draft — 草稿状态
	}
	if err := s.repo.Create(ctx, video); err != nil {
		// Rollback: delete video from MinIO — 数据库写入失败时回滚MinIO文件
		_ = storage.DeleteVideo(ctx, objectName)
		if coverObjectName != "" {
			_ = storage.DeleteVideo(ctx, coverObjectName)
		}
		return nil, fmt.Errorf("video.service.UploadVideo: %w", err)
	}

	// 触发转码（封面失败不影响转码）
	if coverErr != nil {
		s.triggerTranscode(video.ID)
		return toVideoResp(ctx, video), fmt.Errorf("video.service.UploadVideo: %w", coverErr)
	}
	s.triggerTranscode(video.ID)
	return toVideoResp(ctx, video), nil
}

// GetVideo 获取单个视频的详细信息，同时递增播放量。
// 非公开状态的视频仅作者本人可见。
func (s *Service) GetVideo(ctx context.Context, id uint, viewerID uint) (*videomodel.VideoResp, error) {
	// 查询视频
	video, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video.service.GetVideo: %w", err)
	}
	if video == nil {
		return nil, fmt.Errorf("video.service.GetVideo: %w", newError(errcode.VideoNotFound))
	}

	// status != 1 only visible to author — 非公开视频仅作者可见
	if video.Status != 1 && video.UserID != viewerID {
		return nil, fmt.Errorf("video.service.GetVideo: %w", newError(errcode.VideoNotFound))
	}

	// Increment views — 递增播放量
	_ = s.repo.IncrementViews(ctx, id)
	video.Views++

	resp := toVideoResp(ctx, video)

	// Generate cover presigned URL if cover exists — 为封面生成公网URL
	if video.CoverURL != "" {
		resp.CoverURL = storage.GetObjectURL(video.CoverURL)
	}

	return resp, nil
}

// GetPlayURL 获取视频的播放预签名URL（有效期1小时）。
// 非公开状态的视频仅作者本人可获取。
func (s *Service) GetPlayURL(ctx context.Context, videoID uint, viewerID uint) (string, error) {
	// 查询视频
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", err)
	}
	if video == nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", newError(errcode.VideoNotFound))
	}

	// 权限校验
	if video.Status != 1 && video.UserID != viewerID {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", newError(errcode.VideoNotFound))
	}

	// 生成预签名URL（有效期1小时）
	url, err := storage.GetPresignedURL(ctx, video.VideoURL, time.Hour)
	if err != nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", err)
	}

	return url, nil
}

// UpdateVideo 更新视频信息（标题、描述、分类、状态）。
// 仅视频作者可以编辑。
// 当状态变更为"发布"(1)时，会阻塞等待转码完成。
func (s *Service) UpdateVideo(ctx context.Context, userID uint, videoID uint, req *videomodel.UpdateVideoReq) (*videomodel.VideoResp, error) {
	// 查询视频
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", err)
	}
	if video == nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.VideoNotFound))
	}
	// 权限校验：仅作者
	if video.UserID != userID {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.Forbidden))
	}

	// 按需更新字段
	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if req.CategoryID != nil {
		video.CategoryID = *req.CategoryID
	}
	if req.Status != nil {
		// If publishing (status=1), block until transcode is complete
		// 发布前检查转码状态
		if *req.Status == 1 {
			var task tmodel.TranscodeTask
			if err := s.db.WithContext(ctx).Where("video_id = ?", videoID).First(&task).Error; err == nil {
				if task.Status != tmodel.StatusDone {
					switch task.Status {
					case tmodel.StatusFailed:
						return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.TranscodeFailed))
					default:
						return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.TranscodeNotReady))
					}
				}
			}
			// No transcode task at all — ffmpeg not available, allow publish with raw mp4
			// 无转码任务（ffmpeg不可用），允许直接发布原始mp4
		}
		video.Status = *req.Status
	}

	// 持久化更新
	if err := s.repo.Update(ctx, video); err != nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", err)
	}

	return toVideoResp(ctx, video), nil
}

// DeleteVideo 软删除视频（将状态标记为3）。
// 仅视频作者可以删除。
func (s *Service) DeleteVideo(ctx context.Context, userID uint, videoID uint) error {
	// 查询视频
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", err)
	}
	if video == nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", newError(errcode.VideoNotFound))
	}
	// 权限校验：仅作者
	if video.UserID != userID {
		return fmt.Errorf("video.service.DeleteVideo: %w", newError(errcode.Forbidden))
	}

	// 软删除：状态置为3
	video.Status = 3
	if err := s.repo.Update(ctx, video); err != nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", err)
	}

	return nil
}

// ListVideos 分页获取公开视频列表，支持按分类筛选。
func (s *Service) ListVideos(ctx context.Context, page, pageSize int, categoryID uint) (*videomodel.VideoListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询视频列表
	videos, total, err := s.repo.List(ctx, page, pageSize, categoryID)
	if err != nil {
		return nil, fmt.Errorf("video.service.ListVideos: %w", err)
	}

	// 组装响应
	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = *toVideoResp(ctx, &v)
	}

	// 为封面生成公网URL
	s.presignCovers(ctx, items)

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListUserVideos 分页获取指定用户的视频列表，支持按状态筛选。
func (s *Service) ListUserVideos(ctx context.Context, userID uint, statusFilter *int8, page, pageSize int) (*videomodel.VideoListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	// 查询该用户的视频
	videos, total, err := s.repo.ListByUser(ctx, userID, statusFilter, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("video.service.ListUserVideos: %w", err)
	}

	// 组装响应
	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = *toVideoResp(ctx, &v)
	}

	// 为封面生成公网URL
	s.presignCovers(ctx, items)

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// toVideoResp 将视频模型转换为对外响应结构，包含作者信息和头像预签名。
func toVideoResp(ctx context.Context, v *videomodel.Video) *videomodel.VideoResp {
	resp := &videomodel.VideoResp{
		ID:          v.ID,
		Title:       v.Title,
		Description: v.Description,
		CoverURL:    v.CoverURL,
		VideoURL:    v.VideoURL,
		Duration:    v.Duration,
		FileSize:    v.FileSize,
		CategoryID:  v.CategoryID,
		Status:      v.Status,
		Views:       v.Views,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}

	// 填充作者简要信息
	if v.User.ID != 0 {
		resp.User = &usermodel.UserBrief{
			ID:       v.User.ID,
			Username: v.User.Username,
			Nickname: v.User.Nickname,
		}
		if v.User.Avatar != "" {
			resp.User.Avatar = storage.GetObjectURL(v.User.Avatar)
		}
	}

	return resp
}

// getFileExt 从文件名中提取扩展名（含点号 "."）。
// 若文件名不含点号则返回空字符串。
func getFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

// presignCovers 批量将视频列表中的封面路径替换为公网访问URL。
// 失败时静默处理——封面URL置空。
// presignCovers replaces each non-empty CoverURL with a presigned URL.
// Failures are silently handled — the cover URL is set to empty if presigning fails.
func (s *Service) presignCovers(ctx context.Context, videos []videomodel.VideoResp) {
	for i := range videos {
		if videos[i].CoverURL != "" {
			videos[i].CoverURL = storage.GetObjectURL(videos[i].CoverURL)
		}
	}
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

// progressReader 包装 io.Reader，在每次读取时回调进度函数。
// progressReader wraps an io.Reader and calls fn with progress.
type progressReader struct {
	reader   io.Reader
	total    int64
	uploaded int64
	fn       func(uploaded, total int64)
}

// HotVideos 获取热门视频排行榜。使用 Redis 缓存排序结果（有效期10分钟）。
// 排序算法：综合播放量和发布时间，近期高播放视频得分更高。
func (s *Service) HotVideos(ctx context.Context, page, pageSize int) (*videomodel.VideoListResp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	cacheKey := "video:hot"

	// Try cache — 优先从Redis缓存读取
	if s.rdb != nil {
		length, err := s.rdb.LLen(ctx, cacheKey).Result()
		if err == nil && length > 0 {
			start := int64((page - 1) * pageSize)
			stop := start + int64(pageSize) - 1

			ids, err := s.rdb.LRange(ctx, cacheKey, start, stop).Result()
			if err == nil && len(ids) > 0 {
				// 根据缓存的ID列表查询视频详情
				items := make([]videomodel.VideoResp, 0, len(ids))
				for _, idStr := range ids {
					id, parseErr := strconv.ParseUint(idStr, 10, 64)
					if parseErr != nil {
						continue
					}
					video, repoErr := s.repo.FindByID(ctx, uint(id))
					if repoErr != nil || video == nil {
						continue
					}
					if video.Status != 1 {
						continue
					}
					items = append(items, *toVideoResp(ctx, video))
				}

					s.presignCovers(ctx, items)

					total, _ := s.rdb.LLen(ctx, cacheKey).Result()
					return &videomodel.VideoListResp{
						Items:    items,
						Total:    total,
						Page:     page,
						PageSize: pageSize,
					}, nil
			}
		}
	}

	// Build from DB — 缓存未命中，从数据库计算热门排行
	videos, err := s.repo.ListAllPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("video.service.HotVideos: %w", err)
	}

	now := time.Now()
	type scored struct {
		id    uint
		score float64
	}
	// 计算热度得分：播放量 * 0.5 - 发布小时数 * 0.1
	scoredList := make([]scored, len(videos))
	for i, v := range videos {
		hoursSincePublish := now.Sub(v.CreatedAt).Hours()
		score := float64(v.Views)*0.5 - hoursSincePublish*0.1
		scoredList[i] = scored{id: v.ID, score: score}
	}

	// Simple sort by score DESC (bubble for simplicity since list is usually small)
	// 按得分降序排列（冒泡排序，因列表通常不大）
	for i := 0; i < len(scoredList); i++ {
		for j := i + 1; j < len(scoredList); j++ {
			if scoredList[j].score > scoredList[i].score {
				scoredList[i], scoredList[j] = scoredList[j], scoredList[i]
			}
		}
	}

	// Cache to Redis — 将排序结果缓存到Redis（有效期10分钟）
	if s.rdb != nil {
		s.rdb.Del(ctx, cacheKey)
		pipe := s.rdb.Pipeline()
		for _, sc := range scoredList {
			pipe.RPush(ctx, cacheKey, strconv.FormatUint(uint64(sc.id), 10))
		}
		pipe.Expire(ctx, cacheKey, 600*time.Second)
		pipe.Exec(ctx)
	}

	// Build response — 分页构建响应
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(scoredList) {
		end = len(scoredList)
	}

	items := make([]videomodel.VideoResp, 0, end-start)
	for i := start; i < end; i++ {
		video, repoErr := s.repo.FindByID(ctx, scoredList[i].id)
		if repoErr != nil || video == nil {
			continue
		}
		items = append(items, *toVideoResp(ctx, video))
	}

	s.presignCovers(ctx, items)

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    int64(len(scoredList)),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Ranking 获取排行榜视频，支持 day（日榜）、week（周榜）、total（总榜）三种周期。
// 使用 Redis ZSet 缓存排序结果（有效期10分钟）。
func (s *Service) Ranking(ctx context.Context, period string, page, pageSize int) (*model_ranking_resp, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	cacheKey := "video:ranking:" + period
	now := time.Now()

	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1

	// 优先从Redis ZSet缓存读取
	if s.rdb != nil {
		total, err := s.rdb.ZCard(ctx, cacheKey).Result()
		if err == nil && total > 0 {
			ids, err := s.rdb.ZRevRange(ctx, cacheKey, start, stop).Result()
			if err == nil {
				items := make([]videomodel.VideoResp, 0, len(ids))
				for _, idStr := range ids {
					id, parseErr := strconv.ParseUint(idStr, 10, 64)
					if parseErr != nil {
						continue
					}
					video, repoErr := s.repo.FindByID(ctx, uint(id))
					if repoErr != nil || video == nil {
						continue
					}
					items = append(items, *toVideoResp(ctx, video))
				}

					s.presignCovers(ctx, items)

					return &model_ranking_resp{
						Items:    items,
						Total:    total,
						Page:     page,
						PageSize: pageSize,
						Period:   period,
					}, nil
			}
		}
	}

	// Build from DB — 缓存未命中，从数据库计算
	videos, err := s.repo.ListAllPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("video.service.Ranking: %w", err)
	}

	// 按周期过滤
	var filtered []videomodel.Video
	cutoff := time.Time{}

	switch period {
	case "day":
		// 当日 0:00 起
		cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		// 本周一 0:00 起
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		cutoff = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	default: // total — 全部
		cutoff = time.Time{}
	}

	for _, v := range videos {
		if period == "total" || v.CreatedAt.After(cutoff) || v.CreatedAt.Equal(cutoff) {
			filtered = append(filtered, v)
		}
	}

	// Sort by views DESC — 按播放量降序排列
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Views > filtered[i].Views {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Cache to Redis — 缓存结果到Redis（有效期10分钟，最多缓存1000条）
	if s.rdb != nil {
		s.rdb.Del(ctx, cacheKey)
		pipe := s.rdb.Pipeline()
		for i, v := range filtered {
			if i >= 1000 {
				break
			}
			pipe.ZAdd(ctx, cacheKey, redis.Z{
				Score:  float64(v.Views),
				Member: strconv.FormatUint(uint64(v.ID), 10),
			})
		}
		pipe.Expire(ctx, cacheKey, 600*time.Second)
		pipe.Exec(ctx)
	}

	// 分页截取
	sIdx := int(start)
	eIdx := int(math.Min(float64(stop+1), float64(len(filtered))))
	if sIdx >= len(filtered) {
		sIdx = 0
		eIdx = 0
	}

	items := make([]videomodel.VideoResp, 0)
	for i := sIdx; i < eIdx; i++ {
		items = append(items, *toVideoResp(ctx, &filtered[i]))
	}

	s.presignCovers(ctx, items)

	return &model_ranking_resp{
		Items:    items,
		Total:    int64(len(filtered)),
		Page:     page,
		PageSize: pageSize,
		Period:   period,
	}, nil
}

// model_ranking_resp 排行榜响应结构。
type model_ranking_resp struct {
	Items    []videomodel.VideoResp `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Period   string                 `json:"period"`
}

// Read 实现 io.Reader 接口，在读取数据时回调进度函数。
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.uploaded += int64(n)
	if pr.fn != nil {
		pr.fn(pr.uploaded, pr.total)
	}
	return n, err
}

// coverExtByMIME 根据 MIME 类型返回对应的文件扩展名。
// coverExtByMIME returns the file extension for common image MIME types.
func coverExtByMIME(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

// uploadCoverToStorage 将封面图上传到MinIO并返回对象键名。
// uploadCoverToStorage uploads a cover image to MinIO and returns the object name.
func uploadCoverToStorage(ctx context.Context, userID uint, coverFile multipart.File, coverHeader *multipart.FileHeader) (string, error) {
	contentType := coverHeader.Header.Get("Content-Type")
	ext := coverExtByMIME(contentType)
	if ext == "" {
		ext = getFileExt(coverHeader.Filename)
	}
	objectName := fmt.Sprintf("%d/cover_%s%s", userID, uuid.New().String(), ext)

	if err := storage.UploadVideo(ctx, objectName, coverFile, coverHeader.Size, contentType); err != nil {
		return "", fmt.Errorf("uploadCoverToStorage: %w", err)
	}
	return objectName, nil
}

// triggerTranscode 发布转码任务。非阻塞，错误仅记录日志。
// 若RabbitMQ不可用则直接回退到本地goroutine处理。
// triggerTranscode publishes a transcode task. Non-blocking, errors are logged.
func (s *Service) triggerTranscode(videoID uint) {
	if s.rmqPublish != nil {
		go func() {
			if err := s.rmqPublish(videoID); err != nil {
				logger.Warn("trigger transcode publish failed, falling back to direct processing", zap.Uint("video_id", videoID), zap.Error(err))
				// Fallback: process directly in a goroutine — 回退到直接本地处理
				if s.db != nil {
					go worker.ProcessVideo(videoID, s.db)
				}
			}
		}()
	} else {
		// No RabbitMQ at all, process directly — 无消息队列时直接本地处理
		if s.db != nil {
			go worker.ProcessVideo(videoID, s.db)
		}
	}
}
