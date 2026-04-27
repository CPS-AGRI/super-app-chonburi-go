package domain

import (
	"time"

	"github.com/google/uuid"
)

type Municipality struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CityNameTh     string    `json:"cityNameTh"`
	CityNameEn     string    `json:"cityNameEn"`
	CityAddressTh  string    `gorm:"type:text" json:"cityAddressTh"`
	CityAddressEn  string    `gorm:"type:text" json:"cityAddressEn"`
	CityPhone      string    `json:"cityPhone"`
	CityLogoUrl    string    `json:"cityLogoUrl"`
	CityLat        float64   `json:"cityLat"`
	CityLng        float64   `json:"cityLng"`
	Status         string    `gorm:"default:'active'" json:"status"`
	CreatedBy      string    `json:"createdBy"`
	UpdatedBy      string    `json:"updatedBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
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
