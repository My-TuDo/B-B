package auth

import (
	"context"
	"fmt"
	"time"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"github.com/My-TuDo/B-B/backend/internal/repository/auth"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/jwt"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repo *auth.Repository
	rdb  *redis.Client
	db   *gorm.DB
}

func NewService(repo *auth.Repository, rdb *redis.Client, db *gorm.DB) *Service {
	return &Service{repo: repo, rdb: rdb, db: db}
}

func (s *Service) Register(ctx context.Context, req *usermodel.RegisterReq) (*usermodel.LoginResp, string, error) {
	// Check username uniqueness
	existing, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}
	if existing != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", newError(errcode.UserExists))
	}

	// Check email uniqueness
	existing, err = s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}
	if existing != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", newError(errcode.UserExists))
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	user := &usermodel.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Nickname:     req.Username,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	// Generate JWT
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	// Store in Redis whitelist
	redisKey := fmt.Sprintf("auth:token:%d", user.ID)
	if err := s.rdb.Set(ctx, redisKey, token, 7*24*time.Hour).Err(); err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	// Create default favorites folder
	var favCount int64
	s.db.WithContext(ctx).Table("favorites").Where("user_id = ? AND name = ?", user.ID, "默认收藏夹").Count(&favCount)
	if favCount == 0 {
		s.db.WithContext(ctx).Exec("INSERT INTO favorites (user_id, name, is_public, created_at) VALUES (?, ?, 1, NOW())", user.ID, "默认收藏夹")
	}

	resp := &usermodel.LoginResp{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   presignAvatar(ctx, user.Avatar),
	}

	return resp, token, nil
}

func (s *Service) Login(ctx context.Context, account, password string) (*usermodel.LoginResp, string, error) {
	user, err := s.repo.FindByUsernameOrEmail(ctx, account)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Login: %w", err)
	}
	if user == nil {
		return nil, "", fmt.Errorf("auth.service.Login: %w", newError(errcode.UserNotFound))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("auth.service.Login: %w", newError(errcode.PasswordWrong))
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Login: %w", err)
	}

	redisKey := fmt.Sprintf("auth:token:%d", user.ID)
	if err := s.rdb.Set(ctx, redisKey, token, 7*24*time.Hour).Err(); err != nil {
		return nil, "", fmt.Errorf("auth.service.Login: %w", err)
	}

	resp := &usermodel.LoginResp{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   presignAvatar(ctx, user.Avatar),
	}

	return resp, token, nil
}

func (s *Service) Logout(ctx context.Context, userID uint) error {
	redisKey := fmt.Sprintf("auth:token:%d", userID)
	if err := s.rdb.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("auth.service.Logout: %w", err)
	}
	return nil
}

func (s *Service) GetMe(ctx context.Context, userID uint) (*usermodel.UserResp, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth.service.GetMe: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("auth.service.GetMe: %w", newError(errcode.UserNotFound))
	}

	return &usermodel.UserResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    presignAvatar(ctx, user.Avatar),
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, oldToken string) (string, error) {
	claims, err := jwt.ParseTokenUnverified(oldToken)
	if err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", newError(errcode.TokenInvalid))
	}

	// Verify Redis whitelist
	redisKey := fmt.Sprintf("auth:token:%d", claims.UserID)
	storedToken, err := s.rdb.Get(ctx, redisKey).Result()
	if err != nil || storedToken != oldToken {
		return "", fmt.Errorf("auth.service.Refresh: %w", newError(errcode.Unauthorized))
	}

	// Generate new token
	newToken, err := jwt.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", err)
	}

	// Update Redis
	if err := s.rdb.Set(ctx, redisKey, newToken, 7*24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", err)
	}

	return newToken, nil
}

// Service error type for error code propagation
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

func presignAvatar(ctx context.Context, avatar string) string {
	if avatar == "" {
		return ""
	}
	if url, err := storage.GetPresignedURL(ctx, avatar, time.Hour); err == nil {
		return url
	}
	return "" // fallback: don't return unreadable raw key
}
