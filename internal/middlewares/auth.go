package middlewares

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"go-pg-demo/internal/pkgs"
)

// AuthMiddleware JWT验证中间件
func AuthMiddleware(config *pkgs.Config, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			pkgs.Error(c, 401, "Authorization header is required")
			return
		}

		// 检查Bearer前缀
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:] // "Bearer " 长度为7
		} else {
			pkgs.Error(c, 401, "Authorization header must start with 'Bearer '")
			return
		}

		// 解析和验证JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.JWT.Secret), nil
		})

		if err != nil {
			logger.Error("Failed to parse JWT token", zap.Error(err))
			pkgs.Error(c, 401, "Invalid token")
			return
		}

		// 检查token是否有效
		if !token.Valid {
			pkgs.Error(c, 401, "Invalid token")
			return
		}

		// 将用户信息存储到上下文中
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
		}

		c.Next()
	}
}
