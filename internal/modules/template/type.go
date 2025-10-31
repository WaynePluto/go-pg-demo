package template

import (
	"time"
)

// 数据库表Template的表结构
type TemplateEntity struct {
	ID        string    `db:"id" label:"模板ID"`
	CreatedAt time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt time.Time `db:"updated_at" label:"更新时间"`
	Name      string    `db:"name" label:"模板名称"`
	Num       *int      `db:"num" label:"模板数量"`
}

// 创建模板的请求 DTO
type CreateTemplateReq struct {
	Name string `json:"name" validate:"required" label:"模板名称"`
	Num  *int   `json:"num,omitempty" validate:"omitempty,min=1,max=1000" label:"模板数量"`
}

// 批量创建模板的请求体
type CreateTemplatesReq struct {
	Templates []CreateTemplateReq `json:"templates" validate:"required,min=1,dive" label:"模板列表"`
}

// 更新模板的请求体
type UpdateTemplateReq struct {
	Name *string `json:"name,omitempty" validate:"omitempty" label:"模板名称"`
	Num  *int    `json:"num,omitempty" validate:"omitempty,min=1,max=1000" label:"模板数量"`
}

// 批量删除模板的请求体
type DeleteTemplatesReq struct {
	IDs []string `json:"ids" validate:"required,min=1" label:"模板ID列表"`
}

// 查询模板的请求体
type QueryTemplateReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"模板名称"`
}

// 模板响应
type TemplateRes struct {
	ID        string `json:"id" label:"模板ID"`
	Name      string `json:"name" label:"模板名称"`
	Num       *int   `json:"num,omitempty" label:"模板数量"`
	CreatedAt string `json:"created_at" label:"创建时间"`
	UpdatedAt string `json:"updated_at" label:"更新时间"`
}

// 分页列表模板响应
type TemplateListRes struct {
	List  []TemplateRes `json:"list"`
	Total int64         `json:"total"`
}
