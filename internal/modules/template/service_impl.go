package template

import (
	"context"
)

// TemplateServiceImpl 模板服务实现
type TemplateServiceImpl struct {
	repo ITemplateRepo
}

// NewTemplateService 创建模板服务实例
func NewTemplateService(repo ITemplateRepo) *TemplateServiceImpl {
	return &TemplateServiceImpl{
		repo: repo,
	}
}

// Create 创建模板
func (s *TemplateServiceImpl) Create(ctx context.Context, template *TemplateEntity) (*TemplateEntity, error) {
	err := s.repo.Create(ctx, template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// GetByID 根据ID获取模板
func (s *TemplateServiceImpl) GetByID(ctx context.Context, id string) (*TemplateEntity, error) {
	return s.repo.GetByID(ctx, id)
}

// Update 更新模板
func (s *TemplateServiceImpl) Update(ctx context.Context, template *TemplateEntity) (*TemplateEntity, error) {
	err := s.repo.Update(ctx, template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// Delete 删除模板
func (s *TemplateServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List 分页查询模板列表
func (s *TemplateServiceImpl) List(ctx context.Context, page, size int) ([]*TemplateEntity, int64, error) {
	return s.repo.List(ctx, page, size)
}