package video

import (
	"context"
	"fmt"
	"io"
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/google/uuid"
)

type Service struct {
	repo *videorepo.Repository
}

func NewService(repo *videorepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UploadVideo(ctx context.Context, userID uint, file io.Reader, fileName string, fileSize int64, contentType string, title, description string, categoryID uint, progressFn func(uploaded, total int64)) (*videomodel.VideoResp, error) {
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

	// Create DB record
	video := &videomodel.Video{
		UserID:      userID,
		Title:       title,
		Description: description,
		VideoURL:    objectName,
		FileSize:    uint64(fileSize),
		CategoryID:  categoryID,
		Status:      0, // draft
	}
	if err := s.repo.Create(ctx, video); err != nil {
		// Rollback: delete from MinIO
		_ = storage.DeleteVideo(ctx, objectName)
		return nil, fmt.Errorf("video.service.UploadVideo: %w", err)
	}

	return toVideoResp(video), nil
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

	return toVideoResp(video), nil
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

	return toVideoResp(video), nil
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
		items[i] = *toVideoResp(&v)
	}

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
		items[i] = *toVideoResp(&v)
	}

	return &videomodel.VideoListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func toVideoResp(v *videomodel.Video) *videomodel.VideoResp {
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
			Avatar:   v.User.Avatar,
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

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.uploaded += int64(n)
	if pr.fn != nil {
		pr.fn(pr.uploaded, pr.total)
	}
	return n, err
}
