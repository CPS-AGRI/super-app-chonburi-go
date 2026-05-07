package domain

import (
	"time"

	"github.com/google/uuid"
)

type CityShift struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OfficerName       string    `gorm:"not null" json:"officerName"`
	Position          string    `gorm:"not null" json:"position"`
	ContactNumber     string    `gorm:"not null" json:"contactNumber"`
	WorkingDay        string    `gorm:"not null" json:"workingDay"`
	WorkingHoursStart string    `gorm:"not null" json:"workingHoursStart"`
	WorkingHoursEnd   string    `gorm:"not null" json:"workingHoursEnd"`
	Status            string    `gorm:"default:'active'" json:"status"`
	CreatedBy         string    `json:"createdBy"`
	UpdatedBy         string    `json:"updatedBy"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CityShiftRepository interface {
	GetAll() ([]CityShift, error)
	GetByID(id uuid.UUID) (*CityShift, error)
	GetByDay(day string) ([]CityShift, error)
	Create(shift *CityShift) error
	Update(shift *CityShift) error
	Delete(id uuid.UUID) error
}

type CityShiftUseCase interface {
	GetAllShifts() ([]CityShift, error)
	SaveShift(shift *CityShift) error
	DeleteShift(id uuid.UUID) error
}
