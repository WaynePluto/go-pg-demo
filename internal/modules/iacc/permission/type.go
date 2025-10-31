package permission

import (
	"encoding/json"
	"time"
)

// 数据库表Permission的表结构
type PermissionEntity struct {
	ID        string          `db:"id"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
	Name      string          `db:"name" label:"权限名称"`
	Type      string          `db:"type" label:"权限类型"`
	Metadata  json.RawMessage `db:"metadata" label:"权限元数据"`
}

// 创建权限的请求体
type CreatePermissionReq struct {
	Name     string          `json:"name" validate:"required" label:"权限名称"`
	Type     string          `json:"type" validate:"required" label:"权限类型"`
	Metadata json.RawMessage `json:"metadata,omitempty" label:"权限元数据"`
}

// 更新权限的请求体
type UpdatePermissionReq struct {
	Name     *string          `json:"name,omitempty" label:"权限名称"`
	Type     *string          `json:"type,omitempty" label:"权限类型"`
	Metadata *json.RawMessage `json:"metadata,omitempty" label:"权限元数据"`
}

// 查询权限的请求体
type QueryPermissionReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"权限名称"`
	Type     string `form:"type,omitempty" validate:"omitempty" label:"权限类型"`
}

// 权限响应
type PermissionRes struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// 分页列表权限响应
type PermissionListRes struct {
	List  []PermissionRes `json:"list"`
	Total int64           `json:"total"`
}
