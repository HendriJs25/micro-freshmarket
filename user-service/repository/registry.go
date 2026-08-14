package repository

import (
	"user-service/repository/role"
	"user-service/repository/user"
	"user-service/repository/userrole"
	"user-service/repository/verificationtoken"

	"gorm.io/gorm"
)

type Registry struct {
	User              user.Repository
	Role              role.Repository
	UserRole          userrole.Repository
	VerificationToken verificationtoken.Repository
	Transaction       TransactionManager
}

func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{
		User:              user.NewRepository(db),
		Role:              role.NewRepository(db),
		UserRole:          userrole.NewRepository(db),
		VerificationToken: verificationtoken.NewRepository(db),
		Transaction:       NewTransactionManager(db),
	}
}
