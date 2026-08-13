package seeders

import (
	"errors"
	"fmt"
	"user-service/constants"
	"user-service/domain/model"

	"gorm.io/gorm"
)

func seedRoles(tx *gorm.DB) (model.Role, error) {
	roleNames := []string{
		constants.RoleSuperAdmin,
		constants.RoleCustomer,
	}

	var superAdminRole model.Role

	for _, name := range roleNames {
		role, err := findOrCreateRole(tx, name)
		if err != nil {
			return model.Role{}, err
		}

		if role.Name == constants.RoleSuperAdmin {
			superAdminRole = role
		}
	}

	if superAdminRole.ID == 0 {
		return model.Role{}, fmt.Errorf("super admin role was not seeded")
	}

	return superAdminRole, nil
}

func findOrCreateRole(tx *gorm.DB, name string) (model.Role, error) {
	var role model.Role

	err := tx.Unscoped().Where("name = ?", name).First(&role).Error

	switch {
	case err == nil:
		if role.DeletedAt.Valid {
			if err := tx.Unscoped().Model(&role).Update("deleted_at", nil).Error; err != nil {
				return model.Role{}, fmt.Errorf("restore role %q: %w", name, err)
			}
			role.DeletedAt = gorm.DeletedAt{}
		}
		return role, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		role = model.Role{
			Name: name,
		}
		if err := tx.Create(&role).Error; err != nil {
			return model.Role{}, fmt.Errorf("Create role %q: %w", name, err)
		}
		return role, nil

	default:
		return model.Role{}, fmt.Errorf("find role %q: %w", name, err)
	}

}
