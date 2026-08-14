package userrole

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
	Create(context.Context, *model.UserRole) error
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, userRole *model.UserRole) error {
	err := r.db.WithContext(ctx).Create(userRole).Error

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create user role: %w", apperror.ErrAlreadyExists)
		}
		return fmt.Errorf("create user role: %w", err)
	}

	return nil
}
