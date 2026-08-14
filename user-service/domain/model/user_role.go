package model

import (
	"time"

	"gorm.io/gorm"
)

type UserRole struct {
	ID        int64
	UserID    int64
	RoleID    int64
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt
}

func (UserRole) TableName() string {
	return "user_role"
}
