package customer

import (
	"context"
	"errors"
	"fmt"
	apperror "user-service/common/error"
	"user-service/constants"
	"user-service/domain/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderBy string

const (
	OrderByID        OrderBy = "id"
	OrderByName      OrderBy = "name"
	OrderByEmail     OrderBy = "email"
	OrderByPhone     OrderBy = "phone"
	OrderByCreatedAt OrderBy = "created_at"
	OrderByUpdatedAt OrderBy = "updated_at"
)

type ListQuery struct {
	Search  string
	Limit   int
	Offset  int
	OrderBy OrderBy
	Desc    bool
}

type ListItem struct {
	ID    int64
	Name  string
	Email string
	Phone *string
	Photo *string
}

type Detail struct {
	ID      int64
	RoleID  int64
	Name    string
	Email   string
	Phone   *string
	Address *string
	Photo   *string
	Lat     *string
	Lng     *string
}

type UpdateFields struct {
	Name  string
	Email string
	Phone string

	Address *string
	Photo   *string
	Lat     *string
	Lng     *string
}

type repository struct {
	db *gorm.DB
}

type Repository interface {
	FindAll(context.Context, ListQuery) ([]ListItem, int64, error)
	FindByID(context.Context, int64) (*Detail, error)
	Update(context.Context, int64, UpdateFields) error
	Delete(context.Context, int64) error
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) FindAll(ctx context.Context, query ListQuery) ([]ListItem, int64, error) {
	var totalCount int64

	if err := r.customerQuery(ctx, query.Search).Distinct("users.id").Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	customers := make([]ListItem, 0)

	listQuery := r.customerQuery(ctx, query.Search).
		Select(`users.id,
					users.name,
					users.email,
					users.phone,
					users.photo`).
		Order(clause.OrderByColumn{
			Column: orderColumn(query.OrderBy),
			Desc:   query.Desc,
		})

	if query.OrderBy != OrderByID {
		listQuery = listQuery.Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: "users",
				Name:  "id",
			},
			Desc: query.Desc,
		})
	}

	if err := listQuery.Limit(query.Limit).Offset(query.Offset).Scan(&customers).Error; err != nil {
		return nil, 0, fmt.Errorf("find customers: %w", err)
	}
	return customers, totalCount, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Detail, error) {
	var customer Detail

	err := r.customerQuery(ctx, "").
		Select(`users.id,
						ur.role_id,
						users.name,
						users.email,
						users.phone,
						users.address,
						users.photo,
						users.lat,
						users.lng`).
		Where("users.id = ?", id).Take(&customer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find customer by id: %w", apperror.ErrNotFound)
		}

		return nil, fmt.Errorf("find customer by id: %w", err)
	}

	return &customer, nil
}

func (r *repository) Update(ctx context.Context, id int64, updateFields UpdateFields) error {
	updates := map[string]any{
		"name":  updateFields.Name,
		"email": updateFields.Email,
		"phone": updateFields.Phone,
	}

	if updateFields.Address != nil {
		updates["address"] = updateFields.Address
	}

	if updateFields.Photo != nil {
		updates["photo"] = updateFields.Photo
	}

	if updateFields.Lat != nil {
		updates["lat"] = updateFields.Lat
	}

	if updateFields.Lng != nil {
		updates["lng"] = updateFields.Lng
	}

	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).
		Where(`EXISTS (
					SELECT 1 
					FROM user_role ur 
					JOIN roles r 
					ON r.id = ur.role_id 
					AND r.deleted_at IS NULL 
					WHERE ur.user_id = users.id 
					AND ur.deleted_at IS NULL 
					AND r.name = ?)`, constants.RoleCustomer).Updates(updates)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("update customer: %w", apperror.ErrAlreadyExists)
		}
		return fmt.Errorf("update customer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update customer: %w", apperror.ErrNotFound)
	}
	return nil
}

func (r *repository) customerQuery(ctx context.Context, search string) *gorm.DB {
	query := r.db.WithContext(ctx).Table("users").
		Joins(`JOIN user_role ur
					ON ur.user_id = users.id
					AND ur.deleted_at IS NULL`).
		Joins(`JOIN roles r
					ON r.id = ur.role_id
					AND r.deleted_at IS NULL`).
		Where("users.deleted_at IS NULL").
		Where("r.name = ?", constants.RoleCustomer)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(`(users.name ILIKE ? OR users.email ILIKE ? OR users.phone ILIKE ?)`, searchPattern, searchPattern, searchPattern)
	}

	return query
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).
		Where(`EXISTS (
					SELECT 1
					FROM user_role ur 
					JOIN roles r ON r.id = ur.role_id 
					AND r.deleted_at IS NULL 
					WHERE ur.user_id = users.id 
					AND ur.deleted_at IS NULL 
					AND r.name = ?)`, constants.RoleCustomer).Delete(&model.User{})
	if result.Error != nil {
		return fmt.Errorf("delete customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete customer: %w", apperror.ErrNotFound)
	}

	return nil
}

func orderColumn(orderBy OrderBy) clause.Column {
	switch orderBy {
	case OrderByID:
		return clause.Column{
			Table: "users",
			Name:  "id",
		}
	case OrderByName:
		return clause.Column{
			Table: "users",
			Name:  "name",
		}
	case OrderByEmail:
		return clause.Column{
			Table: "users",
			Name:  "email",
		}
	case OrderByPhone:
		return clause.Column{
			Table: "users",
			Name:  "phone",
		}
	case OrderByUpdatedAt:
		return clause.Column{
			Table: "users",
			Name:  "created_at",
		}
	case OrderByCreatedAt:
		fallthrough
	default:
		return clause.Column{
			Table: "users",
			Name:  "created_at",
		}
	}
}
