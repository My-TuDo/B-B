package user

import (
	"context"
	"fmt"
	"io"
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	userrepo "github.com/My-TuDo/B-B/backend/internal/repository/user"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
)

type Service struct {
	repo *userrepo.Repository
}

func NewService(repo *userrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUser(ctx context.Context, id uint) (*usermodel.UserResp, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.GetUser: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.GetUser: %w", newError(errcode.UserNotFound))
	}

	// Generate presigned URL for avatar if set
	avatar := user.Avatar
	if user.Avatar != "" {
		if url, err := storage.GetPresignedURL(ctx, user.Avatar, time.Hour); err == nil {
			avatar = url
		}
	}

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uint, req *usermodel.UpdateUserReq) (*usermodel.UserResp, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", newError(errcode.UserNotFound))
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", err)
	}

	// Generate presigned URL for avatar if set
	avatar := user.Avatar
	if user.Avatar != "" {
		if url, err := storage.GetPresignedURL(ctx, user.Avatar, time.Hour); err == nil {
			avatar = url
		}
	}

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    avatar,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) UploadAvatar(ctx context.Context, id uint, reader io.Reader, size int64, contentType string, ext string) (*usermodel.UserResp, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", newError(errcode.UserNotFound))
	}

	objectKey := fmt.Sprintf("avatars/%d.%s", id, ext)

	if err := storage.UploadFile(ctx, objectKey, reader, size, contentType); err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}

	user.Avatar = objectKey

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}

	// Generate presigned URL for the response (1 hour expiry)
	avatarURL := user.Avatar
	if url, err := storage.GetPresignedURL(ctx, objectKey, time.Hour); err == nil {
		avatarURL = url
	}

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    avatarURL,
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}, nil
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
