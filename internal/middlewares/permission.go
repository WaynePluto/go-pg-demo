package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"go-pg-demo/pkgs"
)

// PermissionMiddleware 接口权限校验
// 规则（与实际实现保持同步）：
// 1. 白名单直接放行：swagger 文档、/v1/auth/login、/v1/auth/refresh-token；以及公共接口前缀 /v1/template*（无需登录 / 权限）。
// 2. 仅对 /v1/ 开头的接口做权限控制，其他路径直接放行。
// 3. 必须先通过 AuthMiddleware 将 user_id 写入 context；若不存在或为空 -> 返回 401 业务码 (HTTP 仍 200)。
// 4. 先查询权限元数据表(iacc_permission) 是否存在(method+path) 精确记录：
//   - 若不存在：说明该接口尚未纳入权限体系 -> 放行（便于灰度 / 临时接口 / 忘记录入时不中断功能）。
//   - 若存在：进入用户权限校验。
//
// 5. 通过用户角色关联 (iacc_user_role -> iacc_role_permission -> iacc_permission) 拉取用户拥有的全部权限(method+path) 列表。
// 6. 匹配策略：
//   - 先按 method 精确一致；
//   - path 完全相等直接通过；
//   - 若权限表 path 含 :param 形式（例如 /v1/user/:id），按段数一致且静态段逐一相等视为匹配（仅做“占位符”精确匹配，不做通配 / 前缀模糊）。
//
// 7. 未匹配 -> 返回 403 业务码；所有错误响应使用 HTTP 200 包装（统一前端处理）。
// 8. 未来可优化点：
//   - 缓存用户权限集合减少每次查询；
//   - 预编译路径模板提升匹配效率；
//   - 后台管理端自动同步/生成权限元数据，降低人工遗漏。
type PermissionMiddleware gin.HandlerFunc

func NewPermissionMiddleware(db *sqlx.DB, logger *zap.Logger) PermissionMiddleware {
	return func(c *gin.Context) {
		// 白名单（与鉴权一致，可根据需要补充），无需权限校验
		authWhitelist := []string{
			"/v1/auth/login",
			"/v1/auth/refresh-token",
		}
		if strings.Contains(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}
		for _, w := range authWhitelist {
			if c.Request.URL.Path == w {
				c.Next()
				return
			}
		}
		// 兼容 AuthMiddleware 中未解析 token 的公共接口（如 /v1/template/list）
		if strings.HasPrefix(c.Request.URL.Path, "/v1/template") {
			c.Next()
			return
		}

		// 未授权直接拒绝
		v, ok := c.Get("user_id")
		if !ok {
			pkgs.Error(c, http.StatusUnauthorized, "未授权")
			return
		}
		userID, _ := v.(string)
		if userID == "" {
			pkgs.Error(c, http.StatusUnauthorized, "未授权")
			return
		}

		method := c.Request.Method
		path := c.Request.URL.Path

		// 仅校验 /v1 开头接口
		if !strings.HasPrefix(path, "/v1/") {
			c.Next()
			return
		}

		// 先查权限表是否有该接口
		var permCount int
		metaQuery := `SELECT COUNT(1) FROM iacc_permission WHERE metadata->>'method' = $1 AND metadata->>'path' = $2`
		if err := db.GetContext(c.Request.Context(), &permCount, metaQuery, method, path); err != nil {
			logger.Error("查询权限表失败", zap.Error(err))
			pkgs.Error(c, http.StatusInternalServerError, "权限校验失败")
			return
		}
		// 权限表查不到该接口，直接放行
		if permCount == 0 {
			c.Next()
			return
		}

		// 权限表有记录，校验用户是否有权限
		var perms []struct {
			Method *string `db:"method"`
			Path   *string `db:"path"`
		}
		query := `SELECT (p.metadata->>'method') AS method, (p.metadata->>'path') AS path
			FROM iacc_permission p
			INNER JOIN iacc_role_permission rp ON p.id = rp.permission_id
			INNER JOIN iacc_user_role ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1`
		if err := db.SelectContext(c.Request.Context(), &perms, query, userID); err != nil {
			logger.Error("查询用户权限失败", zap.Error(err))
			pkgs.Error(c, http.StatusInternalServerError, "权限校验失败")
			return
		}

		allowed := false
		for _, p := range perms {
			if p.Method == nil || p.Path == nil {
				continue
			}
			if *p.Method != method {
				continue
			}
			permPath := *p.Path
			if permPath == path {
				allowed = true
				break
			}
			if strings.Contains(permPath, ":") {
				permSegs := strings.Split(strings.Trim(permPath, "/"), "/")
				pathSegs := strings.Split(strings.Trim(path, "/"), "/")
				if len(permSegs) == len(pathSegs) {
					matchAll := true
					for i := range permSegs {
						if strings.HasPrefix(permSegs[i], ":") {
							continue
						}
						if permSegs[i] != pathSegs[i] {
							matchAll = false
							break
						}
					}
					if matchAll {
						allowed = true
						break
					}
				}
			}
		}

		if !allowed {
			pkgs.Error(c, http.StatusForbidden, "无接口访问权限")
			return
		}

		c.Next()
	}
}
