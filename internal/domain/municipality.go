package domain

import (
	"time"

	"github.com/google/uuid"
)

type Municipality struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	CityNameTh     string    `gorm:"column:cityNameTh" json:"cityNameTh"`
	CityNameEn     string    `gorm:"column:cityNameEn" json:"cityNameEn"`
	CityAddressTh  string    `gorm:"type:text;column:cityAddressTh" json:"cityAddressTh"`
	CityAddressEn  string    `gorm:"type:text;column:cityAddressEn" json:"cityAddressEn"`
	CityPhone      string    `gorm:"column:cityPhone" json:"cityPhone"`
	CityLogoUrl    string    `gorm:"column:cityLogoUrl" json:"cityLogoUrl"`
	CityLat        float64   `gorm:"column:cityLat" json:"cityLat"`
	CityLng        float64   `gorm:"column:cityLng" json:"cityLng"`
	Status         string    `gorm:"column:status;default:'active'" json:"status"`
	CreatedBy      string    `gorm:"column:createdBy" json:"createdBy"`
	UpdatedBy      string    `gorm:"column:updatedBy" json:"updatedBy"`
	CreatedAt      time.Time `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Municipality) TableName() string {
	return "Municipality"
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
	// Helper for current standalone city info
	GetCurrent() (*Municipality, error)
}
