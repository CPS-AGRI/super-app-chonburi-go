package repository

import (
	"errors"

	"github.com/google/uuid"
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
	err := r.db.Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *adminRoleRepository) GetByID(id uuid.UUID) (*domain.AdminRole, error) {
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
