package role

import (
	"context"
	"errors"
	"fmt"
	apperror "user-service/common/error"
	"user-service/domain/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type Repository interface {
	FindAll(context.Context, string) ([]model.Role, error)
	FindByID(context.Context, int64) (*model.Role, error)
	FindByName(context.Context, string) (*model.Role, error)
	FindByUserID(context.Context, int64) ([]model.Role, error)
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, search string) ([]model.Role, error) {
	var roles []model.Role

	query := r.db.WithContext(ctx).Model(&model.Role{})

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Order("id asc").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("find roles: %w", err)
	}

	return roles, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find role by id: %w", apperror.ErrNotFound)
		}

		return nil, fmt.Errorf("find role by id: %w", err)
	}

	return &role, nil
}

func (r *repository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role

	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find role by name: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("find role by name: %w", err)
	}
	return &role, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID int64) ([]model.Role, error) {
	var roles []model.Role

	err := r.db.WithContext(ctx).Model(&model.Role{}).Joins(
		`JOIN user_role 
				ON user_role.role_id = roles.id
				AND user_role.deleted_at IS NULL`).Where("user_role.user_id = ?", userID).Order("roles.id ASC").Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("find roles by user id: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("find roles by user id: %w", apperror.ErrNotFound)
	}

	return roles, nil
}
