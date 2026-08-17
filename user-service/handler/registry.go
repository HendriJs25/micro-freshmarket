package handler

import (
	healthhandler "user-service/handler/health"
	userhandler "user-service/handler/user"
	"user-service/services"
)

type Registry struct {
	Health *healthhandler.Handler
	User   *userhandler.Handler
}

func NewRegistry(services *services.Registry) *Registry {
	return &Registry{
		Health: healthhandler.NewHandler(),
		User:   userhandler.NewHandler(services.User),
	}
}
