package template

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTemplateRepo 模拟模板仓库接口
type MockTemplateRepo struct {
	mock.Mock
}

func (m *MockTemplateRepo) Create(ctx context.Context, template *TemplateEntity) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockTemplateRepo) GetByID(ctx context.Context, id string) (*TemplateEntity, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*TemplateEntity), args.Error(1)
}

func (m *MockTemplateRepo) Update(ctx context.Context, template *TemplateEntity) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockTemplateRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTemplateRepo) List(ctx context.Context, page, size int) ([]*TemplateEntity, int64, error) {
	args := m.Called(ctx, page, size)
	result := args.Get(0)
	if result == nil {
		return nil, 0, args.Error(2)
	}
	return result.([]*TemplateEntity), args.Get(1).(int64), args.Error(2)
}

func TestTemplateService_Create(t *testing.T) {
	mockRepo := new(MockTemplateRepo)
	service := NewTemplateService(mockRepo)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}

	mockRepo.On("Create", mock.Anything, template).Return(nil)

	result, err := service.Create(context.Background(), template)

	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
}

func TestTemplateService_GetByID(t *testing.T) {
	mockRepo := new(MockTemplateRepo)
	service := NewTemplateService(mockRepo)

	now := time.Now()
	expected := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}

	mockRepo.On("GetByID", mock.Anything, "test-id").Return(expected, nil)

	result, err := service.GetByID(context.Background(), "test-id")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestTemplateService_Update(t *testing.T) {
	mockRepo := new(MockTemplateRepo)
	service := NewTemplateService(mockRepo)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Updated Template",
		Num:       intPtr(20),
	}

	mockRepo.On("Update", mock.Anything, template).Return(nil)

	result, err := service.Update(context.Background(), template)

	assert.NoError(t, err)
	assert.Equal(t, template, result)
	mockRepo.AssertExpectations(t)
}

func TestTemplateService_Delete(t *testing.T) {
	mockRepo := new(MockTemplateRepo)
	service := NewTemplateService(mockRepo)

	id := uuid.New().String()

	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := service.Delete(context.Background(), id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTemplateService_List(t *testing.T) {
	mockRepo := new(MockTemplateRepo)
	service := NewTemplateService(mockRepo)

	now := time.Now()
	expected := []*TemplateEntity{
		{
			ID:        "test-id-1",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "Test Template 1",
			Num:       intPtr(10),
		},
		{
			ID:        "test-id-2",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "Test Template 2",
			Num:       intPtr(20),
		},
	}
	total := int64(2)

	mockRepo.On("List", mock.Anything, 1, 10).Return(expected, total, nil)

	list, totalCount, err := service.List(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, expected, list)
	assert.Equal(t, total, totalCount)
	mockRepo.AssertExpectations(t)
}
