package video

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strconv"
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	repo    *videorepo.Repository
	rdb     *redis.Client
	rmqPublish func(videoID uint) error
}

func NewService(repo *videorepo.Repository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithRedis(repo *videorepo.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

// SetTranscodePublisher sets the function used to publish transcode tasks after upload.
func (s *Service) SetTranscodePublisher(fn func(videoID uint) error) {
	s.rmqPublish = fn
}

func (s *Service) UploadVideo(ctx context.Context, userID uint, file io.Reader, fileName string, fileSize int64, contentType string, title, description string, categoryID uint, progressFn func(uploaded, total int64), coverFile multipart.File, coverHeader *multipart.FileHeader) (*videomodel.VideoResp, error) {
	// Validate file size (500MB)
	maxSize := int64(500 * 1024 * 1024)
	if fileSize > maxSize {
		return nil, fmt.Errorf("video.service.UploadVideo: %w", newError(errcode.FileTooLarge))
	}

	// Generate object name
	ext := getFileExt(fileName)
	objectName := fmt.Sprintf("%d/%s%s", userID, uuid.New().String(), ext)

	// Create progress reader
	reader := &progressReader{
		reader: file,
		total:  fileSize,
		fn:     progressFn,
	}

	// Upload to MinIO
	if err := storage.UploadVideo(ctx, objectName, reader, fileSize, contentType); err != nil {
		return nil, fmt.Errorf("video.service.UploadVideo: %w", err)
	}

	// Upload cover if present (failure does not block video)
	var coverObjectName string
	var coverErr error
	if coverFile != nil && coverHeader != nil {
		coverObjectName, coverErr = uploadCoverToStorage(ctx, userID, coverFile, coverHeader)
	}

	// Create DB record
	video := &videomodel.Video{
		UserID:      userID,
		Title:       title,
		Description: description,
		CoverURL:    coverObjectName,
		VideoURL:    objectName,
		FileSize:    uint64(fileSize),
		CategoryID:  categoryID,
		Status:      0, // draft
	}
	if err := s.repo.Create(ctx, video); err != nil {
		// Rollback: delete video from MinIO
		_ = storage.DeleteVideo(ctx, objectName)
		if coverObjectName != "" {
			_ = storage.DeleteVideo(ctx, coverObjectName)
		}
		return nil, fmt.Errorf("video.service.UploadVideo: %w", err)
	}

	if coverErr != nil {
		// Trigger transcode regardless of cover error
		s.triggerTranscode(video.ID)
		return toVideoResp(ctx, video), fmt.Errorf("video.service.UploadVideo: %w", coverErr)
	}
	s.triggerTranscode(video.ID)
	return toVideoResp(ctx, video), nil
}

func (s *Service) GetVideo(ctx context.Context, id uint, viewerID uint) (*videomodel.VideoResp, error) {
	video, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video.service.GetVideo: %w", err)
	}
	if video == nil {
		return nil, fmt.Errorf("video.service.GetVideo: %w", newError(errcode.VideoNotFound))
	}

	// status != 1 only visible to author
	if video.Status != 1 && video.UserID != viewerID {
		return nil, fmt.Errorf("video.service.GetVideo: %w", newError(errcode.VideoNotFound))
	}

	// Increment views
	_ = s.repo.IncrementViews(ctx, id)
	video.Views++

	resp := toVideoResp(ctx, video)

	// Generate cover presigned URL if cover exists
	if video.CoverURL != "" {
		url, err := storage.GetPresignedURL(ctx, video.CoverURL, time.Hour)
		if err == nil {
			resp.CoverURL = url
		}
	}

	return resp, nil
}

func (s *Service) GetPlayURL(ctx context.Context, videoID uint, viewerID uint) (string, error) {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", err)
	}
	if video == nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", newError(errcode.VideoNotFound))
	}

	if video.Status != 1 && video.UserID != viewerID {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", newError(errcode.VideoNotFound))
	}

	url, err := storage.GetPresignedURL(ctx, video.VideoURL, time.Hour)
	if err != nil {
		return "", fmt.Errorf("video.service.GetPlayURL: %w", err)
	}

	return url, nil
}

func (s *Service) UpdateVideo(ctx context.Context, userID uint, videoID uint, req *videomodel.UpdateVideoReq) (*videomodel.VideoResp, error) {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", err)
	}
	if video == nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.VideoNotFound))
	}
	if video.UserID != userID {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", newError(errcode.Forbidden))
	}

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
		video.Status = *req.Status
	}

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, fmt.Errorf("video.service.UpdateVideo: %w", err)
	}

	return toVideoResp(ctx, video), nil
}

func (s *Service) DeleteVideo(ctx context.Context, userID uint, videoID uint) error {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", err)
	}
	if video == nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", newError(errcode.VideoNotFound))
	}
	if video.UserID != userID {
		return fmt.Errorf("video.service.DeleteVideo: %w", newError(errcode.Forbidden))
	}
	if video.Status != 0 {
		return fmt.Errorf("video.service.DeleteVideo: only drafts can be deleted")
	}

	video.Status = 3 // soft delete
	if err := s.repo.Update(ctx, video); err != nil {
		return fmt.Errorf("video.service.DeleteVideo: %w", err)
	}

	return nil
}

func (s *Service) ListVideos(ctx context.Context, page, pageSize int, categoryID uint) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	videos, total, err := s.repo.List(ctx, page, pageSize, categoryID)
	if err != nil {
		return nil, fmt.Errorf("video.service.ListVideos: %w", err)
	}

	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = *toVideoResp(ctx, &v)
	}

	s.presignCovers(ctx, items)

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) ListUserVideos(ctx context.Context, userID uint, statusFilter *int8, page, pageSize int) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	videos, total, err := s.repo.ListByUser(ctx, userID, statusFilter, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("video.service.ListUserVideos: %w", err)
	}

	items := make([]videomodel.VideoResp, len(videos))
	for i, v := range videos {
		items[i] = *toVideoResp(ctx, &v)
	}

	s.presignCovers(ctx, items)

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

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

	if v.User.ID != 0 {
		resp.User = &usermodel.UserBrief{
			ID:       v.User.ID,
			Username: v.User.Username,
			Nickname: v.User.Nickname,
		}
		if v.User.Avatar != "" {
			if url, err := storage.GetPresignedURL(ctx, v.User.Avatar, time.Hour); err == nil {
				resp.User.Avatar = url
			}
		}
	}

	return resp
}

func getFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

// presignCovers replaces each non-empty CoverURL with a presigned URL.
// Failures are silently handled — the cover URL is set to empty if presigning fails.
func (s *Service) presignCovers(ctx context.Context, videos []videomodel.VideoResp) {
	for i := range videos {
		if videos[i].CoverURL != "" {
			url, err := storage.GetPresignedURL(ctx, videos[i].CoverURL, time.Hour)
			if err != nil {
				videos[i].CoverURL = ""
			} else {
				videos[i].CoverURL = url
			}
		}
	}
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

// progressReader wraps an io.Reader and calls fn with progress.
type progressReader struct {
	reader   io.Reader
	total    int64
	uploaded int64
	fn       func(uploaded, total int64)
}

func (s *Service) HotVideos(ctx context.Context, page, pageSize int) (*videomodel.VideoListResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 12
	}

	cacheKey := "video:hot"

	// Try cache
	if s.rdb != nil {
		length, err := s.rdb.LLen(ctx, cacheKey).Result()
		if err == nil && length > 0 {
			start := int64((page - 1) * pageSize)
			stop := start + int64(pageSize) - 1

			ids, err := s.rdb.LRange(ctx, cacheKey, start, stop).Result()
			if err == nil && len(ids) > 0 {
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

	// Build from DB
	videos, err := s.repo.ListAllPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("video.service.HotVideos: %w", err)
	}

	now := time.Now()
	type scored struct {
		id    uint
		score float64
	}
	scoredList := make([]scored, len(videos))
	for i, v := range videos {
		hoursSincePublish := now.Sub(v.CreatedAt).Hours()
		score := float64(v.Views)*0.5 - hoursSincePublish*0.1
		scoredList[i] = scored{id: v.ID, score: score}
	}

	// Simple sort by score DESC (bubble for simplicity since list is usually small)
	for i := 0; i < len(scoredList); i++ {
		for j := i + 1; j < len(scoredList); j++ {
			if scoredList[j].score > scoredList[i].score {
				scoredList[i], scoredList[j] = scoredList[j], scoredList[i]
			}
		}
	}

	// Cache to Redis
	if s.rdb != nil {
		s.rdb.Del(ctx, cacheKey)
		pipe := s.rdb.Pipeline()
		for _, sc := range scoredList {
			pipe.RPush(ctx, cacheKey, strconv.FormatUint(uint64(sc.id), 10))
		}
		pipe.Expire(ctx, cacheKey, 600*time.Second)
		pipe.Exec(ctx)
	}

	// Build response
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

func (s *Service) Ranking(ctx context.Context, period string, page, pageSize int) (*model_ranking_resp, error) {
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

	// Build from DB
	videos, err := s.repo.ListAllPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("video.service.Ranking: %w", err)
	}

	var filtered []videomodel.Video
	cutoff := time.Time{}

	switch period {
	case "day":
		cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		cutoff = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	default: // total
		cutoff = time.Time{}
	}

	for _, v := range videos {
		if period == "total" || v.CreatedAt.After(cutoff) || v.CreatedAt.Equal(cutoff) {
			filtered = append(filtered, v)
		}
	}

	// Sort by views DESC
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Views > filtered[i].Views {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Cache to Redis
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

type model_ranking_resp struct {
	Items    []videomodel.VideoResp `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Period   string                 `json:"period"`
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.uploaded += int64(n)
	if pr.fn != nil {
		pr.fn(pr.uploaded, pr.total)
	}
	return n, err
}

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

// triggerTranscode publishes a transcode task. Non-blocking, errors are logged.
func (s *Service) triggerTranscode(videoID uint) {
	if s.rmqPublish != nil {
		go func() {
			if err := s.rmqPublish(videoID); err != nil {
				logger.Warn("trigger transcode publish failed", zap.Uint("video_id", videoID), zap.Error(err))
				// Fallback: process directly in a goroutine
				// We'd need db access here, so just log the warning
			}
		}()
	}
}
