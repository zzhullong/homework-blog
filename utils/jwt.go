package utils

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const defaultJWTSecret = "your-secret-key"

var (
	jwtSecretOnce sync.Once
	jwtSecret     []byte
)

// getJWTSecret 延迟读取 JWT 密钥。
// 包初始化早于 main 中的 godotenv.Load()，因此不能在包级变量中直接读环境变量。
func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = defaultJWTSecret
			logrus.Warn("JWT_SECRET is not set, using insecure default secret")
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 只接受 HMAC 签名，防止 alg 混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
