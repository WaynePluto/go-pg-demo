package auth

import (
	"database/sql"
	"go-pg-demo/pkgs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/samber/mo"
	"go.uber.org/zap"
)

type Repository struct {
	db     *sqlx.DB
	logger *zap.Logger
	config *pkgs.Config
}

func NewRepository(db *sqlx.DB, logger *zap.Logger, config *pkgs.Config) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (r *Repository) Login(c *gin.Context) func(*LoginReq) mo.Result[LoginRes] {
	return func(req *LoginReq) mo.Result[LoginRes] {
		// 查询用户（用户名唯一）
		var user UserEntity
		query := `SELECT id, username, password, phone, profile, created_at, updated_at FROM iacc_user WHERE username = $1`
		err := r.db.GetContext(c.Request.Context(), &user, query, req.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[LoginRes](pkgs.NewApiError(http.StatusUnauthorized, "用户名或密码错误"))
			}
			r.logger.Error("查询用户失败", zap.Error(err))
			return mo.Err[LoginRes](pkgs.NewApiError(http.StatusInternalServerError, "登录失败"))
		}

		// 简单密码校验（后续可引入加密）
		if user.Password != req.Password {
			return mo.Err[LoginRes](pkgs.NewApiError(http.StatusUnauthorized, "用户名或密码错误"))
		}

		// 生成访问令牌
		accessToken, err := r.generateToken(user.ID, r.config.JWT.AccessTokenExpire)
		if err != nil {
			r.logger.Error("生成访问令牌失败", zap.Error(err))
			return mo.Err[LoginRes](pkgs.NewApiError(http.StatusInternalServerError, "登录失败"))
		}

		// 生成刷新令牌
		refreshToken, err := r.generateToken(user.ID, r.config.JWT.RefreshTokenExpire)
		if err != nil {
			r.logger.Error("生成刷新令牌失败", zap.Error(err))
			return mo.Err[LoginRes](pkgs.NewApiError(http.StatusInternalServerError, "登录失败"))
		}

		return mo.Ok(LoginRes{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(r.config.JWT.AccessTokenExpire.Seconds()),
		})
	}
}

func (r *Repository) RefreshToken(c *gin.Context) func(*RefreshTokenReq) mo.Result[RefreshTokenRes] {
	return func(req *RefreshTokenReq) mo.Result[RefreshTokenRes] {
		// 解析刷新 token
		token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(r.config.JWT.Secret), nil
		})
		if err != nil || !token.Valid {
			return mo.Err[RefreshTokenRes](pkgs.NewApiError(http.StatusUnauthorized, "刷新令牌无效"))
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return mo.Err[RefreshTokenRes](pkgs.NewApiError(http.StatusUnauthorized, "刷新令牌无效"))
		}

		userID, _ := claims["user_id"].(string)
		if userID == "" {
			return mo.Err[RefreshTokenRes](pkgs.NewApiError(http.StatusUnauthorized, "刷新令牌无效"))
		}

		// 生成新的访问令牌
		accessToken, err := r.generateToken(userID, r.config.JWT.AccessTokenExpire)
		if err != nil {
			r.logger.Error("生成访问令牌失败", zap.Error(err))
			return mo.Err[RefreshTokenRes](pkgs.NewApiError(http.StatusInternalServerError, "刷新失败"))
		}

		// 生成新的刷新令牌
		newRefreshToken, err := r.generateToken(userID, r.config.JWT.RefreshTokenExpire)
		if err != nil {
			r.logger.Error("生成刷新令牌失败", zap.Error(err))
			return mo.Err[RefreshTokenRes](pkgs.NewApiError(http.StatusInternalServerError, "刷新失败"))
		}

		return mo.Ok(RefreshTokenRes{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int64(r.config.JWT.AccessTokenExpire.Seconds()),
		})
	}
}

func (r *Repository) UserDetail(c *gin.Context) func(string) mo.Result[UserDetailRes] {
	return func(userID string) mo.Result[UserDetailRes] {
		// 查询用户基本信息
		var user UserEntity
		queryUser := `SELECT id, username, phone, profile, created_at, updated_at FROM iacc_user WHERE id = $1`
		err := r.db.GetContext(c.Request.Context(), &user, queryUser, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[UserDetailRes](pkgs.NewApiError(http.StatusNotFound, "用户不存在"))
			}
			r.logger.Error("查询用户失败", zap.Error(err))
			return mo.Err[UserDetailRes](pkgs.NewApiError(http.StatusInternalServerError, "查询失败"))
		}

		// 查询角色列表
		var roles []RoleEntity
		queryRoles := `SELECT r.id, r.name, r.description FROM iacc_role r INNER JOIN iacc_user_role ur ON r.id = ur.role_id WHERE ur.user_id = $1`
		if err = r.db.SelectContext(c.Request.Context(), &roles, queryRoles, userID); err != nil {
			r.logger.Error("查询角色失败", zap.Error(err))
			return mo.Err[UserDetailRes](pkgs.NewApiError(http.StatusInternalServerError, "查询失败"))
		}

		// 查询权限列表
		var perms []PermissionEntity
		queryPerms := `SELECT p.id, p.name, p.type, p.metadata FROM iacc_permission p INNER JOIN iacc_role_permission rp ON p.id = rp.permission_id INNER JOIN iacc_user_role ur ON rp.role_id = ur.role_id WHERE ur.user_id = $1`
		if err = r.db.SelectContext(c.Request.Context(), &perms, queryPerms, userID); err != nil {
			r.logger.Error("查询权限失败", zap.Error(err))
			return mo.Err[UserDetailRes](pkgs.NewApiError(http.StatusInternalServerError, "查询失败"))
		}

		// 转换角色数据
		var userRoles []UserRoleRes
		for _, role := range roles {
			userRoles = append(userRoles, UserRoleRes{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
			})
		}

		// 去重权限（可能多角色拥有同一权限）
		permMap := map[string]UserPermRes{}
		for _, p := range perms {
			code := ""
			path := ""
			method := ""

			// 处理 metadata 字段
			if metadata, ok := p.Metadata.(map[string]interface{}); ok {
				if codeVal, ok := metadata["code"]; ok {
					if codeStr, ok := codeVal.(string); ok {
						code = codeStr
					}
				}
				if pathVal, ok := metadata["path"]; ok {
					if pathStr, ok := pathVal.(string); ok {
						path = pathStr
					}
				}
				if methodVal, ok := metadata["method"]; ok {
					if methodStr, ok := methodVal.(string); ok {
						method = methodStr
					}
				}
			}

			if _, ok := permMap[p.ID]; !ok {
				permMap[p.ID] = UserPermRes{
					ID:     p.ID,
					Name:   p.Name,
					Type:   p.Type,
					Code:   code,
					Path:   path,
					Method: method,
				}
			}
		}
		var permList []UserPermRes
		for _, v := range permMap {
			permList = append(permList, v)
		}

		// 处理手机号
		phone := ""
		if user.Phone != nil {
			phone = *user.Phone
		}

		return mo.Ok(UserDetailRes{
			ID:          user.ID,
			Username:    user.Username,
			Phone:       phone,
			Profile:     user.Profile,
			Roles:       userRoles,
			Permissions: permList,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		})
	}
}

// generateToken 生成 JWT 令牌
func (r *Repository) generateToken(userID string, expire time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expire).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(r.config.JWT.Secret))
}
