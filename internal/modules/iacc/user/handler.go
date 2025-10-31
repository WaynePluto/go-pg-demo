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
//	  /user/{id}:
//	    get: GetByID
//	    put: UpdateByID
//	    delete: DeleteByID
//	  /user/list:
//	    get: QueryList
//	  /user/{id}/role:
//	    post: AssignRole
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
//	@Summary      创建一个新的用户账户
//	@Description  通过提供用户名、手机号、密码等信息创建一个新的用户账户。成功后返回新创建用户的唯一标识符(UUID)。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        request  body      CreateUserReq              true  "创建用户所需的请求体参数"
//	@Success      200      {object}  pkgs.Response{data=string} "成功创建用户，返回用户ID"
//	@Failure      400      {object}  pkgs.Response              "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response              "服务器内部错误，无法创建用户"
//	@Router       /user [post]
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
		Username: req.Username,
		Phone:    req.Phone,
		Password: req.Password,
		Profile:  req.Profile,
	}

	// 数据库操作
	query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES (:username, :phone, :password, :profile) RETURNING id, created_at, updated_at`
	stmt, err := h.db.PrepareNamedContext(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("创建用户语句准备失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "创建用户失败")
		return
	}
	defer stmt.Close()

	err = stmt.GetContext(c.Request.Context(), entity, entity)
	if err != nil {
		h.logger.Error("创建用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

// GetByID 根据ID获取用户
//
//	@Summary      根据用户ID获取用户详情
//	@Description  通过指定的用户唯一标识符(UUID)来检索特定用户的详细信息。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string                     true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Success      200  {object}  pkgs.Response{data=UserRes} "成功获取用户信息"
//	@Failure      400  {object}  pkgs.Response              "提供的用户ID格式无效"
//	@Failure      404  {object}  pkgs.Response              "未找到指定ID的用户"
//	@Failure      500  {object}  pkgs.Response              "服务器内部错误，无法获取用户信息"
//	@Router       /user/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "用户ID格式无效")
		return
	}

	// 数据库操作
	var entity UserEntity
	query := `SELECT id, username, phone, profile, created_at, updated_at FROM "iacc_user" WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			pkgs.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		h.logger.Error("获取用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "获取用户失败")
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
//	@Summary      根据用户ID更新用户信息
//	@Description  通过指定的用户唯一标识符(UUID)来更新特定用户的密码和个人信息。只会更新请求中包含的字段。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id       path      string          true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Param        request  body      UpdateUserReq   true  "需要更新的用户信息"
//	@Success      200      {object}  pkgs.Response   "成功更新用户信息"
//	@Failure      400      {object}  pkgs.Response   "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response   "服务器内部错误，无法更新用户信息"
//	@Router       /user/{id} [put]
func (h *Handler) UpdateByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "用户ID格式无效")
		return
	}

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

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	query := `UPDATE "iacc_user" SET ` + strings.Join(setClauses, ", ") + " WHERE id = :id"

	// 执行数据库操作
	result, err := h.db.NamedExecContext(c.Request.Context(), query, params)
	if err != nil {
		h.logger.Error("更新用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "更新用户失败")
		return
	}
	
	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		pkgs.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
}

// DeleteByID 根据ID删除用户
//
//	@Summary      根据用户ID删除用户
//	@Description  通过指定的用户唯一标识符(UUID)来删除特定用户。这是一个永久性操作，请谨慎使用。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id   path      string                    true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Success      200  {object}  pkgs.Response{data=int64} "成功删除用户，返回受影响的行数"
//	@Failure      400  {object}  pkgs.Response             "提供的用户ID格式无效"
//	@Failure      500  {object}  pkgs.Response             "服务器内部错误，无法删除用户"
//	@Router       /user/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(id); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "用户ID格式无效")
		return
	}

	// 数据库操作
	query := `DELETE FROM "iacc_user" WHERE id = :id`
	res, err := h.db.NamedExecContext(c.Request.Context(), query, map[string]interface{}{"id": id})
	if err != nil {
		h.logger.Error("删除用户失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "删除用户失败")
		return
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		h.logger.Error("获取用户删除影响行数失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "删除用户失败")
		return
	}

	// 返回结果
	pkgs.Success(c, affectedRows)
}

// QueryList 获取用户列表
//
//	@Summary      获取用户列表（支持分页和筛选）
//	@Description  获取系统中的用户列表，支持按手机号模糊搜索，并提供分页功能。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        page      query     int                        false  "页码，从1开始计算"  minimum(1)  default(1)
//	@Param        pageSize  query     int                        false  "每页条目数"        minimum(1)  maximum(100)  default(10)
//	@Param        phone     query     string                     false  "手机号模糊搜索关键字"
//	@Success      200       {object}  pkgs.Response{data=UserListRes}  "成功获取用户列表"
//	@Failure      400       {object}  pkgs.Response                  "请求参数验证失败或格式不正确"
//	@Failure      500       {object}  pkgs.Response                  "服务器内部错误，无法获取用户列表"
//	@Router       /user/list [get]
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

	baseQuery := `FROM "iacc_user" WHERE 1=1`
	params := make(map[string]interface{})

	if req.Phone != "" {
		baseQuery += " AND phone ILIKE :phone"
		params["phone"] = "%" + req.Phone + "%"
	}

	// 查询总数
	countQuery := "SELECT count(*) " + baseQuery
	nstmt, err := h.db.PrepareNamedContext(c.Request.Context(), countQuery)
	if err != nil {
		h.logger.Error("准备用户命名计数查询失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}
	defer nstmt.Close()
	err = nstmt.GetContext(c.Request.Context(), &total, params)
	if err != nil {
		h.logger.Error("统计用户数量失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	if total == 0 {
		pkgs.Success(c, gin.H{"list": []UserEntity{}, "total": total})
		return
	}

	// 查询列表
	listQuery := `SELECT id, username, phone, profile, created_at, updated_at ` + baseQuery + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`
	params["limit"] = req.PageSize
	params["offset"] = (req.Page - 1) * req.PageSize

	nstmt, err = h.db.PrepareNamedContext(c.Request.Context(), listQuery)
	if err != nil {
		h.logger.Error("准备用户命名列表查询失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}
	defer nstmt.Close()
	err = nstmt.SelectContext(c.Request.Context(), &entities, params)
	if err != nil {
		h.logger.Error("查询用户列表失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "查询用户失败")
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
//	@Summary      为用户分配角色
//	@Description  为指定用户分配一个或多个角色。该操作会完全替换用户当前的所有角色关系。
//	@Tags         用户管理
//	@Accept       json
//	@Produce      json
//	@Param        id       path      string                 true  "用户唯一标识符(UUID)"  Format(UUID)
//	@Param        request  body      AssignRolesToUserReq   true  "要分配给用户的角色ID列表"
//	@Success      200      {object}  pkgs.Response          "成功为用户分配角色"
//	@Failure      400      {object}  pkgs.Response          "请求参数验证失败或格式不正确"
//	@Failure      500      {object}  pkgs.Response          "服务器内部错误，无法为用户分配角色"
//	@Router       /user/{id}/role [post]
func (h *Handler) AssignRole(c *gin.Context) {
	// 获取用户ID
	userID := c.Param("id")

	// 验证ID是否为有效的UUID
	if _, err := uuid.Parse(userID); err != nil {
		pkgs.Error(c, http.StatusBadRequest, "用户ID格式无效")
		return
	}

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
		h.logger.Error("为分配角色开启事务失败", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "分配角色失败")
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
				h.logger.Error("提交分配角色事务失败", zap.Error(err))
			}
		}
	}()

	// 删除用户已有角色
	_, err = tx.ExecContext(c.Request.Context(), `DELETE FROM "iacc_user_role" WHERE user_id = $1`, userID)
	if err != nil {
		h.logger.Error("删除用户已有角色失败", zap.String("userID", userID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "分配角色失败")
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

	_, err = tx.NamedExecContext(c.Request.Context(), `INSERT INTO "iacc_user_role" (user_id, role_id) VALUES (:user_id, :role_id)`, userRoles)
	if err != nil {
		h.logger.Error("为用户插入新角色失败", zap.String("userID", userID), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "分配角色失败")
		return
	}

	// 返回结果
	pkgs.Success(c, nil)
}