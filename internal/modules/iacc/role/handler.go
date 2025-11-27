// Package role API.
//
// The API for managing role.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
package role

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

func NewRoleHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
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

// Create 创建角色
//
//	@Summary  创建角色
//	@Description  创建角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateReq true  "创建角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=CreateRes}  "创建成功，返回角色ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role [post]
func (h *Handler) Create(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[CreateReq](c),
		result.FlatMap(pkgs.ValidateV2[CreateReq](h.validator)),
		result.FlatMap(h.repository.Create(c)),
	).Match(
		pkgs.HandleSuccess[CreateRes](c),
		pkgs.HandleError[CreateRes](c),
	)
}

// BatchCreate 批量创建角色
//
//	@Summary  批量创建角色
//	@Description  批量创建角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    request body  BatchCreateReq  true  "批量创建角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=BatchCreateRes}  "创建成功，返回角色ID列表"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /role/batch-create [post]
func (h *Handler) BatchCreate(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[BatchCreateReq](c),
		result.FlatMap(pkgs.ValidateV2[BatchCreateReq](h.validator)),
		result.FlatMap(h.repository.BatchCreate(c)),
	).Match(
		pkgs.HandleSuccess[BatchCreateRes](c),
		pkgs.HandleError[BatchCreateRes](c),
	)
}

// GetByID 根据ID获取角色
//
//	@Summary  根据ID获取角色
//	@Description  根据ID获取角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "角色ID"
//	@Success  200 {object}  pkgs.Response{data=GetByIDRes}  "获取成功，返回角色信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "角色不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /role/{id} [get]
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

// UpdateByID 根据ID更新角色
//
//	@Summary  根据ID更新角色
//	@Description  根据ID更新角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "角色ID"
//	@Param    request body  UpdateByIDReq true  "更新角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=UpdateByIDRes}       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUriAndJSON[UpdateByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[UpdateByIDReq](h.validator)),
		result.FlatMap(h.repository.UpdateByID(c)),
	).Match(
		pkgs.HandleSuccess[UpdateByIDRes](c),
		pkgs.HandleError[UpdateByIDRes](c),
	)
}

// DeleteByID 根据ID删除角色
//
//	@Summary  根据ID删除角色
//	@Description  根据ID删除角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "角色ID"
//	@Success  200 {object}  pkgs.Response{data=DeleteByIDRes} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role/{id} [delete]
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

// BatchDelete 批量删除角色
//
//	@Summary  批量删除角色
//	@Description  批量删除角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    request body  DeleteRolesReq  true  "批量删除角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=BatchDeleteRes}       "删除成功，返回影响行数"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role/batch-delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[DeleteRolesReq](c),
		result.FlatMap(pkgs.ValidateV2[DeleteRolesReq](h.validator)),
		result.FlatMap(h.repository.BatchDelete(c)),
	).Match(
		pkgs.HandleSuccess[BatchDeleteRes](c),
		pkgs.HandleError[BatchDeleteRes](c),
	)
}

// QueryList 获取角色列表
//
//	@Summary  获取角色列表
//	@Description  获取角色列表
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    pageSize  query int   false "每页数量"  default(10)
//	@Param    name    query string  false "角色名称"
//	@Success  200     {object}  pkgs.Response{data=QueryListRes}  "获取成功，返回角色列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Router   /role/list [get]
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

// AssignPermission 为角色分配权限
//
//	@Summary  为角色分配权限
//	@Description  清空角色现有权限，并重新关联新的权限列表
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id      path      string  true  "角色ID"
//	@Param    request body      AssignPermissionsReq true  "分配权限的请求参数"
//	@Success  200     {object}  pkgs.Response{data=AssignPermissionsRes} "分配成功"
//	@Failure  400     {object}  pkgs.Response "请求参数错误"
//	@Failure  500     {object}  pkgs.Response "服务器内部错误"
//	@Router   /role/{id}/permission [post]
func (h *Handler) AssignPermission(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUriAndJSON[AssignPermissionsByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[AssignPermissionsByIDReq](h.validator)),
		result.FlatMap(h.repository.AssignPermissions(c)),
	).Match(
		pkgs.HandleSuccess[AssignPermissionsRes](c),
		pkgs.HandleError[AssignPermissionsRes](c),
	)
}

// GetPermissions 获取角色权限列表
//
//	@Summary  获取角色权限列表
//	@Description  获取指定角色的权限列表
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "角色ID"
//	@Success  200 {object}  pkgs.Response{data=GetRolePermissionsRes}  "获取成功，返回权限列表"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "角色不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /role/{id}/permission [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUri[GetRolePermissionsReq](c),
		result.FlatMap(pkgs.ValidateV2[GetRolePermissionsReq](h.validator)),
		result.FlatMap(h.repository.GetPermissions(c)),
	).Match(
		pkgs.HandleSuccess[GetRolePermissionsRes](c),
		pkgs.HandleError[GetRolePermissionsRes](c),
	)
}
