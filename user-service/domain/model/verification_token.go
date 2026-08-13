package model

import (
	"time"

	"gorm.io/gorm"
)

type VerificationToken struct {
	ID        int64
	UserID    int64
	Token     string
	TokenType string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt

	User User `gorm:"foreignKey:UserID"`
}

func (VerificationToken) TableName() string {
	return "verification_tokens"
}
