package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Options jwt.RegisteredClaims

type AuthClaims struct {
	Guard string `json:"guard"` // 授权守卫
	jwt.RegisteredClaims
}

func NewNumericDate(t time.Time) *jwt.NumericDate {
	return jwt.NewNumericDate(t)
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(guard string, secret string, ops *Options) string {

	claims := AuthClaims{
		Guard: guard,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"every"},
			ExpiresAt: ops.ExpiresAt,
			ID:        ops.ID,
			IssuedAt:  ops.IssuedAt,
			Issuer:    "jwt_yh",
			NotBefore: ops.NotBefore,
			Subject:   "YH",
		},
	}

	tokenString, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))

	return tokenString
}

// ParseToken 解析 JWT Token
func ParseToken(token string, secret string) (*AuthClaims, error) {

	data, err := jwt.ParseWithClaims(token, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if claims, ok := data.Claims.(*AuthClaims); ok && data.Valid {
		return claims, nil
	}

	return nil, err
}

// GetJwtToken 获取登录授权 token
func GetJwtToken(c *gin.Context) string {

	token := c.GetHeader("Authorization")
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))

	// Headers 中没有授权信息则读取 url 中的 token
	if token == "" {
		token = c.DefaultQuery("token", "")
	}

	if token == "" {
		token = c.DefaultPostForm("token", "")
	}

	return token
}
