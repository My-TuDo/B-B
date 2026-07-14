// Package user 提供用户信息相关的业务逻辑服务，
// 包括用户详情查询、个人信息更新以及头像上传等功能。
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

// Service 用户服务，封装用户信息相关的业务逻辑。
type Service struct {
	repo *userrepo.Repository
}

// NewService 创建用户服务实例。
func NewService(repo *userrepo.Repository) *Service {
	return &Service{repo: repo}
}

// GetUser 根据用户ID获取用户公开信息。
func (s *Service) GetUser(ctx context.Context, id uint) (*usermodel.UserResp, error) {
	// 查询用户
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.GetUser: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.GetUser: %w", newError(errcode.UserNotFound))
	}

	// 生成头像预签名URL
	avatar := ""
	if user.Avatar != "" {
		avatar = storage.GetObjectURL(user.Avatar)
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

// UpdateUser 更新用户个人信息（昵称、头像key、简介）。
// 仅更新请求中提供的非 nil 字段。
func (s *Service) UpdateUser(ctx context.Context, id uint, req *usermodel.UpdateUserReq) (*usermodel.UserResp, error) {
	// 查询用户
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", newError(errcode.UserNotFound))
	}

	// 按需更新各字段
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	// 持久化更新
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user.service.UpdateUser: %w", err)
	}

	// 生成头像预签名URL
	avatar := ""
	if user.Avatar != "" {
		avatar = storage.GetObjectURL(user.Avatar)
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

// UploadAvatar 上传用户头像到MinIO存储，更新用户记录并返回带预签名URL的用户信息。
func (s *Service) UploadAvatar(ctx context.Context, id uint, reader io.Reader, size int64, contentType string, ext string) (*usermodel.UserResp, error) {
	// 校验用户存在
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", newError(errcode.UserNotFound))
	}

	// 构建对象存储路径：avatars/{userID}.{ext}
	objectKey := fmt.Sprintf("avatars/%d.%s", id, ext)

	// 上传到MinIO
	if err := storage.UploadFile(ctx, objectKey, reader, size, contentType); err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}

	// 更新用户头像字段
	user.Avatar = objectKey

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user.service.UploadAvatar: %w", err)
	}

	// Generate presigned URL for the response (1 hour expiry)
	// 为响应生成预签名URL（有效期1小时）
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
