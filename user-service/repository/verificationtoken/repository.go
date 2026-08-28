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
	FindByToken(context.Context, string) (*model.VerificationToken, error)
	DeleteByID(context.Context, int64) error
	DeleteByUserID(context.Context, int64) error
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

func (r *repository) FindByToken(ctx context.Context, token string) (*model.VerificationToken, error) {
	var verificationToken model.VerificationToken

	err := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find verification token: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("find verification token: %w", err)
	}

	return &verificationToken, nil
}

func (r *repository) DeleteByID(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.VerificationToken{})

	if result.Error != nil {
		return fmt.Errorf("delete verification token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete verification token: %w", apperror.ErrNotFound)
	}

	return nil
}

func (r *repository) DeleteByUserID(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user id must be greater than zero", apperror.ErrInvalidArgument)
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.VerificationToken{}).Error; err != nil {
		return fmt.Errorf("delete verification token: %w", err)
	}

	return nil
}
