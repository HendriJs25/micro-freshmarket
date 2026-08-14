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
	FindByName(context.Context, string) (*model.Role, error)
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
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
