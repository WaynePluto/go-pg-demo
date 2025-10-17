package auth

import (
	"time"
)

// --- 数据库模型 ---

// UserRole 用户角色关联模型
type UserRole struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	UserID    string    `db:"user_id"`
	RoleID    string    `db:"role_id"`
}

// --- 输入 DTO ---

// LoginRequest 用户登录请求 DTO
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest 刷新token请求 DTO
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AssignRoleRequest 分配角色请求 DTO (在auth模块中也可能使用)
type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required"`
}

// --- 输出 DTO ---

// LoginResponse 登录响应 DTO
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// UserDTO 用户信息 DTO (用于认证响应)
type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

// RefreshTokenResponse 刷新token响应 DTO
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserInfoResponse 当前用户信息响应 DTO
type UserInfoResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Profile   *string   `json:"profile,omitempty"`
	Roles     []RoleDTO `json:"roles"`
}

// RoleDTO 角色信息 DTO (用于用户信息响应)
type RoleDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UserPermissionsResponse 用户权限响应 DTO
type UserPermissionsResponse struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
}