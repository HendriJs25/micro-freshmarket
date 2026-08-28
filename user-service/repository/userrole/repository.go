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
	DeleteByUserID(context.Context, int64) error
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

func (r *repository) DeleteByUserID(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user id must be greater than zero", apperror.ErrInvalidArgument)
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("delete user role by user id: %w", err)
	}
	return nil
}
