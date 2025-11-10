package v1

import (
	"go-pg-demo/api/v1/intf"

	"github.com/gin-gonic/gin"
)

// v1路由
type Router struct {
	Engine            *gin.Engine
	RouterGroup       *gin.RouterGroup
	TemplateHandler   intf.ITemplateHandler
	UserHandler       intf.UserHandler
	RoleHandler       intf.RoleHandler
	AuthHandler       intf.AuthHandler
	PermissionHandler intf.PermissionHandler
}

func NewRouter(
	engine *gin.Engine,
	templateHandler intf.ITemplateHandler,
	userHandler intf.UserHandler,
	roleHandler intf.RoleHandler,
	authHandler intf.AuthHandler,
	permissionHandler intf.PermissionHandler,
) *Router {
	return &Router{
		Engine:            engine,
		TemplateHandler:   templateHandler,
		UserHandler:       userHandler,
		RoleHandler:       roleHandler,
		AuthHandler:       authHandler,
		PermissionHandler: permissionHandler,
	}
}

func (r *Router) Register() {
	r.RouterGroup = r.Engine.Group("/v1")
	r.RegisterTemplate()
	r.RegisterIACCPermission()
	r.RegisterIACCUser()
	r.RegisterIACCRole()
	r.RegisterIACCAuth()
}

func (r *Router) RegisterTemplate() {
	templates := r.RouterGroup.Group("/template")
	{
		templates.POST("", r.TemplateHandler.Create)
		templates.GET("/:id", r.TemplateHandler.GetByID)
		templates.PUT("/:id", r.TemplateHandler.UpdateByID)
		templates.DELETE("/:id", r.TemplateHandler.DeleteByID)
		templates.POST("/batch-create", r.TemplateHandler.BatchCreate)
		templates.GET("/list", r.TemplateHandler.QueryList)
		templates.POST("/batch-delete", r.TemplateHandler.BatchDelete)
	}
}

func (r *Router) RegisterIACCPermission() {
	permissions := r.RouterGroup.Group("/permission")
	{
		permissions.POST("", r.PermissionHandler.Create)
		permissions.GET("/:id", r.PermissionHandler.GetByID)
		permissions.PUT("/:id", r.PermissionHandler.UpdateByID)
		permissions.DELETE("/:id", r.PermissionHandler.DeleteByID)
		permissions.GET("/list", r.PermissionHandler.QueryList)
	}
}

func (r *Router) RegisterIACCRole() {
	roles := r.RouterGroup.Group("/role")
	{
		roles.POST("", r.RoleHandler.Create)
		roles.GET("/:id", r.RoleHandler.GetByID)
		roles.PUT("/:id", r.RoleHandler.UpdateByID)
		roles.DELETE("/:id", r.RoleHandler.DeleteByID)
		roles.GET("/list", r.RoleHandler.QueryList)
		roles.POST("/:id/permission", r.RoleHandler.AssignPermission)
	}
}

func (r *Router) RegisterIACCUser() {
	users := r.RouterGroup.Group("/user")
	{
		users.POST("", r.UserHandler.Create)
		users.POST("/batch-create", r.UserHandler.BatchCreate)
		users.GET("/:id", r.UserHandler.GetByID)
		users.PUT("/:id", r.UserHandler.UpdateByID)
		users.DELETE("/:id", r.UserHandler.DeleteByID)
		users.POST("/batch-delete", r.UserHandler.BatchDelete)
		users.GET("/list", r.UserHandler.QueryList)
		users.POST("/:id/role", r.UserHandler.AssignRole)
	}
}

func (r *Router) RegisterIACCAuth() {
	auth := r.RouterGroup.Group("/auth")
	{
		auth.POST("/login", r.AuthHandler.Login)
		auth.POST("/refresh-token", r.AuthHandler.RefreshToken)
		auth.GET("/user-detail", r.AuthHandler.UserDetail)
	}
}
