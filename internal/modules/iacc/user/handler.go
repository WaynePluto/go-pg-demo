package user

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-pg-demo/internal/modules/iacc/service"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Handler struct {
	db                *sqlx.DB
	logger            *zap.Logger
	validator         *pkgs.RequestValidator
	permissionService *service.PermissionService
}

func NewUserHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator, permissionService *service.PermissionService) *Handler {
	return &Handler{
		db:                db,
		logger:            logger,
		validator:         validator,
		permissionService: permissionService,
	}
}

// CreateUser 创建用户
//
//	@Summary  创建用户
//	@Description  创建用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateUserRequest true  "创建用户请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "创建成功，返回用户ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  409   {object}  pkgs.Response       "用户名或手机号已存在"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user [post]
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求参数
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 创建实体

	userData := User{
		Username: req.Username,
		Phone:    &req.Phone,
		Password: req.Password, // 注意：实际项目中需要加密存储
		Profile:  req.Profile,
	}

	// 数据库操作，使用 Named prepared statement 获取 RETURNING
	query := `INSERT INTO iacc_user (username, phone, password, profile) 
      VALUES (:username, :phone, :password, :profile) RETURNING id`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to prepare insert user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}
	defer stmt.Close()

	params := map[string]interface{}{
		"username": userData.Username,
		"phone":    userData.Phone,
		"password": userData.Password,
		"profile":  userData.Profile,
	}
	err = stmt.GetContext(c.Request.Context(), &userData.ID, params)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		// 检查是否违反唯一约束
		if strings.Contains(err.Error(), "duplicate key value") {
			if strings.Contains(err.Error(), "username") {
				pkgs.Error(c, http.StatusConflict, "用户名已存在")
				return
			}
			if strings.Contains(err.Error(), "phone") {
				pkgs.Error(c, http.StatusConflict, "手机号已存在")
				return
			}
		}
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// 返回结果
	pkgs.Success(c, userData.ID)
}

// UpdateUser 更新用户
//
//	@Summary  更新用户
//	@Description  更新用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string        true  "用户ID"
//	@Param    request body  UpdateUserRequest true  "更新用户请求参数"
//	@Success  200   {object}  pkgs.Response{data=UserResponse}  "更新成功，返回用户信息"
//	@Failure  400   {object}  pkgs.Response           "请求参数错误"
//	@Failure  404   {object}  pkgs.Response           "用户不存在"
//	@Failure  500   {object}  pkgs.Response           "服务器内部错误"
//	@Router   /user/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 绑定请求参数
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 构建更新 map，仅包含需要更新的列（使用指针判断是否传入）
	now := time.Now()
	data := map[string]interface{}{
		"updated_at": now,
		"id":         id,
	}
	if req.Username != nil {
		data["username"] = *req.Username
	}
	if req.Phone != nil {
		data["phone"] = *req.Phone
	}
	// 只有当前端显式传入 profile 时才更新（支持传空字符串或 null）
	if req.Profile != nil {
		data["profile"] = req.Profile
	}

	// 构建动态 SET 列表，列名为白名单（由 data 的 key 控制，这里仅允许上述字段）
	// 列白名单，避免非预期列被更新
	allowedCols := map[string]bool{
		"username":   true,
		"phone":      true,
		"profile":    true,
		"updated_at": true,
	}
	cols := []string{}
	for k := range data {
		if k == "id" {
			continue
		}
		if !allowedCols[k] {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s = :%s", k, k))
	}
	query := fmt.Sprintf("UPDATE iacc_user SET %s WHERE id = :id", strings.Join(cols, ", "))

	// 数据库操作（命名参数）
	result, err := h.db.NamedExecContext(c.Request.Context(), query, data)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// 检查是否有记录被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}
	if rowsAffected == 0 {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 查询更新后的用户信息
	var updatedUser User
	query = `SELECT id, created_at, updated_at, username, phone, profile FROM iacc_user WHERE id=$1`
	err = h.db.GetContext(c.Request.Context(), &updatedUser, query, id)
	if err != nil {
		h.logger.Error("Failed to get updated user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get updated user")
		return
	}

	// 返回结果
	respPhone := ""
	if updatedUser.Phone != nil {
		respPhone = *updatedUser.Phone
	}
	response := UserResponse{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Username:  updatedUser.Username,
		Phone:     respPhone,
		Profile:   updatedUser.Profile,
	}
	pkgs.Success(c, response)
}

// GetUserByID 根据ID获取用户
//
//	@Summary  根据ID获取用户
//	@Description  根据ID获取用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "用户ID"
//	@Success  200 {object}  pkgs.Response{data=UserResponse}  "获取成功，返回用户信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "用户不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /user/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 数据库操作
	var entity User
	query := `SELECT id, created_at, updated_at, username, phone, profile FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkgs.Error(c, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to get user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// 返回结果
	respPhone := ""
	if entity.Phone != nil {
		respPhone = *entity.Phone
	}
	response := UserResponse{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Username:  entity.Username,
		Phone:     respPhone,
		Profile:   entity.Profile,
	}
	pkgs.Success(c, response)
}

// ListUsers 获取用户列表
//
//	@Summary  获取用户列表
//	@Description  获取用户列表
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    page_size query int   false "每页数量"  default(10)
//	@Param    username  query string  false "用户名"
//	@Param    phone   query string  false "手机号"
//	@Success  200 {object}  pkgs.Response{data=ListUsersResponse} "获取成功，返回用户列表"
//	@Failure  400 {object}  pkgs.Response             "请求参数错误"
//	@Failure  500 {object}  pkgs.Response             "服务器内部错误"
//	@Router   /user/list [get]
func (h *Handler) List(c *gin.Context) {
	// 绑定请求参数
	var req ListUsersRequest
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

	// 查询用户列表
	var users []User
	query := `SELECT id, created_at, updated_at, username, phone, profile FROM iacc_user`
	countQuery := `SELECT COUNT(*) FROM iacc_user`

	// 添加过滤条件，使用命名参数并 BindNamed -> Rebind
	params := map[string]interface{}{}
	if req.Username != "" {
		params["username"] = "%" + req.Username + "%"
	}
	if req.Phone != "" {
		params["phone"] = "%" + req.Phone + "%"
	}
	params["limit"] = limit
	params["offset"] = offset

	if req.Username != "" {
		query += " WHERE username LIKE :username"
		countQuery += " WHERE username LIKE :username"
	}
	if req.Phone != "" {
		if req.Username == "" {
			query += " WHERE phone LIKE :phone"
			countQuery += " WHERE phone LIKE :phone"
		} else {
			query += " AND phone LIKE :phone"
			countQuery += " AND phone LIKE :phone"
		}
	}

	// 添加分页
	query += " ORDER BY created_at DESC LIMIT :limit OFFSET :offset"

	// BindNamed -> Rebind -> SelectContext
	q, args, err := h.db.BindNamed(query, params)
	if err != nil {
		h.logger.Error("Failed to bind named query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to list users")
		return
	}
	q = h.db.Rebind(q)
	err = h.db.SelectContext(c.Request.Context(), &users, q, args...)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// 查询总数（不包含分页参数）
	var totalCount int64
	countParams := map[string]interface{}{}
	if req.Username != "" {
		countParams["username"] = "%" + req.Username + "%"
	}
	if req.Phone != "" {
		countParams["phone"] = "%" + req.Phone + "%"
	}
	cq, carg, err := h.db.BindNamed(countQuery, countParams)
	if err != nil {
		h.logger.Error("Failed to bind named count query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get users count")
		return
	}
	cq = h.db.Rebind(cq)
	err = h.db.GetContext(c.Request.Context(), &totalCount, cq, carg...)
	if err != nil {
		h.logger.Error("Failed to get users count", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get users count")
		return
	}

	// 转换为响应格式
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		phone := ""
		if user.Phone != nil {
			phone = *user.Phone
		}
		userResponses[i] = UserResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
			Phone:     phone,
			Profile:   user.Profile,
		}
	}

	// 返回结果
	response := ListUsersResponse{
		Users:      userResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: totalCount,
	}
	pkgs.Success(c, response)
}

// AssignRole 分配角色
//
//	@Summary  分配角色
//	@Description  为用户分配角色
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string        true  "用户ID"
//	@Param    request body  AssignRoleRequest true  "分配角色请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "分配成功，返回用户角色关联ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user/{id}/role [post]
func (h *Handler) AssignRole(c *gin.Context) {
	// 获取用户ID
	userID := c.Param("id")

	// 验证用户ID是否为有效的UUID
	if _, err := uuid.Parse(userID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

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

	// 验证角色ID是否为有效的UUID
	if _, err := uuid.Parse(req.RoleID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 检查用户是否存在
	var userCount int
	userQuery := `SELECT COUNT(*) FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &userCount, userQuery, userID)
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
	err = h.db.GetContext(c.Request.Context(), &userRoleCount, userRoleQuery, userID, req.RoleID)
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
		"user_id":    userID,
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

// RemoveRole 移除角色
//
//	@Summary  移除角色
//	@Description  为用户移除角色
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string  true  "用户ID"
//	@Param    role_id path  string  true  "角色ID"
//	@Success  200   {object}  pkgs.Response "移除成功"
//	@Failure  400   {object}  pkgs.Response "请求参数错误"
//	@Failure  404   {object}  pkgs.Response "用户或角色不存在"
//	@Failure  500   {object}  pkgs.Response "服务器内部错误"
//	@Router   /user/{id}/role/{role_id} [delete]
func (h *Handler) RemoveRole(c *gin.Context) {
	// 获取用户ID和角色ID
	userID := c.Param("id")
	roleID := c.Param("role_id")

	// 验证用户ID是否为有效的UUID
	if _, err := uuid.Parse(userID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 验证角色ID是否为有效的UUID
	if _, err := uuid.Parse(roleID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// 检查用户是否存在
	var userCount int
	userQuery := `SELECT COUNT(*) FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &userCount, userQuery, userID)
	if err != nil {
		h.logger.Error("Failed to check user existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to remove role")
		return
	}
	if userCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 检查角色是否存在
	var roleCount int
	roleQuery := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
	err = h.db.GetContext(c.Request.Context(), &roleCount, roleQuery, roleID)
	if err != nil {
		h.logger.Error("Failed to check role existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to remove role")
		return
	}
	if roleCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "Role not found")
		return
	}

	// 数据库操作
	query := `DELETE FROM iacc_user_role WHERE user_id = $1 AND role_id = $2`
	result, err := h.db.ExecContext(c.Request.Context(), query, userID, roleID)
	if err != nil {
		h.logger.Error("Failed to remove role from user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to remove role")
		return
	}

	// 检查是否有记录被删除
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to remove role")
		return
	}
	if rowsAffected == 0 {
		pkgs.Error(c, http.StatusNotFound, "User role not found")
		return
	}

	// 返回结果
	pkgs.Success(c, "Role removed successfully")
}

// Delete 删除用户
//
//	@Summary  删除用户
//	@Description  删除用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "用户ID"
//	@Success  200 {object}  pkgs.Response "删除成功"
//	@Failure  400 {object}  pkgs.Response "请求参数错误"
//	@Failure  404 {object}  pkgs.Response "用户不存在"
//	@Failure  500 {object}  pkgs.Response "服务器内部错误"
//	@Router   /user/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 数据库操作
	query := `DELETE FROM iacc_user WHERE id = $1`
	result, err := h.db.ExecContext(c.Request.Context(), query, id)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// 检查是否有记录被删除
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get rows affected", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}
	if rowsAffected == 0 {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 返回结果
	pkgs.Success(c, "User deleted successfully")
}

// GetUserPermissions 获取用户权限
//
//	@Summary  获取用户权限
//	@Description  获取用户所有有效权限
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "用户ID"
//	@Success  200 {object}  pkgs.Response{data=[]string}  "获取成功，返回权限列表"
//	@Failure  400 {object}  pkgs.Response         "请求参数错误"
//	@Failure  500 {object}  pkgs.Response         "服务器内部错误"
//	@Router   /user/{id}/permissions [get]
func (h *Handler) GetUserPermissions(c *gin.Context) {
	// 获取用户ID
	userID := c.Param("id")

	// 验证用户ID是否为有效的UUID
	if _, err := uuid.Parse(userID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 检查用户是否存在
	var userCount int
	userQuery := `SELECT COUNT(*) FROM iacc_user WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &userCount, userQuery, userID)
	if err != nil {
		h.logger.Error("Failed to check user existence", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user permissions")
		return
	}
	if userCount == 0 {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 计算用户有效权限集
	permissions, err := h.permissionService.CalculateEffectivePermissionsForUser(userID)
	if err != nil {
		h.logger.Error("Failed to calculate user permissions", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user permissions")
		return
	}

	// 返回结果
	pkgs.Success(c, permissions)
}
