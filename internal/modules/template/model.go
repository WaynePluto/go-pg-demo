package template

import (
	"time"
)

// TemplateEntity 定义领域模型
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

// UpdateTemplateRequest 更新模板的请求 DTO
// 使用指针来区分"未提供"和"提供空值"
type UpdateTemplateRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty"`
	Num  *int    `json:"num,omitempty" validate:"omitempty"`
}

// --- 输出 DTO ---

// TemplateResponse 返回给客户端的模板信息 DTO
// 它只关心要展示什么，并隐藏了敏感信息
type TemplateResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Num       *int   `json:"num,omitempty"`
	CreatedAt string `json:"created_at"` // 格式化为字符串，更利于前端处理
	UpdatedAt string `json:"updated_at"`
}

// TemplateListResponse 模板列表的响应 DTO，通常包含分页信息
type TemplateListResponse struct {
	List  []TemplateResponse `json:"list"`
	Total int64              `json:"total"`
}