// Package auth API.
//
// The API for managing authentication.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
package auth

import (
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/samber/mo/result"
	"go.uber.org/zap"
)

type Handler struct {
	db         *sqlx.DB
	logger     *zap.Logger
	validator  *pkgs.RequestValidator
	repository *Repository
}

func NewAuthHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator, config *pkgs.Config) *Handler {
	return &Handler{
		db:         db,
		logger:     logger,
		validator:  validator,
		repository: NewRepository(db, logger, config),
	}
}

// Login 用户登录
//
//	@Summary  用户登录
//	@Description  用户登录，获取访问令牌和刷新令牌
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Param    request body  LoginReq true  "登录请求参数"
//	@Success  200   {object}  pkgs.Response{data=LoginRes}  "登录成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  401   {object}  pkgs.Response       "用户名或密码错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[LoginReq](c),
		result.FlatMap(pkgs.ValidateV2[LoginReq](h.validator)),
		result.FlatMap(h.repository.Login(c)),
	).Match(
		pkgs.HandleSuccess[LoginRes](c),
		pkgs.HandleError[LoginRes](c),
	)
}

// RefreshToken 刷新访问令牌
//
//	@Summary  刷新访问令牌
//	@Description  通过刷新令牌获取新的访问令牌
//	@Tags   auth
//	@Accept   json
//	@Produce  json
//	@Param    request body  RefreshTokenReq true  "刷新令牌请求参数"
//	@Success  200   {object}  pkgs.Response{data=RefreshTokenRes}  "刷新成功"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  401   {object}  pkgs.Response         "刷新令牌无效"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[RefreshTokenReq](c),
		result.FlatMap(pkgs.ValidateV2[RefreshTokenReq](h.validator)),
		result.FlatMap(h.repository.RefreshToken(c)),
	).Match(
		pkgs.HandleSuccess[RefreshTokenRes](c),
		pkgs.HandleError[RefreshTokenRes](c),
	)
}

// UserDetail 获取当前用户详情
//
//	@Summary  获取当前用户详情
//	@Description  返回用户基本信息、角色列表、权限列表
//	@Tags   auth
//	@Produce  json
//	@Success  200 {object}  pkgs.Response{data=UserDetailRes}  "成功"
//	@Failure  401 {object}  pkgs.Response           "未授权"
//	@Failure  404 {object}  pkgs.Response           "用户不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /auth/user-detail [get]
func (h *Handler) UserDetail(c *gin.Context) {
	v, exists := c.Get("user_id")
	if !exists {
		pkgs.Error(c, 401, "未授权")
		return
	}
	userID, _ := v.(string)
	if userID == "" {
		pkgs.Error(c, 401, "未授权")
		return
	}

	h.repository.UserDetail(c)(userID).Match(
		pkgs.HandleSuccess[UserDetailRes](c),
		pkgs.HandleError[UserDetailRes](c),
	)
}

// GetMe 兼容路由接口，内部复用 UserDetail 逻辑
//
//	@Summary  当前用户详情
//	@Description  返回当前登录用户的信息、角色和权限列表
//	@Tags   auth
//	@Produce  json
//	@Success  200 {object}  pkgs.Response{data=UserDetailRes}  "成功"
//	@Router   /auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	h.UserDetail(c)
}
