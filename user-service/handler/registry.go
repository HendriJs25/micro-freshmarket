package handler

import healthhandler "user-service/handler/health"

type Registry struct {
	Health *healthhandler.Handler
}

func NewRegistry() *Registry {
	return &Registry{
		Health: healthhandler.NewHandler(),
	}
}
