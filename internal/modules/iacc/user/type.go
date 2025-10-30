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

// UserEntity 数据库表 user 的表结构
type UserEntity struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	// 手机号
	Phone string `db:"phone"`
	// 密码
	Password string `db:"password"`
	// 个人信息
	Profile Profile `db:"profile"`
}

// CreateUserReq 创建用户的请求 DTO
type CreateUserReq struct {
	// 手机号
	Phone string `json:"phone" validate:"required,e164" message:"手机号格式不正确"`
	// 密码，前端进行MD5加密
	Password string `json:"password" validate:"required,md5" message:"密码必须是MD5格式"`
	// 个人信息
	Profile Profile `json:"profile,omitempty"`
}

// UpdateUserReq 更新用户的请求 DTO
type UpdateUserReq struct {
	// 密码，前端进行MD5加密
	Password *string `json:"password,omitempty" validate:"omitempty,md5" message:"密码必须是MD5格式"`
	// 个人信息
	Profile Profile `json:"profile,omitempty"`
}

// QueryUserReq 查询用户的请求 DTO
type QueryUserReq struct {
	Page     int    `form:"page,default=1" validate:"min=1" message:"页码必须大于等于1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100" message:"每页大小必须在1到100之间"`
	Phone    string `form:"phone,omitempty" validate:"omitempty,e164" message:"手机号格式不正确"`
}

// AssignRolesToUserReq 给用户分配角色的请求 DTO
type AssignRolesToUserReq struct {
	// 角色ID列表
	RoleIDs []string `json:"role_ids" validate:"required,min=1,dive,uuid" message:"角色ID列表不能为空或包含无效的ID"`
}

// UserRes 用户信息的响应 DTO
type UserRes struct {
	ID        string  `json:"id"`
	Phone     string  `json:"phone"`
	Profile   Profile `json:"profile,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// UserListRes 用户列表的响应 DTO
type UserListRes struct {
	List  []UserRes `json:"list"`
	Total int64     `json:"total"`
}
