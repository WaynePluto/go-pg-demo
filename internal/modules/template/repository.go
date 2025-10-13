package template

import (
	"context"
)

// ITemplateRepo 定义模板数据持久化接口
type ITemplateRepo interface {
	// Create 创建模板
	Create(ctx context.Context, template *TemplateEntity) error
	
	// GetByID 根据ID获取模板
	GetByID(ctx context.Context, id string) (*TemplateEntity, error)
	
	// Update 更新模板
	Update(ctx context.Context, template *TemplateEntity) error
	
	// Delete 删除模板
	Delete(ctx context.Context, id string) error
	
	// List 分页查询模板列表
	List(ctx context.Context, page, size int) ([]*TemplateEntity, int64, error)
}