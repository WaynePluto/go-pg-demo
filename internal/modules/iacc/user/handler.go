package user

import (
	"database/sql"
	"net/http"
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
	entity := &User{
		Username: req.Username,
		Phone:    req.Phone,
		Password: req.Password, // 注意：实际项目中需要加密存储
		Profile:  req.Profile,
	}

	// 数据库操作
	query := `INSERT INTO iacc_user (username, phone, password, profile) 
            VALUES (:username, :phone, :password, :profile) RETURNING id`
	err := h.db.QueryRowxContext(c.Request.Context(), query, entity).Scan(&entity.ID)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
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

	// 更新实体
	now := time.Now()
	entity := &User{
		ID:        id,
		UpdatedAt: now,
		Username:  *req.Username,
		Phone:     *req.Phone,
		Profile:   req.Profile,
	}

	// 数据库操作
	query := `UPDATE iacc_user SET updated_at=:updated_at, username=:username, phone=:phone, profile=:profile 
            WHERE id=:id`
	result, err := h.db.NamedExecContext(c.Request.Context(), query, entity)
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
	response := UserResponse{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Username:  updatedUser.Username,
		Phone:     updatedUser.Phone,
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
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to get user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// 返回结果
	response := UserResponse{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Username:  entity.Username,
		Phone:     entity.Phone,
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

	// 添加过滤条件
	var args []interface{}
	argIndex := 1
	if req.Username != "" {
		query += " WHERE username LIKE $" + string(rune(argIndex+'0'))
		countQuery += " WHERE username LIKE $" + string(rune(argIndex+'0'))
		args = append(args, "%"+req.Username+"%")
		argIndex++
	}
	if req.Phone != "" {
		if argIndex == 1 {
			query += " WHERE phone LIKE $" + string(rune(argIndex+'0'))
			countQuery += " WHERE phone LIKE $" + string(rune(argIndex+'0'))
		} else {
			query += " AND phone LIKE $" + string(rune(argIndex+'0'))
			countQuery += " AND phone LIKE $" + string(rune(argIndex+'0'))
		}
		args = append(args, "%"+req.Phone+"%")
		argIndex++
	}

	// 添加分页
	query += " ORDER BY created_at DESC LIMIT $" + string(rune(argIndex+'0')) + " OFFSET $" + string(rune(argIndex+1+'0'))
	args = append(args, limit, offset)

	// 执行查询
	err := h.db.SelectContext(c.Request.Context(), &users, query, args...)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// 查询总数
	var totalCount int64
	err = h.db.GetContext(c.Request.Context(), &totalCount, countQuery, args[:len(args)-2]...)
	if err != nil {
		h.logger.Error("Failed to get users count", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get users count")
		return
	}

	// 转换为响应格式
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
			Phone:     user.Phone,
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
func (h *Handler) AssignRoles(c *gin.Context) {
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
func (h *Handler) Delete(c *gin.Context) {
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
