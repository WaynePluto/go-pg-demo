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
	"database/sql"
	"fmt"
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

func NewRoleHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
	}
}

// Create 创建角色
//
//	@Summary      创建角色
//	@Description  创建一个新角色
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        request body  CreateRoleReq true  "创建角色的请求参数"
//	@Success      200  {object}  pkgs.Response{data=string}  "创建成功，返回角色ID"
//	@Failure      400  {object}  pkgs.Response "请求参数错误"
//	@Failure      500  {object}  pkgs.Response "服务器内部错误"
//	@Router       /role [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	entity := &RoleEntity{
		Name:        req.Name,
		Description: req.Description,
	}

	query := `INSERT INTO role (name, description) VALUES (:name, :description) RETURNING id, created_at, updated_at`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to prepare named statement for create role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}
	defer stmt.Close()

	err = stmt.GetContext(c.Request.Context(), entity, entity)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}

	pkgs.Success(c, entity.ID)
}

// Get 根据ID获取角色
//
//	@Summary      获取角色详情
//	@Description  根据ID获取单个角色的详细信息
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string  true  "角色ID"
//	@Success      200  {object}  pkgs.Response{data=RoleRes} "获取成功，返回角色信息"
//	@Failure      404  {object}  pkgs.Response "角色不存在"
//	@Failure      500  {object}  pkgs.Response "服务器内部错误"
//	@Router       /role/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	var entity RoleEntity
	query := `SELECT id, name, description, created_at, updated_at FROM role WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "Role not found")
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get role")
		return
	}

	response := RoleRes{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		CreatedAt:   entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   entity.UpdatedAt.Format(time.RFC3339),
	}
	pkgs.Success(c, response)
}

// Update 更新角色
//
//	@Summary      更新角色信息
//	@Description  根据ID更新角色的名称或描述
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string  true  "角色ID"
//	@Param        request body  UpdateRoleReq true  "更新角色的请求参数"
//	@Success      200  {object}  pkgs.Response "更新成功"
//	@Failure      400  {object}  pkgs.Response "请求参数错误"
//	@Failure      500  {object}  pkgs.Response "服务器内部错误"
//	@Router       /role/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	params := map[string]interface{}{"id": id}
	var setClauses []string

	if req.Name != nil {
		params["name"] = *req.Name
		setClauses = append(setClauses, "name = :name")
	}
	if req.Description != nil {
		params["description"] = *req.Description
		setClauses = append(setClauses, "description = :description")
	}

	if len(setClauses) == 0 {
		pkgs.Success(c, nil)
		return
	}

	query := "UPDATE role SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

	_, err := h.db.NamedExecContext(c.Request.Context(), query, params)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update role")
		return
	}

	pkgs.Success(c, nil)
}

// Delete 删除角色
//
//	@Summary      删除角色
//	@Description  根据ID删除一个角色
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string  true  "角色ID"
//	@Success      200  {object}  pkgs.Response{data=int64} "删除成功，返回影响行数"
//	@Failure      500  {object}  pkgs.Response "服务器内部错误"
//	@Router       /role/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM role WHERE id = :id`
	res, err := h.db.NamedExecContext(c.Request.Context(), query, map[string]interface{}{"id": id})
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get affected rows for role deletion", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	pkgs.Success(c, affectedRows)
}

// List 分页查询角色
//
//	@Summary      分页查询角色
//	@Description  根据条件分页获取角色列表
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        page     query     int     false  "页码"
//	@Param        pageSize query     int     false  "每页数量"
//	@Param        name     query     string  false  "按名称模糊查询"
//	@Success      200      {object}  pkgs.Response{data=RoleListRes} "获取成功，返回角色列表和总数"
//	@Failure      400      {object}  pkgs.Response "请求参数错误"
//	@Failure      500      {object}  pkgs.Response "服务器内部错误"
//	@Router       /role/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	var req QueryRoleReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	var entities []RoleEntity
	var total int64

	baseQuery := "FROM role WHERE 1=1"
	params := make(map[string]interface{})

	if req.Name != "" {
		baseQuery += " AND name ILIKE :name"
		params["name"] = "%" + req.Name + "%"
	}

	countQuery := "SELECT count(*) " + baseQuery
	nstmt, err := h.db.PrepareNamedContext(c.Request.Context(), countQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named count query for roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query roles")
		return
	}
	defer nstmt.Close()
	err = nstmt.GetContext(c.Request.Context(), &total, params)
	if err != nil {
		h.logger.Error("Failed to count roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query roles")
		return
	}

	if total == 0 {
		pkgs.Success(c, RoleListRes{List: []RoleRes{}, Total: 0})
		return
	}

	listQuery := `SELECT id, name, description, created_at, updated_at ` + baseQuery + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`
	params["limit"] = req.PageSize
	params["offset"] = (req.Page - 1) * req.PageSize

	nstmt, err = h.db.PrepareNamedContext(c.Request.Context(), listQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named list query for roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query roles")
		return
	}
	defer nstmt.Close()
	err = nstmt.SelectContext(c.Request.Context(), &entities, params)
	if err != nil {
		h.logger.Error("Failed to select roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query roles")
		return
	}

	var responseEntities []RoleRes
	for _, entity := range entities {
		responseEntities = append(responseEntities, RoleRes{
			ID:          entity.ID,
			Name:        entity.Name,
			Description: entity.Description,
			CreatedAt:   entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   entity.UpdatedAt.Format(time.RFC3339),
		})
	}

	pkgs.Success(c, RoleListRes{List: responseEntities, Total: total})
}

// AssignPermission 为角色分配权限
//
//	@Summary      为角色分配权限
//	@Description  清空角色现有权限，并重新关联新的权限列表
//	@Tags         role
//	@Accept       json
//	@Produce      json
//	@Param        id      path      string  true  "角色ID"
//	@Param        request body      AssignPermissionsReq true  "分配权限的请求参数"
//	@Success      200     {object}  pkgs.Response "分配成功"
//	@Failure      400     {object}  pkgs.Response "请求参数错误"
//	@Failure      500     {object}  pkgs.Response "服务器内部错误"
//	@Router       /role/{id}/permission [post]
func (h *Handler) AssignPermission(c *gin.Context) {
	roleID := c.Param("id")
	var req AssignPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	tx, err := h.db.BeginTxx(c.Request.Context(), nil)
	if err != nil {
		h.logger.Error("Failed to begin transaction for assigning permissions", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign permissions")
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
				h.logger.Error("Failed to commit transaction for assigning permissions", zap.Error(err))
			}
		}
	}()

	// 删除旧的关联
	deleteQuery := `DELETE FROM role_permission WHERE role_id = $1`
	if _, err = tx.ExecContext(c.Request.Context(), deleteQuery, roleID); err != nil {
		h.logger.Error("Failed to delete old permissions for role", zap.String("roleID", roleID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign permissions")
		return
	}

	if len(req.PermissionIDs) == 0 {
		pkgs.Success(c, nil)
		return
	}

	// 插入新的关联
	var values []string
	var params []interface{}
	params = append(params, roleID)
	for i, permID := range req.PermissionIDs {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		params = append(params, permID)
	}

	insertQuery := `INSERT INTO role_permission (role_id, permission_id) VALUES ` + strings.Join(values, ",")
	if _, err = tx.ExecContext(c.Request.Context(), insertQuery, params...); err != nil {
		h.logger.Error("Failed to insert new permissions for role", zap.String("roleID", roleID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign permissions")
		return
	}

	pkgs.Success(c, nil)
}
