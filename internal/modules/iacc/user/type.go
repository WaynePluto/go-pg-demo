package user

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Profile 是一个自定义类型，用于处理 JSONB 数据
type Profile map[string]interface{}

// Value - 实现 driver.Valuer 接口
func (p Profile) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan - 实现 sql.Scanner 接口
func (p *Profile) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), &p)
	}
	return json.Unmarshal(b, &p)
}

// 数据库表 iacc_user 的表结构
type UserEntity struct {
	ID        string    `db:"id" label:"用户ID"`
	CreatedAt time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt time.Time `db:"updated_at" label:"更新时间"`
	Username  string    `db:"username" label:"用户名"`
	Phone     *string   `db:"phone" label:"手机号"`
	Password  string    `db:"password" label:"密码"`
	Profile   Profile   `db:"profile" label:"个人信息"`
}

// 创建用户的请求 DTO
type CreateReq struct {
	Username string  `json:"username" validate:"required" label:"用户名"`
	Phone    string  `json:"phone" validate:"required,min=11,max=11" label:"手机号"`
	Password string  `json:"password" validate:"required" label:"密码"`
	Profile  Profile `json:"profile,omitempty" label:"个人信息"`
}

// 创建用户的响应 DTO
type CreateRes string

// 批量创建用户的请求体
type BatchCreateReq struct {
	Users []CreateReq `json:"users" validate:"required,min=1,dive" label:"用户列表"`
}

// 批量创建用户的响应体
type BatchCreateRes []string

// 根据ID获取用户的参数
type GetByIDReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"用户ID"`
}

// 根据ID获取用户的响应体
type GetByIDRes struct {
	ID        string  `json:"id" label:"用户ID"`
	Username  string  `json:"username" label:"用户名"`
	Phone     string  `json:"phone" label:"手机号"`
	Profile   Profile `json:"profile,omitempty" label:"个人信息"`
	CreatedAt string  `json:"created_at" label:"创建时间"`
	UpdatedAt string  `json:"updated_at" label:"更新时间"`
}

// 更新用户的请求体
type UpdateByIDReq struct {
	ID       string   `uri:"id" validate:"required,uuid" label:"用户ID"`
	Username *string  `json:"username,omitempty" validate:"omitempty" label:"用户名"`
	Phone    *string  `json:"phone,omitempty" validate:"omitempty,min=11,max=11" label:"手机号"`
	Password *string  `json:"password,omitempty" validate:"omitempty" label:"密码"`
	Profile  *Profile `json:"profile,omitempty" label:"个人信息"`
}

// 更新用户的响应体
type UpdateByIDRes = int64

// 根据ID删除用户的请求参数
type DeleteByIDReq struct {
	ID string `uri:"id" binding:"required,uuid" label:"用户ID"`
}

// 根据ID删除用户的响应
type DeleteByIDRes = int64

// 批量删除用户请求参数
type DeleteUsersReq struct {
	IDs []string `json:"ids" validate:"required,min=1,dive,uuid" label:"用户ID列表"`
}

// 批量删除用户响应
type BatchDeleteRes = int64

// 查询用户的请求体
type QueryListReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Phone    string `form:"phone,omitempty" validate:"omitempty" label:"手机号"`
	Username string `form:"username,omitempty" validate:"omitempty" label:"用户名"`
	OrderBy  string `form:"orderBy,default=id" validate:"omitempty" label:"排序字段"`
	Order    string `form:"order,default=desc" validate:"omitempty" label:"排序顺序"`
}

// 用户响应
type UserItem struct {
	ID        string  `json:"id" label:"用户ID"`
	Username  string  `json:"username" label:"用户名"`
	Phone     string  `json:"phone" label:"手机号"`
	Profile   Profile `json:"profile,omitempty" label:"个人信息"`
	CreatedAt string  `json:"created_at" label:"创建时间"`
	UpdatedAt string  `json:"updated_at" label:"更新时间"`
}

// 查询用户的响应体
type QueryListRes struct {
	List  []UserItem `json:"list"`
	Total int64      `json:"total"`
}

// 给用户分配角色的请求 DTO
type AssignRolesReq struct {
	ID      string   `uri:"id" validate:"required,uuid" label:"用户ID"`
	RoleIDs []string `json:"role_ids" validate:"required,min=1,dive,uuid" label:"角色ID列表"`
}

// 给用户分配角色的响应 DTO
type AssignRolesRes = int64

// 获取用户角色列表的请求参数
type GetRolesReq struct {
	ID string `uri:"id" validate:"required,uuid" label:"用户ID"`
}

// 角色项结构体
type RoleItem struct {
	ID          string  `json:"id" label:"角色ID"`
	Name        string  `json:"name" label:"角色名称"`
	Description *string `json:"description,omitempty" label:"角色描述"`
	CreatedAt   string  `json:"created_at" label:"创建时间"`
	UpdatedAt   string  `json:"updated_at" label:"更新时间"`
}

// 获取用户角色列表的响应体
type GetRolesRes struct {
	List  []RoleItem `json:"list"`
	Total int64      `json:"total"`
}
