package routes

import (
	"user-service/handler"
	healthroutes "user-service/routes/health"
	userroutes "user-service/routes/user"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	router   *gin.Engine
	handlers *handler.Registry
}

func NewRegistry(router *gin.Engine, handlers *handler.Registry) *Registry {
	return &Registry{
		router:   router,
		handlers: handlers,
	}
}

func (r *Registry) Register() {
	healthroutes.Register(r.router, r.handlers.Health)
	userroutes.Register(r.router, r.handlers.User)
}
