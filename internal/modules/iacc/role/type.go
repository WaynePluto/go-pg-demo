package role

import "time"

// 数据库表 iacc_role 的表结构
type RoleEntity struct {
	ID          string    `db:"id" label:"角色ID"`
	CreatedAt   time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt   time.Time `db:"updated_at" label:"更新时间"`
	Name        string    `db:"name" label:"角色名称"`
	Description *string   `db:"description" label:"角色描述"`
}

// 创建角色的请求 DTO
type CreateReq struct {
	Name        string  `json:"name" validate:"required" label:"角色名称"`
	Description *string `json:"description" label:"角色描述"`
}

// 创建角色的响应 DTO
type CreateRes string

// 批量创建角色的请求体
type BatchCreateReq struct {
	Roles []CreateReq `json:"roles" validate:"required,min=1,dive" label:"角色列表"`
}

// 批量创建角色的响应体
type BatchCreateRes []string

// 根据ID获取角色的参数
type GetByIDReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"角色ID"`
}

// 根据ID获取角色的响应体
type GetByIDRes struct {
	ID          string  `json:"id" label:"角色ID"`
	Name        string  `json:"name" label:"角色名称"`
	Description *string `json:"description,omitempty" label:"角色描述"`
	CreatedAt   string  `json:"created_at" label:"创建时间"`
	UpdatedAt   string  `json:"updated_at" label:"更新时间"`
}

// 更新角色的请求体
type UpdateByIDReq struct {
	ID          string  `uri:"id" validate:"required,uuid" label:"角色ID"`
	Name        *string `json:"name,omitempty" validate:"omitempty" label:"角色名称"`
	Description *string `json:"description,omitempty" validate:"omitempty" label:"角色描述"`
}

// 更新角色的响应体
type UpdateByIDRes = int64

// 根据ID删除角色的请求参数
type DeleteByIDReq struct {
	ID string `uri:"id" binding:"required,uuid" label:"角色ID"`
}

// 根据ID删除角色的响应
type DeleteByIDRes = int64

// 批量删除角色请求参数
type DeleteRolesReq struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid" label:"角色ID列表"`
}

// 批量删除角色响应
type BatchDeleteRes = int64

// 查询角色的请求体
type QueryListReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Name     string `form:"name,omitempty" validate:"omitempty" label:"角色名称"`
	OrderBy  string `form:"orderBy,default=id" validate:"omitempty" label:"排序字段"`
	Order    string `form:"order,default=desc" validate:"omitempty" label:"排序顺序"`
}

// 角色响应
type RoleItem struct {
	ID          string  `json:"id" label:"角色ID"`
	Name        string  `json:"name" label:"角色名称"`
	Description *string `json:"description,omitempty" label:"角色描述"`
	CreatedAt   string  `json:"created_at" label:"创建时间"`
	UpdatedAt   string  `json:"updated_at" label:"更新时间"`
}

// 查询角色的响应体
type QueryListRes struct {
	List  []RoleItem `json:"list"`
	Total int64      `json:"total"`
}

// 给角色分配权限的请求体
type AssignPermissionsReq struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1" label:"权限ID列表"`
}

// 给角色分配权限的请求参数（包含角色ID）
type AssignPermissionsByIDReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"角色ID"`
	AssignPermissionsReq
}

// 给角色分配权限的响应体
type AssignPermissionsRes = int64
