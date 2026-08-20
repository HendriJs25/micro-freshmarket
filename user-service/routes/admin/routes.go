package admin

import (
	"user-service/constants"
	adminhandler "user-service/handler/admin"
	"user-service/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, handler *adminhandler.Handler, authentication *middleware.Authentication) {
	admin := router.Group("/admin")
	admin.Use(authentication.Handle(), middleware.RequireRole(constants.RoleSuperAdmin))

	admin.GET("/check", handler.Check)
}
