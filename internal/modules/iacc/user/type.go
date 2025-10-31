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

// 数据库表 user 的表结构
type UserEntity struct {
	ID        string    `db:"id" label:"用户ID"`
	CreatedAt time.Time `db:"created_at" label:"创建时间"`
	UpdatedAt time.Time `db:"updated_at" label:"更新时间"`
	Username  string    `db:"username" label:"用户名"`
	Phone     string    `db:"phone" label:"手机号"`
	Password  string    `db:"password" label:"密码"`
	Profile   Profile   `db:"profile" label:"个人信息"`
}

// 创建用户的请求 DTO
type CreateUserReq struct {
	Username string  `json:"username" validate:"required" label:"用户名"`
	Phone    string  `json:"phone" validate:"required,min=11,max=11" label:"手机号"`
	Password string  `json:"password" validate:"required" label:"密码"`
	Profile  Profile `json:"profile,omitempty" label:"个人信息"`
}

// 更新用户的请求 DTO
type UpdateUserReq struct {
	// 密码，前端进行MD5加密
	Password *string `json:"password,omitempty" validate:"omitempty" label:"密码"`
	// 个人信息
	Profile Profile `json:"profile,omitempty" label:"个人信息"`
}

// 查询用户的请求 DTO
type QueryUserReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" label:"页码"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" label:"每页大小"`
	Phone    string `form:"phone,omitempty" validate:"omitempty" label:"手机号"`
}

// 给用户分配角色的请求 DTO
type AssignRolesToUserReq struct {
	// 角色ID列表
	RoleIDs []string `json:"role_ids" validate:"required,min=1,dive,uuid" label:"角色ID列表"`
}

// 用户信息的响应 DTO
type UserRes struct {
	ID        string  `json:"id" label:"用户ID"`
	Phone     string  `json:"phone" label:"手机号"`
	Profile   Profile `json:"profile,omitempty" label:"个人信息"`
	CreatedAt string  `json:"created_at" label:"创建时间"`
	UpdatedAt string  `json:"updated_at" label:"更新时间"`
}

// UserListRes 用户列表的响应 DTO
type UserListRes struct {
	List  []UserRes `json:"list"`
	Total int64     `json:"total"`
}
