package auth

import (
	"net/http"
	"time"

	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Handler struct {
	db        *sqlx.DB
	logger    *zap.Logger
	validator *pkgs.RequestValidator
	config    *pkgs.Config
}

func NewAuthHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator, config *pkgs.Config) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
		config:    config,
	}
}

// Login 用户登录
//
//	@Summary  用户登录
//	@Description  用户登录
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Param    request body  LoginRequest  true  "用户登录请求参数"
//	@Success  200   {object}  pkgs.Response{data=LoginResponse} "登录成功，返回访问令牌等信息"
//	@Failure  400   {object}  pkgs.Response           "请求参数错误"
//	@Failure  401   {object}  pkgs.Response           "用户名或密码错误"
//	@Failure  500   {object}  pkgs.Response           "服务器内部错误"
//	@Router   /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	// 绑定请求参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 验证用户名和密码
	var user UserDTO
	query := `SELECT id, username, phone FROM iacc_user WHERE username = $1 AND password = $2`
	err := h.db.GetContext(c.Request.Context(), &user, query, req.Username, req.Password)
	if err != nil {
		pkgs.Error(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// 生成访问令牌和刷新令牌
	accessToken := "fake_access_token_" + uuid.Must(uuid.NewV7()).String()
	refreshToken := "fake_refresh_token_" + uuid.Must(uuid.NewV7()).String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 返回结果
	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	}
	pkgs.Success(c, response)
}

// RefreshToken 刷新token
//
//	@Summary  刷新token
//	@Description  刷新访问令牌
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Param    request body  RefreshTokenRequest true  "刷新令牌请求参数"
//	@Success  200   {object}  pkgs.Response{data=RefreshTokenResponse}  "刷新成功，返回新的访问令牌等信息"
//	@Failure  400   {object}  pkgs.Response               "请求参数错误"
//	@Failure  401   {object}  pkgs.Response               "无效的刷新令牌"
//	@Failure  500   {object}  pkgs.Response               "服务器内部错误"
//	@Router   /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	// 绑定请求参数
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 验证刷新令牌（简化实现）
	if req.RefreshToken == "" || len(req.RefreshToken) < 10 {
		pkgs.Error(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// 生成新的访问令牌和刷新令牌
	accessToken := "fake_access_token_" + uuid.Must(uuid.NewV7()).String()
	refreshToken := "fake_refresh_token_" + uuid.Must(uuid.NewV7()).String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 返回结果
	response := RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	pkgs.Success(c, response)
}

// GetUserInfo 获取当前用户信息
//
//	@Summary  获取当前用户信息
//	@Description  获取当前登录用户的信息
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Success  200 {object}  pkgs.Response{data=UserInfoResponse}  "获取成功，返回用户信息"
//	@Failure  401 {object}  pkgs.Response             "未授权"
//	@Failure  500 {object}  pkgs.Response             "服务器内部错误"
//	@Router   /auth/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// 从上下文中获取用户ID（简化实现，实际应从JWT中获取）
	// 这里使用一个假的用户ID
	userID := "018e4097-4f9d-7b4f-b0a4-7c9d7e0a0a0a"

	// 查询用户信息
	var user UserInfoResponse
	query := `SELECT id, created_at, updated_at, username, phone, profile FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &user, query, userID)
	if err != nil {
		h.logger.Error("Failed to get user info", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// 查询用户角色
	var roles []RoleDTO
	query = `SELECT r.id, r.name, r.description 
           FROM iacc_role r 
           JOIN iacc_user_role ur ON r.id = ur.role_id 
           WHERE ur.user_id = $1`
	err = h.db.SelectContext(c.Request.Context(), &roles, query, userID)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user roles")
		return
	}
	user.Roles = roles

	// 返回结果
	pkgs.Success(c, user)
}

// AssignRole 分配角色
//
//	@Summary  分配角色
//	@Description  为用户分配角色
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Param    request body  AssignRoleRequest true  "分配角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "分配成功，返回用户角色关联ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /auth/assign-role [post]
func (h *Handler) AssignRole(c *gin.Context) {
	// 绑定请求参数
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 验证用户ID和角色ID是否为有效的UUID
	if _, err := uuid.Parse(req.UserID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if _, err := uuid.Parse(req.RoleID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 检查用户是否存在
	var userCount int
	userQuery := `SELECT COUNT(*) FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &userCount, userQuery, req.UserID)
	if err != nil {
		h.logger.Error("Failed to check user existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign role")
		return
	}
	if userCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 检查角色是否存在
	var roleCount int
	roleQuery := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
	err = h.db.GetContext(c.Request.Context(), &roleCount, roleQuery, req.RoleID)
	if err != nil {
		h.logger.Error("Failed to check role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign role")
		return
	}
	if roleCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 检查用户角色关联是否已存在
	var userRoleCount int
	userRoleQuery := `SELECT COUNT(*) FROM iacc_user_role WHERE user_id = $1 AND role_id = $2`
	err = h.db.GetContext(c.Request.Context(), &userRoleCount, userRoleQuery, req.UserID, req.RoleID)
	if err != nil {
		h.logger.Error("Failed to check user role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign role")
		return
	}
	if userRoleCount > 0 {
		pkgs.Error(c, http.StatusBadRequest, "User role already exists")
		return
	}

	// 创建用户角色关联实体
	id := uuid.Must(uuid.NewV7()).String()
	now := time.Now()
	userRole := map[string]interface{}{
		"id":         id,
		"created_at": now,
		"updated_at": now,
		"user_id":    req.UserID,
		"role_id":    req.RoleID,
	}

	// 数据库操作
	query := `INSERT INTO iacc_user_role (id, created_at, updated_at, user_id, role_id) 
            VALUES (:id, :created_at, :updated_at, :user_id, :role_id)`
	_, err = h.db.NamedExecContext(c.Request.Context(), query, userRole)
	if err != nil {
		h.logger.Error("Failed to assign role to user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign role")
		return
	}

	// 返回结果
	pkgs.Success(c, id)
}
