package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type transactionManager struct {
	db *gorm.DB
}

type TransactionManager interface {
	WithinTransaction(context.Context, func(*Registry) error) error
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{
		db: db,
	}
}

func (m *transactionManager) WithinTransaction(ctx context.Context, fn func(*Registry) error) error {
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactionRepository := NewRegistry(tx)

		return fn(transactionRepository)
	})

	if err != nil {
		return fmt.Errorf("execute database transaction: %w", err)
	}

	return nil
}
