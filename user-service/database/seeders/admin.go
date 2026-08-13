package seeders

import (
	"errors"
	"fmt"
	"user-service/config"
	"user-service/domain/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(tx *gorm.DB, superAdminRole model.Role, cfg config.Seed) error {
	var admin model.User

	err := tx.Unscoped().Where("email = ?", cfg.AdminEmail).First(&admin).Error

	switch {
	case err == nil:
		updates := map[string]any{}

		if admin.DeletedAt.Valid {
			updates["deleted_at"] = nil
		}

		if !admin.IsVerified {
			updates["is_verified"] = true
		}

		if len(updates) > 0 {
			if err := tx.Unscoped().Model(&admin).Updates(updates).Error; err != nil {
				return fmt.Errorf("restore admin user: %w", err)
			}
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generate password hash: %w", err)
		}

		admin = model.User{
			Name:         cfg.AdminName,
			Email:        cfg.AdminEmail,
			PasswordHash: string(passwordHash),
			IsVerified:   true,
		}

		if err := tx.Create(&admin).Error; err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
	default:
		return fmt.Errorf("find admin user: %w", err)
	}

	if err := ensureActiveUserRole(tx, admin.ID, superAdminRole.ID); err != nil {
		return err
	}

	return nil
}

func ensureActiveUserRole(tx *gorm.DB, userID int64, roleID int64) error {
	var userRole model.UserRole

	err := tx.Where("user_id = ? AND role_id = ?", userID, roleID).First(&userRole).Error

	switch {
	case err == nil:
		return nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("find admin role assignment: %w", err)
	}

	userRole = model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := tx.Create(&userRole).Error; err != nil {
		return fmt.Errorf("create admin role assignment: %w", err)
	}

	return nil
}
