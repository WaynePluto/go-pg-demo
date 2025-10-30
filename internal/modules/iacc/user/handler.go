// Package user API.
//
// The API for managing user.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http
package user

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

func NewUserHandler(db *sqlx.DB, logger *zap.Logger, validator *pkgs.RequestValidator) *Handler {
	return &Handler{
		db:        db,
		logger:    logger,
		validator: validator,
	}
}

// Create 创建用户
//
//	@Summary  创建用户
//	@Description  创建用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    request body  CreateUserReq true  "创建用户请求参数"
//	@Success  200   {object}  pkgs.Response{data=string}  "创建成功，返回用户ID"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user [post]
func (h *Handler) Create(c *gin.Context) {
	// 绑定请求参数
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 创建实体
	entity := &UserEntity{
		Phone:    req.Phone,
		Password: req.Password,
		Profile:  req.Profile,
	}

	// 数据库操作
	query := `INSERT INTO "user" (phone, password, profile) VALUES (:phone, :password, :profile) RETURNING id, created_at, updated_at`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to prepare named statement for create user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}
	defer stmt.Close()

	err = stmt.GetContext(c.Request.Context(), entity, entity)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

// GetByID 根据ID获取用户
//
//	@Summary  根据ID获取用户
//	@Description  根据ID获取用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "用户ID"
//	@Success  200 {object}  pkgs.Response{data=UserRes}  "获取成功，返回用户信息"
//	@Failure  400 {object}  pkgs.Response           "请求参数错误"
//	@Failure  404 {object}  pkgs.Response           "用户不存在"
//	@Failure  500 {object}  pkgs.Response           "服务器内部错误"
//	@Router   /user/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusNotFound, "User not found")
		return
	}

	// 数据库操作
	var entity UserEntity
	query := `SELECT id, phone, profile, created_at, updated_at FROM "user" WHERE id = $1`
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
	response := UserRes{
		ID:        entity.ID,
		Phone:     entity.Phone,
		Profile:   entity.Profile,
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
	}
	pkgs.Success(c, response)
}

// UpdateByID 根据ID更新用户
//
//	@Summary  根据ID更新用户
//	@Description  根据ID更新用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "用户ID"
//	@Param    request body  UpdateUserReq true  "更新用户请求参数"
//	@Success  200   {object}  pkgs.Response       "更新成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 绑定请求参数
	var req UpdateUserReq
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

	if req.Password != nil {
		params["password"] = *req.Password
		setClauses = append(setClauses, "password = :password")
	}
	if req.Profile != nil {
		params["profile"] = req.Profile
		setClauses = append(setClauses, "profile = :profile")
	}

	// 如果没有需要更新的字段，直接返回成功
	if len(setClauses) == 0 {
		pkgs.Success(c, nil)
		return
	}

	query := `UPDATE "user" SET ` + strings.Join(setClauses, ", ") + " WHERE id = :id"

	// 执行数据库操作
	_, err := h.db.NamedExecContext(c.Request.Context(), query, params)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
}

// DeleteByID 根据ID删除用户
//
//	@Summary  根据ID删除用户
//	@Description  根据ID删除用户
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id  path  string  true  "用户ID"
//	@Success  200 {object}  pkgs.Response{data=int64} "删除成功，返回影响行数"
//	@Failure  400 {object}  pkgs.Response       "请求参数错误"
//	@Failure  500 {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 数据库操作
	query := `DELETE FROM "user" WHERE id = :id`
	res, err := h.db.NamedExecContext(c.Request.Context(), query, map[string]interface{}{"id": id})
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("Failed to get affected rows for user deletion", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// 返回结果
	pkgs.Success(c, affectedRows)
}

// QueryList 获取用户列表
//
//	@Summary  获取用户列表
//	@Description  获取用户列表
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    page    query int   false "页码"  default(1)
//	@Param    pageSize  query int   false "每页数量"  default(10)
//	@Param    phone    query string  false "手机号"
//	@Success  200     {object}  pkgs.Response{data=UserListRes}  "获取成功，返回用户列表"
//	@Failure  400     {object}  pkgs.Response               "请求参数错误"
//	@Failure  500     {object}  pkgs.Response               "服务器内部错误"
//	@Router   /user/list [get]
func (h *Handler) QueryList(c *gin.Context) {
	// 绑定请求参数
	var req QueryUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 构建查询
	var entities []UserEntity
	var total int64

	baseQuery := `FROM "user" WHERE 1=1`
	params := make(map[string]interface{})

	if req.Phone != "" {
		baseQuery += " AND phone ILIKE :phone"
		params["phone"] = "%" + req.Phone + "%"
	}

	// 查询总数
	countQuery := "SELECT count(*) " + baseQuery
	nstmt, err := h.db.PrepareNamedContext(c.Request.Context(), countQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named count query for users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query users")
		return
	}
	defer nstmt.Close()
	err = nstmt.GetContext(c.Request.Context(), &total, params)
	if err != nil {
		h.logger.Error("Failed to count users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query users")
		return
	}

	if total == 0 {
		pkgs.Success(c, gin.H{"list": []UserEntity{}, "total": 0})
		return
	}

	// 查询列表
	listQuery := `SELECT id, phone, profile, created_at, updated_at ` + baseQuery + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`
	params["limit"] = req.PageSize
	params["offset"] = (req.Page - 1) * req.PageSize

	nstmt, err = h.db.PrepareNamedContext(c.Request.Context(), listQuery)
	if err != nil {
		h.logger.Error("Failed to prepare named list query for users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query users")
		return
	}
	defer nstmt.Close()
	err = nstmt.SelectContext(c.Request.Context(), &entities, params)
	if err != nil {
		h.logger.Error("Failed to select users", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query users")
		return
	}

	// 返回结果
	var responseEntities []UserRes
	for _, entity := range entities {
		responseEntities = append(responseEntities, UserRes{
			ID:        entity.ID,
			Phone:     entity.Phone,
			Profile:   entity.Profile,
			CreatedAt: entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
		})
	}

	pkgs.Success(c, gin.H{"list": responseEntities, "total": total})
}

// AssignRole 为用户分配角色
//
//	@Summary  为用户分配角色
//	@Description  为用户分配角色，此操作会覆盖用户已有的所有角色
//	@Tags   user
//	@Accept   json
//	@Produce  json
//	@Param    id    path  string          true  "用户ID"
//	@Param    request body  AssignRoleReq true  "分配角色请求参数"
//	@Success  200   {object}  pkgs.Response       "分配成功"
//	@Failure  400   {object}  pkgs.Response       "请求参数错误"
//	@Failure  500   {object}  pkgs.Response       "服务器内部错误"
//	@Router   /user/{id}/role [post]
func (h *Handler) AssignRole(c *gin.Context) {
	// 获取用户ID
	userID := c.Param("id")

	// 绑定请求参数
	var req AssignRolesToUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证请求参数
	if err := h.validator.Validate(c, &req); err != nil {
		return
	}

	// 开启事务
	tx, err := h.db.BeginTxx(c.Request.Context(), nil)
	if err != nil {
		h.logger.Error("Failed to begin transaction for assigning roles", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign roles")
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
				h.logger.Error("Failed to commit transaction for assigning roles", zap.Error(err))
			}
		}
	}()

	// 删除用户已有角色
	_, err = tx.ExecContext(c.Request.Context(), `DELETE FROM user_role WHERE user_id = $1`, userID)
	if err != nil {
		h.logger.Error("Failed to delete existing roles for user", zap.String("userID", userID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign roles")
		return
	}

	// 如果没有需要分配的角色，直接返回
	if len(req.RoleIDs) == 0 {
		pkgs.Success(c, nil)
		return
	}

	// 分配新角色
	var userRoles []map[string]interface{}
	for _, roleID := range req.RoleIDs {
		userRoles = append(userRoles, map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		})
	}

	_, err = tx.NamedExecContext(c.Request.Context(), `INSERT INTO user_role (user_id, role_id) VALUES (:user_id, :role_id)`, userRoles)
	if err != nil {
		h.logger.Error("Failed to insert new roles for user", zap.String("userID", userID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to assign roles")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
}
