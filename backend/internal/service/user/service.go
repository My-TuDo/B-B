package user

import (
	"context"
	"fmt"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	userrepo "github.com/My-TuDo/B-B/backend/internal/repository/user"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
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

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
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

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
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
