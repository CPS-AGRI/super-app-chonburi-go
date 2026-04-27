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
	err := r.db.Order("\"createdAt\" ASC").Find(&permissions).Error
	return permissions, err
}
