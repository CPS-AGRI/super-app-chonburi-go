package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Department").First(&admin, "email = ?", email).Error
	return &admin, err
}

func (r *adminRepository) GetByID(id uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Department").First(&admin, "id = ?", id).Error
	return &admin, err
}
