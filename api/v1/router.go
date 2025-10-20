package v1

import (
	"go-pg-demo/api/v1/intf"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
)

// v1路由
type Router struct {
	Engine               *gin.Engine
	RouterGroup          *gin.RouterGroup
	TemplateHandler      intf.ITemplateHandler
	UserHandler          intf.UserHandler
	RoleHandler          intf.RoleHandler
	AuthHandler          intf.AuthHandler
	PermissionMiddleware middlewares.PermissionMiddleware
}

func NewRouter(
	engine *gin.Engine,
	templateHandler intf.ITemplateHandler,
	userHandler intf.UserHandler,
	roleHandler intf.RoleHandler,
	authHandler intf.AuthHandler,
	permissionMiddleware middlewares.PermissionMiddleware,
) *Router {
	return &Router{
		Engine:               engine,
		TemplateHandler:      templateHandler,
		UserHandler:          userHandler,
		RoleHandler:          roleHandler,
		AuthHandler:          authHandler,
		PermissionMiddleware: permissionMiddleware,
	}
}

func (r *Router) Register() {
	r.RouterGroup = r.Engine.Group("/v1")
	r.RegisterTemplate()
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

func (r *Router) RegisterIACCUser() {
	users := r.RouterGroup.Group("/user")
	{
		users.POST("", r.PermissionMiddleware(pkgs.Permissions.UserCreate.Key), r.UserHandler.Create)
		users.GET("/:id", r.PermissionMiddleware(pkgs.Permissions.UserView.Key), r.UserHandler.Get)
		users.PUT("/:id", r.PermissionMiddleware(pkgs.Permissions.UserUpdate.Key), r.UserHandler.Update)
		users.DELETE("/:id", r.PermissionMiddleware(pkgs.Permissions.UserDelete.Key), r.UserHandler.Delete)
		users.GET("/list", r.PermissionMiddleware(pkgs.Permissions.UserList.Key), r.UserHandler.List)
		users.POST("/:id/role", r.PermissionMiddleware(pkgs.Permissions.UserAssignRole.Key), r.UserHandler.AssignRole)
		users.DELETE("/:id/role/:role_id", r.PermissionMiddleware(pkgs.Permissions.UserDelete.Key), r.UserHandler.RemoveRole)
	}
}

func (r *Router) RegisterIACCRole() {
	roles := r.RouterGroup.Group("/role")
	{
		roles.POST("", r.PermissionMiddleware(pkgs.Permissions.RoleCreate.Key), r.RoleHandler.Create)
		roles.GET("/:id", r.PermissionMiddleware(pkgs.Permissions.RoleView.Key), r.RoleHandler.Get)
		roles.PUT("/:id", r.PermissionMiddleware(pkgs.Permissions.RoleUpdate.Key), r.RoleHandler.Update)
		roles.DELETE("/:id", r.PermissionMiddleware(pkgs.Permissions.RoleDelete.Key), r.RoleHandler.Delete)
		roles.GET("/list", r.PermissionMiddleware(pkgs.Permissions.RoleList.Key), r.RoleHandler.List)
	}
}

func (r *Router) RegisterIACCAuth() {
	auth := r.RouterGroup.Group("/auth")
	{
		auth.POST("/login", r.AuthHandler.Login)
		auth.POST("/refresh-token", r.AuthHandler.RefreshToken)
		auth.GET("/profile", r.AuthHandler.GetProfile)
	}
}
