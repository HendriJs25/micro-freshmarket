package customer

import (
	"context"
	"fmt"
	"strings"
	apperror "user-service/common/error"
	customerrepository "user-service/repository/customer"
)

const (
	defaultPage  int64 = 1
	defaultLimit int64 = 10
	maxLimit     int64 = 100
)

type ListInput struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderType string
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
}

type Service interface {
	GetAll(context.Context, ListInput) (*ListResult, error)
	GetByID(context.Context, int64) (*Customer, error)
}

func NewService(customerRepository customerrepository.Repository) Service {
	return &service{
		customerRepository: customerRepository,
	}
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
