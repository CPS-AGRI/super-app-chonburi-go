package domain

import (
	"time"

	"github.com/google/uuid"
)

type Municipality struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	CityNameTh    string    `gorm:"type:text;not null;column:name_th" json:"nameTh"`
	CityNameEn    string    `gorm:"type:text;not null;column:name_en" json:"nameEn"`
	CityAddressTh string    `gorm:"type:text;column:address_th" json:"addressTh"`
	CityAddressEn string    `gorm:"type:text;column:address_en" json:"addressEn"`
	CityPhone     string    `gorm:"type:text;column:phone" json:"phone"`
	CityLogoUrl   string    `gorm:"type:text;column:logo_url" json:"logoUrl"`
	CityLat       float64   `gorm:"type:float8;column:latitude" json:"latitude"`
	CityLng       float64   `gorm:"type:float8;column:longitude" json:"longitude"`
	Status        string    `gorm:"type:text;default:'active';column:status" json:"status"`
	CreatedBy     string    `gorm:"type:text;column:created_by;default:''" json:"createdBy"`
	UpdatedBy     string    `gorm:"type:text;column:updated_by;default:''" json:"updatedBy"`
	CreatedAt     time.Time `gorm:"type:timestamptz;column:created_date" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;column:updated_date" json:"updatedAt"`
	ComplaintMode string    `gorm:"type:text;column:complaint_mode;default:'direct'" json:"complaintMode"`
}

func (Municipality) TableName() string {
	return "municipalities"
}

type MunicipalityRepository interface {
	GetList() ([]Municipality, error)
	GetByID(id uuid.UUID) (*Municipality, error)
	GetFirst() (*Municipality, error)
	Create(municipality *Municipality) error
	Update(municipality *Municipality) error
	Delete(id uuid.UUID) error
}

type MunicipalityUseCase interface {
	GetList() ([]Municipality, error)
	GetDetail(id uuid.UUID) (*Municipality, error)
	Create(municipality *Municipality) error
	Update(municipality *Municipality) error
	Delete(id uuid.UUID) error

	GetCurrent() (*Municipality, error)
}
