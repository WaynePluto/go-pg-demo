package template

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go-pg-demo/internal/pkgs"
)

type TemplateHandler struct {
	service ITemplateService
	logger  *zap.Logger
}

func NewTemplateHandler(service ITemplateService, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTemplate godoc
// @Summary 创建模板
// @Description 创建一个新的模板
// @Tags template
// @Accept json
// @Produce json
// @Param request body CreateTemplateRequest true "创建模板请求参数"
// @Success 200 {object} TemplateResponse
// @Router /templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 创建模板实体
	entity := &TemplateEntity{
		Name: req.Name,
		Num:  req.Num,
	}

	// 调用服务层创建模板
	created, err := h.service.Create(c.Request.Context(), entity)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to create template")
		return
	}

	// 转换为响应DTO
	response := TemplateResponse{
		ID:        created.ID,
		Name:      created.Name,
		CreatedAt: created.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: created.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果Num字段存在，则设置它
	if created.Num != nil {
		response.Num = created.Num
	}

	pkgs.Success(c, response)
}

// GetTemplate godoc
// @Summary 获取模板详情
// @Description 根据ID获取模板详情
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} TemplateResponse
// @Router /templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get template", zap.String("id", id), zap.Error(err))
		pkgs.Error(c, http.StatusNotFound, "Template not found")
		return
	}

	response := TemplateResponse{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: entity.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if entity.Num != nil {
		response.Num = entity.Num
	}

	pkgs.Success(c, response)
}

// UpdateTemplate godoc
// @Summary 更新模板
// @Description 根据ID更新模板
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param request body UpdateTemplateRequest true "更新模板请求参数"
// @Success 200 {object} TemplateResponse
// @Router /templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgs.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取现有实体
	existing, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get template", zap.String("id", id), zap.Error(err))
		pkgs.Error(c, http.StatusNotFound, "Template not found")
		return
	}

	// 更新字段
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Num != nil {
		existing.Num = req.Num
	}

	// 调用服务层更新模板
	updated, err := h.service.Update(c.Request.Context(), existing)
	if err != nil {
		h.logger.Error("Failed to update template", zap.String("id", id), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to update template")
		return
	}

	// 转换为响应DTO
	response := TemplateResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		CreatedAt: updated.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: updated.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if updated.Num != nil {
		response.Num = updated.Num
	}

	pkgs.Success(c, response)
}

// DeleteTemplate godoc
// @Summary 删除模板
// @Description 根据ID删除模板
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 204
// @Router /templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete template", zap.String("id", id), zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to delete template")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTemplates godoc
// @Summary 获取模板列表
// @Description 分页获取模板列表
// @Tags template
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} TemplateListResponse
// @Router /templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	// 获取分页参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size <= 0 || size > 100 {
		size = 10
	}

	// 调用服务层获取模板列表
	list, total, err := h.service.List(c.Request.Context(), page, size)
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		pkgs.Error(c, http.StatusInternalServerError, "Failed to list templates")
		return
	}

	// 转换为响应DTO
	responses := make([]TemplateResponse, len(list))
	for i, item := range list {
		responses[i] = TemplateResponse{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: item.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if item.Num != nil {
			responses[i].Num = item.Num
		}
	}

	response := TemplateListResponse{
		List:  responses,
		Total: total,
	}

	pkgs.Success(c, response)
}