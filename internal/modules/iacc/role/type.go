package role

import (
	"time"
)

// Role 数据库模型
type Role struct {
	ID          string    `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Permissions *string   `db:"permissions"` // JSONB格式存储权限键数组
}

// --- 输入 DTO ---

// CreateRoleRequest 创建角色的请求 DTO
type CreateRoleRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=50"`
	Description string  `json:"description,omitempty"`
	Permissions *string `json:"permissions,omitempty"`
}

// UpdateRoleRequest 更新角色的请求 DTO
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=50"`
	Description *string `json:"description,omitempty"`
	Permissions *string `json:"permissions,omitempty"`
}

// AssignPermissionRequest 分配权限的请求 DTO
type AssignPermissionRequest struct {
	PermissionKey string `json:"permission_key" validate:"required"`
}

// RemovePermissionRequest 移除权限的请求 DTO (URL参数)
type RemovePermissionRequest struct {
	RoleID        string `uri:"id" validate:"required"`
	PermissionKey string `uri:"permission_key" validate:"required"`
}

// GetRoleRequest 获取角色详情的请求 DTO (URL参数)
type GetRoleRequest struct {
	ID string `uri:"id" validate:"required"`
}

// ListRolesRequest 获取角色列表的请求 DTO
type ListRolesRequest struct {
	Page        int    `form:"page,default=1" validate:"min=1"`
	PageSize    int    `form:"page_size,default=10" validate:"min=1,max=100"`
	Name        string `form:"name,omitempty"`
	Description string `form:"description,omitempty"`
}

// --- 输出 DTO ---

// RoleResponse 角色信息响应 DTO
type RoleResponse struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions *string   `json:"permissions,omitempty"`
}

// ListRolesResponse 角色列表响应 DTO
type ListRolesResponse struct {
	Roles       []RoleResponse `json:"roles"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
	TotalCount  int64         `json:"total_count"`
}