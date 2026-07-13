package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type municipalityBankRepository struct {
	db *gorm.DB
}

func NewMunicipalityBankRepository(db *gorm.DB) domain.MunicipalityBankRepository {
	return &municipalityBankRepository{db: db}
}

func (r *municipalityBankRepository) GetAll() ([]domain.MunicipalityBank, error) {
	var banks []domain.MunicipalityBank
	err := r.db.Where("status = 'active'").Find(&banks).Error
	return banks, err
}

func (r *municipalityBankRepository) GetActive() (*domain.MunicipalityBank, error) {
	var bank domain.MunicipalityBank
	err := r.db.Where("bank_status = 'active' AND status = 'active'").First(&bank).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}

func (r *municipalityBankRepository) GetByID(id uuid.UUID) (*domain.MunicipalityBank, error) {
	var bank domain.MunicipalityBank
	err := r.db.First(&bank, id).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}

func (r *municipalityBankRepository) Create(bank *domain.MunicipalityBank) error {
	return r.db.Create(bank).Error
}

func (r *municipalityBankRepository) Update(bank *domain.MunicipalityBank) error {
	return r.db.Save(bank).Error
}

func (r *municipalityBankRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.MunicipalityBank{}).Where("id = ?", id).Update("status", "inactive").Error
}
