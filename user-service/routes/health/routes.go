package health

import (
	healthhandler "user-service/handler/health"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRoutes, handler *healthhandler.Handler) {
	router.GET("/health", handler.Check)
}
