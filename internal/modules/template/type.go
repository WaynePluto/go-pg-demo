package template

import (
	"time"
)

type TemplateEntity struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Num       *int      `db:"num"`
}

// --- 输入 DTO ---

// CreateTemplateRequest 创建模板的请求 DTO
type CreateTemplateRequest struct {
	Name string `json:"name" validate:"required"`
	Num  *int   `json:"num,omitempty" validate:"omitempty"`
}

// CreateTemplatesRequest 批量创建模板的请求 DTO
type CreateTemplatesRequest struct {
	Templates []CreateTemplateRequest `json:"templates" validate:"required,min=1,dive"`
}

// UpdateTemplateRequest 更新模板的请求 DTO
type UpdateTemplateRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty"`
	Num  *int    `json:"num,omitempty" validate:"omitempty"`
}

// DeleteTemplatesRequest 批量删除模板的请求 DTO
type DeleteTemplatesRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// QueryTemplateRequest 查询模板的请求 DTO
type QueryTemplateRequest struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100"`
	Name     string `form:"name,omitempty" validate:"omitempty"`
}

// --- 输出 DTO ---
type TemplateResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Num       *int   `json:"num,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TemplateListResponse struct {
	List  []TemplateResponse `json:"list"`
	Total int64              `json:"total"`
}
