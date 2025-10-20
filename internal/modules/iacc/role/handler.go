package role

import (
	"database/sql"
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
}

func NewRoleHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
	}
}

// CreateRole 创建角色
//
//	@Summary  创建角色
//	@Description  创建角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateRoleRequest true  "创建角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "创建成功，返回角色ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role [post]
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求参数
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 创建实体
	entity := &Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	// 数据库操作
	query := `INSERT INTO iacc_role (name, description, permissions) 
            VALUES ($1, $2, $3) RETURNING id`
	err := h.db.QueryRowContext(c.Request.Context(), query, entity.Name, entity.Description, entity.Permissions).Scan(&entity.ID)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

// UpdateRole 更新角色
//
//	@Summary  更新角色
//	@Description  更新角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string        true  "角色ID"
//	@Param    request body  UpdateRoleRequest true  "更新角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=RoleResponse}  "更新成功，返回角色信息"
//	@Failure  400   {object}  pkgs.Response           "请求参数错误"
//	@Failure  404   {object}  pkgs.Response           "角色不存在"
//	@Failure  500   {object}  pkgs.Response           "服务器内部错误"
//	@Router   /role/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 绑定请求参数
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 更新实体
	now := time.Now()
	entity := &Role{
		ID:          id,
		UpdatedAt:   now,
		Name:        *req.Name,
		Description: *req.Description,
		Permissions: req.Permissions,
	}

	// 数据库操作
	query := `UPDATE iacc_role SET updated_at=:updated_at, name=:name, description=:description, permissions=:permissions 
            WHERE id=:id`
	result, err := h.db.NamedExecContext(c.Request.Context(), query, entity)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update role")
		return
	}

	// 检查是否有记录被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update role")
		return
	}
	if rowsAffected == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 查询更新后的角色信息
	var updatedRole Role
	query = `SELECT id, created_at, updated_at, name, description, permissions FROM iacc_role WHERE id=$1`
	err = h.db.GetContext(c.Request.Context(), &updatedRole, query, id)
	if err != nil {
		h.logger.Error("Failed to get updated role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get updated role")
		return
	}

	// 返回结果
	response := (RoleResponse)(updatedRole)
	pkgs.Success(c, response)
}

// GetRoleByID 根据ID获取角色
//
//	@Summary  根据ID获取角色
//	@Description  根据ID获取角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "角色ID"
//	@Success  200 {object}  pkgs.Response{data=RoleResponse}  "获取成功，返回角色信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "角色不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /role/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 数据库操作
	var entity Role
	query := `SELECT id, created_at, updated_at, name, description, permissions FROM iacc_role WHERE id = $1`
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

	// 返回结果
	response := (RoleResponse)(entity)
	pkgs.Success(c, response)
}

// ListRoles 获取角色列表
//
//	@Summary  获取角色列表
//	@Description  获取角色列表
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    page_size query int   false "每页数量"  default(10)
//	@Param    name    query string  false "角色名"
//	@Param    description query string  false "角色描述"
//	@Success  200 {object}  pkgs.Response{data=ListRolesResponse} "获取成功，返回角色列表"
//	@Failure  400 {object}  pkgs.Response             "请求参数错误"
//	@Failure  500 {object}  pkgs.Response             "服务器内部错误"
//	@Router   /role/list [get]
func (h *Handler) List(c *gin.Context) {
	// 绑定请求参数
	var req ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 构建查询条件
	offset := (req.Page - 1) * req.PageSize
	limit := req.PageSize

	// 查询角色列表
	var roles []Role
	query := `SELECT id, created_at, updated_at, name, description, permissions FROM iacc_role`
	countQuery := `SELECT COUNT(*) FROM iacc_role`

	// 添加过滤条件
	var args []interface{}
	argIndex := 1
	if req.Name != "" {
		query += " WHERE name LIKE $" + string(rune(argIndex+'0'))
		countQuery += " WHERE name LIKE $" + string(rune(argIndex+'0'))
		args = append(args, "%"+req.Name+"%")
		argIndex++
	}
	if req.Description != "" {
		if argIndex == 1 {
			query += " WHERE description LIKE $" + string(rune(argIndex+'0'))
			countQuery += " WHERE description LIKE $" + string(rune(argIndex+'0'))
		} else {
			query += " AND description LIKE $" + string(rune(argIndex+'0'))
			countQuery += " AND description LIKE $" + string(rune(argIndex+'0'))
		}
		args = append(args, "%"+req.Description+"%")
		argIndex++
	}

	// 添加分页
	query += " ORDER BY created_at DESC LIMIT $" + string(rune(argIndex+'0')) + " OFFSET $" + string(rune(argIndex+1+'0'))
	args = append(args, limit, offset)

	// 执行查询
	err := h.db.SelectContext(c.Request.Context(), &roles, query, args...)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	// 查询总数
	var totalCount int64
	err = h.db.GetContext(c.Request.Context(), &totalCount, countQuery, args[:len(args)-2]...)
	if err != nil {
		h.logger.Error("Failed to get roles count", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get roles count")
		return
	}

	// 转换为响应格式
	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = (RoleResponse)(role)
	}

	// 返回结果
	response := ListRolesResponse{
		Roles:      roleResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: totalCount,
	}
	pkgs.Success(c, response)
}

// AssignPermission 分配权限
//
//	@Summary  分配权限
//	@Description  为角色分配权限
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "角色ID"
//	@Param    request body  AssignPermissionRequest true  "分配权限请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}    "分配成功"
//	@Failure  400   {object}  pkgs.Response         "请求参数错误"
//	@Failure  500   {object}  pkgs.Response         "服务器内部错误"
//	@Router   /role/{id}/permission [post]
func (h *Handler) AssignPermission(c *gin.Context) {
	// 获取角色ID
	roleID := c.Param("id")

	// 验证角色ID是否为有效的UUID
	if _, err := uuid.Parse(roleID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 绑定请求参数
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 检查角色是否存在
	var roleCount int
	roleQuery := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &roleCount, roleQuery, roleID)
	if err != nil {
		h.logger.Error("Failed to check role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign permission")
		return
	}
	if roleCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 获取角色当前权限
	var currentPermissions sql.NullString
	query := `SELECT permissions FROM iacc_role WHERE id = $1`
	err = h.db.GetContext(c.Request.Context(), &currentPermissions, query, roleID)
	if err != nil {
		h.logger.Error("Failed to get role permissions", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign permission")
		return
	}

	// 返回结果
	pkgs.Success(c, "Permission assigned successfully")
}

// RemovePermission 移除权限
//
//	@Summary  移除权限
//	@Description  为角色移除权限
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id        path  string  true  "角色ID"
//	@Param    permission_key  path  string  true  "权限键"
//	@Success  200   {object}  pkgs.Response "移除成功"
//	@Failure  400   {object}  pkgs.Response "请求参数错误"
//	@Failure  404   {object}  pkgs.Response "角色或权限不存在"
//	@Failure  500   {object}  pkgs.Response "服务器内部错误"
//	@Router   /role/{id}/permission/{permission_key} [delete]
func (h *Handler) RemovePermission(c *gin.Context) {
	// 获取角色ID和权限键
	roleID := c.Param("id")

	// 验证角色ID是否为有效的UUID
	if _, err := uuid.Parse(roleID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 检查角色是否存在
	var roleCount int
	roleQuery := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &roleCount, roleQuery, roleID)
	if err != nil {
		h.logger.Error("Failed to check role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to remove permission")
		return
	}
	if roleCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 返回结果
	pkgs.Success(c, "Permission removed successfully")
}

// DeleteRole 根据ID删除角色
//
//	@Summary  根据ID删除角色
//	@Description  根据ID删除角色
//	@Tags   role
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "角色ID"
//	@Success  200 {object}  pkgs.Response{data=int64} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  404 {object}  pkgs.Response       "角色不存在"
//	@Failure  409 {object}  pkgs.Response       "角色正在被使用，无法删除"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Router   /role/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 检查角色是否存在
	var roleCount int
	checkQuery := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &roleCount, checkQuery, id)
	if err != nil {
		h.logger.Error("Failed to check role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	if roleCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 检查是否有用户拥有该角色
	var userRoleCount int
	userRoleQuery := `SELECT COUNT(*) FROM iacc_user_role WHERE role_id = $1`
	err = h.db.GetContext(c.Request.Context(), &userRoleCount, userRoleQuery, id)
	if err != nil {
		h.logger.Error("Failed to check user role associations", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	if userRoleCount > 0 {
		pkgs.Error(c, http.StatusConflict, "该角色正在被使用，无法删除")
		return
	}

	// 删除角色
	query := `DELETE FROM iacc_role WHERE id = $1`
	res, err := h.db.ExecContext(c.Request.Context(), query, id)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get affected rows", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	// 返回结果
	pkgs.Success(c, affectedRows)
}
