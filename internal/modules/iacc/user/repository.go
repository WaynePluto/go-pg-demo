package user

import (
	"database/sql"
	"go-pg-demo/pkgs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/samber/mo"
	"go.uber.org/zap"
)

type Repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func (r *Repository) Create(c *gin.Context) func(*CreateReq) mo.Result[CreateRes] {
	return func(req *CreateReq) mo.Result[CreateRes] {
		// 创建实体
		entity := &UserEntity{
			Username: req.Username,
			Phone:    req.Phone,
			Password: req.Password,
			Profile:  req.Profile,
		}
		// 数据库操作
		query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES (:username, :phone, :password, :profile) RETURNING id, created_at, updated_at`
		stmt, err := r.db.PrepareNamedContext(c.Request.Context(), query)
		if err != nil {
			r.logger.Error("创建用户语句准备失败", zap.Error(err))
			return mo.Err[CreateRes](pkgs.NewApiError(http.StatusInternalServerError, "创建用户失败"))
		}
		defer stmt.Close()

		err = stmt.GetContext(c.Request.Context(), entity, entity)
		if err != nil {
			r.logger.Error("创建用户失败", zap.Error(err))
			return mo.Err[CreateRes](pkgs.NewApiError(http.StatusInternalServerError, "创建用户失败"))
		}
		// 返回结果
		return mo.Ok(CreateRes(entity.ID))
	}
}

func (r *Repository) BatchCreate(c *gin.Context) func(*BatchCreateReq) mo.Result[BatchCreateRes] {
	return func(req *BatchCreateReq) mo.Result[BatchCreateRes] {
		// 准备批量插入的实体
		var entities []UserEntity
		for _, u := range req.Users {
			entities = append(entities, UserEntity{
				Username: u.Username,
				Phone:    u.Phone,
				Password: u.Password,
				Profile:  u.Profile,
			})
		}

		// 开启事务
		tx, err := r.db.BeginTxx(c.Request.Context(), nil)
		if err != nil {
			r.logger.Error("开启事务失败", zap.Error(err))
			return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建用户失败"))
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
					r.logger.Error("Failed to commit transaction", zap.Error(err))
					return
				}
			}
		}()

		// 数据库操作
		query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES (:username, :phone, :password, :profile) RETURNING id`
		stmt, err := tx.PrepareNamedContext(c.Request.Context(), query)
		if err != nil {
			r.logger.Error("准备命名语句失败", zap.Error(err))
			return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建用户失败"))
		}
		defer stmt.Close()

		var createdIDs []string
		for _, entity := range entities {
			var id string
			err = stmt.GetContext(c.Request.Context(), &id, entity)
			if err != nil {
				r.logger.Error("批量创建用户失败", zap.Error(err))
				return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建用户失败"))
			}
			createdIDs = append(createdIDs, id)
		}

		return mo.Ok(BatchCreateRes(createdIDs))
	}
}

func (r *Repository) GetByID(c *gin.Context) func(*GetByIDReq) mo.Result[GetByIDRes] {
	return func(req *GetByIDReq) mo.Result[GetByIDRes] {

		// 数据库操作
		var entity UserEntity
		query := `SELECT id, username, phone, profile, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err := r.db.GetContext(c.Request.Context(), &entity, query, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusNotFound, "用户不存在"))
			}
			r.logger.Error("获取用户失败", zap.Error(err))
			return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "获取用户失败"))
		}

		// 返回结果
		response := GetByIDRes{
			ID:        entity.ID,
			Username:  entity.Username,
			Phone:     entity.Phone,
			Profile:   entity.Profile,
			CreatedAt: entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
		}
		return mo.Ok(response)
	}
}

func (r *Repository) UpdateByID(c *gin.Context) func(*UpdateByIDReq) mo.Result[UpdateByIDRes] {
	return func(req *UpdateByIDReq) mo.Result[UpdateByIDRes] {
		// 动态构建更新语句
		params := map[string]any{"id": req.ID}
		var setClauses []string

		if req.Username != nil {
			params["username"] = *req.Username
			setClauses = append(setClauses, "username = :username")
		}
		if req.Phone != nil {
			params["phone"] = *req.Phone
			setClauses = append(setClauses, "phone = :phone")
		}
		if req.Password != nil {
			params["password"] = *req.Password
			setClauses = append(setClauses, "password = :password")
		}
		if req.Profile != nil {
			params["profile"] = *req.Profile
			setClauses = append(setClauses, "profile = :profile")
		}

		// 如果没有需要更新的字段，直接返回成功
		if len(setClauses) == 0 {
			return mo.Ok(UpdateByIDRes(0))
		}

		setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
		query := "UPDATE \"iacc_user\" SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

		// 执行数据库操作
		res, err := r.db.NamedExecContext(c.Request.Context(), query, params)
		if err != nil {
			r.logger.Error("更新用户失败", zap.Error(err))
			return mo.Err[UpdateByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "更新用户失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[UpdateByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "更新用户失败"))
		}
		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) DeleteByID(c *gin.Context) func(*DeleteByIDReq) mo.Result[DeleteByIDRes] {
	return func(req *DeleteByIDReq) mo.Result[DeleteByIDRes] {
		// 数据库操作
		query := `DELETE FROM "iacc_user" WHERE id = $1`
		res, err := r.db.ExecContext(c.Request.Context(), query, req.ID)
		if err != nil {
			r.logger.Error("删除用户失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除用户失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除用户失败"))
		}

		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) BatchDelete(c *gin.Context) func(*DeleteUsersReq) mo.Result[BatchDeleteRes] {
	return func(req *DeleteUsersReq) mo.Result[BatchDeleteRes] {
		query, args, err := sqlx.In(`DELETE FROM "iacc_user" WHERE id IN (?)`, req.IDs)
		if err != nil {
			r.logger.Error("构建批量删除查询失败", zap.Error(err))
			return mo.Err[BatchDeleteRes](pkgs.NewApiError(http.StatusInternalServerError, "构建批量删除查询失败"))
		}

		query = r.db.Rebind(query)
		res, err := r.db.ExecContext(c.Request.Context(), query, args...)
		if err != nil {
			r.logger.Error("批量删除用户失败", zap.Error(err))
			return mo.Err[BatchDeleteRes](pkgs.NewApiError(http.StatusInternalServerError, "批量删除用户失败"))
		}

		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[BatchDeleteRes](pkgs.NewApiError(http.StatusInternalServerError, "获取影响行数失败"))
		}

		return mo.Ok(affectedRows)
	}
}

func (r *Repository) QueryList(c *gin.Context) func(*QueryListReq) mo.Result[QueryListRes] {
	return func(req *QueryListReq) mo.Result[QueryListRes] {
		// 校验排序字段
		validOrderBy := map[string]bool{
			"id":         true,
			"username":   true,
			"phone":      true,
			"created_at": true,
			"updated_at": true,
		}
		if !validOrderBy[req.OrderBy] {
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusBadRequest, "排序字段不存在"))
		}

		// 校验排序顺序
		upperOrder := strings.ToUpper(req.Order)
		if upperOrder != "ASC" && upperOrder != "DESC" {
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusBadRequest, "排序顺序参数错误"))
		}

		// 构建查询
		params := map[string]any{
			"limit":  req.PageSize,
			"offset": (req.Page - 1) * req.PageSize,
		}

		var whereClauses []string
		if req.Phone != "" {
			whereClauses = append(whereClauses, "phone ILIKE :phone")
			params["phone"] = "%" + req.Phone + "%"
		}
		if req.Username != "" {
			whereClauses = append(whereClauses, "username ILIKE :username")
			params["username"] = "%" + req.Username + "%"
		}

		whereCondition := ""
		if len(whereClauses) > 0 {
			whereCondition = " WHERE " + strings.Join(whereClauses, " AND ")
		}

		// 查询总数
		var total int64
		countQuery := "SELECT count(*) FROM \"iacc_user\"" + whereCondition
		// 使用 NamedQuery 而不是 PrepareNamed
		rows, err := r.db.NamedQueryContext(c.Request.Context(), countQuery, params)
		if err != nil {
			r.logger.Error("准备命名计数查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户列表失败"))
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&total)
		}
		if err != nil {
			r.logger.Error("统计用户数量失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户列表失败"))
		}

		if total == 0 {
			return mo.Ok(QueryListRes{
				List:  []UserItem{},
				Total: 0,
			})
		}

		// 查询列表
		var entities []UserEntity
		listQuery := `SELECT id, username, phone, profile, created_at, updated_at FROM "iacc_user"` + whereCondition + ` ORDER BY ` + req.OrderBy + ` ` + upperOrder + ` LIMIT :limit OFFSET :offset`
		// 使用 NamedQuery 而不是 PrepareNamed
		rows, err = r.db.NamedQueryContext(c.Request.Context(), listQuery, params)
		if err != nil {
			r.logger.Error("准备命名列表查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户列表失败"))
		}
		defer rows.Close()

		for rows.Next() {
			var entity UserEntity
			err = rows.StructScan(&entity)
			if err != nil {
				r.logger.Error("扫描行数据失败", zap.Error(err))
				return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户列表失败"))
			}
			entities = append(entities, entity)
		}
		if err != nil {
			r.logger.Error("查询用户列表失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户列表失败"))
		}

		// 转换并返回结果
		var responseEntities []UserItem
		for _, entity := range entities {
			responseEntities = append(responseEntities, UserItem{
				ID:        entity.ID,
				Username:  entity.Username,
				Phone:     entity.Phone,
				Profile:   entity.Profile,
				CreatedAt: entity.CreatedAt.Format(time.RFC3339),
				UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
			})
		}

		return mo.Ok(QueryListRes{
			List:  responseEntities,
			Total: total,
		})
	}
}

func (r *Repository) AssignRoles(c *gin.Context) func(*AssignRolesReq) mo.Result[AssignRolesRes] {
	return func(req *AssignRolesReq) mo.Result[AssignRolesRes] {
		// 开启事务
		tx, err := r.db.BeginTxx(c.Request.Context(), nil)
		if err != nil {
			r.logger.Error("为分配角色开启事务失败", zap.Error(err))
			return mo.Err[AssignRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "分配角色失败"))
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
					r.logger.Error("提交分配角色事务失败", zap.Error(err))
				}
			}
		}()

		// 删除用户已有角色
		_, err = tx.ExecContext(c.Request.Context(), `DELETE FROM "iacc_user_role" WHERE user_id = $1`, req.ID)
		if err != nil {
			r.logger.Error("删除用户已有角色失败", zap.String("userID", req.ID), zap.Error(err))
			return mo.Err[AssignRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "分配角色失败"))
		}

		// 如果没有需要分配的角色，直接返回
		if len(req.RoleIDs) == 0 {
			return mo.Ok(AssignRolesRes(0))
		}

		// 分配新角色
		var userRoles []map[string]interface{}
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, map[string]interface{}{
				"user_id": req.ID,
				"role_id": roleID,
			})
		}

		_, err = tx.NamedExecContext(c.Request.Context(), `INSERT INTO "iacc_user_role" (user_id, role_id) VALUES (:user_id, :role_id)`, userRoles)
		if err != nil {
			r.logger.Error("为用户插入新角色失败", zap.String("userID", req.ID), zap.Error(err))
			return mo.Err[AssignRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "分配角色失败"))
		}

		// 返回结果
		return mo.Ok(AssignRolesRes(int64(len(req.RoleIDs))))
	}
}

func (r *Repository) GetRoles(c *gin.Context) func(*GetRolesReq) mo.Result[GetRolesRes] {
	return func(req *GetRolesReq) mo.Result[GetRolesRes] {
		// 查询总数
		var total int64
		countQuery := `
			SELECT COUNT(*)
			FROM "iacc_user_role" ur
			JOIN "iacc_role" r ON ur.role_id = r.id
			WHERE ur.user_id = $1
		`
		err := r.db.GetContext(c.Request.Context(), &total, countQuery, req.ID)
		if err != nil {
			r.logger.Error("统计用户角色数量失败", zap.Error(err))
			return mo.Err[GetRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户角色列表失败"))
		}

		// 如果没有角色，返回空列表
		if total == 0 {
			return mo.Ok(GetRolesRes{
				List:  []RoleItem{},
				Total: 0,
			})
		}

		// 查询角色列表
		var roles []RoleItem
		listQuery := `
			SELECT r.id, r.name, r.description, r.created_at, r.updated_at
			FROM "iacc_user_role" ur
			JOIN "iacc_role" r ON ur.role_id = r.id
			WHERE ur.user_id = $1
			ORDER BY r.created_at DESC
		`
		rows, err := r.db.QueryxContext(c.Request.Context(), listQuery, req.ID)
		if err != nil {
			r.logger.Error("查询用户角色列表失败", zap.Error(err))
			return mo.Err[GetRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户角色列表失败"))
		}
		defer rows.Close()

		for rows.Next() {
			var role struct {
				ID          string    `db:"id"`
				Name        string    `db:"name"`
				Description *string   `db:"description"`
				CreatedAt   time.Time `db:"created_at"`
				UpdatedAt   time.Time `db:"updated_at"`
			}
			err = rows.StructScan(&role)
			if err != nil {
				r.logger.Error("扫描角色数据失败", zap.Error(err))
				return mo.Err[GetRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户角色列表失败"))
			}
			roles = append(roles, RoleItem{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				CreatedAt:   role.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   role.UpdatedAt.Format(time.RFC3339),
			})
		}

		if err = rows.Err(); err != nil {
			r.logger.Error("查询用户角色列表失败", zap.Error(err))
			return mo.Err[GetRolesRes](pkgs.NewApiError(http.StatusInternalServerError, "查询用户角色列表失败"))
		}

		// 返回结果
		return mo.Ok(GetRolesRes{
			List:  roles,
			Total: total,
		})
	}
}
