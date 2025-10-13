package middlewares

import (
	"fmt"
	"net/http"
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code": 401,
				"msg":  "Authorization header is required",
				"data": nil,
			})
			return
		}

		// 检查Bearer前缀
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:] // "Bearer " 长度为7
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code": 401,
				"msg":  "Authorization header must start with 'Bearer '",
				"data": nil,
			})
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code": 401,
				"msg":  "Invalid token",
				"data": nil,
			})
			return
		}

		// 检查token是否有效
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code": 401,
				"msg":  "Invalid token",
				"data": nil,
			})
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
