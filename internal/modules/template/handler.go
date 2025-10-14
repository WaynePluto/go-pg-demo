package template

import (
	"net/http"

	"go-pg-demo/internal/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TemplateHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewTemplateHandler(db *sqlx.DB, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		db:     db,
		logger: logger,
	}
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	// 绑定请求参数
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 创建实体
	entity := &TemplateEntity{
		Name: req.Name,
		Num:  req.Num,
	}

	// 数据库操作
	query := `INSERT INTO template (name, num) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	row := h.db.QueryRowContext(c.Request.Context(), query, entity.Name, entity.Num)
	err := row.Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create template")
		return
	}

	// 返回结果
	pkgs.Success(c, entity.ID)
}

func (h *TemplateHandler) CreateTemplateBatch(c *gin.Context) {
	// 绑定请求参数
	var req CreateTemplatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
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
	tx, err := h.db.Beginx()
	if err != nil {
		h.logger.Error("Failed to begin transaction", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
		return
	}
	defer tx.Rollback() // Rollback in case of panic

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
		err := stmt.GetContext(c.Request.Context(), &id, entity)
		if err != nil {
			h.logger.Error("Failed to create template in batch", zap.Error(err))
			pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
			return
		}
		createdIDs = append(createdIDs, id)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		h.logger.Error("Failed to commit transaction", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create templates")
		return
	}

	// 返回结果
	pkgs.Success(c, createdIDs)
}

func (h *TemplateHandler) GetTemplateByID(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 数据库操作
	var entity TemplateEntity
	query := `SELECT id, name, num, created_at, updated_at FROM template WHERE id = $1`
	err := h.db.GetContext(c.Request.Context(), &entity, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkgs.Error(c, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to get template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to get template")
		return
	}

	// 返回结果
	pkgs.Success(c, entity)
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 绑定请求参数
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 动态构建更新语句
	query := "UPDATE template SET "
	params := make(map[string]interface{})
	params["id"] = id
	if req.Name != nil {
		query += "name = :name, "
		params["name"] = *req.Name
	}
	if req.Num != nil {
		query += "num = :num, "
		params["num"] = *req.Num
	}
	query += "updated_at = NOW() WHERE id = :id"

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

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	// 获取 ID
	id := c.Param("id")

	// 数据库操作
	query := `DELETE FROM template WHERE id = $1`
	res, err := h.db.ExecContext(c.Request.Context(), query, id)
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

func (h *TemplateHandler) DeleteTemplateBatch(c *gin.Context) {
	// 绑定请求参数
	var req DeleteTemplatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
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

func (h *TemplateHandler) QueryTemplateList(c *gin.Context) {
	// 绑定请求参数
	var req QueryTemplateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
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
	countQuery, args, err := h.db.BindNamed(countQuery, params)
	if err != nil {
		h.logger.Error("Failed to bind count query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}
	err = h.db.GetContext(c.Request.Context(), &total, countQuery, args...)
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

	listQuery, args, err = h.db.BindNamed(listQuery, params)
	if err != nil {
		h.logger.Error("Failed to bind list query", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}
	err = h.db.SelectContext(c.Request.Context(), &entities, listQuery, args...)
	if err != nil {
		h.logger.Error("Failed to select templates", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to query templates")
		return
	}

	// 返回结果
	pkgs.Success(c, gin.H{"list": entities, "total": total})
}
