package intf

import "github.com/gin-gonic/gin"

// UserHandler 用户管理处理器接口
type UserHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	AssignRole(c *gin.Context)
	RemoveRole(c *gin.Context)
}

// RoleHandler 角色管理处理器接口
type RoleHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// AuthHandler 认证处理器接口
type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetProfile(c *gin.Context)
}