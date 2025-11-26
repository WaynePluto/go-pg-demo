package permission

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

func (r *Repository) Create(c *gin.Context) func(*CreatePermissionReq) mo.Result[CreatePermissionRes] {
	return func(req *CreatePermissionReq) mo.Result[CreatePermissionRes] {
		// 创建实体
		entity := &PermissionEntity{
			Name:     req.Name,
			Type:     req.Type,
			Metadata: req.Metadata,
		}
		// 数据库操作
		query := `INSERT INTO iacc_permission (name, type, metadata) VALUES (:name, :type, :metadata) RETURNING id, created_at, updated_at`
		stmt, err := r.db.PrepareNamedContext(c.Request.Context(), query)
		if err != nil {
			r.logger.Error("创建权限语句准备失败", zap.Error(err))
			return mo.Err[CreatePermissionRes](pkgs.NewApiError(http.StatusInternalServerError, "创建权限失败"))
		}
		defer stmt.Close()

		err = stmt.GetContext(c.Request.Context(), entity, entity)
		if err != nil {
			r.logger.Error("创建权限失败", zap.Error(err))
			return mo.Err[CreatePermissionRes](pkgs.NewApiError(http.StatusInternalServerError, "创建权限失败"))
		}
		// 返回结果
		return mo.Ok(CreatePermissionRes(entity.ID))
	}
}

func (r *Repository) GetByID(c *gin.Context) func(*GetByIDReq) mo.Result[GetByIDRes] {
	return func(req *GetByIDReq) mo.Result[GetByIDRes] {

		// 数据库操作
		var entity PermissionEntity
		query := `SELECT id, name, type, metadata, created_at, updated_at FROM iacc_permission WHERE id = $1`
		err := r.db.GetContext(c.Request.Context(), &entity, query, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusNotFound, "权限不存在"))
			}
			r.logger.Error("获取权限失败", zap.Error(err))
			return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "获取权限失败"))
		}

		// 返回结果
		response := GetByIDRes{
			ID:        entity.ID,
			Name:      entity.Name,
			Type:      entity.Type,
			Metadata:  entity.Metadata,
			CreatedAt: entity.CreatedAt.Format(time.RFC3339),
			UpdatedAt: entity.UpdatedAt.Format(time.RFC3339),
		}
		return mo.Ok(response)
	}
}

func (r *Repository) UpdateByID(c *gin.Context) func(*UpdatePermissionReq) mo.Result[UpdatePermissionRes] {
	return func(req *UpdatePermissionReq) mo.Result[UpdatePermissionRes] {
		// 动态构建更新语句
		params := map[string]any{"id": req.ID}
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
			return mo.Ok(UpdatePermissionRes(0))
		}

		query := "UPDATE iacc_permission SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

		// 执行数据库操作
		res, err := r.db.NamedExecContext(c.Request.Context(), query, params)
		if err != nil {
			r.logger.Error("更新权限失败", zap.Error(err))
			return mo.Err[UpdatePermissionRes](pkgs.NewApiError(http.StatusInternalServerError, "更新权限失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[UpdatePermissionRes](pkgs.NewApiError(http.StatusInternalServerError, "更新权限失败"))
		}
		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) DeleteByID(c *gin.Context) func(*DeleteByIDReq) mo.Result[DeleteByIDRes] {
	return func(req *DeleteByIDReq) mo.Result[DeleteByIDRes] {
		// 数据库操作
		query := `DELETE FROM iacc_permission WHERE id = $1`
		res, err := r.db.ExecContext(c.Request.Context(), query, req.ID)
		if err != nil {
			r.logger.Error("删除权限失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除权限失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除权限失败"))
		}

		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) QueryList(c *gin.Context) func(*QueryListReq) mo.Result[QueryListRes] {
	return func(req *QueryListReq) mo.Result[QueryListRes] {
		// 校验排序字段
		validOrderBy := map[string]bool{
			"id":         true,
			"name":       true,
			"type":       true,
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
		if req.Name != "" {
			whereClauses = append(whereClauses, "name ILIKE :name")
			params["name"] = "%" + req.Name + "%"
		}
		if req.Type != "" {
			whereClauses = append(whereClauses, "type = :type")
			params["type"] = req.Type
		}

		whereCondition := ""
		if len(whereClauses) > 0 {
			whereCondition = " WHERE " + strings.Join(whereClauses, " AND ")
		}

		// 查询总数
		var total int64
		countQuery := "SELECT count(*) FROM iacc_permission" + whereCondition
		// 使用 NamedQuery 而不是 PrepareNamed
		rows, err := r.db.NamedQueryContext(c.Request.Context(), countQuery, params)
		if err != nil {
			r.logger.Error("准备命名计数查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询权限列表失败"))
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&total)
		}
		if err != nil {
			r.logger.Error("统计权限数量失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询权限列表失败"))
		}

		if total == 0 {
			return mo.Ok(QueryListRes{
				List:  []PermissionItem{},
				Total: 0,
			})
		}

		// 查询列表
		var entities []PermissionEntity
		listQuery := `SELECT id, name, type, metadata, created_at, updated_at FROM iacc_permission` + whereCondition + ` ORDER BY ` + req.OrderBy + ` ` + upperOrder + ` LIMIT :limit OFFSET :offset`
		// 使用 NamedQuery 而不是 PrepareNamed
		rows, err = r.db.NamedQueryContext(c.Request.Context(), listQuery, params)
		if err != nil {
			r.logger.Error("准备命名列表查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询权限列表失败"))
		}
		defer rows.Close()

		for rows.Next() {
			var entity PermissionEntity
			err = rows.StructScan(&entity)
			if err != nil {
				r.logger.Error("扫描行数据失败", zap.Error(err))
				return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询权限列表失败"))
			}
			entities = append(entities, entity)
		}
		if err != nil {
			r.logger.Error("查询权限列表失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询权限列表失败"))
		}

		// 转换并返回结果
		var responseEntities []PermissionItem
		for _, entity := range entities {
			responseEntities = append(responseEntities, PermissionItem{
				ID:        entity.ID,
				Name:      entity.Name,
				Type:      entity.Type,
				Metadata:  entity.Metadata,
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
