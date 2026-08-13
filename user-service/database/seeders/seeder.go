package seeders

import (
	"context"
	"fmt"
	"user-service/config"

	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, cfg config.Seed) error {
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		superAdminRole, err := seedRoles(tx)
		if err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}

		if err := SeedAdmin(tx, superAdminRole, cfg); err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("run database seeders: %w", err)
	}

	return nil
}
