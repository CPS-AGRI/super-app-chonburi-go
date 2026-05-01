package repository

import (
	"super-app-chonburi-go/internal/domain"
	"gorm.io/gorm"
)

type systemPermissionRepository struct {
	db *gorm.DB
}

func NewSystemPermissionRepository(db *gorm.DB) domain.SystemPermissionRepository {
	return &systemPermissionRepository{db: db}
}

func (r *systemPermissionRepository) GetAll() ([]domain.SystemPermission, error) {
	var permissions []domain.SystemPermission
	err := r.db.Order("created_at ASC").Find(&permissions).Error
	return permissions, err
}
func (r *systemPermissionRepository) Create(p *domain.SystemPermission) error {
	return r.db.Create(p).Error
}

func (r *systemPermissionRepository) Update(p *domain.SystemPermission) error {
	return r.db.Save(p).Error
}

func (r *systemPermissionRepository) Delete(id string) error {
	return r.db.Delete(&domain.SystemPermission{}, "id = ?", id).Error
}
