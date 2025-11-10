// Package template API.
//
// The API for managing template.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
package template

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

func NewTemplateHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
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

// Create 创建模板
//
//	@Summary  创建模板
//	@Description  创建模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateOneReq true  "创建模板请求参数"
//	@Success  200   {object}  pkgs.Response{data=CreateOneRes}  "创建成功，返回模板ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template [post]
func (h *Handler) Create(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[CreateOneReq](c),
		result.FlatMap(pkgs.ValidateV2[CreateOneReq](h.validator)),
		result.FlatMap(h.repository.Create(c)),
	).Match(
		func(id CreateOneRes) (CreateOneRes, error) {
			pkgs.Success(c, id)
			return id, nil
		},
		func(err error) (CreateOneRes, error) {
			pkgs.HandleError(c, err)
			return "", err
		},
	)
}

// BatchCreate 批量创建模板
//
//	@Summary  批量创建模板
//	@Description  批量创建模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  BatchCreateReq  true  "批量创建模板请求参数"
//	@Success  200   {object}  pkgs.Response{data=BatchCreateRes}  "创建成功，返回模板ID列表"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /template/batch-create [post]
func (h *Handler) BatchCreate(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[BatchCreateReq](c),
		result.FlatMap(pkgs.ValidateV2[BatchCreateReq](h.validator)),
		result.FlatMap(h.repository.BatchCreate(c)),
	).Match(
		func(ids BatchCreateRes) (BatchCreateRes, error) {
			pkgs.Success(c, ids)
			return ids, nil
		},
		func(err error) (BatchCreateRes, error) {
			pkgs.HandleError(c, err)
			return nil, err
		},
	)
}

// GetByID 根据ID获取模板
//
//	@Summary  根据ID获取模板
//	@Description  根据ID获取模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "模板ID"
//	@Success  200 {object}  pkgs.Response{data=TemplateRes}  "获取成功，返回模板信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "模板不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /template/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUri[GetByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[GetByIDReq](h.validator)),
		result.FlatMap(h.repository.GetByID(c)),
	).Match(
		func(res GetByIDRes) (GetByIDRes, error) {
			pkgs.Success(c, res)
			return res, nil
		},
		func(err error) (GetByIDRes, error) {
			pkgs.HandleError(c, err)
			return GetByIDRes{}, err
		},
	)
}

// UpdateByID 根据ID更新模板
//
//	@Summary  根据ID更新模板
//	@Description  根据ID更新模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "模板ID"
//	@Param    request body  UpdateTemplateReq true  "更新模板请求参数"
//	@Success  200   {object}  pkgs.Response       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUriAndJSON[UpdateOneReq](c),
		result.FlatMap(pkgs.ValidateV2[UpdateOneReq](h.validator)),
		result.FlatMap(h.repository.UpdateByID(c)),
	).
		Match(
			func(res UpdateOneRes) (UpdateOneRes, error) {
				pkgs.Success(c, res)
				return res, nil
			},
			func(err error) (UpdateOneRes, error) {
				pkgs.HandleError(c, err)
				return UpdateOneRes{}, err
			},
		)
}

// DeleteByID 根据ID删除模板
//
//	@Summary  根据ID删除模板
//	@Description  根据ID删除模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "模板ID"
//	@Success  200 {object}  pkgs.Response{data=int64} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	result.Pipe2(
		pkgs.BindUri[DeleteByIDReq](c),
		result.FlatMap(pkgs.ValidateV2[DeleteByIDReq](h.validator)),
		result.FlatMap(h.repository.DeleteByID(c)),
	).Match(
		func(res DeleteByIDRes) (DeleteByIDRes, error) {
			pkgs.Success(c, res)
			return res, nil
		},
		func(err error) (DeleteByIDRes, error) {
			pkgs.HandleError(c, err)
			return 0, err
		},
	)
}

// BatchDelete 批量删除模板
//
//	@Summary  批量删除模板
//	@Description  批量删除模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  DeleteTemplatesReq  true  "批量删除模板请求参数"
//	@Success  200   {object}  pkgs.Response{data=int64}       "删除成功，返回影响行数"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template/batch-delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	result.Pipe2(
		pkgs.BindJSON[DeleteTemplatesReq](c),
		result.FlatMap(pkgs.ValidateV2[DeleteTemplatesReq](h.validator)),
		result.FlatMap(h.repository.BatchDelete(c)),
	).Match(
		func(res BatchDeleteRes) (BatchDeleteRes, error) {
			pkgs.Success(c, res)
			return res, nil
		},
		func(err error) (BatchDeleteRes, error) {
			pkgs.HandleError(c, err)
			return 0, err
		},
	)
}

// QueryList 获取模板列表
//
//	@Summary  获取模板列表
//	@Description  获取模板列表
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    pageSize  query int   false "每页数量"  default(10)
//	@Param    name    query string  false "模板名称"
//	@Success  200     {object}  pkgs.Response{data=QueryTemplateRes}  "获取成功，返回模板列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Router   /template/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	result.Pipe2(
		pkgs.BindQuery[QueryListReq](c),
		result.FlatMap(pkgs.ValidateV2[QueryListReq](h.validator)),
		result.FlatMap(h.repository.QueryList(c)),
	).Match(
		func(res QueryListRes) (QueryListRes, error) {
			pkgs.Success(c, res)
			return res, nil
		},
		func(err error) (QueryListRes, error) {
			pkgs.HandleError(c, err)
			return QueryListRes{}, err
		},
	)
}
