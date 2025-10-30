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

func NewTemplateHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
	}
}

// Create 创建模板
//
//	@Summary  创建模板
//	@Description  创建模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateTemplateRequest true  "创建模板请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "创建成功，返回模板ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template [post]
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求参数
	var req CreateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 创建实体
	entity := &TemplateEntity{
		Name: req.Name,
		Num:  req.Num,
	}

	// 数据库操作
	query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id, created_at, updated_at`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to prepare named statement for create", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create template")
		return
	}
	defer stmt.Close()

	err = stmt.GetContext(c.Request.Context(), entity, entity)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create template")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

// BatchCreate 批量创建模板
//
//	@Summary  批量创建模板
//	@Description  批量创建模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateTemplatesRequest  true  "批量创建模板请求参数"
//	@Success  200   {object}  pkgs.Response{data=[]string}  "创建成功，返回模板ID列表"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /template/batch-create [post]
func (h *Handler) BatchCreate(c *gin.Context) {
	// 绑定请求参数
	var req CreateTemplatesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 准备批量插入的实体
	var entities []TemplateEntity
	for _, t := range req.Templates {
		entities = append(entities, TemplateEntity{
			Name: t.Name,
			Num:  t.Num,
		})
	}

	// 开启事务
	tx, err := h.db.BeginTxx(c.Request.Context(), nil)
	if err != nil {
		h.logger.Error("Failed to begin transaction", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				h.logger.Error("Failed to commit transaction", zap.Error(err))
			}
		}
	}()

	// 数据库操作
	query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id`
	stmt, err := tx.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to prepare named statement", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
		return
	}
	defer stmt.Close()

	var createdIDs []string
	for _, entity := range entities {
		var id string
		err = stmt.GetContext(c.Request.Context(), &id, entity)
		if err != nil {
			h.logger.Error("Failed to create template in batch", zap.Error(err))
			pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
			return
		}
		createdIDs = append(createdIDs, id)
	}

	// 返回结果
	pkgs.Success(c, createdIDs)
}

// GetByID 根据ID获取模板
//
//	@Summary  根据ID获取模板
//	@Description  根据ID获取模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "模板ID"
//	@Success  200 {object}  pkgs.Response{data=TemplateResponse}  "获取成功，返回模板信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "模板不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /template/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusNotFound, "Template not found")
		return
	}

	// 数据库操作
	var entity TemplateEntity
	query := `SELECT id, name, num, created_at, updated_at FROM template WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to get template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get template")
		return
	}

	// 返回结果
	response := TemplateRes{
		ID:        entity.ID,
		Name:      entity.Name,
		Num:       entity.Num,
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
	}
	pkgs.Success(c, response)
}

// UpdateByID 根据ID更新模板
//
//	@Summary  根据ID更新模板
//	@Description  根据ID更新模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "模板ID"
//	@Param    request body  UpdateTemplateRequest true  "更新模板请求参数"
//	@Success  200   {object}  pkgs.Response       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 绑定请求参数
	var req UpdateTemplateReq
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
	if req.Num != nil {
		params["num"] = *req.Num
		setClauses = append(setClauses, "num = :num")
	}

	// 如果没有需要更新的字段，直接返回成功
	if len(setClauses) == 0 {
		pkgs.Success(c, nil)
		return
	}

	query := "UPDATE template SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

	// 执行数据库操作
	_, err := h.db.NamedExecContext(c.Request.Context(), query, params)
	if err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update template")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
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
	// 获取 ID
	id := c.Param("id")

	// 数据库操作
	query := `DELETE FROM template WHERE id = :id`
	res, err := h.db.NamedExecContext(c.Request.Context(), query, map[string]interface{}{"id": id})
	if err != nil {
		h.logger.Error("Failed to delete template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete template")
		return
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get affected rows", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete template")
		return
	}

	// 返回结果
	pkgs.Success(c, affectedRows)
}

// BatchDelete 批量删除模板
//
//	@Summary  批量删除模板
//	@Description  批量删除模板
//	@Tags   template
//	@Accept   json
//	@Produce  json
//	@Param    request body  DeleteTemplatesRequest  true  "批量删除模板请求参数"
//	@Success  200   {object}  pkgs.Response       "删除成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /template/batch-delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	// 绑定请求参数
	var req DeleteTemplatesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 数据库操作
	query, args, err := sqlx.In(`DELETE FROM template WHERE id IN (?)`, req.IDs)
	if err != nil {
		h.logger.Error("Failed to build delete query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to build delete query")
		return
	}
	query = h.db.Rebind(query)
	_, err = h.db.ExecContext(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to delete templates", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete templates")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
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
//	@Success  200     {object}  pkgs.Response{data=TemplateListResponse}  "获取成功，返回模板列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Router   /template/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	// 绑定请求参数
	var req QueryTemplateReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 构建查询
	var entities []TemplateEntity
	var total int64

	baseQuery := "FROM template WHERE 1=1"
	params := make(map[string]interface{})

	if req.Name != "" {
		baseQuery += " AND name ILIKE :name"
		params["name"] = "%" + req.Name + "%"
	}

	// 查询总数
	countQuery := "SELECT count(*) " + baseQuery
	nstmt, err := h.db.PrepareNamedContext(c.Request.Context(), countQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named count query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}
	defer nstmt.Close()
	err = nstmt.GetContext(c.Request.Context(), &total, params)
	if err != nil {
		h.logger.Error("Failed to count templates", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}

	if total == 0 {
		pkgs.Success(c, gin.H{"list": []TemplateEntity{}, "total": 0})
		return
	}

	// 查询列表
	listQuery := `SELECT id, name, num, created_at, updated_at ` + baseQuery + ` ORDER BY id DESC LIMIT :limit OFFSET :offset`
	params["limit"] = req.PageSize
	params["offset"] = (req.Page - 1) * req.PageSize

	nstmt, err = h.db.PrepareNamedContext(c.Request.Context(), listQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named list query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}
	defer nstmt.Close()
	err = nstmt.SelectContext(c.Request.Context(), &entities, params)
	if err != nil {
		h.logger.Error("Failed to select templates", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}

	// 返回结果
	var responseEntities []TemplateRes
	for _, entity := range entities {
		responseEntities = append(responseEntities, TemplateRes{
			ID:        entity.ID,
			Name:      entity.Name,
			Num:       entity.Num,
			CreatedAt: entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
		})
	}

	pkgs.Success(c, gin.H{"list": responseEntities, "total": total})
}
