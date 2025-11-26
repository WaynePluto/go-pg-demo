package permission

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Permision元数据类型
type Metadata struct {
	Path   *string `json:"path,omitempty" label:"接口路径"`
	Method *string `json:"method,omitempty" label:"请求方法"`
	Code   *string `json:"code,omitempty" label:"权限编码"`
}

// Value 实现 driver.Valuer 接口，用于将Metadata类型正确存储到数据库中
func (p Metadata) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口，用于从数据库中正确读取Metadata类型
func (p *Metadata) Scan(value interface{}) error {
	if value == nil {
		*p = Metadata{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), &p)
	}
	return json.Unmarshal(b, &p)
}

// 数据库表Permission的表结构
type PermissionEntity struct {
	ID        string    `db:"id" label:"权限ID"`
	CreatedAt time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt time.Time `db:"updated_at" label:"更新时间"`
	Name      string    `db:"name" label:"权限名称"`
	Type      string    `db:"type" label:"权限类型"`
	Metadata  Metadata  `db:"metadata" label:"权限元数据"`
}

// 创建权限的请求 DTO
type CreatePermissionReq struct {
	Name     string   `json:"name" validate:"required" label:"权限名称"`
	Type     string   `json:"type" validate:"required" label:"权限类型"`
	Metadata Metadata `json:"metadata" label:"权限元数据"`
}

// 创建权限的响应 DTO
type CreatePermissionRes string

// 根据ID获取权限的参数
type GetByIDReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"权限ID"`
}

// 根据ID获取权限的响应
type GetByIDRes struct {
	ID        string   `json:"id" label:"权限ID"`
	Name      string   `json:"name" label:"权限名称"`
	Type      string   `json:"type" label:"权限类型"`
	Metadata  Metadata `json:"metadata,omitempty" label:"权限元数据"`
	CreatedAt string   `json:"created_at" label:"创建时间"`
	UpdatedAt string   `json:"updated_at" label:"更新时间"`
}

// 更新权限的请求体
type UpdatePermissionReq struct {
	ID       string    `uri:"id" validate:"required,uuid" label:"权限ID"`
	Name     *string   `json:"name,omitempty" label:"权限名称"`
	Type     *string   `json:"type,omitempty" label:"权限类型"`
	Metadata *Metadata `json:"metadata,omitempty" label:"权限元数据"`
}

// 更新权限的响应体
type UpdatePermissionRes = int64

// 根据ID删除权限的请求参数
type DeleteByIDReq struct {
	ID string `uri:"id" binding:"required,uuid" label:"权限ID"`
}

// 根据ID删除权限的响应
type DeleteByIDRes = int64

// 查询权限的请求体
type QueryListReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"权限名称"`
	Type     string `form:"type,omitempty" validate:"omitempty" label:"权限类型"`
	OrderBy  string `form:"orderBy,default=id" validate:"omitempty" label:"排序字段"`
	Order    string `form:"order,default=desc" validate:"omitempty" label:"排序顺序"`
}

// 权限响应项
type PermissionItem struct {
	ID        string   `json:"id" label:"权限ID"`
	Name      string   `json:"name" label:"权限名称"`
	Type      string   `json:"type" label:"权限类型"`
	Metadata  Metadata `json:"metadata,omitempty" label:"权限元数据"`
	CreatedAt string   `json:"created_at" label:"创建时间"`
	UpdatedAt string   `json:"updated_at" label:"更新时间"`
}

// 分页列表权限响应
type QueryListRes struct {
	List  []PermissionItem `json:"list"`
	Total int64            `json:"total"`
}
