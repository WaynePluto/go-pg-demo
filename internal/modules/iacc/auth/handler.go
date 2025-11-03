package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"go-pg-demo/pkgs"
)

type Handler struct {
	db        *sqlx.DB
	logger    *zap.Logger
	validator *pkgs.RequestValidator
	config    *pkgs.Config
}

func NewAuthHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator, config *pkgs.Config) *Handler {
	return &Handler{db: db, logger: logger, validator: validator, config: config}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录，获取访问令牌和刷新令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginReq true "登录请求参数"
// @Success 200 {object} pkgs.Response{data=LoginRes} "登录成功"
// @Failure 400 {object} pkgs.Response "请求参数错误"
// @Failure 401 {object} pkgs.Response "用户名或密码错误"
// @Failure 500 {object} pkgs.Response "服务器内部错误"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 查询用户（用户名唯一）
	var user struct {
		ID        string         `db:"id"`
		Username  string         `db:"username"`
		Password  string         `db:"password"`
		Phone     sql.NullString `db:"phone"`
		Profile   interface{}    `db:"profile"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
	}
	query := `SELECT id, username, password, phone, profile, created_at, updated_at FROM iacc_user WHERE username = $1`
	err := h.db.GetContext(c.Request.Context(), &user, query, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		h.logger.Error("查询用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}
	// 简单密码校验（后续可引入加密）
	if user.Password != req.Password {
		pkgs.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	accessToken, err := h.generateToken(user.ID, h.config.JWT.AccessTokenExpire)
	if err != nil {
		h.logger.Error("生成访问令牌失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}
	refreshToken, err := h.generateToken(user.ID, h.config.JWT.RefreshTokenExpire)
	if err != nil {
		h.logger.Error("生成刷新令牌失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}
	pkgs.Success(c, LoginRes{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(h.config.JWT.AccessTokenExpire.Seconds())})
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 通过刷新令牌获取新的访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenReq true "刷新令牌请求参数"
// @Success 200 {object} pkgs.Response{data=RefreshTokenRes} "刷新成功"
// @Failure 400 {object} pkgs.Response "请求参数错误"
// @Failure 401 {object} pkgs.Response "刷新令牌无效"
// @Failure 500 {object} pkgs.Response "服务器内部错误"
// @Router /auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 解析刷新 token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(h.config.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		pkgs.Error(c, http.StatusUnauthorized, "刷新令牌无效")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		pkgs.Error(c, http.StatusUnauthorized, "刷新令牌无效")
		return
	}
	userID, _ := claims["user_id"].(string)
	if userID == "" {
		pkgs.Error(c, http.StatusUnauthorized, "刷新令牌无效")
		return
	}
	accessToken, err := h.generateToken(userID, h.config.JWT.AccessTokenExpire)
	if err != nil {
		h.logger.Error("生成访问令牌失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "刷新失败")
		return
	}
	newRefreshToken, err := h.generateToken(userID, h.config.JWT.RefreshTokenExpire)
	if err != nil {
		h.logger.Error("生成刷新令牌失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "刷新失败")
		return
	}
	pkgs.Success(c, RefreshTokenRes{AccessToken: accessToken, RefreshToken: newRefreshToken, ExpiresIn: int64(h.config.JWT.AccessTokenExpire.Seconds())})
}

// UserDetail 获取当前用户详情
// @Summary 获取当前用户详情
// @Description 返回用户基本信息、角色列表、权限列表
// @Tags auth
// @Produce json
// @Success 200 {object} pkgs.Response{data=UserDetailRes} "成功"
// @Failure 401 {object} pkgs.Response "未授权"
// @Failure 404 {object} pkgs.Response "用户不存在"
// @Failure 500 {object} pkgs.Response "服务器内部错误"
// @Router /auth/user-detail [get]
func (h *Handler) UserDetail(c *gin.Context) {
	v, exists := c.Get("user_id")
	if !exists {
		pkgs.Error(c, http.StatusUnauthorized, "未授权")
		return
	}
	userID, _ := v.(string)
	if userID == "" {
		pkgs.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 查询用户基本信息
	var user struct {
		ID        string         `db:"id"`
		Username  string         `db:"username"`
		Phone     sql.NullString `db:"phone"`
		Profile   interface{}    `db:"profile"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
	}
	queryUser := `SELECT id, username, phone, profile, created_at, updated_at FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &user, queryUser, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		h.logger.Error("查询用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 查询角色列表
	var roles []UserRoleRes
	queryRoles := `SELECT r.id, r.name, r.description FROM iacc_role r INNER JOIN iacc_user_role ur ON r.id = ur.role_id WHERE ur.user_id = $1`
	if err = h.db.SelectContext(c.Request.Context(), &roles, queryRoles, userID); err != nil {
		h.logger.Error("查询角色失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 查询权限列表
	var perms []struct {
		ID       string `db:"id"`
		Name     string `db:"name"`
		Type     string `db:"type"`
		Metadata struct {
			Path   *string `json:"path"`
			Method *string `json:"method"`
			Code   *string `json:"code"`
		} `db:"metadata"`
	}
	queryPerms := `SELECT p.id, p.name, p.type, p.metadata FROM iacc_permission p INNER JOIN iacc_role_permission rp ON p.id = rp.permission_id INNER JOIN iacc_user_role ur ON rp.role_id = ur.role_id WHERE ur.user_id = $1`
	if err = h.db.SelectContext(c.Request.Context(), &perms, queryPerms, userID); err != nil {
		h.logger.Error("查询权限失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	// 去重权限（可能多角色拥有同一权限）
	permMap := map[string]UserPermRes{}
	for _, p := range perms {
		code := ""
		path := ""
		method := ""
		if p.Metadata.Code != nil {
			code = *p.Metadata.Code
		}
		if p.Metadata.Path != nil {
			path = *p.Metadata.Path
		}
		if p.Metadata.Method != nil {
			method = *p.Metadata.Method
		}
		if _, ok := permMap[p.ID]; !ok {
			permMap[p.ID] = UserPermRes{ID: p.ID, Name: p.Name, Type: p.Type, Code: code, Path: path, Method: method}
		}
	}
	var permList []UserPermRes
	for _, v := range permMap {
		permList = append(permList, v)
	}

	res := UserDetailRes{
		ID:          user.ID,
		Username:    user.Username,
		Phone:       user.Phone.String,
		Profile:     user.Profile,
		Roles:       roles,
		Permissions: permList,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}
	pkgs.Success(c, res)
}

// GetMe 兼容路由接口，内部复用 UserDetail 逻辑
// @Summary 当前用户详情
// @Description 返回当前登录用户的信息、角色和权限列表
// @Tags auth
// @Produce json
// @Success 200 {object} pkgs.Response{data=UserDetailRes} "成功"
// @Router /auth/me [get]

// generateToken 生成 JWT 令牌
func (h *Handler) generateToken(userID string, expire time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expire).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWT.Secret))
}
