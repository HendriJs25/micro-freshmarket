package admin

import (
	"user-service/constants"
	adminhandler "user-service/handler/admin"
	customerhandler "user-service/handler/customer"
	rolehandler "user-service/handler/role"
	"user-service/middleware"
	customerroutes "user-service/routes/customer"
	roleroutes "user-service/routes/role"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter,
	adminHandler *adminhandler.Handler,
	roleHandler *rolehandler.Handler,
	customerHandler *customerhandler.Handler,
	authentication *middleware.Authentication) {
	admin := router.Group("/admin")
	admin.Use(authentication.Handle(), middleware.RequireRole(constants.RoleSuperAdmin))

	admin.GET("/check", adminHandler.Check)

	roleroutes.Register(admin, roleHandler)
	customerroutes.Register(admin, customerHandler)
}
