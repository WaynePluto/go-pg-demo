package intf

import "github.com/gin-gonic/gin"

// 定义template模块的handler接口
type ITemplateHandler interface {
	Create(*gin.Context)
	GetByID(*gin.Context)
	UpdateByID(*gin.Context)
	DeleteByID(*gin.Context)
	BatchCreate(*gin.Context)
	QueryList(*gin.Context)
	BatchDelete(*gin.Context)
}
