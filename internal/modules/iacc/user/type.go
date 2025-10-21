package user

import (
	"time"
)

// User 数据库模型
type User struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Username  string    `db:"username"`
	Phone     *string   `db:"phone"`
	Password  string    `db:"password"`
	Profile   *string   `db:"profile"` // JSONB格式存储用户扩展信息
}

// --- 输入 DTO ---

// CreateUserRequest 创建用户的请求 DTO
type CreateUserRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Phone    string  `json:"phone" validate:"required"`
	Password string  `json:"password" validate:"required,min=6,max=50"`
	Profile  *string `json:"profile,omitempty"`
}

// UpdateUserRequest 更新用户的请求 DTO
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty"`
	Profile  *string `json:"profile,omitempty"`
}

// AssignRoleRequest 分配角色的请求 DTO
type AssignRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}

// RemoveRoleRequest 移除角色的请求 DTO (URL参数)
type RemoveRoleRequest struct {
	UserID string `uri:"id" validate:"required"`
	RoleID string `uri:"role_id" validate:"required"`
}

// GetUserRequest 获取用户详情的请求 DTO (URL参数)
type GetUserRequest struct {
	ID string `uri:"id" validate:"required"`
}

// ListUsersRequest 获取用户列表的请求 DTO
type ListUsersRequest struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"page_size,default=10" validate:"min=1,max=100"`
	Username string `form:"username,omitempty"`
	Phone    string `form:"phone,omitempty"`
}

// --- 输出 DTO ---

// UserResponse 用户信息响应 DTO
type UserResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Profile   *string   `json:"profile,omitempty"`
}

// ListUsersResponse 用户列表响应 DTO
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalCount int64          `json:"total_count"`
}
