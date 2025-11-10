// Package user API.
//
// 用户管理API接口。
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
//
//	Paths:
//	  /user:
//	    post: Create
//	  /user/batch-create:
//	    post: BatchCreate
//	  /user/{id}:
//	    get: GetByID
//	    put: UpdateByID
//	    delete: DeleteByID
//	  /user/batch-delete:
//	    post: BatchDelete
//	  /user/list:
//	    get: QueryList
//	  /user/{id}/role:
//	    post: AssignRoles
package user

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

func NewUserHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
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

// Create 创建用户
//
//	@Summary      创建一个新的用户账户
//	@Description  通过提供用户名、手机号、密码等信息创建一个新的用户账户。成功后返回新创建用户的唯一标识符(UUID)。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        request  body      CreateReq              true  "创建用户所需的请求体参数"
//	@Success      200      {object}  pkgs.Response{data=CreateRes} "成功创建用户，返回用户ID"
//	@Failure      400      {object}  pkgs.Response              "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response              "服务器内部错误，无法创建用户"
//	@Router       /user [post]
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

// BatchCreate 批量创建用户
//
//	@Summary  批量创建用户
//	@Description  批量创建用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    request body  BatchCreateReq  true  "批量创建用户请求参数"
//	@Success  200   {object}  pkgs.Response{data=BatchCreateRes}  "创建成功，返回用户ID列表"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /user/batch-create [post]
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

// GetByID 根据ID获取用户
//
//	@Summary      根据用户ID获取用户详情
//	@Description  通过指定的用户唯一标识符(UUID)来检索特定用户的详细信息。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string                     true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Success      200  {object}  pkgs.Response{data=GetByIDRes} "成功获取用户信息"
//	@Failure      400  {object}  pkgs.Response              "提供的用户ID格式无效"
//	@Failure      404  {object}  pkgs.Response              "未找到指定ID的用户"
//	@Failure      500  {object}  pkgs.Response              "服务器内部错误，无法获取用户信息"
//	@Router       /user/{id} [get]
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

// UpdateByID 根据ID更新用户
//
//	@Summary      根据用户ID更新用户信息
//	@Description  通过指定的用户唯一标识符(UUID)来更新特定用户的密码和个人信息。只会更新请求中包含的字段。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id       path      string          true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Param        request  body      UpdateByIDReq   true  "需要更新的用户信息"
//	@Success      200      {object}  pkgs.Response{data=UpdateByIDRes}   "成功更新用户信息"
//	@Failure      400      {object}  pkgs.Response   "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response   "服务器内部错误，无法更新用户信息"
//	@Router       /user/{id} [put]
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

// DeleteByID 根据ID删除用户
//
//	@Summary      根据用户ID删除用户
//	@Description  通过指定的用户唯一标识符(UUID)来删除特定用户。这是一个永久性操作，请谨慎使用。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string                    true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Success      200  {object}  pkgs.Response{data=DeleteByIDRes} "成功删除用户，返回受影响的行数"
//	@Failure      400  {object}  pkgs.Response             "提供的用户ID格式无效"
//	@Failure      500  {object}  pkgs.Response             "服务器内部错误，无法删除用户"
//	@Router       /user/{id} [delete]
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

// BatchDelete 批量删除用户
//
//	@Summary  批量删除用户
//	@Description  批量删除用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    request body  DeleteUsersReq  true  "批量删除用户请求参数"
//	@Success  200   {object}  pkgs.Response{data=BatchDeleteRes}       "删除成功，返回影响行数"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user/batch-delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[DeleteUsersReq](c),
		result.FlatMap(pkgs.ValidateV2[DeleteUsersReq](h.validator)),
		result.FlatMap(h.repository.BatchDelete(c)),
	).Match(
		pkgs.HandleSuccess[BatchDeleteRes](c),
		pkgs.HandleError[BatchDeleteRes](c),
	)
}

// QueryList 获取用户列表
//
//	@Summary      获取用户列表（支持分页和筛选）
//	@Description  获取系统中的用户列表，支持按手机号和用户名模糊搜索，并提供分页功能。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        page      query     int                        false  "页码，从1开始计算"  minimum(1)  default(1)
//	@Param        pageSize  query     int                        false  "每页条目数"        minimum(1)  maximum(100)  default(10)
//	@Param        phone     query     string                     false  "手机号模糊搜索关键字"
//	@Param        username  query     string                     false  "用户名模糊搜索关键字"
//	@Success      200       {object}  pkgs.Response{data=QueryListRes}  "成功获取用户列表"
//	@Failure      400       {object}  pkgs.Response                  "请求参数验证失败或格式不正确"
//	@Failure      500       {object}  pkgs.Response                  "服务器内部错误，无法获取用户列表"
//	@Router       /user/list [get]
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

// AssignRole 为用户分配角色
//
//	@Summary      为用户分配角色
//	@Description  为指定用户分配一个或多个角色。该操作会完全替换用户当前的所有角色关系。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id       path      string                 true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Param        request  body      AssignRolesReq   true  "要分配给用户的角色ID列表"
//	@Success      200      {object}  pkgs.Response{data=AssignRolesRes}          "成功为用户分配角色"
//	@Failure      400      {object}  pkgs.Response          "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response          "服务器内部错误，无法为用户分配角色"
//	@Router       /user/{id}/role [post]
func (h *Handler) AssignRole(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUriAndJSON[AssignRolesReq](c),
		result.FlatMap(pkgs.ValidateV2[AssignRolesReq](h.validator)),
		result.FlatMap(h.repository.AssignRoles(c)),
	).Match(
		pkgs.HandleSuccess[AssignRolesRes](c),
		pkgs.HandleError[AssignRolesRes](c),
	)
}
