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
type CreateReq struct {
	Name string `json:"name" validate:"required" label:"模板名称"`
	Num  *int   `json:"num,omitempty" validate:"omitempty,min=1,max=1000" label:"模板数量"`
}

// 创建模板的响应 DTO
type CreateRes string

// 批量创建模板的请求体
type BatchCreateReq struct {
	Templates []CreateReq `json:"templates" validate:"required,min=1,dive" label:"模板列表"`
}

// 批量创建模板的响应体
type BatchCreateRes []string

// 根据ID获取模板的参数
type GetByIDReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"模板ID"`
}

// 根据ID获取模板的响应体
type GetByIDRes struct {
	ID        string `json:"id" label:"模板ID"`
	Name      string `json:"name" label:"模板名称"`
	Num       *int   `json:"num,omitempty" label:"模板数量"`
	CreatedAt string `json:"created_at" label:"创建时间"`
	UpdatedAt string `json:"updated_at" label:"更新时间"`
}

// 更新模板的请求体
type UpdateByIDReq struct {
	ID   string  `uri:"id" validate:"required,uuid" label:"模板ID"`
	Name *string `json:"name,omitempty" validate:"omitempty" label:"模板名称"`
	Num  *int    `json:"num,omitempty" validate:"omitempty,min=1,max=1000" label:"模板数量"`
}

// 更新模板的响应体
type UpdateByIDRes = int64

// 根据ID删除模板的请求参数
type DeleteByIDReq struct {
	ID string `uri:"id" binding:"required,uuid" label:"模板ID"`
}

// 根据ID删除模板的响应
type DeleteByIDRes = int64

// 批量删除模板请求参数
type DeleteTemplatesReq struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid" label:"模板ID列表"`
}

// 批量删除模板响应
type BatchDeleteRes = int64

// 查询模板的请求体
type QueryListReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"模板名称"`
	OrderBy  string `form:"orderBy,default=id" validate:"omitempty" label:"排序字段"`
	Order    string `form:"order,default=desc" validate:"omitempty" label:"排序顺序"`
}

// 模板响应
type TemplateItem struct {
	ID        string `json:"id" label:"模板ID"`
	Name      string `json:"name" label:"模板名称"`
	Num       *int   `json:"num,omitempty" label:"模板数量"`
	CreatedAt string `json:"created_at" label:"创建时间"`
	UpdatedAt string `json:"updated_at" label:"更新时间"`
}

// 查询模板的响应体
type QueryListRes struct {
	List  []TemplateItem `json:"list"`
	Total int64          `json:"total"`
}
