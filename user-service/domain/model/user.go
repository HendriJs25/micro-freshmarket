package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string `gorm:"column:password"`
	Address      *string
	Phone        *string
	Photo        *string
	Lat          *string
	Lng          *string
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    gorm.DeletedAt

	Roles []Role `gorm:"many2many:user_role;"`
}

func (User) TableName() string {
	return "users"
}
