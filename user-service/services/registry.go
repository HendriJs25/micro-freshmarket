package services

import (
	"fmt"
	"user-service/config"
	"user-service/repository"
	sessionrepository "user-service/repository/session"
	jwtservice "user-service/services/jwt"
	userservice "user-service/services/user"
)

type Registry struct {
	User userservice.Service
	JWT  jwtservice.Service
}

func NewRegistry(repositories *repository.Registry, jwtConfig config.JWT, sessionRepository sessionrepository.Repository) (*Registry, error) {
	jwtService, err := jwtservice.NewService(jwtConfig)
	if err != nil {
		return nil, fmt.Errorf("create jwt service: %w", err)
	}

	return &Registry{
		User: userservice.NewService(
			repositories.User,
			repositories.Role,
			repositories.Transaction,
			jwtService,
			sessionRepository),
		JWT: jwtService,
	}, nil
}
