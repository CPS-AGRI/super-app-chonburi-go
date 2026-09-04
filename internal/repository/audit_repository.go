package repository

import (
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db}
}

func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	return r.db.Create(log).Error
}
