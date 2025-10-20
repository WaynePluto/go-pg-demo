package middlewares

import (
	"net/http"

	"go-pg-demo/internal/modules/iacc/service"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PermissionMiddleware func(...string) gin.HandlerFunc

// 检查用户是否具有所需权限的中间件
func NewPermissionMiddleware(permissionService *service.PermissionService, logger *zap.Logger) PermissionMiddleware {
	return func(requiredPermissions ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			// 从上下文中获取用户ID
			userID, exists := c.Get("user_id")
			if !exists {
				pkgs.Error(c, http.StatusUnauthorized, "用户未认证")
				return
			}

			userIDStr, ok := userID.(string)
			if !ok {
				pkgs.Error(c, http.StatusInternalServerError, "用户ID格式无效")
				return
			}

			// 计算用户的有效权限
			effectivePermissions, err := permissionService.CalculateEffectivePermissionsForUser(userIDStr)
			if err != nil {
				logger.Error("权限计算出错", zap.String("user_id", userIDStr), zap.Error(err))
				pkgs.Error(c, http.StatusInternalServerError, "权限校验失败")
				return
			}

			// 检查用户是否拥有所有必需的权限
			userPermissions := make(map[string]struct{}, len(effectivePermissions))
			for _, p := range effectivePermissions {
				userPermissions[p] = struct{}{}
			}

			// 处理权限为空的情况
			if len(userPermissions) == 0 {
				pkgs.Error(c, http.StatusForbidden, "您没有权限访问该资源")
				return
			}

			for _, required := range requiredPermissions {
				if _, ok := userPermissions[required]; !ok {
					pkgs.Error(c, http.StatusForbidden, "您没有权限访问该资源")
					return
				}
			}

			c.Next()
		}
	}

}
