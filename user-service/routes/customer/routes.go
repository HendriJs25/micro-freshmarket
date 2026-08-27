package customer

import (
	customerhandler "user-service/handler/customer"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, handler *customerhandler.Handler) {
	router.GET("/customers", handler.GetAll)
	router.GET("/customers/:id", handler.GetByID)
}
