package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     uint8  `json:"role"`
	jwtlib.RegisteredClaims
}

var jwtSecret []byte

func Init(secret string) {
	jwtSecret = []byte(secret)
}

func GenerateToken(userID uint, username string, role uint8) (string, error) {
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

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt.ParseToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt.ParseToken: invalid token")
	}

	return claims, nil
}

// ParseTokenUnverified parses token without validating expiry (for refresh flow).
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
