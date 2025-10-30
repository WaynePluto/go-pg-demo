package intf

import "github.com/gin-gonic/gin"

// 权限管理处理器接口
type PermissionHandler interface {
	Create(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	DeleteByID(c *gin.Context)
	QueryList(c *gin.Context)
}

// 角色管理处理器接口
type RoleHandler interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	DeleteByID(c *gin.Context)
	QueryList(c *gin.Context)
	AssignPermission(c *gin.Context)
}

// 用户管理处理器接口
type UserHandler interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	DeleteByID(c *gin.Context)
	QueryList(c *gin.Context)
	AssignRole(c *gin.Context)
}

// 认证处理器接口
type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetMe(c *gin.Context)
}
