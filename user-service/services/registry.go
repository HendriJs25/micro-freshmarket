package services

import (
	"user-service/repository"
	userService "user-service/services/user"
)

type Registry struct {
	User userService.Service
}

func NewRegistry(repositories *repository.Registry) *Registry {
	return &Registry{
		User: userService.NewService(
			repositories.User),
	}
}
