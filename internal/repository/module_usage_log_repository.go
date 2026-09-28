package repository

import (
	"time"

	"gorm.io/gorm"
	"super-app-chonburi-go/internal/domain"
)

type moduleUsageLogRepository struct {
	db *gorm.DB
}

func NewModuleUsageLogRepository(db *gorm.DB) domain.ModuleUsageLogRepository {
	return &moduleUsageLogRepository{db: db}
}

func (r *moduleUsageLogRepository) Create(log *domain.ModuleUsageLog) error {
	return r.db.Create(log).Error
}

func (r *moduleUsageLogRepository) GetCountByModuleAndDateRange(moduleCode string, startDate, endDate time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&domain.ModuleUsageLog{})
	if moduleCode != "" {
		query = query.Where("LOWER(module_code) = LOWER(?)", moduleCode)
	}
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	err := query.Count(&count).Error
	return count, err
}
