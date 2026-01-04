package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("xmS10711")

type Claims struct {
	StudentID string `json:"student_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(studentID string, role string) (string, error) {
	// 过期时间
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	// 构建自定义负载
	claims := &Claims{
		StudentID: studentID, // 放入自定义业务信息
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间（必须设置，否则令牌永久有效）
			Issuer:    "LMH",                              // 签发者（可选，标识令牌的发行方）
		},
	}

	// 创建令牌对象，生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用jwtSecret对令牌进行签名，确保令牌不被篡改
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	var claims = new(Claims)
	// 解析令牌
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims, // 用于接收解析后的负载
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil // 返回用于验证签名的密钥
		},
	)

	// 错误处理
	if err != nil {
		return nil, err // 错误可能是：签名无效、令牌过期、格式错误等
	}

	// 4.3 验证令牌是否有效，并提取负载
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil // 令牌有效，返回负载信息
	}

	return nil, errors.New("invalid token")
}
