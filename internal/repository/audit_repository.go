package repository

import (
	"super-app-chonburi-go/internal/domain"
	"gorm.io/gorm"
)

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) domain.ActivityLogRepository {
	return &activityLogRepository{db}
}

func (r *activityLogRepository) Create(log *domain.ActivityLog) error {
	return r.db.Create(log).Error
}
