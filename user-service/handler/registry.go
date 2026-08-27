package handler

import (
	adminhandler "user-service/handler/admin"
	customerhandler "user-service/handler/customer"
	healthhandler "user-service/handler/health"
	rolehandler "user-service/handler/role"
	userhandler "user-service/handler/user"
	"user-service/services"
)

type Registry struct {
	Health   *healthhandler.Handler
	User     *userhandler.Handler
	Admin    *adminhandler.Handler
	Role     *rolehandler.Handler
	Customer *customerhandler.Handler
}

func NewRegistry(services *services.Registry) *Registry {
	return &Registry{
		Health:   healthhandler.NewHandler(),
		User:     userhandler.NewHandler(services.User),
		Admin:    adminhandler.NewHandler(),
		Role:     rolehandler.NewHandler(services.Role),
		Customer: customerhandler.NewHandler(services.Customer),
	}
}
