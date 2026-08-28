package customer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	apperror "user-service/common/error"
	"user-service/constants"
	"user-service/domain/model"
	"user-service/repository"
	customerrepository "user-service/repository/customer"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	defaultPage          int64 = 1
	defaultLimit         int64 = 10
	maxLimit             int64 = 100
	minimumPasswordBytes       = 8
	maximumPasswordBytes       = 72
)

type CreateInput struct {
	Name     string
	Email    string
	Password string
	Phone    string
	Address  string
	Lat      *float64
	Lng      *float64
	Photo    string
	RoleID   int64
}

type ListInput struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderType string
}

type UpdateInput struct {
	ID int64

	Name  string
	Email string
	Phone string

	Address *string
	Photo   *string

	Lat *float64
	Lng *float64
}

type CustomerListItem struct {
	ID    int64
	Name  string
	Email string
	Phone string
	Photo string
}

type Customer struct {
	ID      int64
	RoleID  int64
	Name    string
	Email   string
	Phone   string
	Address string
	Photo   string
	Lat     string
	Lng     string
}

type Pagination struct {
	Page       int64
	TotalCount int64
	PerPage    int64
	TotalPage  int64
}

type ListResult struct {
	Customers  []CustomerListItem
	Pagination Pagination
}

type service struct {
	customerRepository customerrepository.Repository
	transactionManager repository.TransactionManager
}

type Service interface {
	Create(context.Context, CreateInput) error
	GetAll(context.Context, ListInput) (*ListResult, error)
	GetByID(context.Context, int64) (*Customer, error)
	Update(context.Context, UpdateInput) error
}

func NewService(customerRepository customerrepository.Repository, transactionManager repository.TransactionManager) Service {
	return &service{
		customerRepository: customerRepository,
		transactionManager: transactionManager,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) error {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)

	if name == "" || email == "" || phone == "" {
		return fmt.Errorf("%w: required customer data must not be empty", apperror.ErrInvalidArgument)
	}

	passwordLength := len([]byte(input.Password))

	if passwordLength < minimumPasswordBytes || passwordLength > maximumPasswordBytes {
		return fmt.Errorf("%w: password msut be between %d and %d bytes", apperror.ErrInvalidArgument, minimumPasswordBytes, maximumPasswordBytes)
	}

	if input.RoleID <= 0 {
		return fmt.Errorf("%w: role_id must be greater than zero", apperror.ErrInvalidArgument)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf("hash customer password: %w", err)
	}

	err = s.transactionManager.WithinTransaction(ctx, func(repositories *repository.Registry) error {
		customerRole, err := repositories.Role.FindByName(ctx, constants.RoleCustomer)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: customer role is missing", apperror.ErrMisconfigured)
			}

			return fmt.Errorf("find customer role: %w", err)
		}

		if input.RoleID != customerRole.ID {
			return fmt.Errorf("%w: role id must reference Customer role", apperror.ErrInvalidArgument)
		}

		customer := &model.User{
			Name:         name,
			Email:        email,
			PasswordHash: string(passwordHash),
			Phone:        optionalString(phone),
			Address:      optionalString(input.Address),
			Photo:        optionalString(input.Photo),
			Lat:          coordinateString(input.Lat),
			Lng:          coordinateString(input.Lng),
			IsVerified:   true,
		}

		if err := repositories.User.Create(ctx, customer); err != nil {
			return fmt.Errorf("create customer user: %w", err)
		}

		userRole := &model.UserRole{
			UserID: customer.ID,
			RoleID: customerRole.ID,
		}

		if err := repositories.UserRole.Create(ctx, userRole); err != nil {
			return fmt.Errorf("assign customer role: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	return nil
}

func (s *service) GetAll(ctx context.Context, input ListInput) (*ListResult, error) {
	search := strings.TrimSpace(input.Search)
	page := normalizePage(input.Page)
	limit := normalizeLimit(input.Limit)
	orderBy, err := normalizeOrderBy(input.OrderBy)
	if err != nil {
		return nil, err
	}

	desc := normalizeOrderType(input.OrderType)

	offset := (page - 1) * limit

	items, totalCount, err := s.customerRepository.FindAll(ctx, customerrepository.ListQuery{
		Search:  search,
		Limit:   int(limit),
		Offset:  int(offset),
		OrderBy: orderBy,
		Desc:    desc,
	})
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}

	customers := make([]CustomerListItem, 0, len(items))
	for _, item := range items {
		customers = append(customers, CustomerListItem{
			ID:    item.ID,
			Name:  item.Name,
			Email: item.Email,
			Phone: stringValue(item.Phone),
			Photo: stringValue(item.Photo),
		})
	}

	totalPage := calculateTotalPage(totalCount, limit)

	return &ListResult{
		Customers: customers,
		Pagination: Pagination{
			Page:       page,
			TotalCount: totalCount,
			PerPage:    limit,
			TotalPage:  totalPage,
		},
	}, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*Customer, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: customer id must be greater than zero", apperror.ErrInvalidArgument)
	}

	result, err := s.customerRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get customer by id: %w", err)
	}

	return &Customer{
		ID:      result.ID,
		RoleID:  result.RoleID,
		Name:    result.Name,
		Email:   result.Email,
		Phone:   stringValue(result.Phone),
		Address: stringValue(result.Address),
		Photo:   stringValue(result.Photo),
		Lat:     stringValue(result.Lat),
		Lng:     stringValue(result.Lng),
	}, nil
}

func (s *service) Update(ctx context.Context, input UpdateInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("%w: customer id must be greater than zero", apperror.ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)

	if name == "" || email == "" || phone == "" {
		return fmt.Errorf("%w: name, email and phone are required", apperror.ErrInvalidArgument)
	}

	if err := s.customerRepository.Update(ctx, input.ID, customerrepository.UpdateFields{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: normalizeOptionalString(input.Address),
		Photo:   normalizeOptionalString(input.Photo),
		Lat:     coordinateString(input.Lat),
		Lng:     coordinateString(input.Lng),
	}); err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}

func calculateTotalPage(totalCount int64, perPage int64) int64 {
	if totalCount == 0 {
		return 0
	}

	return (totalCount + perPage - 1) / perPage
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func normalizePage(page int64) int64 {
	if page <= 0 {
		return defaultPage
	}

	return page
}

func normalizeLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultLimit
	}

	if limit > maxLimit {
		return maxLimit
	}

	return limit
}

func normalizeOrderType(orderType string) bool {
	switch strings.ToLower(strings.TrimSpace(orderType)) {
	case "asc":
		return false
	case "desc":
		return true
	default:
		return true
	}
}

func normalizeOrderBy(orderBy string) (customerrepository.OrderBy, error) {
	switch strings.ToLower(strings.TrimSpace(orderBy)) {
	case "":
		return customerrepository.OrderByCreatedAt, nil
	case "id":
		return customerrepository.OrderByID, nil
	case "name":
		return customerrepository.OrderByName, nil
	case "email":
		return customerrepository.OrderByEmail, nil
	case "phone":
		return customerrepository.OrderByPhone, nil
	case "created_at":
		return customerrepository.OrderByCreatedAt, nil
	case "updated_at":
		return customerrepository.OrderByUpdatedAt, nil
	default:
		return "", fmt.Errorf("%w: unsupported customer order field", apperror.ErrInvalidArgument)
	}
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	return &normalized
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}
	return &value
}

func coordinateString(value *float64) *string {
	if value == nil {
		return nil
	}

	result := strconv.FormatFloat(*value, 'g', -1, 64)
	return &result
}
