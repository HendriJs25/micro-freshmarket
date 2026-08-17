package user

import (
	userhandler "user-service/handler/user"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, handler *userhandler.Handler) {
	router.POST("/signup", handler.SignUp)
	router.GET("/verify-account", handler.VerifyAccount)
}
