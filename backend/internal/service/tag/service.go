package tag

import (
	"context"
	"fmt"

	tagmodel "github.com/My-TuDo/B-B/backend/internal/model/tag"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	tagrepo "github.com/My-TuDo/B-B/backend/internal/repository/tag"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
)

type Service struct {
	repo      *tagrepo.Repository
	videoRepo *videorepo.Repository
}

func NewService(repo *tagrepo.Repository, videoRepo *videorepo.Repository) *Service {
	return &Service{repo: repo, videoRepo: videoRepo}
}

func (s *Service) List(ctx context.Context) ([]tagmodel.TagResp, error) {
	tags, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("tag.service.List: %w", err)
	}

	resp := make([]tagmodel.TagResp, len(tags))
	for i, t := range tags {
		resp[i] = tagmodel.TagResp{ID: t.ID, Name: t.Name}
	}

	return resp, nil
}

func (s *Service) SetVideoTags(ctx context.Context, userID, videoID uint, req *tagmodel.SetVideoTagsReq) error {
	video, err := s.videoRepo.FindByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", err)
	}
	if video == nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", newError(errcode.VideoNotFound))
	}
	if video.UserID != userID {
		return fmt.Errorf("tag.service.SetVideoTags: %w", newError(errcode.Forbidden))
	}

	if err := s.repo.ReplaceVideoTags(ctx, videoID, req.TagIDs); err != nil {
		return fmt.Errorf("tag.service.SetVideoTags: %w", err)
	}

	return nil
}

func (s *Service) Create(ctx context.Context, name string) (*tagmodel.TagResp, error) {
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("tag.service.Create: %w", err)
	}
	if existing != nil {
		return &tagmodel.TagResp{ID: existing.ID, Name: existing.Name}, nil
	}

	tag := &tagmodel.Tag{Name: name}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("tag.service.Create: %w", err)
	}

	return &tagmodel.TagResp{ID: tag.ID, Name: tag.Name}, nil
}

func (s *Service) GetVideoTags(ctx context.Context, videoID uint) ([]tagmodel.TagResp, error) {
	tags, err := s.repo.GetVideoTags(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("tag.service.GetVideoTags: %w", err)
	}

	resp := make([]tagmodel.TagResp, len(tags))
	for i, t := range tags {
		resp[i] = tagmodel.TagResp{ID: t.ID, Name: t.Name}
	}

	return resp, nil
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
