package permission

import (
	"encoding/json"
	"time"
)

// 数据库表Permission的表结构
type PermissionEntity struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	// 权限名称
	Name string `db:"name"`
	// 权限类型
	Type string `db:"type"`
	// 权限元数据
	Metadata json.RawMessage `db:"metadata"`
}

// 创建权限的请求 DTO
type CreatePermissionReq struct {
	// 权限名称
	Name string `json:"name" validate:"required"`
	// 权限类型
	Type string `json:"type" validate:"required"`
	// 权限元数据
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// 更新权限的请求体
type UpdatePermissionReq struct {
	// 权限名称
	Name *string `json:"name,omitempty"`
	// 权限类型
	Type *string `json:"type,omitempty"`
	// 权限元数据
	Metadata *json.RawMessage `json:"metadata,omitempty"`
}

// 查询权限的请求体
type QueryPermissionReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" error:"页码必须大于等于1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" error:"每页大小必须在1到100之间"`
	Name     string `form:"name,omitempty" validate:"omitempty"`
	Type     string `form:"type,omitempty" validate:"omitempty"`
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
