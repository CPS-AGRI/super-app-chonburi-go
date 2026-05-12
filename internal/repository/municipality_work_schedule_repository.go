package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type municipalityWorkScheduleRepository struct {
	db *gorm.DB
}

func NewMunicipalityWorkScheduleRepository(db *gorm.DB) domain.MunicipalityWorkScheduleRepository {
	return &municipalityWorkScheduleRepository{db: db}
}

func (r *municipalityWorkScheduleRepository) GetAll() ([]domain.MunicipalityWorkSchedule, error) {
	var schedules []domain.MunicipalityWorkSchedule
	err := r.db.Where("status = 'active'").Order("working_day, working_hours_start").Find(&schedules).Error
	return schedules, err
}

func (r *municipalityWorkScheduleRepository) GetByID(id uuid.UUID) (*domain.MunicipalityWorkSchedule, error) {
	var schedule domain.MunicipalityWorkSchedule
	err := r.db.First(&schedule, id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *municipalityWorkScheduleRepository) Create(schedule *domain.MunicipalityWorkSchedule) error {
	return r.db.Create(schedule).Error
}

func (r *municipalityWorkScheduleRepository) Update(schedule *domain.MunicipalityWorkSchedule) error {
	return r.db.Save(schedule).Error
}

func (r *municipalityWorkScheduleRepository) Delete(id uuid.UUID) error {
	// Soft delete by updating status
	return r.db.Model(&domain.MunicipalityWorkSchedule{}).Where("id = ?", id).Update("status", "inactive").Error
}
