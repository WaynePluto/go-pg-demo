package role

import "time"

// RoleEntity 数据库表 role 的表结构
type RoleEntity struct {
	ID          string    `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

// CreateRoleReq 创建角色的请求 DTO
type CreateRoleReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// UpdateRoleReq 更新角色的请求体
type UpdateRoleReq struct {
	Name        *string `json:"name,omitempty" validate:"omitempty"`
	Description *string `json:"description,omitempty" validate:"omitempty"`
}

// QueryRoleReq 查询角色的请求体
type QueryRoleReq struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100"`
	Name     string `form:"name,omitempty" validate:"omitempty"`
}

// AssignPermissionsReq 给角色分配权限的请求体
type AssignPermissionsReq struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1"`
}

// RoleRes 角色响应
type RoleRes struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RoleListRes 分页列表角色响应
type RoleListRes struct {
	List  []RoleRes `json:"list"`
	Total int64     `json:"total"`
}
