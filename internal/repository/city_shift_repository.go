package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cityShiftRepository struct {
	db *gorm.DB
}

func NewCityShiftRepository(db *gorm.DB) domain.CityShiftRepository {
	return &cityShiftRepository{db: db}
}

func (r *cityShiftRepository) GetAll() ([]domain.CityShift, error) {
	var shifts []domain.CityShift
	err := r.db.Where("status = ?", "active").Order("created_at ASC").Find(&shifts).Error
	return shifts, err
}

func (r *cityShiftRepository) GetByID(id uuid.UUID) (*domain.CityShift, error) {
	var shift domain.CityShift
	err := r.db.Where("status = ?", "active").First(&shift, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *cityShiftRepository) GetByDay(day string) ([]domain.CityShift, error) {
	var shifts []domain.CityShift
	err := r.db.Where("status = ?", "active").Where("working_day = ?", day).Find(&shifts).Error
	return shifts, err
}

func (r *cityShiftRepository) Create(shift *domain.CityShift) error {
	return r.db.Create(shift).Error
}

func (r *cityShiftRepository) Update(shift *domain.CityShift) error {
	return r.db.Save(shift).Error
}

func (r *cityShiftRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.CityShift{}).Where("id = ?", id).Update("status", "inactive").Error
}
