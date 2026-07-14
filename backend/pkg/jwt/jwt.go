// Package jwt 提供 JWT Token 的生成、解析与验证功能。
// 使用 HS256 签名算法，Token 有效期 7 天。
// 支持普通解析（验证签名+有效期）和未验证解析（用于刷新流程）。
package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 载荷，包含用户基本信息和标准注册声明。
type Claims struct {
	UserID   uint   `json:"user_id"`  // 用户 ID
	Username string `json:"username"` // 用户名
	Role     uint8  `json:"role"`     // 角色：0=普通用户，1=管理员
	jwtlib.RegisteredClaims
}

// jwtSecret 全局 JWT 签名密钥，由 Init 设置。
var jwtSecret []byte

// Init 初始化 JWT 签名密钥。
func Init(secret string) {
	jwtSecret = []byte(secret)
}

// GenerateToken 为用户生成签名的 JWT Token 字符串。
// Token 包含用户 ID、用户名、角色，有效期 7 天。
func GenerateToken(userID uint, username string, role uint8) (string, error) {
	// 构造 Claims，设置签发者、过期时间和签发时间
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "B-B",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}

	// 使用 HS256 算法签名
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并验证 JWT Token，返回 Claims。
// 会验证签名算法和 Token 有效性（含过期时间）。
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		// 验证签名算法必须为 HMAC
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt.ParseToken: %w", err)
	}

	// 类型断言并检查 Token 有效性
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt.ParseToken: invalid token")
	}

	return claims, nil
}

// ParseTokenUnverified 不验证签名和有效期，仅解析 Token 载荷。
// 用于刷新 Token 等场景，调用方应自行验证。
func ParseTokenUnverified(tokenString string) (*Claims, error) {
	token, _, err := new(jwtlib.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("jwt.ParseTokenUnverified: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("jwt.ParseTokenUnverified: invalid claims type")
	}

	return claims, nil
}
