package user

import (
	userhandler "user-service/handler/user"
	"user-service/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, handler *userhandler.Handler, authentication *middleware.Authentication) {
	router.POST("/signup", handler.SignUp)
	router.GET("/verify-account", handler.VerifyAccount)
	router.POST("/signin", handler.SignIn)

	authenticated := router.Group("/auth")
	authenticated.Use(authentication.Handle())
	authenticated.GET("/profile", handler.GetProfile)
}
