package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrScheduleOverlap = errors.New("schedule overlap")

type MunicipalityWorkSchedule struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	OfficerName       string    `gorm:"type:text;not null;column:officer_name;default:''" json:"officerName"`
	Position          string    `gorm:"type:text;not null;column:position" json:"position"`
	ContactNumber     string    `gorm:"type:text;not null;column:contact_number" json:"contactNumber"`
	WorkingDay        string    `gorm:"type:text;not null;column:working_day" json:"workingDay"`
	WorkingHoursStart string    `gorm:"type:text;not null;column:working_hours_start" json:"workingHoursStart"`
	WorkingHoursEnd   string    `gorm:"type:text;not null;column:working_hours_end" json:"workingHoursEnd"`
	Status            string    `gorm:"type:text;default:'active';column:status" json:"status"`
	CreatedBy         string    `gorm:"type:text;column:created_by;default:''" json:"createdBy"`
	UpdatedBy         string    `gorm:"type:text;column:updated_by;default:''" json:"updatedBy"`
	CreatedAt         time.Time `gorm:"autoCreateTime;column:created_date" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime;column:updated_date" json:"updatedAt"`
}

func (MunicipalityWorkSchedule) TableName() string {
	return "municipality_work_schedules"
}

type MunicipalityWorkScheduleRepository interface {
	GetAll() ([]MunicipalityWorkSchedule, error)
	GetByID(id uuid.UUID) (*MunicipalityWorkSchedule, error)
	Create(schedule *MunicipalityWorkSchedule) error
	Update(schedule *MunicipalityWorkSchedule) error
	Delete(id uuid.UUID) error
}

type MunicipalityWorkScheduleUseCase interface {
	GetAllShifts() ([]MunicipalityWorkSchedule, error)
	SaveShift(schedule *MunicipalityWorkSchedule) error
	DeleteShift(id uuid.UUID) error
}
