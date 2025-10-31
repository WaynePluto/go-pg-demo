package role

import "time"

// 数据库表 role 的表结构
type RoleEntity struct {
	ID          string    `db:"id" label:"角色ID"`
	CreatedAt   time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt   time.Time `db:"updated_at" label:"更新时间"`
	Name        string    `db:"name" label:"角色名称"`
	Description *string   `db:"description" label:"角色描述"`
}

// 创建角色的请求 DTO
type CreateRoleReq struct {
	Name        string  `json:"name" validate:"required" label:"角色名称"`
	Description *string `json:"description" label:"角色描述"`
}

// 更新角色的请求体
type UpdateRoleReq struct {
	Name        *string `json:"name,omitempty" validate:"omitempty" label:"角色名称"`
	Description *string `json:"description,omitempty" validate:"omitempty" label:"角色描述"`
}

// 查询角色的请求体
type QueryRoleReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"角色名称"`
}

// 给角色分配权限的请求体
type AssignPermissionsReq struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1" label:"权限ID列表"`
}

// 角色响应
type RoleRes struct {
	ID          string  `json:"id" label:"角色ID"`
	Name        string  `json:"name" label:"角色名称"`
	Description *string `json:"description,omitempty" label:"角色描述"`
	CreatedAt   string  `json:"created_at" label:"创建时间"`
	UpdatedAt   string  `json:"updated_at" label:"更新时间"`
}

// 分页列表角色响应
type RoleListRes struct {
	List  []RoleRes `json:"list" label:"角色列表"`
	Total int64     `json:"total" label:"总数"`
}
