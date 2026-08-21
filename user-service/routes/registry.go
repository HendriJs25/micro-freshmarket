package routes

import (
	"user-service/handler"
	"user-service/middleware"
	adminroutes "user-service/routes/admin"
	healthroutes "user-service/routes/health"
	userroutes "user-service/routes/user"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	router         *gin.Engine
	handlers       *handler.Registry
	authentication *middleware.Authentication
}

func NewRegistry(router *gin.Engine, handlers *handler.Registry, authentication *middleware.Authentication) *Registry {
	return &Registry{
		router:         router,
		handlers:       handlers,
		authentication: authentication,
	}
}

func (r *Registry) Register() {
	healthroutes.Register(r.router, r.handlers.Health)
	userroutes.Register(r.router, r.handlers.User, r.authentication)
	adminroutes.Register(r.router, r.handlers.Admin, r.handlers.Role, r.authentication)
}
