package verificationtoken

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
	Create(context.Context, *model.VerificationToken) error
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, token *model.VerificationToken) error {
	err := r.db.WithContext(ctx).Create(token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create verification token: %w", apperror.ErrAlreadyExists)
		}
		return fmt.Errorf("create verification token: %w", err)
	}

	return nil
}
