package middlewares

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"go-pg-demo/pkgs"
)

// JWT验证中间件
type AuthMiddleware gin.HandlerFunc

func NewAuthMiddleware(config *pkgs.Config, logger *zap.Logger) AuthMiddleware {
	return func(c *gin.Context) {
		// 白名单
		if strings.Contains(c.Request.URL.Path, "/swagger") ||
			strings.Contains(c.Request.URL.Path, "/v1/template") ||
			strings.Contains(c.Request.URL.Path, "/v1/auth/login") ||
			strings.Contains(c.Request.URL.Path, "/v1/auth/refresh-token") {
			c.Next()
			return
		}

		// 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			pkgs.Error(c, 401, "请求头缺少 Authorization 字段")
			return
		}

		// 检查Bearer前缀
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:] // "Bearer " 长度为7
		} else {
			pkgs.Error(c, 401, "Authorization 字段必须以 'Bearer ' 开头")
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
			pkgs.Error(c, 401, "无效的令牌")
			return
		}

		// 检查token是否有效
		if !token.Valid {
			pkgs.Error(c, 401, "无效的令牌")
			return
		}

		// 将用户信息存储到上下文中
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			pkgs.Error(c, 500, "token类型出错")
		}
		c.Set("user_id", claims["user_id"])

		c.Next()
	}
}
