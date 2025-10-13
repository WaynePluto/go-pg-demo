package template

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTemplateService 是ITemplateService接口的模拟实现
type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) Create(ctx context.Context, template *TemplateEntity) (*TemplateEntity, error) {
	args := m.Called(ctx, template)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*TemplateEntity), args.Error(1)
}

func (m *MockTemplateService) GetByID(ctx context.Context, id string) (*TemplateEntity, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*TemplateEntity), args.Error(1)
}

func (m *MockTemplateService) Update(ctx context.Context, template *TemplateEntity) (*TemplateEntity, error) {
	args := m.Called(ctx, template)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*TemplateEntity), args.Error(1)
}

func (m *MockTemplateService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTemplateService) List(ctx context.Context, page, size int) ([]*TemplateEntity, int64, error) {
	args := m.Called(ctx, page, size)
	list := args.Get(0)
	if list == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return list.([]*TemplateEntity), args.Get(1).(int64), args.Error(2)
}

func TestTemplateHandler_CreateTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTemplateService)
	logger, _ := zap.NewDevelopment()
	handler := NewTemplateHandler(mockService, logger)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}

	reqBody := CreateTemplateRequest{
		Name: "Test Template",
		Num:  intPtr(10),
	}

	jsonValue, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/templates", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("Create", mock.Anything, mock.MatchedBy(func(t *TemplateEntity) bool {
		return t.Name == reqBody.Name && t.Num == reqBody.Num
	})).Return(template, nil)

	handler.CreateTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTemplateHandler_GetTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTemplateService)
	logger, _ := zap.NewDevelopment()
	handler := NewTemplateHandler(mockService, logger)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}

	req, _ := http.NewRequest(http.MethodGet, "/templates/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.AddParam("id", "test-id")

	mockService.On("GetByID", mock.Anything, "test-id").Return(template, nil)

	handler.GetTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTemplateHandler_UpdateTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTemplateService)
	logger, _ := zap.NewDevelopment()
	handler := NewTemplateHandler(mockService, logger)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Updated Template",
		Num:       intPtr(20),
	}

	reqBody := UpdateTemplateRequest{
		Name: stringPtr("Updated Template"),
		Num:  intPtr(20),
	}

	jsonValue, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/templates/test-id", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.AddParam("id", "test-id")

	// First call to GetByID
	mockService.On("GetByID", mock.Anything, "test-id").Return(&TemplateEntity{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}, nil)

	// Second call to Update
	mockService.On("Update", mock.Anything, mock.MatchedBy(func(t *TemplateEntity) bool {
		return t.Name == "Updated Template" && *t.Num == 20
	})).Return(template, nil)

	handler.UpdateTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTemplateHandler_DeleteTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTemplateService)
	logger, _ := zap.NewDevelopment()
	handler := NewTemplateHandler(mockService, logger)

	req, _ := http.NewRequest(http.MethodDelete, "/templates/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.AddParam("id", "test-id")

	mockService.On("Delete", mock.Anything, "test-id").Return(nil)

	handler.DeleteTemplate(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestTemplateHandler_ListTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTemplateService)
	logger, _ := zap.NewDevelopment()
	handler := NewTemplateHandler(mockService, logger)

	now := time.Now()
	templates := []*TemplateEntity{
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

	req, _ := http.NewRequest(http.MethodGet, "/templates?page=1&size=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	mockService.On("List", mock.Anything, 1, 10).Return(templates, total, nil)

	handler.ListTemplates(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// stringPtr 是一个辅助函数，用于创建string指针
func stringPtr(s string) *string {
	return &s
}

// intPtr 是一个辅助函数，用于创建int指针
func intPtr(i int) *int {
	return &i
}