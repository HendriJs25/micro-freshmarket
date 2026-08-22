package role

import (
	rolehandler "user-service/handler/role"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, handler *rolehandler.Handler) {
	router.GET("/roles", handler.GetAll)
	router.GET("/roles/:id", handler.GetByID)
	router.POST("/roles", handler.Create)
}
