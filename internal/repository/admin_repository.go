package repository

import (
	"super-app-chonburi-go/internal/domain"
	"gorm.io/gorm"
)

type AdminRepository interface {
	GetByEmail(email string) (*domain.Admin, error)
	GetAdminWithDepartment(adminID uint) (*domain.Admin, error)
	CreateAdmin(admin *domain.Admin) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Department").Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetAdminWithDepartment(adminID uint) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Department").First(&admin, adminID).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) CreateAdmin(admin *domain.Admin) error {
	return r.db.Create(admin).Error
}
