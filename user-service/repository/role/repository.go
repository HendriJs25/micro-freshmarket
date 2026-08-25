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
	Create(context.Context, *model.Role) error
	Update(context.Context, int64, string) error
	FindAll(context.Context, string) ([]model.Role, error)
	FindByID(context.Context, int64) (*model.Role, error)
	FindByName(context.Context, string) (*model.Role, error)
	FindByUserID(context.Context, int64) ([]model.Role, error)
	DeleteIfUnused(context.Context, int64) error
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, role *model.Role) error {
	err := r.db.WithContext(ctx).Create(role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create role: %w", apperror.ErrAlreadyExists)
		}

		return fmt.Errorf("create role: %w", err)
	}

	return nil
}

func (r *repository) Update(ctx context.Context, id int64, roleName string) error {
	result := r.db.WithContext(ctx).Model(model.Role{}).Where("id = ?", id).Update("name", roleName)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("update role name: %w", apperror.ErrAlreadyExists)
		}
		return fmt.Errorf("update role name: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update role name: %w", apperror.ErrNotFound)
	}
	return nil
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

func (r *repository) DeleteIfUnused(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).
		Where(`NOT EXISTS (SELECT 1 FROM user_role ur WHERE ur.role_id = roles.id AND ur.deleted_at IS NULL)`).Delete(&model.Role{})

	if result.Error != nil {
		return fmt.Errorf("delete role: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		return nil
	}

	_, err := r.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return fmt.Errorf("delete role: %w", apperror.ErrNotFound)
		}
		return fmt.Errorf("check role after failed delete: %w", err)
	}

	return fmt.Errorf("delete role: %w", apperror.ErrConflict)

}
