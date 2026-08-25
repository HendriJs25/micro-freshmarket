package role

import (
	"context"
	"fmt"
	"strings"
	apperror "user-service/common/error"
	"user-service/domain/model"
	rolerepository "user-service/repository/role"
)

type CreateInput struct {
	Name string
}

type UpdateInput struct {
	ID   int64
	Name string
}

type Role struct {
	ID   int64
	Name string
}

type service struct {
	roleRepository rolerepository.Repository
}

type Service interface {
	Create(context.Context, CreateInput) error
	Update(ctx context.Context, input UpdateInput) error
	GetAll(context.Context, string) ([]Role, error)
	GetByID(context.Context, int64) (*Role, error)
	Delete(context.Context, int64) error
}

func NewService(roleRepository rolerepository.Repository) Service {
	return &service{
		roleRepository: roleRepository,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) error {
	name := strings.TrimSpace(input.Name)

	if name == "" {
		return fmt.Errorf("%w: role name must not be empty", apperror.ErrInvalidArgument)
	}

	role := &model.Role{
		Name: name,
	}

	if err := s.roleRepository.Create(ctx, role); err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	return nil
}

func (s *service) Update(ctx context.Context, input UpdateInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("%w: role id must be greater than zero", apperror.ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("%w: role name must not be empty", apperror.ErrInvalidArgument)
	}

	err := s.roleRepository.Update(ctx, input.ID, name)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	return nil
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

func (s *service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w:role id must be greater than zero", apperror.ErrInvalidArgument)
	}

	if err := s.roleRepository.DeleteIfUnused(ctx, id); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	return nil
}
