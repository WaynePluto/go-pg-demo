package template

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TemplateRepoImplPostgre 模板仓库PostgreSQL实现
type TemplateRepoImplPostgre struct {
	db *sqlx.DB
}

// NewTemplateRepo 创建模板仓库实例
func NewTemplateRepo(db *sqlx.DB) *TemplateRepoImplPostgre {
	return &TemplateRepoImplPostgre{
		db: db,
	}
}

// Create 创建模板
func (r *TemplateRepoImplPostgre) Create(ctx context.Context, template *TemplateEntity) error {
	query := `INSERT INTO template (name, num) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query, template.Name, template.Num)
	err := row.Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetByID 根据ID获取模板
func (r *TemplateRepoImplPostgre) GetByID(ctx context.Context, id string) (*TemplateEntity, error) {
	var template TemplateEntity
	query := `SELECT id, created_at, updated_at, name, num FROM template WHERE id = $1`
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get template by id: %w", err)
	}
	return &template, nil
}

// Update 更新模板
func (r *TemplateRepoImplPostgre) Update(ctx context.Context, template *TemplateEntity) error {
	query := `UPDATE template SET updated_at = :updated_at, name = :name, num = :num WHERE id = :id`
	result, err := r.db.NamedExecContext(ctx, query, template)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %w", sql.ErrNoRows)
	}

	return nil
}

// Delete 删除模板
func (r *TemplateRepoImplPostgre) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM template WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		// 检查是否是外键约束错误
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return fmt.Errorf("cannot delete template due to foreign key constraint: %w", err)
		}
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %w", sql.ErrNoRows)
	}

	return nil
}

// List 分页查询模板列表
func (r *TemplateRepoImplPostgre) List(ctx context.Context, page, size int) ([]*TemplateEntity, int64, error) {
	// 计算偏移量
	offset := (page - 1) * size

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM template`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count template: %w", err)
	}

	// 查询列表
	templates := make([]*TemplateEntity, 0)
	listQuery := `SELECT id, created_at, updated_at, name, num FROM template ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err = r.db.SelectContext(ctx, &templates, listQuery, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list template: %w", err)
	}

	return templates, total, nil
}
