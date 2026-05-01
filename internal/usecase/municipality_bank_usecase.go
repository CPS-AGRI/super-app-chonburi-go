package usecase

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type municipalityBankUseCase struct {
	bankRepo domain.MunicipalityBankRepository
}

func NewMunicipalityBankUseCase(bankRepo domain.MunicipalityBankRepository) domain.MunicipalityBankUseCase {
	return &municipalityBankUseCase{
		bankRepo: bankRepo,
	}
}

func (u *municipalityBankUseCase) GetActiveBank() (*domain.MunicipalityBank, error) {
	return u.bankRepo.GetActive()
}

func (u *municipalityBankUseCase) GetAllBanks() ([]domain.MunicipalityBank, error) {
	return u.bankRepo.GetAll()
}

func (u *municipalityBankUseCase) SaveBank(bank *domain.MunicipalityBank) error {
	if bank.ID == uuid.Nil {
		// New record - ensure default status
		if bank.Status == "" {
			bank.Status = "active"
		}
		return u.bankRepo.Create(bank)
	}

	// Update record - fetch existing to preserve fields like status, createdBy
	existing, err := u.bankRepo.GetByID(bank.ID)
	if err != nil {
		return err
	}

	// Preserve fields if not provided
	if bank.Status == "" {
		bank.Status = existing.Status
	}
	if bank.CreatedBy == "" {
		bank.CreatedBy = existing.CreatedBy
	}
	if bank.CreatedAt.IsZero() {
		bank.CreatedAt = existing.CreatedAt
	}

	return u.bankRepo.Update(bank)
}

func (u *municipalityBankUseCase) DeleteBank(id uuid.UUID) error {
	return u.bankRepo.Delete(id)
}
