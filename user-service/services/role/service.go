package role

import (
	"context"
	"fmt"
	"strings"
	apperror "user-service/common/error"
	rolerepository "user-service/repository/role"
)

type Role struct {
	ID   int64
	Name string
}

type service struct {
	roleRepository rolerepository.Repository
}

type Service interface {
	GetAll(context.Context, string) ([]Role, error)
	GetByID(context.Context, int64) (*Role, error)
}

func NewService(roleRepository rolerepository.Repository) Service {
	return &service{
		roleRepository: roleRepository,
	}
}

func (s *service) GetAll(ctx context.Context, search string) ([]Role, error) {
	search = strings.TrimSpace(search)

	modelRoles, err := s.roleRepository.FindAll(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}

	roles := make([]Role, 0, len(modelRoles))

	for _, modelRole := range modelRoles {
		roles = append(roles, Role{
			ID:   modelRole.ID,
			Name: modelRole.Name,
		})
	}

	return roles, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*Role, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w:role id must be greater than zero", apperror.ErrInvalidArgument)
	}

	modelRole, err := s.roleRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get role by id: %w", err)
	}

	return &Role{
		ID:   modelRole.ID,
		Name: modelRole.Name,
	}, nil

}
