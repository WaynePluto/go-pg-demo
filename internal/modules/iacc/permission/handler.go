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
	"database/sql"
	"net/http"
	"strings"
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
}

func NewPermissionHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
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
//	@Success  200   {object}  pkgs.Response{data=string}  "创建成功，返回权限ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission [post]
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求参数
	var req CreatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 创建实体
	entity := &PermissionEntity{
		Name:     req.Name,
		Type:     req.Type,
		Metadata: req.Metadata,
	}

	// 数据库操作
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES (:name, :type, :metadata) RETURNING id, created_at, updated_at`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("创建权限语句准备失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "创建权限失败")
		return
	}
	defer stmt.Close()

	err = stmt.GetContext(c.Request.Context(), entity, entity)
	if err != nil {
		h.logger.Error("创建权限失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "创建权限失败")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

// GetByID 根据ID获取权限
//
//	@Summary  根据ID获取权限
//	@Description  根据ID获取权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "权限ID"
//	@Success  200 {object}  pkgs.Response{data=PermissionRes}  "获取成功，返回权限信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "权限不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusNotFound, "权限不存在")
		return
	}

	// 数据库操作
	var entity PermissionEntity
	query := `SELECT id, name, type, metadata, created_at, updated_at FROM iacc_permission WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "权限不存在")
			return
		}
		h.logger.Error("获取权限失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "获取权限失败")
		return
	}

	// 返回结果
	response := PermissionRes{
		ID:        entity.ID,
		Name:      entity.Name,
		Type:      entity.Type,
		Metadata:  entity.Metadata,
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
	}
	pkgs.Success(c, response)
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
//	@Success  200   {object}  pkgs.Response       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 绑定请求参数
	var req UpdatePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 动态构建更新语句
	params := map[string]interface{}{"id": id}
	var setClauses []string

	if req.Name != nil {
		params["name"] = *req.Name
		setClauses = append(setClauses, "name = :name")
	}
	if req.Type != nil {
		params["type"] = *req.Type
		setClauses = append(setClauses, "type = :type")
	}
	if req.Metadata != nil {
		params["metadata"] = *req.Metadata
		setClauses = append(setClauses, "metadata = :metadata")
	}

	// 如果没有需要更新的字段，直接返回成功
	if len(setClauses) == 0 {
		pkgs.Success(c, nil)
		return
	}

	query := "UPDATE iacc_permission SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

	// 执行数据库操作
	_, err := h.db.NamedExecContext(c.Request.Context(), query, params)
	if err != nil {
		h.logger.Error("更新权限失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "更新权限失败")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
}

// DeleteByID 根据ID删除权限
//
//	@Summary  根据ID删除权限
//	@Description  根据ID删除权限
//	@Tags   permission
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "权限ID"
//	@Success  200 {object}  pkgs.Response{data=int64} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Security JWT
//	@Router   /permission/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 数据库操作
	query := `DELETE FROM iacc_permission WHERE id = :id`
	res, err := h.db.NamedExecContext(c.Request.Context(), query, map[string]interface{}{"id": id})
	if err != nil {
		h.logger.Error("删除权限失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "删除权限失败")
		return
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("获取影响行数失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "删除权限失败")
		return
	}

	// 返回结果
	pkgs.Success(c, affectedRows)
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
//	@Success  200     {object}  pkgs.Response{data=PermissionListRes}  "获取成功，返回权限列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Security JWT
//	@Router   /permission/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	// 绑定请求参数
	var req QueryPermissionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 构建查询
	var entities []PermissionEntity
	var total int64

	baseQuery := "FROM iacc_permission WHERE 1=1"
	params := make(map[string]interface{})

	if req.Name != "" {
		baseQuery += " AND name ILIKE :name"
		params["name"] = "%" + req.Name + "%"
	}
	if req.Type != "" {
		baseQuery += " AND type = :type"
		params["type"] = req.Type
	}

	// 查询总数
	countQuery := "SELECT count(*) " + baseQuery
	nstmt, err := h.db.PrepareNamedContext(c.Request.Context(), countQuery)
	if err != nil {
		h.logger.Error("准备命名计数查询失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询权限失败")
		return
	}
	defer nstmt.Close()
	err = nstmt.GetContext(c.Request.Context(), &total, params)
	if err != nil {
		h.logger.Error("统计权限数量失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询权限失败")
		return
	}

	if total == 0 {
		pkgs.Success(c, PermissionListRes{List: []PermissionRes{}, Total: 0})
		return
	}

	// 查询列表
	listQuery := `SELECT id, name, type, metadata, created_at, updated_at ` + baseQuery + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`
	params["limit"] = req.PageSize
	params["offset"] = (req.Page - 1) * req.PageSize

	nstmt, err = h.db.PrepareNamedContext(c.Request.Context(), listQuery)
	if err != nil {
		h.logger.Error("准备命名列表查询失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询权限失败")
		return
	}
	defer nstmt.Close()
	err = nstmt.SelectContext(c.Request.Context(), &entities, params)
	if err != nil {
		h.logger.Error("查询权限列表失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询权限失败")
		return
	}

	// 返回结果
	var responseEntities []PermissionRes
	for _, entity := range entities {

		responseEntities = append(responseEntities, PermissionRes{
			ID:        entity.ID,
			Name:      entity.Name,
			Type:      entity.Type,
			Metadata:  entity.Metadata,
			CreatedAt: entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
		})
	}

	pkgs.Success(c, PermissionListRes{List: responseEntities, Total: total})
}