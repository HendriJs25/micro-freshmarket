package user

import (
	"context"
	"fmt"
	"strings"
	apperror "user-service/common/error"
	"user-service/domain/model"
	userRepository "user-service/repository/user"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	maxPasswordLenght = 72
)

type CreateInput struct {
	Name     string
	Email    string
	Password string
}

type service struct {
	userRepository userRepository.Repository
}

type Service interface {
	GetByID(context.Context, int64) (*model.User, error)
	GetByEmail(context.Context, string) (*model.User, error)
	Create(context.Context, CreateInput) (*model.User, error)
}

func NewService(userRepository userRepository.Repository) Service {
	return &service{
		userRepository: userRepository,
	}
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if id < 1 {
		return nil, fmt.Errorf("%w: user id must be greater than zero", apperror.ErrInvalidArgument)
	}

	user, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	email = normalizeEmail(email)

	if email == "" {
		return nil, fmt.Errorf("%w: user email must not be empty", apperror.ErrInvalidArgument)
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*model.User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)

	if name == "" {
		return nil, fmt.Errorf("%w: user name must not be empty", apperror.ErrInvalidArgument)
	}

	if email == "" {
		return nil, fmt.Errorf("%w: user email must not be empty", apperror.ErrInvalidArgument)
	}

	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash user password: %w", err)
	}

	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		IsVerified:   false,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validatePassword(password string) error {
	passwordLength := len([]byte(password))

	if passwordLength < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d bytes", apperror.ErrInvalidArgument, minPasswordLength)
	}

	if passwordLength > maxPasswordLenght {
		return fmt.Errorf("%w: password must not exceed %d bytes", apperror.ErrInvalidArgument, maxPasswordLenght)
	}

	return nil
}
