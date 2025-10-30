package template

import (
	"time"
)

// 数据库表Template的表结构
type TemplateEntity struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	// 模板名称
	Name string `db:"name"`
	// 模板数量，可能为NULL
	Num *int `db:"num"`
}

// 创建模板的请求 DTO
type CreateTemplateReq struct {
	// 模板名称
	Name string `json:"name" validate:"required"`
	// 模板数量
	Num *int `json:"num,omitempty" validate:"omitempty,min=1,max=1000" message:"Num必须在1到1000之间"`
}

// 批量创建模板的请求体
type CreateTemplatesReq struct {
	Templates []CreateTemplateReq `json:"templates" validate:"required,min=1,dive"`
}

// 更新模板的请求体
type UpdateTemplateReq struct {
	Name *string `json:"name,omitempty" validate:"omitempty"`
	Num  *int    `json:"num,omitempty" validate:"omitempty,min=1,max=1000" message:"Num必须在1到1000之间"`
}

// 批量删除模板的请求体
type DeleteTemplatesReq struct {
	IDs []string `json:"ids" validate:"required,min=1" error:"至少需要提供一个ID"`
}

// 查询模板的请求体
type QueryTemplateReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" error:"页码必须大于等于1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" error:"每页大小必须在1到100之间"`
	Name     string `form:"name,omitempty" validate:"omitempty" error:"名称搜索项不能包含非法字符"`
}

// 模板响应
type TemplateRes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Num       *int   `json:"num,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// 分页列表模板响应
type TemplateListRes struct {
	List  []TemplateRes `json:"list"`
	Total int64         `json:"total"`
}
