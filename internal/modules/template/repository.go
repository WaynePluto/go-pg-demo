package template

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
		entity := &TemplateEntity{
			Name: req.Name,
			Num:  req.Num,
		}
		// 数据库操作
		query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id, created_at, updated_at`
		stmt, err := r.db.PrepareNamedContext(c.Request.Context(), query)
		if err != nil {
			r.logger.Error("创建模板语句准备失败", zap.Error(err))
			return mo.Err[CreateRes](pkgs.NewApiError(http.StatusInternalServerError, "创建模板失败"))
		}
		defer stmt.Close()

		err = stmt.GetContext(c.Request.Context(), entity, entity)
		if err != nil {
			r.logger.Error("创建模板失败", zap.Error(err))
			return mo.Err[CreateRes](pkgs.NewApiError(http.StatusInternalServerError, "创建模板失败"))
		}
		// 返回结果
		return mo.Ok(CreateRes(entity.ID))
	}
}

func (r *Repository) BatchCreate(c *gin.Context) func(*BatchCreateReq) mo.Result[BatchCreateRes] {
	return func(req *BatchCreateReq) mo.Result[BatchCreateRes] {
		// 准备批量插入的实体
		var entities []TemplateEntity
		for _, t := range req.Templates {
			entities = append(entities, TemplateEntity{
				Name: t.Name,
				Num:  t.Num,
			})
		}

		// 开启事务
		tx, err := r.db.BeginTxx(c.Request.Context(), nil)
		if err != nil {
			r.logger.Error("开启事务失败", zap.Error(err))
			return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建模板失败"))
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
		query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id`
		stmt, err := tx.PrepareNamedContext(c.Request.Context(), query)
		if err != nil {
			r.logger.Error("准备命名语句失败", zap.Error(err))
			return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建模板失败"))
		}
		defer stmt.Close()

		var createdIDs []string
		for _, entity := range entities {
			var id string
			err = stmt.GetContext(c.Request.Context(), &id, entity)
			if err != nil {
				r.logger.Error("批量创建模板失败", zap.Error(err))
				return mo.Err[BatchCreateRes](pkgs.NewApiError(http.StatusInternalServerError, "批量创建模板失败"))
			}
			createdIDs = append(createdIDs, id)
		}

		return mo.Ok(BatchCreateRes(createdIDs))
	}
}

func (r *Repository) GetByID(c *gin.Context) func(*GetByIDReq) mo.Result[GetByIDRes] {
	return func(req *GetByIDReq) mo.Result[GetByIDRes] {

		// 数据库操作
		var entity TemplateEntity
		query := `SELECT id, name, num, created_at, updated_at FROM template WHERE id = $1`
		err := r.db.GetContext(c.Request.Context(), &entity, query, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusNotFound, "模板不存在"))
			}
			r.logger.Error("获取模板失败", zap.Error(err))
			return mo.Err[GetByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "获取模板失败"))
		}

		// 返回结果
		response := GetByIDRes{
			ID:        entity.ID,
			Name:      entity.Name,
			Num:       entity.Num,
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
			return mo.Ok(UpdateByIDRes(0))
		}

		query := "UPDATE template SET " + strings.Join(setClauses, ", ") + " WHERE id = :id"

		// 执行数据库操作
		res, err := r.db.NamedExecContext(c.Request.Context(), query, params)
		if err != nil {
			r.logger.Error("更新模板失败", zap.Error(err))
			return mo.Err[UpdateByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "更新模板失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[UpdateByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "更新模板失败"))
		}
		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) DeleteByID(c *gin.Context) func(*DeleteByIDReq) mo.Result[DeleteByIDRes] {
	return func(req *DeleteByIDReq) mo.Result[DeleteByIDRes] {
		// 数据库操作
		query := `DELETE FROM template WHERE id = $1`
		res, err := r.db.ExecContext(c.Request.Context(), query, req.ID)
		if err != nil {
			r.logger.Error("删除模板失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除模板失败"))
		}
		affectedRows, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("获取影响行数失败", zap.Error(err))
			return mo.Err[DeleteByIDRes](pkgs.NewApiError(http.StatusInternalServerError, "删除模板失败"))
		}

		// 返回结果
		return mo.Ok(affectedRows)
	}
}

func (r *Repository) BatchDelete(c *gin.Context) func(*DeleteTemplatesReq) mo.Result[BatchDeleteRes] {
	return func(req *DeleteTemplatesReq) mo.Result[BatchDeleteRes] {
		query, args, err := sqlx.In(`DELETE FROM template WHERE id IN (?)`, req.IDs)
		if err != nil {
			r.logger.Error("构建批量删除查询失败", zap.Error(err))
			return mo.Err[BatchDeleteRes](pkgs.NewApiError(http.StatusInternalServerError, "构建批量删除查询失败"))
		}

		query = r.db.Rebind(query)
		res, err := r.db.ExecContext(c.Request.Context(), query, args...)
		if err != nil {
			r.logger.Error("批量删除模板失败", zap.Error(err))
			return mo.Err[BatchDeleteRes](pkgs.NewApiError(http.StatusInternalServerError, "批量删除模板失败"))
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
			"name":       true,
			"num":        true,
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

		whereCondition := ""
		if len(whereClauses) > 0 {
			whereCondition = " WHERE " + strings.Join(whereClauses, " AND ")
		}

		// 查询总数
		var total int64
		countQuery := "SELECT count(*) FROM template" + whereCondition
		// 使用 NamedExec 而不是 PrepareNamed
		rows, err := r.db.NamedQueryContext(c.Request.Context(), countQuery, params)
		if err != nil {
			r.logger.Error("准备命名计数查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询模板列表失败"))
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&total)
		}
		if err != nil {
			r.logger.Error("统计模板数量失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询模板列表失败"))
		}

		if total == 0 {
			return mo.Ok(QueryListRes{
				List:  []TemplateItem{},
				Total: 0,
			})
		}

		// 查询列表
		var entities []TemplateEntity
		listQuery := `SELECT id, name, num, created_at, updated_at FROM template` + whereCondition + ` ORDER BY ` + req.OrderBy + ` ` + upperOrder + ` LIMIT :limit OFFSET :offset`
		// 使用 NamedQuery 而不是 PrepareNamed
		rows, err = r.db.NamedQueryContext(c.Request.Context(), listQuery, params)
		if err != nil {
			r.logger.Error("准备命名列表查询失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询模板列表失败"))
		}
		defer rows.Close()

		for rows.Next() {
			var entity TemplateEntity
			err = rows.StructScan(&entity)
			if err != nil {
				r.logger.Error("扫描行数据失败", zap.Error(err))
				return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询模板列表失败"))
			}
			entities = append(entities, entity)
		}
		if err != nil {
			r.logger.Error("查询模板列表失败", zap.Error(err))
			return mo.Err[QueryListRes](pkgs.NewApiError(http.StatusInternalServerError, "查询模板列表失败"))
		}

		// 转换并返回结果
		var responseEntities []TemplateItem
		for _, entity := range entities {
			responseEntities = append(responseEntities, TemplateItem{
				ID:        entity.ID,
				Name:      entity.Name,
				Num:       entity.Num,
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
