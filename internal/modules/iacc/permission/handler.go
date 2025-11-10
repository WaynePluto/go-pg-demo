// Package permission API.
//
// The API for managing permission.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
//
//	Security:
//	- JWT:
//
//	SecurityDefinitions:
//	 JWT:
//	   type: apiKey
//	   name: Authorization
//	   in: header
//	   description: JWT token for authentication
package permission

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

func NewPermissionHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
		repository: &Repository{
			db:     db,
			logger: logger,
		},
	}
}

// Create 创建权限
//
//	@Summary  创建权限
//	@Description  创建权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreatePermissionReq true  "创建权限请求参数"
//	@Success  200   {object}  pkgs.Response{data=CreatePermissionRes}  "创建成功，返回权限ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission [post]
func (h *Handler) Create(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[CreatePermissionReq](c),
		result.FlatMap(pkgs.ValidateV2[CreatePermissionReq](h.validator)),
		result.FlatMap(h.repository.Create(c)),
	).Match(
		pkgs.HandleSuccess[CreatePermissionRes](c),
		pkgs.HandleError[CreatePermissionRes](c),
	)
}

// GetByID 根据ID获取权限
//
//	@Summary  根据ID获取权限
//	@Description  根据ID获取权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "权限ID"
//	@Success  200 {object}  pkgs.Response{data=GetByIDRes}  "获取成功，返回权限信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "权限不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUri[GetByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[GetByIDReq](h.validator)),
		result.FlatMap(h.repository.GetByID(c)),
	).Match(
		pkgs.HandleSuccess[GetByIDRes](c),
		pkgs.HandleError[GetByIDRes](c),
	)
}

// UpdateByID 根据ID更新权限
//
//	@Summary  根据ID更新权限
//	@Description  根据ID更新权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "权限ID"
//	@Param    request body  UpdatePermissionReq true  "更新权限请求参数"
//	@Success  200   {object}  pkgs.Response{data=UpdatePermissionRes}       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUriAndJSON[UpdatePermissionReq](c),
		result.FlatMap(pkgs.ValidateV2[UpdatePermissionReq](h.validator)),
		result.FlatMap(h.repository.UpdateByID(c)),
	).Match(
		pkgs.HandleSuccess[UpdatePermissionRes](c),
		pkgs.HandleError[UpdatePermissionRes](c),
	)
}

// DeleteByID 根据ID删除权限
//
//	@Summary  根据ID删除权限
//	@Description  根据ID删除权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "权限ID"
//	@Success  200 {object}  pkgs.Response{data=DeleteByIDRes} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUri[DeleteByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[DeleteByIDReq](h.validator)),
		result.FlatMap(h.repository.DeleteByID(c)),
	).Match(
		pkgs.HandleSuccess[DeleteByIDRes](c),
		pkgs.HandleError[DeleteByIDRes](c),
	)
}

// QueryList 获取权限列表
//
//	@Summary  获取权限列表
//	@Description  获取权限列表
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    pageSize  query int   false "每页数量"  default(10)
//	@Param    name    query string  false "权限名称"
//	@Param    type    query string  false "权限类型"
//	@Success  200     {object}  pkgs.Response{data=QueryListRes}  "获取成功，返回权限列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Security JWT
//	@Router   /permission/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	result.Pipe2(
		pkgs.BindQuery[QueryListReq](c),
		result.FlatMap(pkgs.ValidateV2[QueryListReq](h.validator)),
		result.FlatMap(h.repository.QueryList(c)),
	).Match(
		pkgs.HandleSuccess[QueryListRes](c),
		pkgs.HandleError[QueryListRes](c),
	)
}