package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
	apperror "user-service/common/error"
	"user-service/constants"
	"user-service/domain/model"
	"user-service/repository"
	userRepository "user-service/repository/user"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72

	emailVerificationTokenTTL = time.Hour
)

type AuthenticateInput struct {
	Email    string
	Password string
}

type AuthenticatedUser struct {
	ID    int64
	Name  string
	Email string
}
type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

type service struct {
	userRepository     userRepository.Repository
	transactionManager repository.TransactionManager
}

type Service interface {
	GetByID(context.Context, int64) (*model.User, error)
	GetByEmail(context.Context, string) (*model.User, error)
	SignUp(context.Context, SignUpInput) error
	VerifyAccount(context.Context, string) error
	Authenticate(context.Context, AuthenticateInput) (*AuthenticatedUser, error)
}

func NewService(userRepository userRepository.Repository, transactionManager repository.TransactionManager) Service {
	return &service{
		userRepository:     userRepository,
		transactionManager: transactionManager,
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

func (s *service) SignUp(ctx context.Context, input SignUpInput) error {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)

	if name == "" {
		return fmt.Errorf("%w: user name must not be empty", apperror.ErrInvalidArgument)
	}

	if email == "" {
		return fmt.Errorf("%w: user email must not be empty", apperror.ErrInvalidArgument)
	}

	if err := validatePassword(input.Password); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash user password: %w", err)
	}

	verificationToken := generateVerificationToken()

	expiresAt := time.Now().UTC().Add(emailVerificationTokenTTL)

	err = s.transactionManager.WithinTransaction(ctx, func(repositories *repository.Registry) error {
		customerRole, err := repositories.Role.FindByName(ctx, constants.RoleCustomer)
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return fmt.Errorf("%w: required role %q not found", apperror.ErrMisconfigured, constants.RoleCustomer)
			}
			return fmt.Errorf("find customer role: %w", err)
		}

		user := &model.User{
			Name:         name,
			Email:        email,
			PasswordHash: string(passwordHash),
			IsVerified:   false,
		}

		if err := repositories.User.Create(ctx, user); err != nil {
			return fmt.Errorf("create signup user: %w", err)
		}

		userRole := &model.UserRole{
			UserID: user.ID,
			RoleID: customerRole.ID,
		}

		if err := repositories.UserRole.Create(ctx, userRole); err != nil {
			return fmt.Errorf("assign customer role: %w", err)
		}

		token := &model.VerificationToken{
			UserID:    user.ID,
			Token:     verificationToken,
			TokenType: constants.TokenTypeEmailVerification,
			ExpiresAt: expiresAt,
		}

		if err := repositories.VerificationToken.Create(ctx, token); err != nil {
			if errors.Is(err, apperror.ErrAlreadyExists) {
				return fmt.Errorf("generated verification token already exists")
			}
			return fmt.Errorf("create verification token: %w", err)
		}

		return nil

	})

	if err != nil {
		return fmt.Errorf("sign up user: %w", err)
	}

	return nil
}

func (s *service) VerifyAccount(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("%w: verification token must not be empty", apperror.ErrInvalidToken)
	}

	err := s.transactionManager.WithinTransaction(ctx, func(repositories *repository.Registry) error {
		verificationToken, err := repositories.VerificationToken.FindByToken(ctx, token)
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return fmt.Errorf("%w: verification token not found", apperror.ErrInvalidToken)
			}
			return fmt.Errorf("find verification token: %w", err)
		}

		if verificationToken.TokenType != constants.TokenTypeEmailVerification {
			return fmt.Errorf("%w: unexpected verification token type", apperror.ErrInvalidToken)
		}

		now := time.Now().UTC()

		if !now.Before(verificationToken.ExpiresAt) {
			return fmt.Errorf("%w: verification token has expired", apperror.ErrInvalidToken)
		}

		if err := repositories.User.MarkVerified(ctx, verificationToken.UserID); err != nil {
			return fmt.Errorf("verify token user: %w", err)
		}

		if err := repositories.VerificationToken.DeleteByID(ctx, verificationToken.ID); err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return fmt.Errorf("%w: verification token already used", apperror.ErrInvalidToken)
			}
			return fmt.Errorf("invalidate verification token: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("verify account: %w", err)
	}
	return nil
}

func (s *service) Authenticate(ctx context.Context, input AuthenticateInput) (*AuthenticatedUser, error) {
	email := normalizeEmail(input.Email)

	if email == "" || input.Password == "" {
		return nil, fmt.Errorf("%w: email and password must not be empty", apperror.ErrInvalidCredentials)
	}

	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, fmt.Errorf("%w", apperror.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("find authentication user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, fmt.Errorf("%w", apperror.ErrInvalidCredentials)
		}

		return nil, fmt.Errorf("compare user password hash: %w", err)
	}

	if !user.IsVerified {
		return nil, fmt.Errorf("%w", apperror.ErrAccountNotVerified)
	}

	return &AuthenticatedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func generateVerificationToken() string {
	return rand.Text()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validatePassword(password string) error {
	passwordLength := len([]byte(password))

	if passwordLength < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d bytes", apperror.ErrInvalidArgument, minPasswordLength)
	}

	if passwordLength > maxPasswordLength {
		return fmt.Errorf("%w: password must not exceed %d bytes", apperror.ErrInvalidArgument, maxPasswordLength)
	}

	return nil
}
