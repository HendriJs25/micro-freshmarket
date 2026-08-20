package handler

import (
	adminhandler "user-service/handler/admin"
	healthhandler "user-service/handler/health"
	userhandler "user-service/handler/user"
	"user-service/services"
)

type Registry struct {
	Health *healthhandler.Handler
	User   *userhandler.Handler
	Admin  *adminhandler.Handler
}

func NewRegistry(services *services.Registry) *Registry {
	return &Registry{
		Health: healthhandler.NewHandler(),
		User:   userhandler.NewHandler(services.User),
		Admin:  adminhandler.NewHandler(),
	}
}
