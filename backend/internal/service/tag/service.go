// Package tag 提供标签相关的业务逻辑服务，
// 包括标签列表查询、创建标签、为视频设置标签以及获取视频标签等功能。
package tag

import (
	"context"
	"fmt"

	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	tagrepo "github.com/My-TuDo/B-B/backend/internal/repository/tag"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
)

// Service 标签服务，封装标签相关的业务逻辑。
type Service struct {
	repo      *tagrepo.Repository
	videoRepo *videorepo.Repository
}

// NewService 创建标签服务实例。
func NewService(repo *tagrepo.Repository, videoRepo *videorepo.Repository) *Service {
	return &Service{repo: repo, videoRepo: videoRepo}
}

// List 获取全部标签列表。
func (s *Service) List(ctx context.Context) ([]tagmodel.TagResp, error) {
	// 查询所有标签
	tags, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("tag.service.List: %w", err)
	}

	// 转换为响应格式
	resp := make([]tagmodel.TagResp, len(tags))
	for i, t := range tags {
		resp[i] = tagmodel.TagResp{ID: t.ID, Name: t.Name}
	}

	return resp, nil
}

// SetVideoTags 为指定视频设置标签（全量替换模式）。
// 仅视频所有者可以操作。
func (s *Service) SetVideoTags(ctx context.Context, userID, videoID uint, req *tagmodel.SetVideoTagsReq) error {
	// 校验视频存在
	video, err := s.videoRepo.FindByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", err)
	}
	if video == nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", newError(errcode.VideoNotFound))
	}
	// 权限校验：仅视频所有者
	if video.UserID != userID {
		return fmt.Errorf("tag.service.SetVideoTags: %w", newError(errcode.Forbidden))
	}

	// 全量替换标签关联
	if err := s.repo.ReplaceVideoTags(ctx, videoID, req.TagIDs); err != nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", err)
	}

	return nil
}

// Create 创建新标签，若标签已存在则直接返回已有标签。
func (s *Service) Create(ctx context.Context, name string) (*tagmodel.TagResp, error) {
	// 检查是否已存在同名标签
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("tag.service.Create: %w", err)
	}
	if existing != nil {
		// 已存在则直接返回
		return &tagmodel.TagResp{ID: existing.ID, Name: existing.Name}, nil
	}

	// 创建新标签
	tag := &tagmodel.Tag{Name: name}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("tag.service.Create: %w", err)
	}

	return &tagmodel.TagResp{ID: tag.ID, Name: tag.Name}, nil
}

// GetVideoTags 获取指定视频的所有标签。
func (s *Service) GetVideoTags(ctx context.Context, videoID uint) ([]tagmodel.TagResp, error) {
	// 查询视频关联的标签
	tags, err := s.repo.GetVideoTags(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("tag.service.GetVideoTags: %w", err)
	}

	// 转换为响应格式
	resp := make([]tagmodel.TagResp, len(tags))
	for i, t := range tags {
		resp[i] = tagmodel.TagResp{ID: t.ID, Name: t.Name}
	}

	return resp, nil
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
