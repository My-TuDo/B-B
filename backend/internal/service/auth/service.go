// Package auth 提供用户认证相关的业务逻辑服务，
// 包括注册、登录、登出、令牌刷新以及当前用户信息获取等功能。
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

// Service 认证服务，封装注册、登录、登出及令牌管理等业务逻辑。
type Service struct {
	repo *auth.Repository
	rdb  *redis.Client
	db   *gorm.DB
}

// NewService 创建认证服务实例。
func NewService(repo *auth.Repository, rdb *redis.Client, db *gorm.DB) *Service {
	return &Service{repo: repo, rdb: rdb, db: db}
}

// Register 用户注册：校验用户名和邮箱唯一性、加密密码、生成JWT、
// 将令牌存入Redis白名单，并自动创建默认收藏夹。
// 返回登录响应和JWT令牌字符串。
func (s *Service) Register(ctx context.Context, req *usermodel.RegisterReq) (*usermodel.LoginResp, string, error) {
	// Check username uniqueness — 检查用户名是否已被占用
	existing, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}
	if existing != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", newError(errcode.UserExists))
	}

	// Check email uniqueness — 检查邮箱是否已被占用
	existing, err = s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}
	if existing != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", newError(errcode.UserExists))
	}

	// Hash password — 使用 bcrypt 算法对密码进行哈希
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

	// Generate JWT — 为新用户生成访问令牌
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	// Store in Redis whitelist — 将令牌存入Redis白名单，7天有效
	redisKey := fmt.Sprintf("auth:token:%d", user.ID)
	if err := s.rdb.Set(ctx, redisKey, token, 7*24*time.Hour).Err(); err != nil {
		return nil, "", fmt.Errorf("auth.service.Register: %w", err)
	}

	// Create default favorites folder — 为新用户创建默认收藏夹
	var favCount int64
	s.db.WithContext(ctx).Table("favorites").Where("user_id = ? AND name = ?", user.ID, "默认收藏夹").Count(&favCount)
	if favCount == 0 {
		s.db.WithContext(ctx).Exec("INSERT INTO favorites (user_id, name, is_public, created_at) VALUES (?, ?, 1, NOW())", user.ID, "默认收藏夹")
	}

	resp := &usermodel.LoginResp{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   presignAvatar(user.Avatar),
	}

	return resp, token, nil
}

// Login 用户登录：通过用户名或邮箱查找用户、验证密码、
// 生成JWT并写入Redis白名单。返回登录响应和JWT令牌字符串。
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
		Avatar:   presignAvatar(user.Avatar),
	}

	return resp, token, nil
}

// Logout 用户登出：从Redis白名单中删除当前用户的令牌。
func (s *Service) Logout(ctx context.Context, userID uint) error {
	redisKey := fmt.Sprintf("auth:token:%d", userID)
	if err := s.rdb.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("auth.service.Logout: %w", err)
	}
	return nil
}

// GetMe 获取当前登录用户的个人信息。
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
		Avatar:    presignAvatar(user.Avatar),
		Bio:       user.Bio,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Refresh 刷新令牌：解析旧令牌、校验Redis白名单中的一致性、
// 生成新令牌并更新白名单。返回新的JWT令牌字符串。
func (s *Service) Refresh(ctx context.Context, oldToken string) (string, error) {
	// 解析旧令牌（不验证过期，仅提取 claims）
	claims, err := jwt.ParseTokenUnverified(oldToken)
	if err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", newError(errcode.TokenInvalid))
	}

	// Verify Redis whitelist — 验证旧令牌是否在白名单中
	redisKey := fmt.Sprintf("auth:token:%d", claims.UserID)
	storedToken, err := s.rdb.Get(ctx, redisKey).Result()
	if err != nil || storedToken != oldToken {
		return "", fmt.Errorf("auth.service.Refresh: %w", newError(errcode.Unauthorized))
	}

	// Generate new token — 生成新的JWT令牌
	newToken, err := jwt.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", err)
	}

	// Update Redis — 用新令牌覆盖Redis白名单
	if err := s.rdb.Set(ctx, redisKey, newToken, 7*24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("auth.service.Refresh: %w", err)
	}

	return newToken, nil
}

// Error 服务层错误类型，携带错误码以支持在HTTP层映射为合适的响应。
// Service error type for error code propagation
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

// presignAvatar 对用户头像路径生成预签名URL，空字符串直接返回。
func presignAvatar(avatar string) string {
	if avatar == "" {
		return ""
	}
	return storage.GetObjectURL(avatar)
}
