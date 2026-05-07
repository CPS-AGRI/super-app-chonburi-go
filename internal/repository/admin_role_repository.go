package repository

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"gorm.io/gorm"
)

type adminRoleRepository struct {
	db *gorm.DB
}

func NewAdminRoleRepository(db *gorm.DB) domain.AdminRoleRepository {
	return &adminRoleRepository{db: db}
}

func (r *adminRoleRepository) GetAll() ([]domain.AdminRole, error) {
	var roles []domain.AdminRole
	err := r.db.Order("name_th ASC").Find(&roles).Error
	return roles, err
}

func (r *adminRoleRepository) GetByID(id string) (*domain.AdminRole, error) {
	var role domain.AdminRole
	err := r.db.First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}
func (r *adminRoleRepository) Create(role *domain.AdminRole) error {
	return r.db.Create(role).Error
}

func (r *adminRoleRepository) Update(role *domain.AdminRole) error {
	return r.db.Save(role).Error
}

func (r *adminRoleRepository) Delete(id string) error {
	return r.db.Delete(&domain.AdminRole{}, "id = ?", id).Error
}
