package domain

import (
	"time"

	"github.com/google/uuid"
)

type MunicipalityBank struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	BankName          string    `gorm:"not null" json:"bankName"`
	BankAccountName   string    `gorm:"not null" json:"bankAccountName"`
	BankAccountNumber string    `gorm:"not null" json:"bankAccountNumber"`
	BankBranch        string    `json:"bankBranch"`
	BankType          string    `json:"bankType"` // Saving, Current, etc.
	PromptPayNumber   string    `json:"promptPayNumber"`
	BankQrCodeUrl     string    `json:"bankQrCodeUrl"` // Kept for legacy support or manual override
	IsActive          bool      `gorm:"default:true" json:"isActive"`
	Status            string    `gorm:"default:'active'" json:"status"`
	CreatedBy         string    `json:"createdBy"`
	UpdatedBy         string    `json:"updatedBy"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type MunicipalityBankRepository interface {
	GetAll() ([]MunicipalityBank, error)
	GetActive() (*MunicipalityBank, error)
	GetByID(id uuid.UUID) (*MunicipalityBank, error)
	Create(bank *MunicipalityBank) error
	Update(bank *MunicipalityBank) error
	Delete(id uuid.UUID) error
}

type MunicipalityBankUseCase interface {
	GetActiveBank() (*MunicipalityBank, error)
	GetAllBanks() ([]MunicipalityBank, error)
	SaveBank(bank *MunicipalityBank) error
	DeleteBank(id uuid.UUID) error
}
